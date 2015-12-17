# DCM (Docker-Compose Manager)

[![Build Status](https://travis-ci.org/beanworks/dcm.svg)](https://travis-ci.org/beanworks/dcm)

DCM is a wrapper for docker-compose. It enables one click setup, build && run process for a set
micro services with docker. DCM also provides a couple of neat shorthand commands.

## Getting started

To install DCM, first checkout DCM on your local file system, and create an enhanced version
docker-compose config.

```
git clone git@github.com:beanworks/dcm.git
# Note here the name of the config needs to be same as DCM project name
touch pi314.yml
```

Then add the following lines to your bashrc/zshrc

```
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

```
service:
  labels:
    dcm.repository: git@github.com:username/repository.git
```

#### `dcm.initscript` (optional)

If this option is given, `dcm run` command will run the init script automatically right after
`docker-compose up` process is finished.

The value of the `dcm.initscript` is relative to the service's folder.

```
service:
  build: "./srv/project/service/"
  labels:
    dcm.initscript: "dcm/init.bash"
```

In the example above, DCM will the init script `$DCM_DIR/srv/project/service/dcm/init.bash`.

#### `dcm.branch` (optional)

IF this option is given, DCM will switch to the git branch provided right after it clones
the repo during the setup process.

```
service:
  labels:
    dcm.branch: default-branch-name
```

## One click setup, build && run

For your first time setup, run the following commands. They will checkout all the repositories
for different micro services, build all the images and spin up the docker containers.

```
dcm setup && dcm run
```

Generally in your day to day development process, you should only need to run either `dcm run`
(shorthand version `dcm r`) or `dcm build && dcm run` (shorthand version `dcm b && dcm r`).

## Setting up multi instance

#### 1. Create YAML configuration files for multiple instances

```
touch instance1.yml instance2.yml instance3.yml
```

Note that if you are running multi instance setup for the same set of services, you will need to
assign different public facing ports to in `ports` options for those containers that need to be
directly accessed from the host machine.

For example, if you have nginx as load balancer in instance1, instance2 and instance3, you will
probably need `- "8081:80"` for instance1.yml, `- "8082:80"` for instance2.yml and `- "8083:80"`
for instance3.yml.

#### 2. Initial setup, build && run

```
export DCM_PROJECT=instance1
dcm setup && dcm run

export DCM_PROJECT=instance2
dcm setup && dcm run

export DCM_PROJECT=instance3
dcm setup && dcm run
```

Note that you can always set the env variable within the same command like this:

```
DCM_PROJECT=instance1 dcm setup
DCM_PROJECT=instance1 dcm run
```

The choices are yours :)

#### 3. Subsequent rebuild && rerun

```
export DCM_PROJECT=instance1
dcm build && dcm run

export DCM_PROJECT=instance2
dcm build && dcm run

export DCM_PROJECT=instance3
dcm build && dcm run
```

## All available DCM commands

The follow menu can be viewed in command line by entering `dcm` or `dcm help` commands.

```
Docker Compose Manager

Usage:
  dcm help                Show this message
                          shorthand ver.: `dcm h`
  dcm setup               Check out all the repositories for API, UI and services
  dcm run [<args>]        Run docker-compose commands. If <args> is not given, by
                          default DCM will run `up` command.
                          <args>: up, build, start, stop, restart
                          shorthand ver.: `dcm r [<args>]`
  dcm run build           Run build command that (re)create all the service images
                          shorthand ver.: `dcm build` or `dcm b`
  dcm shell <service>     Log into a given service container
                          <service>: api, ui, postgres, mongo, redis, nginx, php
                          shorthand ver.: `dcm sh <service>`
  dcm purge [<type>]      Remove either all the containers or all the images or
                          everything. If <type> is not given, by default DCM will
                          purge everything
                          <type>: images|img, containers|con, all
                          shorthand ver.: `dcm rm [<type>]`
  dcm branch <service>    Display the current branch for the given service name
                          <service>: api, ui, postgres, mongo, redis, nginx, php
                          shorthand ver.: `dcm br <service>`
  dcm goto [<service>]    Go to the service's folder. If <service> is not given,
                          by default DCM will go to $DCM_DIR
                          <service>: api, ui, postgres, mongo, redis, nginx, php
                          shorthand ver.: `dcm gt [<service>]`
  dcm update [<service>]  Update DCM and dependent services (PostgrSQL, MongoDB,
                          Redis, Nginx and Base PHP). If <service> is not given,
                          by default DCM will update everything except api and ui.
                          <service>: postgres, mongo, redis, nginx, php
                          shorthand ver.: `dcm u`

Example:
  Initial setup
    dcm setup
    dcm run

  Rebuild API or UI after switching branch
    dcm build
    dcm run

  Log into different service containers
    dcm shell api
    dcm shell ui
    ...
```

## License

Copyright (c) 2015, Beanworks Solutions Inc. <engpartnership@beanworks.com>
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
