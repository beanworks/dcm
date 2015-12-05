package main

import (
	"os"
	"os/exec"
)

func getMapVal(v map[interface{}]interface{}, keys ...string) interface{} {
	if len(keys) == 0 {
		return v
	}

	if len(keys) == 1 {
		return v[keys[0]]
	}

	v, ok := v[keys[0]].(map[interface{}]interface{})
	if !ok {
		panic("Error asserting the type of yaml config at key: " + keys[0])
	}

	return getMapVal(v, keys[1:]...)
}

func cmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd
}
