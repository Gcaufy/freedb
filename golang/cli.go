package main

import (
	"encoding/json"
	"fmt"
	"freedb/helper"
	"freedb/kv"
	"reflect"
	"strings"

	prompt "github.com/c-bata/go-prompt"
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
	m["APPEND"] = &Command{
		Args: 2,
		Exec: cli.append,
	}
	m["KEYS"] = &Command{
		Args: 0,
		Exec: cli.keys,
	}
	m["CONFIG"] = &Command{
		Args: 2,
		Exec: cli.config,
	}
	m["USE"] = &Command{
		Args: 1,
		Exec: cli.use,
	}
	m["DELETE"] = &Command{
		Args: 1,
		Exec: cli.delete,
	}
	m["EXIT"] = &Command{
		Args: 0,
		Exec: func(args []string) {},
	}
	cli.commands = m
	cli.initCLI()
	return cli
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
				if c.kv == nil {
					c.log.Error("Please use -h to config host")
					return
				}
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

type instruct struct {
	text string
	desc string
}

var commandInstruct = []*instruct{
	&instruct{
		text: "APPEND", desc: "Append a value to a key",
	},
	&instruct{
		text: "CONFIG", desc: "Config options",
	},
	&instruct{
		text: "DELETE", desc: "Delete a key",
	},
	&instruct{
		text: "GET", desc: "Get the value of a key",
	},
	&instruct{
		text: "SET", desc: "Set value to a key",
	},
	&instruct{
		text: "KEYS", desc: "List all keys",
	},
	&instruct{
		text: "USE", desc: "Change database",
	},
}
var configInstruct = []*instruct{
	&instruct{
		text: "HOST", desc: "It's a https/ssh git clone link",
	},
	&instruct{
		text: "TOKEN", desc: "Git OAuth access token",
	},
	&instruct{
		text: "BRANCH", desc: "Git branch",
	},
}

func (c *Cli) completer(in prompt.Document) []prompt.Suggest {
	var rst []prompt.Suggest
	text := in.TextBeforeCursor()
	args := strings.Split(text, " ")
	commandName := args[0]

	var filterArr []*instruct
	var filterStr string

	l := len(args)

	if l == 1 {
		filterArr = commandInstruct
		filterStr = args[0]
	} else if l == 2 {
		if strings.ToUpper(commandName) == "CONFIG" {
			filterArr = configInstruct
			filterStr = args[1]
		}
	}
	for _, com := range filterArr {
		if strings.Index(com.text, strings.ToUpper(filterStr)) != -1 {
			rst = append(rst, prompt.Suggest{
				Text:        com.text,
				Description: com.desc,
			})
		}
	}
	return rst
}

func (c *Cli) run() {
	fmt.Print("freedb shell version v1.0.0\n")

	p := prompt.New(
		func(in string) { c.execLine(in) },
		c.completer,
		prompt.OptionShowCompletionAtStart(),
	)
	p.Run()
}

func (c *Cli) execMultipleLine(line string) bool {
	lines := strings.Split(line, ";")
	for _, s := range lines {
		s = strings.TrimSpace(s)
		if s != "" {
			c.execLine(s)
		}
	}
	return false
}

func (c *Cli) execLine(line string) bool {
	if strings.Index(line, ";") > -1 {
		return c.execMultipleLine(line)
	}
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

	if strings.ToUpper(line) == "EXIT" {
		fmt.Println("Thanks for using freedb")
		return true
	}
	return false
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

func (c *Cli) outputList(krl *[]*kv.KeyRecord) {
	if len(*krl) == 0 {
		fmt.Println("[]")
		return
	}

	var b []byte
	var err error
	if c.conf.shortOutput {
		var shortlist []string
		for _, kr := range *krl {
			shortlist = append(shortlist, kr.Name)
		}
		b, err = json.MarshalIndent(shortlist, "", "  ")
	} else {
		b, err = json.MarshalIndent(krl, "", "  ")
	}
	if err != nil {
		c.log.Error(fmt.Sprintln(err))
		return
	}
	fmt.Println(string(b))
}
