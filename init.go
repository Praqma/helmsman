package main

import (
	"log"
	"os"
)

// init is executed before the main is executed.
// It checks if Helm exists and configures the connection to the k8s cluster.
func init() {
	if !toolExists("helm") {
		log.Fatal("ERROR: helm is not installed/configured correctly. Aborting!")
		os.Exit(1)
	}

	if !toolExists("kubectl") {
		log.Fatal("ERROR: kubectl is not installed/configured correctly. Aborting!")
		os.Exit(1)
	}

	// TODO : setup cluster connection

}

// toolExists returns true if the tool is present in the environment and false otherwise.
// It takes as input the tool's command to check if it is recognizable or not. e.g. helm or kubectl
func toolExists(tool string) bool {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", tool},
		Description: "validating that " + tool + " is installed.",
	}

	exitCode, _ := cmd.exec(debug)

	if exitCode != 0 {
		return false
	}

	return true
}
