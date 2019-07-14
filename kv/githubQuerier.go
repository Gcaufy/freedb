package kv

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GithubQuerier is a querier for github
type GithubQuerier struct {
	baseURL   string
	option    *QuerierOption
	shaCache  shaMap
	retryMap  retryCounter
	committer *Committer
}

type githubCommiter struct {
	name  string `json:"name"`
	email string `json:"email"`
}
type githubError struct {
	Message string `json:"message"`
	URL     string `json:"document_url"`
	Code    int    `json:"code"`
}

func (e *githubError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

type githubKeyRecord struct {
	Content  string `json:"content"`
	Name     string `json:"name"`
	Size     int    `json:"size"`
	RawURL   string `json:"download_url"`
	HTMLURL  string `json:"html_url"`
	Commit   string `json:"commit"`
	Path     string `json:"path"`
	Sha      string `json:"sha"`
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
}
type githubCommitInfo struct {
	Sha string `json:"sha"`
}
type githubPutResult struct {
	Content *githubKeyRecord  `json:"content"`
	Commit  *githubCommitInfo `json:"commit"`
}

type githubPutOption struct {
	Message   string     `json:"message"`
	Content   string     `json:"content"`
	Sha       string     `json:"sha,omitempty"`
	Branch    string     `json:"branch"`
	Committer *Committer `json:"committer"`
}

type shaMap map[string]string
type retryCounter map[string]int

// NewGithubQuerier is a querier constructor
func NewGithubQuerier(option *QuerierOption) *GithubQuerier {
	return &GithubQuerier{
		baseURL:   fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", option.user, option.repo),
		option:    option,
		committer: option.committer,
		shaCache:  make(map[string]string),
		retryMap:  make(map[string]int),
	}
}

// Keys is a function to list all keys
func (q *GithubQuerier) Keys() (*[]*KeyRecord, error) {
	record, err := q.listReq()
	if err != nil {
		if err.Code == 404 {
			return nil, fmt.Errorf("Repository not found")
		}
		return nil, err
	}
	return record, nil
}

// Get is a function to read a key
func (q *GithubQuerier) Get(key string) (*KeyRecord, error) {
	record, err := q.getReq(key)
	if err != nil {
		if err.Code == 404 {
			return &KeyRecord{}, nil
		}
		return nil, err
	}
	decodeBytes, _ := base64.StdEncoding.DecodeString(record.Content)
	record.Content = string(decodeBytes)
	q.shaCache[record.Name] = record.Sha
	return record.transfer(), nil
}

// Set is a function to set a key
func (q *GithubQuerier) Set(key string, value string) (*KeyRecord, error) {
	encoded := base64.StdEncoding.EncodeToString([]byte(value))
	gpo := &githubPutOption{
		Content:   encoded,
		Branch:    q.option.branch,
		Message:   "freedb update a key from golang client",
		Committer: q.committer,
	}

	sha, ok := q.shaCache[key]
	if ok { // Try to update a exist record
		gpo.Sha = sha
	} else { // Set a new record
		gpo.Message = "freedb create a key from golang client"
	}
	record, err := q.putReq(key, gpo)
	if err != nil {
		// 409: [409] xxx does not match. which mean sha is wrong
		// 422: [422] "sha" wasn't supplied.
		if err.Code == 409 || err.Code == 422 {
			_, ok := q.retryMap[key]
			if ok {
				delete(q.retryMap, key)
				return nil, fmt.Errorf("Update key \"%s\" failed: %s", key, err)
			}
			q.retryMap[key] = 1
			_, getErr := q.Get(key) // Update sha for the key
			if getErr != nil {
				return nil, fmt.Errorf("Get key \"%s\" failed: %s", key, getErr)
			}
			kr, setErr := q.Set(key, value)
			delete(q.retryMap, key)
			return kr, setErr
		}
		return nil, err
	}
	q.shaCache[record.Name] = record.Sha
	return record.transfer(), nil
}

// Delete is a function to delete a key
func (q *GithubQuerier) Delete(key string) (*KeyRecord, error) {
	gpo := &githubPutOption{
		Branch:    q.option.branch,
		Message:   "freedb delete a key from golang client",
		Committer: q.committer,
	}
	sha, ok := q.shaCache[key]
	if ok {
		gpo.Sha = sha
	}

	record, err := q.deleteReq(key, gpo)
	if err != nil {
		// 409: [409] xxx does not match. which mean sha is wrong
		// 422: [422] "sha" wasn't supplied.
		if err.Code == 409 || err.Code == 422 {
			_, ok := q.retryMap[key]
			if ok {
				delete(q.retryMap, key)
				return nil, fmt.Errorf("Update key \"%s\" failed: %s", key, err)
			}
			q.retryMap[key] = 1
			getKr, getErr := q.Get(key) // Update sha for the key
			if getErr != nil {
				return nil, fmt.Errorf("Get key \"%s\" failed: %s", key, getErr)
			}
			if getKr.Name == "" { // The key do not exist, can not delete it

				delete(q.retryMap, key)
				return &KeyRecord{}, nil
			}
			kr, setErr := q.Delete(key)
			delete(q.retryMap, key)
			return kr, setErr
		}
		return nil, err
	}
	delete(q.shaCache, key)
	return record.transfer(), nil
}

func (q *GithubQuerier) setHost(user, repo string) {
	q.baseURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", user, repo)
	q.option.user = user
	q.option.repo = repo
}
func (q *GithubQuerier) setBranch(branch string) {
	q.option.branch = branch
}
func (q *GithubQuerier) use(db string) {
	q.option.db = db
}
func (q *GithubQuerier) setToken(token string) {
	q.option.token = token
}

func (q *GithubQuerier) listReq() (*[]*KeyRecord, *githubError) {
	body, err := q.query("", "GET", nil)
	if err != nil {
		return nil, err
	}
	var gkrl []*githubKeyRecord
	decodeErr := json.Unmarshal(*body, &gkrl)
	if decodeErr != nil {
		return nil, &githubError{Message: decodeErr.Error()}
	}
	var krl []*KeyRecord
	for _, gkr := range gkrl {
		q.shaCache[gkr.Name] = gkr.Sha
		krl = append(krl, gkr.transfer())
	}
	return &krl, nil
}
func (q *GithubQuerier) getReq(key string) (*githubKeyRecord, *githubError) {
	body, err := q.query(key, "GET", nil)
	if err != nil {
		return nil, err
	}
	kr := &githubKeyRecord{}
	decodeErr := json.Unmarshal(*body, kr)
	if decodeErr != nil {
		var gkrl []*githubKeyRecord
		decodeErr := json.Unmarshal(*body, &gkrl)
		if decodeErr == nil {
			return nil, &githubError{Message: fmt.Sprintf("'%s' is a folder", key)}
		}
		return nil, &githubError{Message: decodeErr.Error()}
	}
	return kr, nil
}
func (q *GithubQuerier) putReq(key string, gpo *githubPutOption) (*githubKeyRecord, *githubError) {
	body, err := q.query(key, "PUT", gpo)
	if err != nil {
		return nil, err
	}
	gpr := &githubPutResult{}
	decodeErr := json.Unmarshal(*body, gpr)
	if decodeErr != nil {
		return nil, &githubError{Message: decodeErr.Error()}
	}
	return gpr.transfer(), nil
}

func (q *GithubQuerier) deleteReq(key string, gpo *githubPutOption) (*githubKeyRecord, *githubError) {
	body, err := q.query(key, "DELETE", gpo)
	if err != nil {
		return nil, err
	}
	gpr := &githubPutResult{}
	decodeErr := json.Unmarshal(*body, gpr)
	if decodeErr != nil {
		return nil, &githubError{Message: decodeErr.Error()}
	}
	return &githubKeyRecord{
		Commit: gpr.Commit.Sha,
	}, nil
}

func (q *GithubQuerier) query(key string, method string, data *githubPutOption) (*[]byte, *githubError) {
	urlStr := q.baseURL + "/" + q.option.db
	if key != "" {
		urlStr += "/" + key
	}
	var req *http.Request
	var err error
	if data != nil {
		body := new(bytes.Buffer)
		json.NewEncoder(body).Encode(data)
		req, err = http.NewRequest(method, urlStr, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, urlStr, nil)
	}
	if err != nil {
		return nil, &githubError{Message: err.Error()}
	}
	req.Header.Set("User-Agent", "freedb")
	req.Header.Set("Authorization", "token "+q.option.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, &githubError{Message: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return nil, &githubError{Code: 401, Message: "Invalid token"}
	} else if resp.StatusCode == 404 {
		if key == "" {
			return nil, &githubError{Code: 404, Message: "Invalid repository"}
		}
		return nil, &githubError{Code: 404, Message: "Invalid repository or invalid key"}
	}
	if resp.StatusCode > 299 {
		gitErr := &githubError{}
		respBody, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(respBody, &gitErr)
		gitErr.Code = resp.StatusCode
		if err != nil {
			return nil, gitErr
		}
		return nil, gitErr
	}
	respBody, _ := ioutil.ReadAll(resp.Body)

	/*
		fmt.Println(urlStr)
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Body:", string(respBody))
	*/

	return &respBody, nil
}

func (gpr *githubPutResult) transfer() *githubKeyRecord {
	return &githubKeyRecord{
		Name:    gpr.Content.Name,
		Size:    gpr.Content.Size,
		RawURL:  gpr.Content.RawURL,
		HTMLURL: gpr.Content.HTMLURL,
		Commit:  gpr.Commit.Sha,
	}
}

func (gkr *githubKeyRecord) transfer() *KeyRecord {
	return &KeyRecord{
		Content: gkr.Content,
		Name:    gkr.Name,
		Size:    gkr.Size,
		RawURL:  gkr.RawURL,
		HTMLURL: gkr.HTMLURL,
		Commit:  gkr.Commit,
	}
}
