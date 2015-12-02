package main

import "os"

func main() {
	args := os.Args[1:]
	conf := NewConfigFile()
	dcm := NewDcm(conf, args)

	if len(args) < 1 {
		dcm.Usage()
		return
	}

	dcm.Command()
}
