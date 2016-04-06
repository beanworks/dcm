package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecDcmCmd(t *testing.T) {
	file := helperCreateTestFile(t, "good_yaml", yamlFixtureGood)
	defer os.Remove(file)
	os.Setenv("DCM_CONFIG_FILE", file)
	code, err := execDcmCmd()
	require.Equal(t, 0, code)
	require.NoError(t, err)
}
