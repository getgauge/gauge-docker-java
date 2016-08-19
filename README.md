Gauge-Docker-Java
========================

**Status: Unstable | Not ready for production use**

This is an **experimental** plugin for running [Java language plugin](http://getgauge.io/documentation/user/current/test_code/java/java.html) for [gauge](http://getgauge.io) in a Docker container.

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

## Caveats

TODO

License
-------

![GNU Public License version 3.0](http://www.gnu.org/graphics/gplv3-127x51.png)
Gauge-Ruby is released under [GNU Public License version 3.0](http://www.gnu.org/licenses/gpl-3.0.txt)

Copyright
---------

Copyright 2016 ThoughtWorks, Inc.
