package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	helpStr = "\n    > You have to specify a working directory with the -dir flag (every time you run gaze). This is where gaze stores all config and tv data.\n\n    > TIP: create an alias for 'gaze -dir /path/to/data/dir'.\n"
)

const (
	logo = `
     ______     ______     ______     ______   
    /\  ___\   /\  __ \   /\___  \   /\  ___\  
    \ \ \__ \  \ \  __ \  \/_/  /__  \ \  __\  
     \ \_____\  \ \_\ \_\   /\_____\  \ \_____\
      \/_____/   \/_/\/_/   \/_____/   \/_____/


`
)

var (
	dir = ""
)

func mountDataDir(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	finfo, err := os.Stat(abs)

	// If the path does not exist.
	if err != nil && os.IsNotExist(err) {
		// Create directory.
		err = os.MkdirAll(abs, 0755)
		if err != nil {
			return err
		}
		fmt.Println("Created working directory at", abs)
		dir = abs
	} else {
		// Test if it is a directory.
		if !finfo.IsDir() {
			return errors.New(abs + " expected directory but got file")
		}
		dir = abs
	}

	return err
}

func main() {
	path := flag.String("dir", "", "Sets the working directory of gaze where config and tv data is stored (required).")
	flag.Parse()
	fmt.Print(logo)
	if path == nil || *path == "" {
		bye(helpStr, 0)
	}

	if err := mountDataDir(*path); err != nil {
		bye(fmt.Sprintf("Error mounting working directory: %s", err.Error()), 1)
	}

	loop := cmdLoop{}

	loop.run()
}
