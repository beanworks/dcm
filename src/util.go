package main

import (
	"os"
	"os/exec"
)

func getMapVal(v yamlConfig, keys ...string) interface{} {
	if len(keys) == 0 {
		return v
	}

	if vv, ok := v[keys[0]]; ok {
		if len(keys) == 1 {
			return vv
		}

		if vvv, ok := vv.(yamlConfig); ok {
			return getMapVal(vvv, keys[1:]...)
		}
	}

	return nil
}

func cmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	return cmd
}
