package main

import (
	"fmt"
	"freedb/helper"
	"freedb/kv"
	"strings"
)

func (c *Cli) use(args []string) {
	if c.kv == nil || c.conf.host == nil {
		c.log.Error("Please config your host first")
		return
	}
	c.kv.Use(args[0])
}
func (c *Cli) keys(args []string) {
	if c.kv == nil || c.conf.host == nil {
		c.log.Error("Please config your host first")
		return
	}
	krl, err := c.kv.Keys()

	if err != nil {
		c.log.Error(err.Error())
		return
	}
	c.outputList(krl)
}

func (c *Cli) set(args []string) {
	record, err := c.kv.Set(args[0], args[1])
	if err != nil {
		c.log.Error(err.Error())
		return
	}
	c.output(record)
}
func (c *Cli) append(args []string) {
	record, err := c.kv.Append(args[0], args[1])
	if err != nil {
		c.log.Error(err.Error())
		return
	}
	c.output(record)
}
func (c *Cli) get(args []string) {
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
}
func (c *Cli) delete(args []string) {
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
}

func (c *Cli) config(args []string) {
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
	default:
		c.log.Error("CONFIG command does not recognize key: " + item)
	}
}
