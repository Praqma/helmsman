package app

import (
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"strings"
	"time"
)

// Command type representing all executable commands Helmsman needs
// to execute in order to inspect the environment/ releases/ charts etc.
type Command struct {
	Cmd         string
	Args        []string
	Description string
}

func (c *Command) String() string {
	return c.Cmd + " " + strings.Join(c.Args, " ")
}

// RetryExec runs exec command with retry
func (c *Command) RetryExec(attempts int) (res *ExitStatus, err error) {
	for i := 0; i < attempts; i++ {
		res, err = c.Exec()
		if err == nil {
			return
		}
		if i < (attempts - 1) {
			time.Sleep(time.Duration(math.Pow(2, float64(2+i))) * time.Second)
			log.Infof("Retrying %s due to error: %v", c.Description, err)
		}
	}

	return nil, fmt.Errorf("cmd %s failed after %d attempts: %w", c.Description, attempts, err)
}

// Exec executes the executable command and returns the exit code and execution result
func (c *Command) Exec() (*ExitStatus, error) {
	// Only use non-empty string args
	var args []string
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
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return nil, newExitError(c, exiterr.ExitCode(), stdout, stderr)
		} else {
			log.Fatal("cmd.Wait: " + err.Error())
		}
	}
	return newExitStatus(c, stdout, stderr), nil
}

type ExitStatus struct {
	cmd    string
	output string
	errors string
}

func newExitStatus(cmd *Command, stdout, stderr bytes.Buffer) *ExitStatus {
	return &ExitStatus{
		cmd:    cmd.String(),
		output: strings.TrimSpace(stdout.String()),
		errors: strings.TrimSpace(stderr.String()),
	}
}

func (es *ExitStatus) String() string {
	return fmt.Sprintf("cmd %s exited successfully\noutput: %s", es.cmd, es.output)
}

type exitError struct {
	cmd    string
	code   int
	output string
}

func newExitError(cmd *Command, code int, stdout, stderr bytes.Buffer) *exitError {
	return &exitError{
		cmd:  cmd.String(),
		code: code,
		output: fmt.Sprintf(
			"--- stdout ---\n%s\n--- stderr ---\n%s",
			strings.TrimSpace(stdout.String()),
			strings.TrimSpace(stderr.String())),
	}
}

func (ee *exitError) Error() string {
	return fmt.Sprintf("cmd %s failed with non-zero exit code %d\noutput: %s", ee.cmd, ee.code, ee.output)
}

// ToolExists returns true if the tool is present in the environment and false otherwise.
// It takes as input the tool's command to check if it is recognizable or not. e.g. helm or kubectl
func ToolExists(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}
