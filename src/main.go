package main

import (
	"fmt"
	"os"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintln(os.Stderr, "DCM error:", err)
		}
	}()

	conf := NewConfigFile()
	args := os.Args[1:]

	dcm := NewDcm(conf, args)
	code, err := dcm.Command()
	if err != nil {
		fmt.Fprintln(os.Stderr, "DCM error:", err)
	}
	os.Exit(code)
}
