package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type doForService func(string, yamlConfig) (int, error)

type Dcm struct {
	Config *Config
	Args   []string
}

func NewDcm(c *Config, args []string) *Dcm {
	return &Dcm{c, args}
}

func (d *Dcm) Command() (int, error) {
	if len(d.Args) < 1 {
		d.Usage()
		return 1, nil
	}

	moreArgs := d.Args[1:]

	switch d.Args[0] {
	case "help", "h":
		d.Usage()
		return 0, nil
	case "setup":
		return d.Setup()
	case "run", "r":
		return d.Run(moreArgs...)
	case "build", "b":
		return d.Run("build")
	case "dir", "d":
		return d.Dir(moreArgs...)
	case "shell", "sh":
		return d.Shell(moreArgs...)
	case "branch", "br":
		return d.Branch(moreArgs...)
	case "update", "u":
		return d.Update(moreArgs...)
	case "purge", "rm":
		return d.Purge(moreArgs...)
	default:
		d.Usage()
		return 127, nil
	}
}

func (d *Dcm) Setup() (int, error) {
	if _, err := os.Stat(d.Config.Srv); os.IsNotExist(err) {
		os.MkdirAll(d.Config.Srv, 0777)
	}

	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		repo, ok := getMapVal(configs, "labels", "com.dcm.repository").(string)
		if !ok {
			return 1, errors.New(
				"Error reading git repository config for service: " + service)
		}
		dir := d.Config.Srv + "/" + service
		cmd := cmd("git", "clone", repo, dir)
		cmd.Dir = d.Config.Dir
		if err := cmd.Run(); err != nil {
			return 1, errors.New("Error cloning git repository for service: " + service)
		}

		return 0, nil
	})
}

func (d *Dcm) doForEachService(fn doForService) (int, error) {
	for service, configs := range d.Config.Config {
		service, _ := service.(string)
		configs, ok := configs.(yamlConfig)
		if !ok {
			panic("Error reading configs for service: " + service)
		}

		code, err := fn(service, configs)
		if err != nil {
			return code, err
		}
	}

	return 0, nil
}

func (d *Dcm) Run(args ...string) (int, error) {
	if len(args) == 0 {
		args = append(args, "default")
	}

	switch args[0] {
	case "execute":
		return d.runExecute(args[1:]...)
	case "init":
		fmt.Println("Initializing project [" + d.Config.Project + "]...")
		return d.runInit()
	case "build":
		fmt.Println("Building project [" + d.Config.Project + "]...")
		return d.Run("execute", "build")
	case "start":
		fmt.Println("Starting project [" + d.Config.Project + "]...")
		return d.Run("execute", "start")
	case "stop":
		fmt.Println("Stopping project [" + d.Config.Project + "]...")
		return d.Run("execute", "stop")
	case "restart":
		fmt.Println("Restarting project [" + d.Config.Project + "]...")
		return d.Run("execute", "restart")
	case "up":
		fmt.Println("Bringing up project [" + d.Config.Project + "]...")
		return d.runUp()
	default:
		return d.Run("up")
	}
}

func (d *Dcm) runExecute(args ...string) (int, error) {
	cmd := cmd("docker-compose", args...)
	cmd.Dir = d.Config.Dir
	cmd.Env = append(
		os.Environ(),
		"COMPOSE_PROJECT_NAME="+d.Config.Project,
		"COMPOSE_FILE="+d.Config.File,
	)
	if err := cmd.Run(); err != nil {
		return 1, fmt.Errorf(
			"Error executing docker-compose with args [%s], and envs [%s]",
			strings.Join(args, ", "), strings.Join(cmd.Env, ", "),
		)
	}
	return 0, nil
}

func (d *Dcm) runInit() (int, error) {
	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		init, ok := getMapVal(configs, "labels", "com.dcm.initscript").(string)
		if !ok {
			fmt.Println("Skipping init script for service:", service, "...")
			return 0, nil
		}
		if err := os.Chdir(d.Config.Srv); err != nil {
			return 1, err
		}
		c := cmd("/bin/bash", init)
		c.Dir = d.Config.Srv + "/" + service + "/"
		if err := c.Run(); err != nil {
			return 1, fmt.Errorf(
				"Error executing init script [%s] for service [%s]: %s",
				init, service, err,
			)
		}
		branch, ok := getMapVal(configs, "labels", "com.dcm.branch").(string)
		if ok {
			if err := cmd("git", "checkout", branch).Run(); err != nil {
				return 1, err
			}
		}
		return 0, nil
	})
}

func (d *Dcm) runUp() (int, error) {
	code, err := d.Run("execute", "up", "-d", "--force-recreate")
	if err != nil {
		return code, err
	}
	return d.Run("init")
}

func (d *Dcm) Dir(args ...string) (int, error) {
	var dir string
	if len(args) < 1 {
		dir = d.Config.Dir
	} else {
		dir = d.Config.Srv + "/" + args[0]
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			dir = d.Config.Dir
		}
	}
	fmt.Fprint(os.Stdout, dir)
	return 0, nil
}

func (d *Dcm) Shell(args ...string) (int, error) {
	if len(args) < 1 {
		return 1, errors.New("Error: no service name specified.")
	}

	cid, err := d.getContainerId(args[0])
	if err != nil {
		return 1, err
	}

	err = cmd("docker", "exec", "-it", cid, "bash").Run()
	if err != nil {
		return 1, err
	}

	return 0, nil
}

func (d *Dcm) getContainerId(service string) (string, error) {
	filter := fmt.Sprintf("name=%s_%s_", d.Config.Project, service)
	out, err := exec.Command("docker", "ps", "-q", "-f", filter).Output()
	if err != nil {
		return "", err
	}

	cid := strings.TrimSpace(string(out))
	if cid == "" {
		return "", fmt.Errorf(
			"Error: no running container name starts with %s_%s_",
			d.Config.Project, service,
		)
	}

	return cid, nil
}

func (d *Dcm) getImageRepository(service string) (string, error) {
	first := exec.Command("docker", "images")
	second := exec.Command("awk", fmt.Sprintf(
		"$1 ~ /%s_%s/ { print $1 }", d.Config.Project, service))

	reader, writer := io.Pipe()
	first.Stdout = writer
	second.Stdin = reader
	var buff bytes.Buffer
	second.Stdout = &buff

	if err := first.Start(); err != nil {
		return "", err
	}
	if err := second.Start(); err != nil {
		return "", err
	}
	if err := first.Wait(); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	if err := second.Wait(); err != nil {
		return "", err
	}

	return strings.TrimSpace(buff.String()), nil
}

func (d *Dcm) Branch(args ...string) (int, error) {
	if len(args) < 1 {
		return d.branchForAll()
	} else {
		return d.branchForOne(args[0])
	}
}

func (d *Dcm) branchForAll() (int, error) {
	code, err := d.branchForOne("dcm")
	if err != nil {
		return code, err
	}
	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		return d.branchForOne(service)
	})
}

func (d *Dcm) branchForOne(service string) (int, error) {
	var dir string
	if service == "dcm" {
		dir = d.Config.Dir
	} else {
		dir = d.Config.Srv + "/" + service
	}

	if err := os.Chdir(dir); err != nil {
		return 1, err
	}

	fmt.Print(service + ": ")
	if err := cmd("git", "rev-parse", "--abbrev-ref", "HEAD").Run(); err != nil {
		return 1, err
	}

	return 0, nil
}

func (d *Dcm) Update(args ...string) (int, error) {
	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		fmt.Print(service + ": ")
		if err := os.Chdir(d.Config.Srv + "/" + service); err != nil {
			return 1, err
		}
		branch, ok := getMapVal(configs, "labels", "com.dcm.branch").(string)
		if !ok {
			// When service > labels > com.dcm.branch is not defined in
			// the yaml config file, use "master" as default branch
			branch = "master"
		}
		if err := cmd("git", "checkout", branch).Run(); err != nil {
			return 1, err
		}
		if err := cmd("git", "pull").Run(); err != nil {
			return 1, err
		}
		return 0, nil
	})
}

func (d *Dcm) Purge(args ...string) (int, error) {
	if len(args) == 0 {
		args = append(args, "default")
	}

	switch args[0] {
	case "img", "images":
		return d.purgeImages()
	case "con", "containers":
		return d.purgeContainers()
	case "all":
		return d.purgeAll()
	default:
		return d.Purge("containers")
	}
}

func (d *Dcm) purgeImages() (int, error) {
	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		repo, err := d.getImageRepository(service)
		if err != nil {
			return 1, err
		}
		if err := cmd("docker", "rmi", repo).Run(); err != nil {
			return 1, err
		}
		return 0, nil
	})
}

func (d *Dcm) purgeContainers() (int, error) {
	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		cid, err := d.getContainerId(service)
		if err != nil {
			return 1, err
		}
		if err := cmd("docker", "kill", cid).Run(); err != nil {
			return 1, err
		}
		if err := cmd("docker", "rm", "-v", cid).Run(); err != nil {
			return 1, err
		}
		return 0, nil
	})
}

func (d *Dcm) purgeAll() (int, error) {
	code, err := d.Purge("containers")
	if err != nil {
		return code, err
	}
	return d.Purge("images")
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
