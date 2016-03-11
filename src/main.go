package main

import (
	"fmt"
	"os"
)

func main() {
	if code, err := execDcmCmd(); err != nil {
		fmt.Fprintln(os.Stderr, "DCM error:", err)
		os.Exit(code)
	}
}

func execDcmCmd() (int, error) {
	conf, err := NewConfigFile()
	if err != nil {
		return 1, err
	}
	args := os.Args[1:]
	dcm := NewDcm(conf, args)
	code, err := dcm.Command()
	if err != nil {
		return code, err
	}
	return 0, nil
}
