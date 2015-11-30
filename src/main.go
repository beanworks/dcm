package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		usage()
	}
	conf := NewConfigFile()
	Dcm(conf, args)
}

func usage() {
	fmt.Println("abc")
}
