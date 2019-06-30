package helper

import (
	"errors"
	"regexp"
)

// Host is github host struct
type Host struct {
	Provider string
	User     string
	Repo     string
}

// ParseHost is a function to parse git@github.com:xx/yy.git or https://github.com/xxx/yyy.git
func ParseHost(host string) (*Host, error) {
	regs := [2]*regexp.Regexp{
		regexp.MustCompile(`git@([\w\.]+):([\w_-]+)\/([\w\._-]+)\.git`),
		regexp.MustCompile(`https:\/\/([\w\.]+)\/([\w_-]+)\/([\w\._-]+)\.git`),
	}

	for _, reg := range regs {
		matchs := reg.FindStringSubmatch(host)
		length := len(matchs)
		if length == 0 {
			continue
		} else {
			return &Host{
				Provider: matchs[1],
				User:     matchs[2],
				Repo:     matchs[3],
			}, nil
		}
	}
	return nil, errors.New("Can not reconigize host: " + host)
}
