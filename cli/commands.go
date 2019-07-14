package cli

import (
	"fmt"
	"strings"

	helper "github.com/Gcaufy/freedb/helper"
	kv "github.com/Gcaufy/freedb/kv"
)

func (c *cli) use(args []string) {
	if c.kv == nil || c.conf.host == nil {
		c.log.Error("Please config your host first")
		return
	}
	c.kv.Use(args[0])
}
func (c *cli) keys(args []string) {
	if c.kv == nil || c.conf.host == nil {
		c.log.Error("Please config your host first")
		return
	}
	c.timeUse(func() {

		krl, err := c.kv.Keys()

		if err != nil {
			c.log.Error(err.Error())
			return
		}
		c.outputList(krl)
	})
}

func (c *cli) set(args []string) {
	if c.kv == nil || c.conf.host == nil {
		c.log.Error("Please config your host first")
		return
	}
	c.timeUse(func() {
		record, err := c.kv.Set(args[0], args[1])
		if err != nil {
			c.log.Error(err.Error())
			return
		}
		c.output(record)
	})
}
func (c *cli) append(args []string) {
	if c.kv == nil || c.conf.host == nil {
		c.log.Error("Please config your host first")
		return
	}
	c.timeUse(func() {
		record, err := c.kv.Append(args[0], args[1])
		if err != nil {
			c.log.Error(err.Error())
			return
		}
		c.output(record)
	})
}
func (c *cli) get(args []string) {
	if c.kv == nil || c.conf.host == nil {
		c.log.Error("Please config your host first")
		return
	}
	c.timeUse(func() {

		record, err := c.kv.Get(args[0])

		if err != nil {
			c.log.Error(fmt.Sprintln(err))
			return
		}
		if record.Name == "" {
			c.log.Error(fmt.Sprintf("Key \"%s\" not found", args[0]))
			return
		}
		c.output(record)
	})
}
func (c *cli) delete(args []string) {
	if c.kv == nil || c.conf.host == nil {
		c.log.Error("Please config your host first")
		return
	}
	c.timeUse(func() {

		record, err := c.kv.Delete(args[0])
		if err != nil {
			c.log.Error(fmt.Sprintln(err))
			return
		}
		if record.Name == "" {
			c.log.Error(fmt.Sprintf("Key \"%s\" not found", args[0]))
			return
		}
		c.output(record)
	})
}

func (c *cli) config(args []string) {
	item, value := strings.ToUpper(args[0]), args[1]
	switch item {
	case "HOST":
		host, err := helper.ParseHost(value)
		if err != nil {
			c.log.Error(err.Error())
			return
		}
		if c.conf.host == nil || c.conf.host.Provider != host.Provider {
			c.conf.host = host
			c.kv, err = kv.NewKV(value, c.conf.token)
			if c.conf.db != "" {
				c.kv.Use(c.conf.db)
			}
			if c.conf.branch != "" {
				c.kv.SetBranch(c.conf.branch)
			}
			if c.conf.secret != "" {
				c.kv.SetSecret(c.conf.secret)
			}
			if err != nil {
				c.log.Error(err.Error())
				return
			}
		} else {
			c.kv.SetHost(value)
		}
		break
	case "TOKEN":
		c.conf.token = value
		if c.kv != nil {
			c.kv.SetToken(value)
		}
		break
	case "DB":
		c.conf.db = value
		if c.kv != nil {
			c.kv.Use(value)
		}
		break
	case "BRANCH":
		c.conf.branch = value
		if c.kv != nil {
			c.kv.SetBranch(value)
		}
		break
	case "CACHE":
		s := strings.ToUpper(value)
		if s == "FALSE" {
			c.conf.cache = false
			if c.kv != nil {
				c.kv.UseCache = false
				c.kv.ClearCache()
			}
		} else if s == "TRUE" {
			c.conf.cache = true
			if c.kv != nil {
				c.kv.UseCache = true
			}
		}
		break
	default:
		c.log.Error("CONFIG command does not recognize key: " + item)
	}
}
