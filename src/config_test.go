package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var yamlFixtureGood string = `
foo:
  bar:
    baz: qux
another: value
yet:
  another:
    - value1
    - value2
    - value3
`

var yamlFixtureBad string = `
foo: bar
    - baz: qux
`

func TestCreateNewConfig(t *testing.T) {
	os.Unsetenv("DCM_DIR")
	os.Unsetenv("DCM_PROJECT")

	c := NewConfig()
	wd, _ := os.Getwd()

	assert.Equal(t, &Config{
		Dir:     wd,
		Project: "bean",
		File:    wd + "/bean.yml",
		Srv:     wd + "/srv/bean",
	}, c)
}

func TestCreateNewConfigWithEnvVars(t *testing.T) {
	os.Setenv("DCM_DIR", "/test/dcm/dir")
	os.Setenv("DCM_PROJECT", "testproj")

	c := NewConfig()

	assert.Equal(t, &Config{
		Dir:     "/test/dcm/dir",
		Project: "testproj",
		File:    "/test/dcm/dir/testproj.yml",
		Srv:     "/test/dcm/dir/srv/testproj",
	}, c)
}

func helperCreateTestFile(t *testing.T, prefix, fixture string) string {
	tf, err := ioutil.TempFile("", prefix)
	require.Nil(t, err)

	_, err = tf.WriteString(fixture)
	require.Nil(t, err)

	err = tf.Close()
	require.Nil(t, err)

	return tf.Name()
}

func TestCreateNewConfigFile(t *testing.T) {
	var file string

	// Negative case: Bad YAML file name
	os.Setenv("DCM_CONFIG_FILE", "bad_file_name.yml")

	assert.Panics(t, func() { NewConfigFile() })

	// Negative case: Bad YAML config formatting
	file = helperCreateTestFile(t, "bad_yaml", yamlFixtureBad)
	defer os.Remove(file)
	os.Setenv("DCM_CONFIG_FILE", file)

	assert.Panics(t, func() { NewConfigFile() })

	// Positive case: success
	file = helperCreateTestFile(t, "good_yaml", yamlFixtureGood)
	defer os.Remove(file)
	os.Setenv("DCM_CONFIG_FILE", file)
	c := NewConfigFile()

	assert.Equal(t, yamlConfig{
		"foo": yamlConfig{
			"bar": yamlConfig{
				"baz": "qux",
			},
		},
		"another": "value",
		"yet": yamlConfig{
			"another": []interface{}{
				"value1",
				"value2",
				"value3",
			},
		},
	}, c.Config)
}
