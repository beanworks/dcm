package main

import (
	"fmt"
	"os"
	"os/exec"
)

type Dcm struct {
	Config *Config
	Args   []string
}

func NewDcm(c *Config, args []string) *Dcm {
	return &Dcm{c, args}
}

func (d *Dcm) Command() {
	switch d.Args[0] {
	case "help", "h":
		d.Usage()
	case "setup":
		d.Setup()
	default:
		d.Usage()
		os.Exit(127)
	}
}

func (d *Dcm) Setup() {
	if _, err := os.Stat(d.Config.Srv); os.IsNotExist(err) {
		os.MkdirAll(d.Config.Srv, 0777)
	}

	for service, configs := range d.Config.Config {
		configs, ok := configs.(map[interface{}]interface{})
		if !ok {
			continue
		}

		repo := d.GetMapValue(configs, "labels", "com.dcm.repository")
		dir := d.Config.Srv + "/" + service.(string)
		cmd := exec.Command("git", "clone", repo.(string), dir)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "Error running git clone cmd", err)
			os.Exit(1)
		}
	}
}

func (d *Dcm) GetMapValue(v map[interface{}]interface{}, keys ...string) interface{} {
	if len(keys) == 0 {
		return v
	}

	if len(keys) == 1 {
		return v[keys[0]]
	}

	v, ok := v[keys[0]].(map[interface{}]interface{})
	if !ok {
		fmt.Fprintln(os.Stderr, "Error asserting the type of yaml config")
		os.Exit(1)
	}

	return d.GetMapValue(v, keys[1:]...)
}

func (d *Dcm) Usage() {
	fmt.Println("")
	fmt.Println("Docker Compose Manager")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  dcm help                Show this message")
	fmt.Println("                          shorthand ver.: `dcm h`")
	fmt.Println("  dcm setup               Check out all the repositories for API, UI and services")
	fmt.Println("  dcm run [<args>]        Run docker-compose commands. If <args> is not given, by")
	fmt.Println("                          default DCM will run `up` command.")
	fmt.Println("                          <args>: up, build, start, stop, restart")
	fmt.Println("                          shorthand ver.: `dcm r [<args>]`")
	fmt.Println("  dcm run build           Run build command that (re)create all the service images")
	fmt.Println("                          shorthand ver.: `dcm build` or `dcm b`")
	fmt.Println("  dcm shell <service>     Log into a given service container")
	fmt.Println("                          <service>: api, ui, postgres, mongo, redis, nginx, php")
	fmt.Println("                          shorthand ver.: `dcm sh <service>`")
	fmt.Println("  dcm purge [<type>]      Remove either all the containers or all the images or")
	fmt.Println("                          everything. If <type> is not given, by default DCM will")
	fmt.Println("                          purge everything")
	fmt.Println("                          <type>: images|img, containers|con, all")
	fmt.Println("                          shorthand ver.: `dcm rm [<type>]`")
	fmt.Println("  dcm branch <service>    Display the current branch for the given service name")
	fmt.Println("                          <service>: api, ui, postgres, mongo, redis, nginx, php")
	fmt.Println("                          shorthand ver.: `dcm br <service>`")
	fmt.Println("  dcm goto [<service>]    Go to the service's folder. If <service> is not given,")
	fmt.Println("                          by default DCM will go to $DCM_DIR")
	fmt.Println("                          <service>: api, ui, postgres, mongo, redis, nginx, php")
	fmt.Println("                          shorthand ver.: `dcm gt [<service>]`")
	fmt.Println("  dcm update [<service>]  Update DCM and dependent services (PostgrSQL, MongoDB,")
	fmt.Println("                          Redis, Nginx and Base PHP). If <service> is not given,")
	fmt.Println("                          by default DCM will update everything except api and ui.")
	fmt.Println("                          <service>: postgres, mongo, redis, nginx, php")
	fmt.Println("                          shorthand ver.: `dcm u`")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("  Initial setup")
	fmt.Println("    dcm setup")
	fmt.Println("    dcm run")
	fmt.Println("")
	fmt.Println("  Rebuild API or UI after switching branch")
	fmt.Println("    dcm build")
	fmt.Println("    dcm run")
	fmt.Println("")
	fmt.Println("  Log into different service containers")
	fmt.Println("    dcm shell api")
	fmt.Println("    dcm shell ui")
	fmt.Println("    ...")
	fmt.Println("")
}
