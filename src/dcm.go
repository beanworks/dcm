package main

type Dcm struct {
	config *Config
	args   []string
}

func NewDcm(c *Config, args []string) {
	return &Dcm{c, args}
}

func (d *Dcm) Setup() {
}
