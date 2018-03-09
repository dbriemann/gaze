package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type processor struct {
	reader   *bufio.Reader
	commands []command
}

func newProcessor() *processor {
	p := processor{}
	p.reader = bufio.NewReader(os.Stdin)
	p.commands = []command{
		command{
			long:  "help",
			short: "h",
			desc:  "lists all commands",
			fun:   cmdHelp,
		},
		command{
			long:  "import",
			short: "i",
			desc:  "imports favorites from thetvdb.com",
			fun:   cmdImport,
		},
		command{
			long:  "quit",
			short: "q",
			desc:  "quits gaze",
			fun:   cmdQuit,
		},
		command{
			long:  "auth",
			short: "a",
			desc:  "enter and save your auth data",
			fun:   cmdSetAuth,
		},
	}

	return &p
}

func (p *processor) prompt() []string {
	fmt.Print(pad2 + "< ")
	line, err := p.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			fmt.Println("^D\n")
			cmdQuit(p, nil)
		}
		bye(fmt.Sprintf(pad2+"> something bad happened (%s)\n", err.Error()), 1)
	}
	fmt.Println("")

	line = strings.TrimSpace(line)
	return strings.Fields(line)
}

func (p *processor) run() {
	if !configData.hasAuthData() {
		cmdSetAuth(p, nil)
	}
	if err := tvdbEnsureLogin(); err != nil {
		bye(fmt.Sprintf(pad2+"> something bad happened (%s)\n", err.Error()), 1)
	}
	if len(db.Shows) == 0 {
		fmt.Printf("%s> use command 'import' or 'i' to import all your favorites from thetvdb.com\n\n", pad2)
	} else {
		db.updateAllShows()
	}

	for {
		if err := tvdbEnsureLogin(); err != nil {
			bye(fmt.Sprintf(pad2+"> something bad happened (%s)\n", err.Error()), 1)
		}
		line := p.prompt()
		if len(line) > 0 {
			for _, cmd := range p.commands {
				if line[0] == cmd.long || line[0] == cmd.short {
					err := cmd.fun(p, line[1:])
					if err != nil {
						fmt.Printf("%s> An error occured while executing command '%s': '%s'\n", pad2, cmd.long, err.Error())
					}
					goto COMMAND_EXECUTED
				}
			}
			fmt.Println(pad2 + "> command unkown")
			fmt.Println("")
			cmdHelp(p, line[1:])
		COMMAND_EXECUTED:
			fmt.Println("")
		}
	}
}
