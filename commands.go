package main

import (
	"fmt"
	"strings"
)

type command struct {
	long  string
	short string
	desc  string
	fun   func(*processor, []string) error
}

func cmdHelp(p *processor, args []string) error {
	// TODO settings status
	fmt.Printf("%s> %-10s | %-5s | %s\n", pad2, "command", "short", "description")
	fmt.Println(pad2 + "----------------------------------------------------------------------")
	for _, cmd := range p.commands {
		fmt.Printf("%s> %-10s | %-5s | %s\n", pad2, cmd.long, cmd.short, cmd.desc)
	}
	return nil
}

func cmdImport(p *processor, args []string) error {
	err := tvdbImportFavs()
	return err
}

func cmdQuit(p *processor, args []string) error {
	bye(pad2+"> See you soon.", 0)
	return nil
}

func cmdSetAuth(p *processor, args []string) error {
	fmt.Printf("\n%s> specify your 'thetvdb.com' data\n\n", pad2)
START_OVER:
	if configData.Auth.UserName == "" {
		fmt.Printf("%s> enter your user name\n", pad2)
	} else {
		fmt.Printf("%s> enter your user name (press enter to keep '%s')\n", pad2, configData.Auth.UserName)
	}
	reply := p.prompt()
	if len(reply) == 0 && configData.Auth.UserName == "" {
		goto START_OVER
	}
	if len(reply) != 0 {
		configData.Auth.UserName = strings.Join(reply, " ")
	}

	if configData.Auth.UserKey == "" {
		fmt.Printf("%s> enter your user key\n", pad2)
	} else {
		fmt.Printf("%s> enter your user key (press enter to keep '%s')\n", pad2, configData.Auth.UserKey)
	}
	reply = p.prompt()
	if len(reply) == 0 && configData.Auth.UserKey == "" {
		goto START_OVER
	}
	if len(reply) != 0 {
		configData.Auth.UserKey = reply[0]
	}

	if configData.Auth.ApiKey == "" {
		fmt.Printf("%s> enter your API key\n", pad2)
	} else {
		fmt.Printf("%s> enter your API key (press enter to keep '%s')\n", pad2, configData.Auth.ApiKey)
	}
	reply = p.prompt()
	if len(reply) == 0 && configData.Auth.ApiKey == "" {
		goto START_OVER
	}
	if len(reply) != 0 {
		configData.Auth.ApiKey = reply[0]
	}

ASK_AGAIN:
	fmt.Println(pad2 + "> is this data correct (y/n):")
	fmt.Printf("%s> %-9s: %s\n", pad2, "user name", configData.Auth.UserName)
	fmt.Printf("%s> %-9s: %s\n", pad2, "user key", configData.Auth.UserKey)
	fmt.Printf("%s> %-9s: %s\n", pad2, "api key", configData.Auth.ApiKey)

	reply = p.prompt()
	if len(reply) > 0 {
		if reply[0][0] == 'y' {
			configData.save()
			return nil
		} else if reply[0][0] == 'n' {
			goto START_OVER
		} else {
			goto ASK_AGAIN
		}
	} else {
		goto ASK_AGAIN
	}

	return nil
}
