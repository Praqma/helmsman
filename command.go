package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"syscall"
)

type command struct {
	Cmd         string
	Args        []string
	Description string
}

func (c command) printDescription() {

	fmt.Println(c.Description)
}

func (c command) printFullCommand() {

	fmt.Println(c.Description, " by running : \"", c.Cmd, " ", c.Args, " \"")
}

func (c command) exec(debug bool) (int, string) {

	if debug {
		log.Println("INFO: executing command: " + c.Description)
	}
	cmd := exec.Command(c.Cmd, c.Args...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Start(); err != nil {
		log.Fatalf("ERROR: cmd.Start: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				//log.Printf("Exit Status: %d", status.ExitStatus())
				return status.ExitStatus(), out.String()
			}
		} else {
			log.Fatalf("ERROR: cmd.Wait: %v", err)
		}
	}

	return 0, out.String()
}
