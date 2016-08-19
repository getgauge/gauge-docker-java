// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge-Docker-Java.

// Gauge-Docker-Java is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.

// Gauge-Docker-Java is distributed in the hope that it will be
// useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge-Docker-Java.  If not, see
// <http://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/getgauge/common"
)

const (
	dockerCmd               = "docker"
	stepImplementationClass = "StepImplementation.java"
	skelDir                 = "skel"
	envDir                  = "env"
	windows                 = "windows"
	java                    = "java"
	defaultSrcDir           = "src"
)

var pluginDir = ""
var projectRoot = ""
var start = flag.Bool("start", false, "Start the docker-java runner")
var initialize = flag.Bool("init", false, "Initialize the docker-java runner")

func main() {
	flag.Parse()
	setPluginAndProjectRoots()
	if *start {
		p := pluginVersion("java")
		fmt.Printf("Java Plugin Version: %s\n", p)
		startDockerJava()
	} else if *initialize {
		buildImage()
		initializeProject()
	} else {
		printUsage()
	}
}

func initializeProject() {
	buildImage()
	os.Chdir(projectRoot)
	funcs := []initializerFunc{createSrcDirectory, createEnvDirectory, createLibsDirectory, createStepImplementationClass, createJavaPropertiesFile}
	for _, f := range funcs {
		f()
	}
}

func buildImage() {
	runCommand("docker", "build", "-t", "getgauge/java", ".")
}

func pluginVersion(name string) string {
	out, err := exec.Command("gauge", "-v", "--machine-readable").Output()
	if err != nil {
		log.Fatal(err)
	}
	type Plugin struct {
		name    string
		version string
	}
	type GaugeOutput struct {
		version string
		plugins []Plugin
	}
	var outjson GaugeOutput
	err = json.Unmarshal([]byte(string(out)), &outjson)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Printf("%+v\n", outjson)
	return "0.5.0"
}

func startDockerJava() {
	os.Chdir(projectRoot)

	internalPort := os.Getenv("GAUGE_INTERNAL_PORT")
	apiPort := os.Getenv("GAUGE_API_PORT")

	args := []string{
		"run",
		"--rm",
		"-v", fmt.Sprintf("%s:%s", projectRoot, "/opt/test"),
		"-e", fmt.Sprintf("GAUGE_PROJECT_ROOT=%s", "/opt/test"),
		"-e", fmt.Sprintf("GAUGE_INTERNAL_PORT=%s", internalPort),
		"-e", fmt.Sprintf("GAUGE_API_PORT=%s", apiPort),
		"--net=host",
		"getgauge/java",
		"/bin/sh",
		"-c",
		"set -e; cd /opt/test; cp -rv ~/.gauge/plugins/java/0.5.0/libs/* ./libs/; ~/.gauge/plugins/java/0.5.0/bin/gauge-java --start"}

	fmt.Printf("Running command:\n\t%s %s\n", dockerCmd, args)
	cmd := runCommandAsync(dockerCmd, args)
	listenForKillSignal(cmd)
	go killIfGaugeIsDead(cmd) // Kills gauge-docker-java.go process if gauge process i.e. parent process is already dead.

	err := cmd.Wait()
	if err != nil {
		fmt.Printf("process %s with pid %d quit unexpectedly. %s\n", cmd.Path, cmd.Process.Pid, err.Error())
		os.Exit(1)
	}
}

func listenForKillSignal(cmd *exec.Cmd) {
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, syscall.SIGTERM)
	go func() {
		<-sigc
		cmd.Process.Kill()
	}()
}

func killIfGaugeIsDead(cmd *exec.Cmd) {
	parentProcessID := os.Getppid()
	for {
		if !isProcessRunning(parentProcessID) {
			// fmt.Printf("Parent Gauge process with pid %d has terminated.", parentProcessID)
			err := cmd.Process.Kill()
			if err != nil {
				fmt.Printf("Failed to kill process with pid %d. %s\n", cmd.Process.Pid, err.Error())
			}
			os.Exit(0)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	if runtime.GOOS != windows {
		return process.Signal(syscall.Signal(0)) == nil
	}

	processState, err := process.Wait()
	if err != nil {
		return false
	}
	if processState.Exited() {
		return false
	}

	return true
}

func encoding() string {
	return "-Dfile.encoding=UTF-8"
}

func execName(name string) string {
	if runtime.GOOS == windows {
		return fmt.Sprintf("%s.exe", name)
	}
	return name
}

func setPluginAndProjectRoots() {
	var err error
	pluginDir, err = os.Getwd()
	if err != nil {
		fmt.Printf("Failed to find current working directory: %s \n", err)
		os.Exit(1)
	}
	projectRoot = os.Getenv(common.GaugeProjectRootEnv)
	if projectRoot == "" {
		fmt.Printf("Could not find %s env. Docker-Java Runner exiting...", common.GaugeProjectRootEnv)
		os.Exit(1)
	}
}

func appendClasspath(source *string, classpath string) {
	if len(classpath) == 0 {
		return
	}

	if len(*source) == 0 {
		*source = classpath
	} else {
		*source = fmt.Sprintf("%s%c%s", *source, os.PathListSeparator, classpath)
	}
}

// User set classpath & additional libs will be comma separated
// it could be relative path, but JVM needs full path to be specified
// so this function splits the path, convert them to absolute path forms a classpath
func getClassPathForVariable(envVariableName string) string {
	value := os.Getenv(envVariableName)
	cp := ""
	if len(value) > 0 {
		paths := splitByComma(value)
		for _, p := range paths {
			abs, err := filepath.Abs(p)
			if err == nil {
				appendClasspath(&cp, abs)
			} else {
				appendClasspath(&cp, p)
			}
		}
	}
	return cp
}

type initializerFunc func()

func showMessage(action, filename string) {
	fmt.Printf(" %s  %s\n", action, filename)
}

func createSrcDirectory() {
	createDirectory(filepath.Join(defaultSrcDir, "test", java))
}

func createEnvDirectory() {
	createDirectory(filepath.Join("env", "default"))
}

func createLibsDirectory() {
	createDirectory("libs")
}

func createDirectory(filePath string) {
	showMessage("create", filePath)
	if !common.DirExists(filePath) {
		err := os.MkdirAll(filePath, 0755)
		if err != nil {
			fmt.Printf("Failed to make directory. %s\n", err.Error())
		}
	} else {
		showMessage("skip", filePath)
	}
}

func createStepImplementationClass() {
	javaSrc := filepath.Join(defaultSrcDir, "test", java)
	destFile := filepath.Join(javaSrc, stepImplementationClass)
	showMessage("create", destFile)
	if common.FileExists(destFile) {
		showMessage("skip", destFile)
	} else {
		srcFile := filepath.Join(pluginDir, skelDir, stepImplementationClass)
		if !common.FileExists(srcFile) {
			showMessage("error", fmt.Sprintf("%s Does not exist.\n", stepImplementationClass))
			return
		}
		err := common.CopyFile(srcFile, destFile)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to copy %s. %s \n", srcFile, err.Error()))
		}
	}
}

func createJavaPropertiesFile() {
	destFile := filepath.Join(envDir, "default", "docker-java.properties")
	showMessage("create", destFile)
	if common.FileExists(destFile) {
		showMessage("skip", destFile)
	} else {
		srcFile := filepath.Join(pluginDir, skelDir, envDir, "docker-java.properties")
		if !common.FileExists(srcFile) {
			showMessage("error", fmt.Sprintf("docker-java.properties does not exist at %s. \n", srcFile))
			return
		}
		err := common.CopyFile(srcFile, destFile)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to copy %s. %s \n", srcFile, err.Error()))
		}
	}
}

func printUsage() {
	flag.PrintDefaults()
	os.Exit(2)
}

func runCommand(command string, arg ...string) {
	cmd := exec.Command(command, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Execute %v\n", cmd.Args)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func runCommandAsync(cmdName string, args []string) *exec.Cmd {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var err error
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to start %s. %s\n", cmd.Path, err.Error())
		os.Exit(1)
	}
	return cmd
}

func uniqueFileName() string {
	return fmt.Sprintf("%d", common.GetUniqueID())
}

func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func splitByComma(text string) []string {
	var splits []string
	values := strings.Split(text, ",")
	for _, val := range values {
		splits = append(splits, strings.TrimSpace(val))
	}
	return splits
}
