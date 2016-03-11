package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type yamlConfig map[interface{}]interface{}

type Config struct {
	Dir, File, Project, Srv string
	Config                  yamlConfig
}

func NewConfig() *Config {
	c := &Config{}
	c.Dir, _ = os.Getwd()
	c.Project = "bean"

	return c.loadEnvConfig()
}

func NewConfigFile() (*Config, error) {
	c := NewConfig()
	content, err := ioutil.ReadFile(c.File)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(content, &c.Config); err != nil {
		return nil, fmt.Errorf("Error parsing config file: %s", err)
	}

	return c, nil
}

func (c *Config) loadEnvConfig() *Config {
	if env := os.Getenv("DCM_DIR"); env != "" {
		c.Dir = env
	}
	if env := os.Getenv("DCM_PROJECT"); env != "" {
		c.Project = env
	}

	c.File = c.Dir + "/" + c.Project + ".yml"
	c.Srv = c.Dir + "/srv/" + c.Project

	// This is created for unit test
	if env := os.Getenv("DCM_CONFIG_FILE"); env != "" {
		c.File = env
	}

	return c
}
