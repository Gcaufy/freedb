package kv

import (
	"fmt"

	helper "github.com/Gcaufy/freedb/helper"
)

// DataBaseOption is KV database option
type DataBaseOption struct {
	host   string
	token  string
	db     string
	branch string
}

// KV is a key-value storage
type KV struct {
	querier Querier
}

var querierMap = make(map[string]func(option *QuerierOption) Querier)

// NewKV will create a KV instace
func NewKV(host string, token string) (*KV, error) {

	parsedHost, err := helper.ParseHost(host)

	if err != nil {
		return nil, err
	}

	if len(querierMap) == 0 {
		// TODO: we may support gitlab, bitbucket later
		querierMap["github.com"] = func(option *QuerierOption) Querier {
			return NewGithubQuerier(option)
		}
	}

	con := querierMap[parsedHost.Provider]
	if con == nil {
		return nil, fmt.Errorf("\"%s\" is not supported currently", parsedHost.Provider)
	}
	op := &QuerierOption{
		user:   parsedHost.User,
		repo:   parsedHost.Repo,
		token:  token,
		db:     "default",
		branch: "master",
		committer: &Committer{
			Name:  "freedb",
			Email: "freedb@unknown.email.host",
		},
	}
	querier := con(op)

	return &KV{
		querier: querier,
	}, nil
}

// SetHost will update the host
func (kv *KV) SetHost(host string) (bool, error) {
	parsedHost, err := helper.ParseHost(host)
	if err != nil {
		return false, err
	}
	kv.querier.setHost(parsedHost.User, parsedHost.Repo)
	return true, nil
}

// SetBranch is a function to update the branch
func (kv *KV) SetBranch(branch string) {
	kv.querier.setBranch(branch)
}

// SetToken is a functio to update token
func (kv *KV) SetToken(token string) {
	kv.querier.setToken(token)
}

// Use is a function to change database
func (kv *KV) Use(db string) {
	kv.querier.use(db)
}

// Get is the function to get a key
func (kv *KV) Get(key string) (*KeyRecord, error) {
	record, err := kv.querier.Get(key)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Set is the function to update a key or create a new key
func (kv *KV) Set(key string, value string) (*KeyRecord, error) {
	record, err := kv.querier.Set(key, value)
	if record != nil {
		record.Content = value
	}
	return record, err
}

// Append is the function to append value to a key
func (kv *KV) Append(key string, value string) (*KeyRecord, error) {
	record, err := kv.Get(key)
	if err != nil {
		return nil, err
	}

	value = record.Content + value
	return kv.Set(key, value)
}

// Delete is the function to delete a key
func (kv *KV) Delete(key string) (*KeyRecord, error) {
	record, err := kv.querier.Delete(key)
	return record, err
}

// Keys is the function to list all keys
func (kv *KV) Keys() (*[]*KeyRecord, error) {
	record, err := kv.querier.Keys()
	return record, err
}
