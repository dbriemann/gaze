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
	client   TVDBClient
}

func newProcessor() *processor {
	p := processor{}
	p.client = TVDBClient{}
	p.reader = bufio.NewReader(os.Stdin)
	p.commands = []command{
		command{
			long:  "help",
			short: "h",
			desc:  "shows the settings and lists all commands",
			fun:   cmdHelp,
		},
		command{
			long:  "quit",
			short: "q",
			desc:  "quits gaze",
			fun:   cmdQuit,
		},
		command{
			long:  "setauth",
			short: "sa",
			desc:  "queries the user for his auth data and saves it",
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
	for {
		if err := p.client.ensureLogin(); err != nil {
			bye(fmt.Sprintf(pad2+"> something bad happened (%s)\n", err.Error()), 1)
		}
		line := p.prompt()
		if len(line) > 0 {
			for _, cmd := range p.commands {
				if line[0] == cmd.long || line[0] == cmd.short {
					cmd.fun(p, line[1:]) // TODO err
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
