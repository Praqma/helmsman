package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// release type representing Helm releases which are described in the desired state
type release struct {
	Name        string
	Description string
	Namespace   string
	Enabled     bool
	Chart       string
	Version     string
	ValuesFile  string
	Purge       bool
	Test        bool
	Protected   bool
	Wait        bool
	Priority    int
	Set         map[string]string
}

// validateRelease validates if a release inside a desired state meets the specifications or not.
// check the full specification @ https://github.com/Praqma/helmsman/docs/desired_state_spec.md
func validateRelease(r *release, names map[string]bool, s state) (bool, string) {
	_, err := os.Stat(r.ValuesFile)
	if r.Name == "" || names[r.Name] {
		return false, "release name can't be empty and must be unique."
	} else if nsOverride == "" && r.Namespace == "" {
		return false, "release targeted namespace can't be empty."
	} else if nsOverride == "" && r.Namespace != "" && !checkNamespaceDefined(r.Namespace, s) {
		return false, "release " + r.Name + " is using namespace [ " + r.Namespace + " ] which is not defined in the Namespaces section of your desired state file." +
			" Release [ " + r.Name + " ] can't be installed in that Namespace until its defined."
	} else if r.Chart == "" || !strings.ContainsAny(r.Chart, "/") {
		return false, "chart can't be empty and must be of the format: repo/chart."
	} else if r.Version == "" {
		return false, "version can't be empty."
	} else if r.ValuesFile != "" && (!isOfType(r.ValuesFile, ".yaml") || err != nil) {
		return false, "valuesFile must be a valid file path for a yaml file, Or can be left empty."
	} else if len(r.Set) > 0 {
		for k, v := range r.Set {
			if !strings.HasPrefix(v, "$") {
				return false, "the value for variable [ " + k + " ] must be an environment variable name and start with '$'."
			} else if ok, _ := envVarExists(v); !ok {
				return false, "env variable [ " + v + " ] is not found in the environment."
			}
		}
	} else if r.Priority != 0 && r.Priority > 0 {
		return false, "priority can only be 0 or negative value, positive values are not allowed."
	}

	names[r.Name] = true
	return true, ""
}

// checkNamespaceDefined checks if a given namespace is defined in the namespaces section of the desired state file
func checkNamespaceDefined(ns string, s state) bool {
	_, ok := s.Namespaces[ns]
	if !ok {
		return false
	}
	return true
}

// overrideNamespace overrides a release defined namespace with a new given one
func overrideNamespace(r *release, newNs string) {
	log.Println("INFO: overriding namespace for app:  " + r.Name)
	r.Namespace = newNs
}

// print prints the details of the release
func (r release) print() {
	fmt.Println("")
	fmt.Println("\tname : ", r.Name)
	fmt.Println("\tdescription : ", r.Description)
	fmt.Println("\tnamespace : ", r.Namespace)
	fmt.Println("\tenabled : ", r.Enabled)
	fmt.Println("\tchart : ", r.Chart)
	fmt.Println("\tversion : ", r.Version)
	fmt.Println("\tvaluesFile : ", r.ValuesFile)
	fmt.Println("\tpurge : ", r.Purge)
	fmt.Println("\ttest : ", r.Test)
	fmt.Println("\tprotected : ", r.Protected)
	fmt.Println("\twait : ", r.Wait)
	fmt.Println("\tpriority : ", r.Priority)
	fmt.Println("\tvalues to override from env:")
	printMap(r.Set)
	fmt.Println("------------------- ")
}
