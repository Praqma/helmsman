package app

import (
	"bytes"
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

type exitStatus struct {
	code   int
	errors string
	output string
}

func (c *command) String() string {
	return c.Cmd + " " + strings.Join(c.Args, " ")
}

// exec executes the executable command and returns the exit code and execution result
func (c *command) exec() exitStatus {
	// Only use non-empty string args
	args := []string{}
	for _, str := range c.Args {
		if str != "" {
			args = append(args, str)
		}
	}

	log.Verbose(c.Description)
	log.Debug(c.String())

	cmd := exec.Command(c.Cmd, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		log.Info("cmd.Start: " + err.Error())
		return exitStatus{
			code:   1,
			errors: err.Error(),
		}
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return exitStatus{
					code:   status.ExitStatus(),
					output: stdout.String(),
					errors: stderr.String(),
				}
			}
		} else {
			log.Fatal("cmd.Wait: " + err.Error())
		}
	}
	return exitStatus{
		code:   0,
		output: stdout.String(),
		errors: stderr.String(),
	}
}

// toolExists returns true if the tool is present in the environment and false otherwise.
// It takes as input the tool's command to check if it is recognizable or not. e.g. helm or kubectl
func toolExists(tool string) bool {
	cmd := command{
		Cmd:         tool,
		Args:        []string{},
		Description: "Validating that [ " + tool + " ] is installed",
	}

	result := cmd.exec()

	return result.code == 0
}
