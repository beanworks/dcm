package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Dir, Path, Project, Srv string
	Config                  map[interface{}]interface{}
}

func NewConfig() *Config {
	c := &Config{}
	c.Dir, _ = os.Getwd()
	c.Project = "bean"

	return c.loadEnvConfig()
}

func NewConfigFile() *Config {
	c := NewConfig()
	content, err := ioutil.ReadFile(c.Path)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(content, &c.Config); err != nil {
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

	c.Path = c.Dir + "/" + c.Project + ".yml"
	c.Srv = c.Dir + "/srv/" + c.Project

	return c
}
