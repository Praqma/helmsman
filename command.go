package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
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
func (c command) exec(debug bool, verbose bool) (int, string, string) {

	// Only use non-empty string args
	args := []string{}
	for _, str := range c.Args {
		if str != "" {
			args = append(args, str)
		}
	}

	if debug {
		log.Println("INFO: " + c.Description)
	}
	if verbose {
		log.Println("VERBOSE: " + c.Cmd + " " + strings.Join(args, " "))
	}

	cmd := exec.Command(c.Cmd, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// we need to tell the TILLER to be silent. This will only matter in
	// tillerless mode.
	cmd.Env = append(os.Environ(), "HELM_TILLER_SILENT=true")

	if err := cmd.Start(); err != nil {
		log.Println("ERROR: cmd.Start: " + err.Error())
		return 1, err.Error(), ""
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), stderr.String(), ""
			}
		} else {
			logError("ERROR: cmd.Wait: " + err.Error())
		}
	}
	return 0, stdout.String(), stderr.String()
}
