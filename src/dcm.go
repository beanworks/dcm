package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type doForService func(string, yamlConfig) (int, error)

type Dcm struct {
	Config *Config
	Args   []string
	Cmd    Executable
}

func NewDcm(c *Config, args []string) *Dcm {
	return &Dcm{c, args, NewCmd()}
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
	case "dir":
		return d.Dir(moreArgs...)
	case "shell", "sh":
		return d.Shell(moreArgs...)
	case "branch", "br":
		return d.Branch(moreArgs...)
	case "update":
		return d.Update(moreArgs...)
	case "purge", "rm":
		return d.Purge(moreArgs...)
	case "list", "ls":
		return d.List()
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
		_, ok := getMapVal(configs, "image").(string)
		if ok {
			// If image is defined for the service, then skip
			// checking out the repository
			return 0, nil
		}
		repo, ok := getMapVal(configs, "labels", "dcm.repository").(string)
		if !ok {
			return 1, fmt.Errorf(
				"Error reading git repository config for service [%s]",
				service,
			)
		}
		dir := d.Config.Srv + "/" + service
		if _, err := os.Stat(dir); err == nil {
			fmt.Printf("Skipping git clone for %s. Service folder already exists.\n", service)
			return 0, nil
		}
		c := d.Cmd.Exec("git", "clone", repo, dir).Setdir(d.Config.Dir)
		if err := c.Run(); err != nil {
			return 1, fmt.Errorf(
				"Error cloning git repository for service [%s]: %v",
				service, err,
			)
		}
		branch, ok := getMapVal(configs, "labels", "dcm.branch").(string)
		if ok {
			c = d.Cmd.Exec("git", "checkout", branch).Setdir(dir)
			if err := c.Run(); err != nil {
				return 1, err
			}
		}
		return 0, nil
	})
}

func (d *Dcm) doForEachService(fn doForService) (int, error) {
	for service, configs := range d.Config.Config {
		service, _ := service.(string)
		configs, ok := configs.(yamlConfig)
		if !ok {
			return 1, fmt.Errorf("Error reading configs for service: %s", service)
		}

		code, err := fn(service, configs)
		if err != nil {
			if code == 0 {
				fmt.Println(err)
			} else {
				// Only when error code is not zero and error is not nil
				// then break the iteration and return
				return code, err
			}
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
		fmt.Println("Initializing project:", d.Config.Project, "...")
		return d.runInit()
	case "build":
		fmt.Println("Building project:", d.Config.Project, "...")
		return d.Run("execute", "build")
	case "start":
		fmt.Println("Starting project:", d.Config.Project, "...")
		return d.Run("execute", "start")
	case "stop":
		fmt.Println("Stopping project:", d.Config.Project, "...")
		return d.Run("execute", "stop")
	case "restart":
		fmt.Println("Restarting project:", d.Config.Project, "...")
		return d.Run("execute", "restart")
	case "up":
		fmt.Println("Bringing up project:", d.Config.Project, "...")
		return d.runUp()
	default:
		return d.Run("up")
	}
}

func (d *Dcm) runExecute(args ...string) (int, error) {
	env := append(
		os.Environ(),
		"COMPOSE_PROJECT_NAME="+d.Config.Project,
		"COMPOSE_FILE="+d.Config.File,
	)
	c := d.Cmd.
		Exec("docker-compose", args...).
		Setdir(d.Config.Dir).
		Setenv(env)
	if err := c.Run(); err != nil {
		return 1, fmt.Errorf(
			"Error executing `docker-compose %s`: %v",
			strings.Join(args, " "), err,
		)
	}
	return 0, nil
}

func (d *Dcm) runInit() (int, error) {
	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		init, ok := getMapVal(configs, "labels", "dcm.initscript").(string)
		if !ok {
			fmt.Println("Skipping init script for service:", service, "...")
			return 0, nil
		}
		c := d.Cmd.Exec("/bin/bash", init).Setdir(d.Config.Srv + "/" + service)
		if err := c.Run(); err != nil {
			return 1, fmt.Errorf(
				"Error executing init script [%s] for service [%s]: %v",
				init, service, err,
			)
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

	if err := d.Cmd.Exec("docker", "exec", "-it", cid, "bash").Run(); err != nil {
		return 1, err
	}

	return 0, nil
}

func (d *Dcm) getContainerId(service string) (string, error) {
	filter := fmt.Sprintf("name=%s_%s_", d.Config.Project, service)
	out, err := d.Cmd.Exec("docker", "ps", "-q", "-f", filter).Out()
	if err != nil {
		return "", d.Cmd.FormatError(err, out)
	}

	cid := d.Cmd.FormatOutput(out)
	if cid == "" {
		return "", fmt.Errorf(
			"Error: no running container name starts with %s_%s_",
			d.Config.Project, service,
		)
	}

	return cid, nil
}

func (d *Dcm) getImageRepository(service string) (string, error) {
	repo := d.Config.Project + "_" + service
	out, err := d.Cmd.Exec("docker", "images").Out()
	if err != nil {
		return "", d.Cmd.FormatError(err, out)
	}
	if strings.Contains(string(out), repo+" ") {
		return repo, nil
	}
	return "", nil
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

	fmt.Print(service + ": ")

	if service == "dcm" {
		fmt.Print("branch: ")
		dir = d.Config.Dir
	} else {
		configs, ok := getMapVal(d.Config.Config, service).(yamlConfig)
		if !ok {
			return 0, errors.New("Service not exists.")
		}
		if image, ok := getMapVal(configs, "image").(string); ok {
			fmt.Println("Docker hub image:", image)
			return 0, nil
		}
		if repo, ok := getMapVal(configs, "labels", "dcm.repository").(string); ok {
			fmt.Print("Git repo: ", repo, ", branch: ")
		}
		dir = d.Config.Srv + "/" + service
	}
	if err := os.Chdir(dir); err != nil {
		return 0, err
	}
	if err := d.Cmd.Exec("git", "rev-parse", "--abbrev-ref", "HEAD").Run(); err != nil {
		return 0, err
	}

	return 0, nil
}

func (d *Dcm) Update(args ...string) (int, error) {
	if len(args) < 1 {
		return d.updateForAll()
	} else {
		return d.updateForOne(args[0])
	}
}

func (d *Dcm) updateForAll() (int, error) {
	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		return d.updateForOne(service)
	})
}

func (d *Dcm) updateForOne(service string) (int, error) {
	fmt.Print(service + ": ")

	configs, ok := getMapVal(d.Config.Config, service).(yamlConfig)
	if !ok {
		return 0, errors.New("Service not exists.")
	}

	updateable, ok := getMapVal(configs, "labels", "dcm.updateable").(string)
	if ok && updateable == "false" {
		// Service is flagged as not updateable
		return 0, errors.New("Service not updateable. Skipping the update.")
	}

	image, ok := getMapVal(configs, "image").(string)
	if ok {
		// Service is using docker hub image
		// Pull the latest version from docker hub
		if err := d.Cmd.Exec("docker", "pull", image).Run(); err != nil {
			return 0, err
		}
		return 0, nil
	} else {
		// Service is using a local build
		// Pull the latest version from git
		if err := os.Chdir(d.Config.Srv + "/" + service); err != nil {
			return 0, err
		}
		branch, ok := getMapVal(configs, "labels", "dcm.branch").(string)
		if !ok {
			// When service > labels > dcm.branch is not defined in
			// the yaml config file, use "master" as default branch
			branch = "master"
		}
		if err := d.Cmd.Exec("git", "checkout", branch).Run(); err != nil {
			return 0, err
		}
		if err := d.Cmd.Exec("git", "pull").Run(); err != nil {
			return 0, err
		}
	}

	return 0, nil
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
			return 0, err
		}
		if err = d.Cmd.Exec("docker", "rmi", repo).Run(); err != nil {
			return 0, err
		}
		return 0, nil
	})
}

func (d *Dcm) purgeContainers() (int, error) {
	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		cid, err := d.getContainerId(service)
		if err != nil {
			return 0, err
		}
		if err := d.Cmd.Exec("docker", "kill", cid).Run(); err != nil {
			return 0, err
		}
		if err := d.Cmd.Exec("docker", "rm", "-v", cid).Run(); err != nil {
			return 0, err
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

func (d *Dcm) List() (int, error) {
	return d.doForEachService(func(service string, configs yamlConfig) (int, error) {
		fmt.Fprintln(os.Stdout, service)
		return 0, nil
	})
}

func (d *Dcm) Usage() {
	fmt.Println("")
	fmt.Println("DCM (Docker-Compose Manager)")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  dcm help                Show this help menu.")
	fmt.Println("  dcm setup               Git checkout repositories for the services that require")
	fmt.Println("                          local docker build. It skips the service when the image")
	fmt.Println("                          is from docker hub, or the repo's folder already exists.")
	fmt.Println("  dcm run [<args>]        Run docker-compose commands. If <args> is not given, by")
	fmt.Println("                          default DCM will run `docker-compose up` command.")
	fmt.Println("                          <args>: up, build, start, stop, restart")
	fmt.Println("  dcm build               Docker (re)build service images that require local build.")
	fmt.Println("                          It's the shorthand version of `dcm run build` command.")
	fmt.Println("  dcm shell <service>     Log into a given service container.")
	fmt.Println("  dcm purge [<type>]      Remove either all the containers or all the images. If <type>")
	fmt.Println("                          is not given, by default DCM will purge everything.")
	fmt.Println("                          <type>: images, containers, all")
	fmt.Println("  dcm branch [<service>]  Display the current git branch for the given service that")
	fmt.Println("                          was built locally.")
	fmt.Println("  dcm goto [<service>]    Go to the service's folder. If <service> is not given, by")
	fmt.Println("                          default DCM will go to $DCM_DIR.")
	fmt.Println("  dcm update [<service>]  Update DCM and(or) the given service.")
	fmt.Println("  dcm list                List all the available services.")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Println("  Initial setup")
	fmt.Println("    dcm setup")
	fmt.Println("    dcm run")
	fmt.Println("")
	fmt.Println("  Rebuild")
	fmt.Println("    dcm build")
	fmt.Println("    dcm run")
	fmt.Println("")
	fmt.Println("  Or only Rerun")
	fmt.Println("    dcm run")
	fmt.Println("")
	fmt.Println("  Log into a service's container")
	fmt.Println("    dcm shell service_name")
	fmt.Println("")
}
