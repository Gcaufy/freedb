package cli

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	helper "github.com/Gcaufy/freedb/helper"
	kv "github.com/Gcaufy/freedb/kv"
	prompt "github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

// Config type
type Config struct {
	host        *helper.Host
	hostStr     string
	token       string
	db          string
	branch      string
	execute     string
	shortOutput bool
	cache       bool
}

// Cli type
type cli struct {
	log  *ConsoleLogger
	conf *Config
	kv   *kv.KV
}

// Run is the method to create cli instance and run
func Run() {

	c := &cli{
		log: NewConsoleLogger(),
		conf: &Config{
			db:     "default",
			branch: "master",
			cache:  true,
		},
	}

	c.initDSL()
	c.initCLI()
}

func (c *cli) initCLI() {
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
				c.interact()
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

func (c *cli) completer(in prompt.Document) []prompt.Suggest {
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

func (c *cli) interact() {
	fmt.Print("freedb shell version v1.0.0\n")

	p := prompt.New(
		func(in string) { c.execLine(in) },
		c.completer,
		prompt.OptionShowCompletionAtStart(),
	)
	p.Run()
}

func (c *cli) execMultipleLine(line string) bool {
	lines := strings.Split(line, ";")
	for _, s := range lines {
		s = strings.TrimSpace(s)
		if s != "" {
			c.execLine(s)
		}
	}
	return false
}

func (c *cli) execLine(line string) bool {
	if strings.Index(line, ";") > -1 {
		return c.execMultipleLine(line)
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	arg0 := strings.ToUpper(fields[0])
	args := fields[1:]
	commandName := dslInstructs[arg0]

	if commandName != nil {
		argLen := len(args)
		if commandName.args == argLen {
			commandName.exec(args)
		} else {
			c.log.Error("Command \"%s\" expect %d arguments, but %d arguments got.", arg0, commandName.args, argLen)
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

func (c *cli) timeUse(fn func()) {
	if c.conf.shortOutput {
		fn()
	} else {
		s := int64(time.Now().UnixNano() / (1000 * 1000))
		fn()
		e := int64(time.Now().UnixNano() / (1000 * 1000))
		fmt.Printf("(%d ms)\n", e-s)
	}
}

func (c *cli) output(kr *kv.KeyRecord) {
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

func (c *cli) outputList(krl *[]*kv.KeyRecord) {
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
