package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	banner = " _          _ \n" +
		"| |        | | \n" +
		"| |__   ___| |_ __ ___  ___ _ __ ___   __ _ _ __\n" +
		"| '_ \\ / _ \\ | '_ ` _ \\/ __| '_ ` _ \\ / _` | '_ \\ \n" +
		"| | | |  __/ | | | | | \\__ \\ | | | | | (_| | | | | \n" +
		"|_| |_|\\___|_|_| |_| |_|___/_| |_| |_|\\__,_|_| |_|"
	slogan = "A Helm-Charts-as-Code tool.\n\n"
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
	flag.BoolVar(&verbose, "verbose", false, "show verbose execution logs")
	flag.StringVar(&nsOverride, "ns-override", "", "override defined namespaces with this one")
	flag.BoolVar(&skipValidation, "skip-validation", false, "skip desired state validation")
	flag.BoolVar(&applyLabels, "apply-labels", false, "apply Helmsman labels to Helm state for all defined apps.")
	flag.BoolVar(&keepUntrackedReleases, "keep-untracked-releases", false, "keep releases that are managed by Helmsman and are no longer tracked in your desired state.")

	flag.Parse()

	fmt.Println(banner + "version: " + version + "\n" + slogan)

	if v {
		fmt.Println("Helmsman version: " + version)
		os.Exit(0)
	}

	if help {
		printHelp()
		os.Exit(0)
	}

	//fmt.Println("Helmsman version: " + version)

	if !toolExists("helm") {
		log.Fatal("ERROR: helm is not installed/configured correctly. Aborting!")
	}

	if !toolExists("kubectl") {
		log.Fatal("ERROR: kubectl is not installed/configured correctly. Aborting!")
	}

	// read the TOML/YAML desired state file
	result, msg := fromFile(file, &s)
	if result {
		log.Printf(msg)
	} else {
		log.Fatal(msg)
	}

	if !skipValidation {
		// validate the desired state content
		if result, msg := s.validate(); !result { // syntax validation
			log.Fatal(msg)
		}
	} else {
		log.Println("INFO: desired state validation is skipped.")
	}

	if applyLabels {
		for _, r := range s.Apps {

			labelResource(r)
		}
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

	exitCode, _ := cmd.exec(debug, false)

	if exitCode != 0 {
		return false
	}

	return true
}
