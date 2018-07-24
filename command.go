package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"syscall"
)

// command type representing all executable commands Helmsman needs
// to execute in order to inspect the environment/ releases/ charts etc.
type command struct {
	Cmd         string
	Args        []string
	Description string
}

//printDescription prints the description of a command
func (c command) printDescription() {

	fmt.Println(c.Description)
}

// printFullCommand prints the executable command.
func (c command) printFullCommand() {

	fmt.Println(c.Description, " by running : \"", c.Cmd, " ", c.Args, " \"")
}

// exec executes the executable command and returns the exit code and execution result
func (c command) exec(debug bool, verbose bool) (int, string) {

	if debug {
		log.Println("INFO: " + c.Description)
	}
	if verbose {
		log.Println("VERBOSE: " + c.Args[1])
	}

	cmd := exec.Command(c.Cmd, c.Args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("ERROR: cmd.Start: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), stderr.String()
			}
		} else {
			log.Fatalf("ERROR: cmd.Wait: %v", err)
		}
	}

	return 0, stdout.String()
}
