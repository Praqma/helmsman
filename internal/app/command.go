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

type ExitStatus struct {
	code   int
	errors string
	output string
}

func (e ExitStatus) String() string {
	str := strings.TrimSpace(e.output)
	if errs := strings.TrimSpace(e.errors); errs != "" {
		str = fmt.Sprintf("%s\n--- stderr ---\n%s", str, errs)
	}
	return str
}

func (c *Command) String() string {
	return c.Cmd + " " + strings.Join(c.Args, " ")
}

// RetryExec runs exec command with retry
func (c *Command) RetryExec(attempts int) (ExitStatus, error) {
	var (
		result ExitStatus
		err    error
	)

	for i := 0; i < attempts; i++ {
		result, err = c.Exec()
		if err == nil {
			return result, nil
		}
		if i < (attempts - 1) {
			time.Sleep(time.Duration(math.Pow(2, float64(2+i))) * time.Second)
			log.Infof("Retrying %s due to error: %v", c.Description, err)
		}
	}

	return result, fmt.Errorf("%s, failed after %d attempts with: %w", c.Description, attempts, err)
}

func (c *Command) newExitError(code int, stdout, stderr bytes.Buffer, cause error) error {
	return fmt.Errorf(
		"%s failed with non-zero exit code %d: %w\noutput: %s",
		c.Description, code, cause,
		fmt.Sprintf(
			"\n--- stdout ---\n%s\n--- stderr ---\n%s",
			strings.TrimSpace(stdout.String()),
			strings.TrimSpace(stderr.String()),
		),
	)
}

// Exec executes the executable command and returns the exit code and execution result
func (c *Command) Exec() (ExitStatus, error) {
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
		log.Info("cmd.Start: " + err.Error())
		return ExitStatus{
			code:   1,
			errors: err.Error(),
		}, err
	}

	err := cmd.Wait()
	res := ExitStatus{
		output: strings.TrimSpace(stdout.String()),
		errors: strings.TrimSpace(stderr.String()),
	}
	if err != nil {
		res.code = 1
		err = fmt.Errorf("failed to run %s: %w", c.Description, err)
		if exiterr, ok := err.(*exec.ExitError); ok {
			res.code = exiterr.ExitCode()
			err = c.newExitError(exiterr.ExitCode(), stdout, stderr, err)
		}
	}
	return res, err
}

// ToolExists returns true if the tool is present in the environment and false otherwise.
// It takes as input the tool's command to check if it is recognizable or not. e.g. helm or kubectl
func ToolExists(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}
