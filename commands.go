package main

import (
	"fmt"
	"strings"
	"time"
)

type command struct {
	long  string
	short string
	desc  string
	fun   func(*processor, []string) error
}

func cmdHelp(p *processor, args []string) error {
	// TODO settings status
	fmt.Printf("%s> command    | description\n", pad2)
	fmt.Println(pad2 + strings.Repeat("-", 76))
	for _, cmd := range p.commands {
		fmt.Printf("%s> %-10s | %s\n", pad2, fmt.Sprintf("(%s)%s", cmd.short, cmd.long[1:]), cmd.desc)
	}
	return nil
}

func cmdImport(p *processor, args []string) error {
	db.importFavorites()
	return nil
}

func cmdQuit(p *processor, args []string) error {
	bye(pad2+"> See you soon.", 0)
	return nil
}

func cmdUpcoming(p *processor, args []string) error {
	upcomers := [3][]*episode{}
	upcomers[0] = []*episode{} // today
	upcomers[1] = []*episode{} // tomorrow
	upcomers[2] = []*episode{} // in 2 days

	for _, ep := range db.Episodes {
		// Parse ep data format
		epdate, err := time.Parse("2006-01-02", ep.FirstAired)
		if err != nil {
			// This should be logged if there was a log.
			// We ignore errors here. Many episodes that air in the future
			// have incomplete information, including missing dates.
		} else {
		}

		now := time.Now()
		if epdate.Year() == now.Year() && epdate.Month() == now.Month() {
			if epdate.Day() == now.Day() {
				upcomers[0] = append(upcomers[0], ep)
			} else if epdate.Day() == now.Add(time.Hour*24).Day() {
				upcomers[1] = append(upcomers[1], ep)
			} else if epdate.Day() == now.Add(time.Hour*48).Day() {
				upcomers[2] = append(upcomers[2], ep)
			}
		}
	}

	for i, eps := range upcomers {
		for _, ep := range eps {
			fmt.Printf("%s> %d: %dx%d %v\n", pad2, i, ep.Season, ep.Number, ep.Name)
		}
	}

	return nil
}

func cmdWatch(p *processor, args []string) error {
	// TODO
	fmt.Printf("%s> select shows/episodes that you have watched\n", pad2)
	fmt.Printf("%s> seperate multiple selections with space\n", pad2)
	fmt.Printf("%s(%3d) %50s\n", pad2, 0, "<*> all before today <*>")
	counter := 1
	for _, show := range db.Shows {
		fmt.Printf("%s(%3d) %50s\n", pad2, counter, show.Name)
		fmt.Println(show.EpisodeIDs)
		counter++
	}
	return nil
}

func cmdAuth(p *processor, args []string) error {
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
