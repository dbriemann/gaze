package main

import (
	"fmt"
	"os"
)

const (
	padLeft = "  "
)

func bye(msg string, code int) {
	fmt.Println(msg)
	os.Exit(code)
}
