package main

import (
	"bufio"
	"fmt"
	"freedb/helper"
	"freedb/kv"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
)

// Command type
type Command struct {
	Args int
	Exec func(args []string)
}

// Config type
type Config struct {
	host        *helper.Host
	hostStr     string
	token       string
	db          string
	branch      string
	execute     string
	shortOutput bool
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
	cli.initCLI()
	return cli
}

func (c *Cli) run() {
	fmt.Print("freedb shell version v1.0.0\n")
	fmt.Print("> ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		quit := c.execLine(line)
		if quit {
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

func (c *Cli) execLine(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
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
		return true
	}
	return false
}

func (c *Cli) initCLI() {
	var helpFlag bool
	var rootCmd = &cobra.Command{
		Use: "freedb",
		Run: func(cmd *cobra.Command, args []string) {
			configList := [...]string{"token", "branch", "db"}

			r := reflect.ValueOf(c.conf)
			i := reflect.Indirect(r)
			for _, item := range configList {
				v := i.FieldByName(item).String()
				if v != "" {
					// fmt.Println(v)
					c.execLine(fmt.Sprintf("CONFIG %s %s", strings.ToUpper(item), v))
				}
			}
			if c.conf.hostStr != "" {
				c.execLine("CONFIG HOST " + c.conf.hostStr)
			}

			if c.conf.execute != "" {
				c.execLine(c.conf.execute)
			} else {
				c.run()
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&c.conf.db, "database", "d", "default", "Config using database.")
	rootCmd.PersistentFlags().StringVarP(&c.conf.branch, "branch", "b", "master", "Config using branch.")
	rootCmd.PersistentFlags().StringVarP(&c.conf.token, "token", "t", "", "Access token for the git repository.")
	rootCmd.PersistentFlags().StringVarP(&c.conf.hostStr, "host", "h", "", "Connect to host, which is a https/ssh git clone link.")
	rootCmd.PersistentFlags().StringVarP(&c.conf.execute, "execute", "e", "", "Execute command and quit.")
	rootCmd.PersistentFlags().BoolVarP(&helpFlag, "help", "?", false, "Display the help")
	rootCmd.PersistentFlags().BoolVarP(&c.conf.shortOutput, "short-output", "s", false, "Only output the value")

	rootCmd.Execute()
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
	c.output(record)
}

func (c *Cli) output(kr *kv.KeyRecord) {
	var val string
	val = kr.Short()
	if !c.conf.shortOutput {
		str, err := kr.ToString()
		if err != nil {
			c.log.Error(fmt.Sprintln(err))
			return
		}
		val = str
	}
	fmt.Println(val)
}
