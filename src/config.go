package main

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type yamlConfig map[interface{}]interface{}

type Config struct {
	Dir, File, Project, Srv string
	Config                  yamlConfig
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

	if isDockerComposeVersion2(c.Config) {
		services, ok := getMapVal(c.Config, "services").(yamlConfig)
		if ok {
			c.Config = services
		}
	}

	return c, nil
}

func NewConfig() *Config {
	wd, _ := os.Getwd()
	c := &Config{Dir: wd, Project: "dcm"}
	return c.loadEnvConfig()
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

func isDockerComposeVersion2(config yamlConfig) bool {
	var ok bool
	_, ok = getMapVal(config, "version").(string)
	if !ok {
		return false
	}
	_, ok = getMapVal(config, "services").(yamlConfig)
	if !ok {
		return false
	}
	return true
}
