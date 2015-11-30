package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Dir, Project, Srv string
	Config            map[string]interface{}
}

func NewConfig() *Config {
	c := &Config{}
	c.Dir, _ = os.Getwd()
	c.Project = "bean"

	return c.loadEnvConfig()
}

func NewConfigFile() *Config {
	c := NewConfig()
	path := c.Dir + "/" + c.Project + ".yml"
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(content, c.Config); err != nil {
		panic(fmt.Sprintf("Error parsing config file: %s", err))
	}

	return c
}

func (c *Config) loadEnvConfig() *Config {
	if env := os.Getenv("DCM_DIR"); env != "" {
		c.Dir = env
	}
	if env := os.Getenv("DCM_PROJECT"); env != "" {
		c.Project = env
	}

	c.Srv = c.Dir + "/srv/" + c.Project

	return c
}
