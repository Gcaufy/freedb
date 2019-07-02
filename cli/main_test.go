package cli

import "testing"
import "os"

var host = "git@github.com:Gcaufy-Test/test-database.git"
var token = os.Getenv("TEST_REPO_TOKEN")

var c = createCliInstance()

func createCliInstance() *cli {
	c := &cli{
		log: NewConsoleLogger(),
		conf: &Config{
			db:     "default",
			branch: "master",
		},
	}
	c.initDSL()
	return c
}

func TestConfig(t *testing.T) {
	if token == "" {
		t.Error("Token is not set")
		return
	}
	c.execLine("CONFIG HOST " + host)
	c.execLine("CONFIG TOKEN " + token)
	c.execLine("CONFIG Branch " + "test")

	c.execLine("GET acb")
}

func TestCommandBeforeConfig(t *testing.T) {
	c := createCliInstance()
	if token == "" {
		t.Error("Token is not set")
		return
	}
	c.execLine("APPEND abc 123")
	c.execLine("SET abc 123")
	c.execLine("GET abc")
	c.execLine("KEYS")
}
