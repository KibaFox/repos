//+build mage

// This is the build script for Mage. The install target is all you really need.
// The release target is for generating official releases and is really only
// useful to project admins.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Clean removes the dist/ and out/ directory.
func Clean() (err error) {
	if err = os.RemoveAll("dist"); err != nil {
		return err
	}

	if err = os.RemoveAll("out"); err != nil {
		return err
	}

	return nil
}

// Build repos into the dist/ directory.
func Build() error {
	name := "repos"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	os.MkdirAll("dist", 0777)

	return run("go", "build", "-o", fmt.Sprintf("dist/%s", name),
		"./cmd/repos")
}

// Lint will perform style checks and static analysis on the Go code.
func Lint() error {
	return run("golangci-lint", "run")
}

// Test will run all tests.
func Test() error {
	return run("ginkgo", "-v", "test", "./...")
}

func run(cmd string, args ...string) (err error) {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	fmt.Println("exec:", cmd, strings.Join(args, " "))
	return c.Run()
}
