# busan

A command line utility for creating versioned Docker images and containers with an automated upgrade process.

Busan allows you to autonomously ensure your Docker environment is kept up to date and running based off your source Dockerfiles. In combination with [weave](https://github.com/callumj/weave) you can implement a continous deployment pipeline that keeps your Docker hosts in sync. 

Busan is named after the [largest port in South Korea](http://en.wikipedia.org/wiki/Busan_Port), in the city of Busan.

## How it works

Busan searches your `Dockerfile` for the `VERSION x.xx` comment and will ensure that your target Docker installation has both correct versioned image and correct running versioned container.

Busan maintains the `name:version` naming scheme.

In my `~/DockerStuff/callumjcom-deploy` folder I have the following `Dockerfile`

```
# VERSION 0.11
FROM ubuntu:14.04

MAINTAINER Callum Jones <callum@callumj.com>

EXPOSE 80
ENV DOMAIN callumj.com
.....
```

and my current `docker ps` looks like this

```
root@localhost:~# docker ps
CONTAINER ID        IMAGE                                 COMMAND                CREATED             STATUS              PORTS                NAMES
11255c55312c        callumjcom-deploy:v0.10               /bin/sh -c /bin/bash   49 minutes ago      Up 49 minutes       80/tcp               callumjcom-deploy               
```

By running `busan ~/DockerStuff/callumjcom-deploy` the system will ensure my image and container are brought up to the expected version.


```
root@localhost:~# docker ps
CONTAINER ID        IMAGE                                 COMMAND                CREATED             STATUS              PORTS                NAMES
ec8ca4c8e802        callumjcom-deploy:v0.11               /bin/sh -c /bin/bash   49 minutes ago      Up 49 minutes       80/tcp               callumjcom-deploy               
```

Busan will clean up old images where possible and remove all previous containers (if needed).

## Installation

You can install busan for now via go get

```
go get -u github.com/callumj/busan
```

## Usage

```
Usage:
  busan [OPTIONS] DIRECTORY

Application Options:
  -d, --docker-host= unix:// or tcp:// address to Docker host
  -n, --name=        The name of the configurations

Help Options:
  -h, --help         Show this help message
```

* DIRECTORY: The full path to the directory containing the necessary
* -d: An alternative way to contact the docker server, default is the unix socket.
* -name: An alternative name for the image/container, default is extracted from the final directory name (e.g. callumjcom-deploy)

## attributes.yml

Busan is also aware of `attributes.yml` located in the same directory as the `Dockerfile` and supports the following options

* volumes: A hash of volumes to be mapped. In the form of /container_directory:/host_directory:ro.
* exposed_ports: A array of ports to expose of the container's IP address.

Here is an example used to provide the /app_logs directory mapped to a read/write /opt/callumj_logging on the host. Port 80 is published on the container's IP address.

```
volumes:
  /app_logs: "/opt/callumj_logging:rw"
exposed_ports:
  - 80/tcp
```