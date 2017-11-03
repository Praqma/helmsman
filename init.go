package main

import (
	"log"
	"os"
)

func init() {
	// check helm exists
	if !helmExists() {
		log.Fatal("ERROR: helm is not installed/configured correctly. Aborting!")
		os.Exit(1)
	}

	// TODO : setup cluster connection
}

func helmExists() bool {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm "},
		Description: "validating that helm is installed.",
	}

	exitCode, _ := cmd.exec()

	if exitCode != 0 {
		return false
	}

	return true
}
