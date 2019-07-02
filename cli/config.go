package cli

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

type dslInstruct struct {
	args int
	exec func(args []string)
}

var dslInstructs = make(map[string]*dslInstruct)

func (c *cli) initDSL() {
	dslInstructs["SET"] = &dslInstruct{
		args: 2,
		exec: c.set,
	}

	dslInstructs["GET"] = &dslInstruct{
		args: 1,
		exec: c.get,
	}
	dslInstructs["APPEND"] = &dslInstruct{
		args: 2,
		exec: c.append,
	}
	dslInstructs["KEYS"] = &dslInstruct{
		args: 0,
		exec: c.keys,
	}
	dslInstructs["CONFIG"] = &dslInstruct{
		args: 2,
		exec: c.config,
	}
	dslInstructs["USE"] = &dslInstruct{
		args: 1,
		exec: c.use,
	}
	dslInstructs["DELETE"] = &dslInstruct{
		args: 1,
		exec: c.delete,
	}
	dslInstructs["EXIT"] = &dslInstruct{
		args: 0,
		exec: func(args []string) {},
	}
}
