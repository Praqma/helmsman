package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// init is executed after all package vars are initialized [before the main() func in this case].
// It checks if Helm and Kubectl exist and configures: the connection to the k8s cluster, helm repos, namespaces, etc.
func init() {
	//parsing command line flags
	flag.StringVar(&file, "f", "example.toml", "desired state file name")
	flag.BoolVar(&apply, "apply", false, "apply the plan directly")
	flag.BoolVar(&debug, "debug", false, "show the execution logs")
	flag.BoolVar(&help, "help", false, "show Helmsman help")
	flag.BoolVar(&v, "v", false, "show the version")

	flag.Parse()

	if v {
		fmt.Println("Helmsman version: " + version)
		os.Exit(0)
	}

	if help {
		printHelp()
		os.Exit(0)
	}

	fmt.Println("Helmsman version: " + version)

	if !toolExists("helm") {
		log.Fatal("ERROR: helm is not installed/configured correctly. Aborting!")
	}

	if !toolExists("kubectl") {
		log.Fatal("ERROR: kubectl is not installed/configured correctly. Aborting!")
	}

	// after the init() func is run, read the TOML desired state file
	result, msg := fromTOML(file, &s)
	if result {
		log.Printf(msg)
	} else {
		log.Fatal(msg)
	}

	// validate the desired state content
	if result, msg := s.validate(); !result { // syntax validation
		log.Fatal(msg)
	}

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
