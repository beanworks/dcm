# DCM := "Docker-Compose Manager"

[![Build Status](https://travis-ci.org/beanworks/dcm.svg)](https://travis-ci.org/beanworks/dcm)
[![Coverage Status](https://coveralls.io/repos/beanworks/dcm/badge.svg?service=github)](https://coveralls.io/github/beanworks/dcm)
[![Go Report Card](https://goreportcard.com/badge/github.com/beanworks/dcm)](https://goreportcard.com/report/github.com/beanworks/dcm)

```text

         _____         _____        ______  _______
     ___|\    \    ___|\    \      |      \/       \
    |    |\    \  /    /\    \    /          /\     \
    |    | |    ||    |  |    |  /     /\   / /\     |
    |    | |    ||    |  |____| /     /\ \_/ / /    /|
    |    | |    ||    |   ____ |     |  \|_|/ /    / |
    |    | |    ||    |  |    ||     |       |    |  |
    |____|/____/||\ ___\/    /||\____\       |____|  /
    |    /    | || |   /____/ || |    |      |    | /
    |____|____|/  \|___|    | / \|____|      |____|/
      \(    )/      \( |____|/     \(          )/
       '    '        '   )/         '          '
                         '
```

DCM is a wrapper for docker-compose. It enables one click setup, build && run process for a set
micro services with docker. DCM also provides a couple of neat shorthand commands.

**Prerequisites**
* Linux Distros
  * [Docker Engine](https://docs.docker.com/engine/installation/linux/)
  * [Docker-Compose](https://docs.docker.com/compose/install/)
* Mac OS X / Windows
  * [Docker Toolbox](https://www.docker.com/products/docker-toolbox)

OSX folks can also manually install docker:

* [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
* [Homebrew](http://brew.sh/)
* Docker Client `brew install docker`
* Docker Machine `brew install docker-machine`
* Docker Compose `brew install docker-compose`

**Supported Operating Systems**

* Mac OS X, 64bit (tested)
* Linux, 64bit
  * Ubuntu (tested)
  * Debian
  * Mint (tested)
  * CentOS
  * Red Hat
  * Fedora
  * Gentoo
* FreeBSD, 64bit
* Windows (Cygwin), 64bit

## Getting started

To install DCM, first checkout DCM on your local file system, and create an enhanced version
docker-compose config.

```shell
git clone git@github.com:beanworks/dcm.git
# Note here the name of the config needs to be same as DCM project name
touch dcm/pi314.yml
```

Then add the following lines to your bashrc/zshrc

```shell
export DCM_DIR="/path/to/your/dcm/dir"
export DCM_PROJECT="pi314"

[ -s "$DCM_DIR/dcm.sh" ] && . "$DCM_DIR/dcm.sh"
```

Source your bashrc/zshrc or profile again then you are all set.

## Enhanced docker-compose config

DCM is based on docker-compose, so it supports all the configuration options from compose
(https://docs.docker.com/compose/compose-file/). In addition to those options, DCM extends
docker-compose with a couple of additional options.

All DCM specific options in the YAML configuration file are under `serviceName.labels`.

#### `dcm.repository` (required)

`dcm setup` command will read this option and clone the service's git repository. DCM will
place the repo at `$DCM_DIR/srv/$DCM_PROJECT/[service name]`.

```yaml
service:
  labels:
    dcm.repository: git@github.com:username/repository.git
```

#### `dcm.pre_initscript` (optional)

If this option is given, `dcm run` command will run the pre-init script automatically right before `docker-compose up` process is started.

The value of the `dcm.pre_initscript` is relative to the service's folder.

```yaml
service:
  build: "./srv/project/service/"
  labels:
    dcm.pre_initscript: "dcm/pre-init.bash"
```

In the example above, DCM will the pre-init script `$DCM_DIR/srv/project/service/dcm/pre-init.bash`.

#### `dcm.initscript` (optional)

If this option is given, `dcm run` command will run the init script automatically right after
`docker-compose up` process is finished.

The value of the `dcm.initscript` is relative to the service's folder.

```yaml
service:
  build: "./srv/project/service/"
  labels:
    dcm.initscript: "dcm/init.bash"
```

In the example above, DCM will the init script `$DCM_DIR/srv/project/service/dcm/init.bash`.

#### `dcm.initscript_shell` (optional)

If this option is given, `dcm run` command will run the init script with the value of this shell as executable.
If no value is provided it defaults to `/bin/bash`.

#### `dcm.branch` (optional)

IF this option is given, DCM will switch to the git branch provided right after it clones
the repo during the setup process.

```yaml
service:
  labels:
    dcm.branch: default-branch-name
```

## One click setup, build && run

For your first time setup, run the following commands. They will checkout all the repositories
for different micro services, build all the images and spin up the docker containers.

```shell
dcm setup && dcm run
```

Generally in your day to day development process, you should only need to run either `dcm run`
(shorthand version `dcm r`) or `dcm build && dcm run` (shorthand version `dcm b && dcm r`).

## Update DCM

First, uninstall DCM from bash/zsh

```shell
dcm unload
```

Then, pull the latest version DCM with git

```shell
git pull
```

Lastly, source bashrc/zshrc/profile again to reinstall DCM

```shell
source ~/.profile
# or
source ~/.bash_profile
# or
source ~/.bashrc
# or
zsh
```

## Setting up multi instance

#### 1. Create YAML configuration files for multiple instances

```shell
touch instance1.yml instance2.yml instance3.yml
```

Note that if you are running multi instance setup for the same set of services, you will need to
assign different public facing ports to in `ports` options for those containers that need to be
directly accessed from the host machine.

For example, if you have nginx as load balancer in instance1, instance2 and instance3, you will
probably need `- "8081:80"` for instance1.yml, `- "8082:80"` for instance2.yml and `- "8083:80"`
for instance3.yml.

#### 2. Initial setup, build && run

```shell
export DCM_PROJECT=instance1
dcm setup && dcm run

export DCM_PROJECT=instance2
dcm setup && dcm run

export DCM_PROJECT=instance3
dcm setup && dcm run
```

Note that you can always set the env variable within the same command like this:

```shell
DCM_PROJECT=instance1 dcm setup
DCM_PROJECT=instance1 dcm run
```

The choices are yours :)

#### 3. Subsequent rebuild && rerun

```shell
export DCM_PROJECT=instance1
dcm build && dcm run

export DCM_PROJECT=instance2
dcm build && dcm run

export DCM_PROJECT=instance3
dcm build && dcm run
```

## All available DCM commands

The follow menu can be viewed in command line by entering `dcm` or `dcm help` commands.

```text
DCM (Docker-Compose Manager)

Usage:
  dcm help                Show this help menu.
  dcm setup               Git checkout repositories for the services that require
                          local docker build. It skips the service when the image
                          is from docker hub, or the repo's folder already exists.
  dcm run [<args>]        Run docker-compose commands. If <args> is not given, by
                          default DCM will run `docker-compose up` command.
                          <args>: up, build, start, stop, restart, pre-init, init, execute
  dcm build               Docker (re)build service images that require local build.
                          It's the shorthand version of `dcm run build` command.
  dcm shell <service>     Log into a given service container.
  dcm purge [<type>]      Remove either all the containers or all the images. If <type>
                          is not given, by default DCM will purge everything.
                          <type>: images, containers, all
  dcm branch [<service>]  Display the current git branch for the given service that
                          was built locally.
  dcm goto [<service>]    Go to the service's folder. If <service> is not given, by
                          default DCM will go to $DCM_DIR.
  dcm update [<service>]  Update DCM and(or) the given service.
  dcm list                List all the available services.

Example:
  Initial setup
    dcm setup
    dcm run

  Rebuild
    dcm build
    dcm run

  Or only Rerun
    dcm run

  Log into a service's container
    dcm shell service_name
```

## TODOs

* Support CLI autocomplete for zsh
* Add working examples
  * Containerize an app that involves a couple of micro services
  * Create a YAML config for DCM to setup, build and run the app
* Test on different OS
  * Linux Distros
  * FreeBSD
  * Windows Cygwin

## Contributing

All code needs to be formatted with `gofmt`. `goimports` is more preferred as it also auto-generate
and format import section.

We suggest contributors use vim-go or GoSublime if you are vim lovers or sublime folks. If you use
neither of those editors, having a editor or IDE respects EditorConfig and automatically invoke
`gofmt` or `goimports` on save is highly recommended.

#### Make a development copy

Make sure you have Go 1.6+ installed and GOPATH set (https://golang.org/doc/code.html).

```shell
git clone git@github.com:beanworks/dcm.git $GOPATH/src/github.com/beanworks/dcm
cd $GOPATH/src/github.com/beanworks/dcm
```

Run command `tree -a -I .git` you will see the following folder and file structure:

```text
.
├── .editorconfig
├── .gitignore
├── .travis.yml
├── LICENSE
├── Makefile
├── README.md
├── bin
│   ├── dcm-darwin-amd64
│   ├── dcm-freebsd-amd64
│   ├── dcm-linux-amd64
│   └── dcm-windows-amd64.exe
├── dcm.sh
├── src
│   ├── cmd.go
│   ├── cmd_test.go
│   ├── config.go
│   ├── config_test.go
│   ├── dcm.go
│   ├── dcm_test.go
│   ├── main.go
│   ├── util.go
│   └── util_test.go
└── srv
    └── .gitignore
```

All the source files are located in `src` folder.

#### Running unit tests

```shell
# Run the whole test suite
make test
# Run tests in verbose mode
make vtest
```

#### Generating test coverage report

```shell
make cover
```

#### Build executables

```shell
# Build development executable
make
# Run development executable
bin/dcm
# Cross compile executables for different OS
make cross
# Cleanup
make clean
# Cleanup and remove all the cross compile executables
make cleanall
```

## License

Copyright (c) 2016, Beanworks Solutions Inc. <engpartnership@beanworks.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

* Neither the name of dcm nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
