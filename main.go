package main

import (
	"flag"
	"fmt"
)

const (
	helpStr = "\n    > You have to specify a working directory with the -dir flag (every time you run gaze). This is where gaze stores all config and tv data.\n\n    > TIP: create an alias for 'gaze -dir /path/to/data/dir'.\n"
)

const (
	logo = `
     ______     ______     ______     ______   
    /\  __/_   /\  __ \   /\___  \   /\  ___\  
    \ \ \__ \  \ \  __ \  \/_/  /__  \ \  __\  
     \ \_____\  \ \_\ \_\   /\_____\  \ \_____\
      \/_____/   \/_/\/_/   \/_____/   \/_____/


`
)

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
	loadConfig()

	processor := newProcessor()

	processor.run()
}
