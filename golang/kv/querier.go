package kv

import "encoding/json"

// KeyRecord is the record for a key
type KeyRecord struct {
	Content string `mapstructure: "content" json:"content"`
	Name    string `mapstructure: "name" json:"name"`
	Size    int    `mapstructure: "size" json:"size"`
	RawURL  string `mapstructure: "download_url" json:"raw_url"`
	HTMLURL string `mapstructure: "html_url" json:"html_url"`
	Commit  string `mapstructure: "commit" json:"commit"`
}

// Querier is a interface that to query github
type Querier interface {
	Get(key string) (*KeyRecord, error)
	Set(key string, value string) (*KeyRecord, error)
	Delete(key string) (*KeyRecord, error)
	Keys() (*[]*KeyRecord, error)

	setHost(user, repo string)
	setBranch(branch string)
	use(db string)
	setToken(token string)
}

// QuerierOption is an option pass to Querier constructor
type QuerierOption struct {
	user      string
	repo      string
	db        string
	token     string
	branch    string
	committer *Committer
}

// Committer is a git comitter type
type Committer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Short is a function to get short value fo a KeyRecord
func (k *KeyRecord) Short() string {
	return k.Content
}

// ToString is a function to get a serialized KeyRecord
func (k *KeyRecord) ToString() (string, error) {
	b, err := json.MarshalIndent(k, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
