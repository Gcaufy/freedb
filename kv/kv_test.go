package kv

import (
	"os"
	"testing"
)

func TestGet(t *testing.T) {

	token := os.Getenv("TEST_REPO_TOKEN")
	if token == "" {
		t.Error("Please set a token")
		return
	}
	kv, err := NewKV("git@github.com:Gcaufy-Test/test-database.git", token)
	kv.Use("golang")
	if err != nil {
		t.Error("Should not be error")
		t.Error(err)
	}
	_, err = kv.Set("key-exist", "123")
	if err != nil {
		t.Error(err)
	}
	_, err = kv.Set("key-non-exist", "456")
	if err != nil {
		t.Error(err)
	}
	_, err = kv.Delete("key-non-exist")
	if err != nil {
		t.Error(err)
	}
	_, err = kv.Append("key-exist", "hello-world")
	if err != nil {
		t.Error(err)
	}
	list, err := kv.Keys()
	if err != nil {
		t.Error(err)
	}
	if len(*list) < 1 {
		t.Error("Should have at least one key there")
	}
}
