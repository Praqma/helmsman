package app

import (
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"strings"
	"syscall"
	"time"
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

// runs exec command with retry
func (c *command) retryExec(attempts int) exitStatus {
	var result exitStatus

	for i := 0; i < attempts; i++ {
		result = c.exec()
		if result.code == 0 {
			return result
		}
		if i < (attempts - 1) {
			time.Sleep(time.Duration(math.Pow(2, float64(2+i))) * time.Second)
			log.Info(fmt.Sprintf("Retrying %s due to error: %s", c.Description, result.errors))
		}
	}

	return exitStatus{
		code:   result.code,
		output: result.output,
		errors: fmt.Sprintf("After %d attempts of %s, it failed with: %s", attempts, c.Description, result.errors),
	}
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
