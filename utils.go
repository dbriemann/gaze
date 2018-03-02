package main

import (
	"fmt"
	"os"
)

const (
	pad2 = "  "
)

func bye(msg string, code int) {
	fmt.Println(msg)
	os.Exit(code)
}
