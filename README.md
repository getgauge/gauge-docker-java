Gauge-Docker-Java
========================

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v1.4%20adopted-ff69b4.svg)](CODE_OF_CONDUCT.md)

**Status: Unstable | Not ready for production use**

This is an **experimental** plugin for running [Java language plugin](https://github.com/getgauge/gauge-java) for [gauge](http://gauge.org) in a Docker container.

## How it works

This plugin appears to Gauge as a language runner. Gauge has no knowledge about Docker usage. The plugin is responsible for bringing up Docker and executing the test code in the container.

When a new project is initialized, this plugin builds a Docker image from the Dockerfile shipped with the plugin. The image has Gauge Java installed in it.

When gauge tests are run, the plugin brings up a container from the image that was built previously, mounts the current project root in the container, sets Gauge's API ports as environment variables and invokes the Gauge Java runner inside the container. The Gauge Java runner in the container connects to Gauge running on the host through TCP.

## Caveats

- Works only on Linux at the moment. [OSX needs further abstractions](https://forums.docker.com/t/should-docker-run-net-host-work/14215/4) using `docker-manager`.
- Docker is invoked using `--net=host` option. This does not isolate the network interface of the container from the host. This is necessary because both Gauge and Gauge-Java try to listen and connect on `localhost` for now.
- Currently, Gauge commands need to be run as the same user that runs the Docker daemon. By default, it needs `root`. So, you may need to do `sudo gauge --init docker-java` and `sudo gauge specs/` unless you run [Docker manager as a different user](http://askubuntu.com/q/477551/25541).
- `gauge-java` in the container is not able to figure out the classpaths. So, `jar`s have to explicitly copied over from `.gauge/plugins/java` to the `./libs` directory in the current project.

---

## Build from source

### Requirements

* [Golang](http://golang.org/)
* [Docker](https://docker.com/)
* [Gauge](http://getgauge.io)

### Compiling

```
go run make.go
```

### Installing

After compilation:

```
go run make.go --install
```

### Creating distributable

Run after compiling:

```
go run make.go --distro
```

## Create a project

### Initialize

```
gauge --init docker-java
```

This will build a docker image with Gauge Java and create a sample implementation in the current directory.

### Add tests

Write your tests the way you would write for Gauge Java.

### Run

```
gauge specs/
```

This will start Gauge, Gauge will trigger the `docker-java` runner,
which will in turn bring up a container with the image compiled
previously.

License
-------

![GNU Public License version 3.0](http://www.gnu.org/graphics/gplv3-127x51.png)
Gauge-Ruby is released under [GNU Public License version 3.0](http://www.gnu.org/licenses/gpl-3.0.txt)

Copyright
---------

Copyright 2016 ThoughtWorks, Inc.
