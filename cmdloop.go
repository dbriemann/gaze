package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type cmdLoop struct {
	reader *bufio.Reader
}

func (c *cmdLoop) prompt() []string {
	fmt.Print(padLeft + "< ")
	line, err := c.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			bye("^D\n"+padLeft+"> See you soon.", 0)
		}
		bye(fmt.Sprintf("Something bad happened (%s)\n", err.Error()), 1)
	}

	line = strings.TrimSpace(line)
	if line == "quit" || line == "exit" {
		bye(padLeft+"> See you soon.", 0)
	}

	return strings.Fields(line)
}

func (c *cmdLoop) run() {
	c.reader = bufio.NewReader(os.Stdin)

	for {
		line := c.prompt()
		fmt.Printf(padLeft+"> Echo: %s\n\n", line)
	}
}
