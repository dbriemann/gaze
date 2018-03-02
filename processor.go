package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Processor struct {
	reader   *bufio.Reader
	commands []command
}

func newProcessor() *Processor {
	p := Processor{}
	p.reader = bufio.NewReader(os.Stdin)
	p.commands = []command{
		command{
			long:  "help",
			short: "h",
			desc:  "shows the settings and lists all commands",
			fun:   help,
		},
	}

	return &p
}

func (p *Processor) prompt() []string {
	fmt.Print(pad2 + "< ")
	line, err := p.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			bye("^D\n"+pad2+"> See you soon.", 0)
		}
		bye(fmt.Sprintf("Something bad happened (%s)\n", err.Error()), 1)
	}

	line = strings.TrimSpace(line)
	if line == "quit" || line == "exit" {
		bye(pad2+"> See you soon.", 0)
	}

	return strings.Fields(line)
}

func (p *Processor) run() {
	for {
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
			help(p, line[1:])
		COMMAND_EXECUTED:
			fmt.Println("")
		}
	}
}
