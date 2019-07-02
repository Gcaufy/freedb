package helper

import "testing"

func TestParseHost(t *testing.T) {
	hostHTTPS, errh := ParseHost("https://github.com/Gcaufy-Test/testdb.git")
	hostSSH, errs := ParseHost("git@github.com:Gcaufy-Test/testdb.git")

	if errh != nil || errs != nil {
		t.Error(errh)
		t.Error(errs)
		return
	}
	if hostSSH.Provider != "github.com" || hostSSH.Provider != hostHTTPS.Provider {
		t.Error("host.provider is wrong")
		return
	}

	if hostSSH.User != "Gcaufy-Test" || hostSSH.User != hostHTTPS.User {
		t.Error("host.user is wrong")
		return
	}

	if hostSSH.Repo != "testdb" || hostSSH.Repo != hostHTTPS.Repo {
		t.Error("host.repo is wrong")
	}
}

func TestParseError(t *testing.T) {
	_, err := ParseHost("https://www.baidu.com")

	if err == nil {
		t.Error("Expect parse error")
	}
}
