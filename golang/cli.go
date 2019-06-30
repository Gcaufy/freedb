package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"freedb/helper"
	"freedb/kv"
	"os"
	"strings"
)

// Command type
type Command struct {
	Args int
	Exec func(args []string)
}

// Config type
type Config struct {
	host   *helper.Host
	token  string
	db     string
	branch string
}

// Cli type
type Cli struct {
	log      *ConsoleLogger
	commands map[string]*Command
	conf     *Config
	kv       *kv.KV
}

func newCli() *Cli {

	m := make(map[string]*Command)
	cli := &Cli{
		log: NewConsoleLogger(),
		conf: &Config{
			db:     "default",
			branch: "master",
		},
	}

	m["SET"] = &Command{
		Args: 2,
		Exec: cli.set,
	}

	m["GET"] = &Command{
		Args: 1,
		Exec: cli.get,
	}
	m["CONFIG"] = &Command{
		Args: 2,
		Exec: cli.config,
	}
	m["USE"] = &Command{
		Args: 1,
		Exec: cli.use,
	}
	cli.commands = m
	return cli
}

func (c *Cli) run() {
	fmt.Print("freedb shell version v1.0.0\n")
	fmt.Print("> ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			fmt.Print("> ")
			continue
		}
		arg0 := strings.ToUpper(fields[0])
		args := fields[1:]
		commandName := c.commands[arg0]

		if commandName != nil {
			argLen := len(args)
			if commandName.Args == argLen {
				commandName.Exec(args)
			} else {
				c.log.Error("Command \"%s\" expect %d arguments, but %d arguments got.", arg0, commandName.Args, argLen)
			}
		} else {
			c.log.Error("Invalid command")
		}

		if line == "EXIT" {
			fmt.Println("Thanks for using freedb")
			break
		}

		fmt.Print("> ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error encountered:", err)
	}
}

func (c *Cli) use(args []string) {
	if c.kv == nil || c.conf.host == nil {
		c.log.Error("Please config your host first")
		return
	}
	c.kv.Use(args[0])
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
		c.kv.Use(value)
		break
	case "BRANCH":
		c.conf.branch = value
		c.kv.SetBranch(value)
		break
	default:
		c.log.Error("CONFIG command does not recognize key: " + item)
	}
}
func (c *Cli) set(args []string) {
	record, err := c.kv.Set(args[0], args[1])
	if err != nil {
		c.log.Error(err.Error())
		return
	}
	b, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		c.log.Error(err.Error())
		return
	}
	fmt.Println(string(b))
}
func (c *Cli) append(args []string) {
	record, err := c.kv.Append(args[0], args[1])
	if err != nil {
		c.log.Error(err.Error())
		return
	}
	b, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		c.log.Error(err.Error())
		return
	}
	fmt.Println(string(b))
}
func (c *Cli) get(args []string) {
	record, err := c.kv.Get(args[0])
	if err != nil {
		c.log.Error(fmt.Sprintln(err))
		return
	}
	b, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		c.log.Error(err.Error())
		return
	}
	fmt.Println(string(b))
}
func (c *Cli) delete(args []string) {
	record, err := c.kv.Delete(args[0])
	if err != nil {
		c.log.Error(fmt.Sprintln(err))
		return
	}
	b, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		c.log.Error(err.Error())
		return
	}
	fmt.Println(string(b))
}
