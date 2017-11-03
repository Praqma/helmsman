package main

import (
	"log"
	"os"
)

// init is executed before the main is executed.
// It checks if Helm exists and configures the connection to the k8s cluster.
func init() {
	if !helmExists() {
		log.Fatal("ERROR: helm is not installed/configured correctly. Aborting!")
		os.Exit(1)
	}

	// TODO : setup cluster connection
}

// helmExists return true if Helm is present in the environment and false otherwise.
// It uses the Helm command to check if it is recognizable or not.
func helmExists() bool {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm "},
		Description: "validating that helm is installed.",
	}

	exitCode, _ := cmd.exec(debug)

	if exitCode != 0 {
		return false
	}

	return true
}
