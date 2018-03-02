package main

import "fmt"

type command struct {
	long  string
	short string
	desc  string
	fun   func(*Processor, []string) error
}

func help(p *Processor, args []string) error {
	// TODO settings status
	listCommands(p, nil)

	return nil
}

func listCommands(p *Processor, args []string) {
	fmt.Println(pad2 + "> available commands:")
	for _, cmd := range p.commands {
		fmt.Printf("%s> %s(%s): %s\n", pad2, cmd.long, cmd.short, cmd.desc)
	}
}
