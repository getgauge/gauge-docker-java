# Copyright 2015 ThoughtWorks, Inc.

# This file is part of Gauge-Docker-Java.

# Gauge-Docker-Java is free software: you can redistribute it and/or
# modify it under the terms of the GNU General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.

# Gauge-Docker-Java is distributed in the hope that it will be
# useful, but WITHOUT ANY WARRANTY; without even the implied warranty
# of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with Gauge-Docker-Java.  If not, see
# <http://www.gnu.org/licenses/>.

FROM ubuntu:16.04
MAINTAINER ThoughtWorks, Inc. <studios@thoughtworks.com>

RUN apt-key adv --keyserver hkp://pool.sks-keyservers.net --recv-keys 023EDB0B
RUN echo deb http://dl.bintray.com/gauge/gauge-deb stable main | tee -a /etc/apt/sources.list

RUN apt-get update
RUN apt-get install -y gauge openjdk-8-jdk

RUN gauge_setup
RUN gauge --install java

RUN gauge -v
