package main

import (
	"fmt"
	"os"
	"strings"
)

// release type representing Helm releases which are described in the desired state
type release struct {
	Name        string
	Description string
	Env         string
	Enabled     bool
	Chart       string
	Version     string
	ValuesFile  string
	Purge       bool
	Test        bool
}

// validateRelease validates if a release inside a desired state meets the specifications or not.
// check the full specification @ https://github.com/Praqma/Helmsman/docs/desired_state_spec.md
func validateRelease(r release, names map[string]bool) (bool, string) {
	_, err := os.Stat(r.ValuesFile)
	if r.Name == "" || names[r.Name] {
		return false, "release name can't be empty and must be unique."
	} else if r.Env == "" {
		return false, "env can't be empty."
	} else if r.Chart == "" || !strings.ContainsAny(r.Chart, "/") {
		return false, "chart can't be empty and must be of the format: repo/chart."
	} else if r.Version == "" {
		return false, "version can't be empty."
	} else if r.ValuesFile != "" && (!isOfType(r.ValuesFile, ".yaml") || err != nil) {
		return false, "valuesFile must be a valid file path for a yaml file, Or can be left empty."
	}

	names[r.Name] = true
	return true, ""
}

// print prints the details of the release
func (r release) print() {
	fmt.Println("")
	fmt.Println("\tname : ", r.Name)
	fmt.Println("\tdescription : ", r.Description)
	fmt.Println("\tenv : ", r.Env)
	fmt.Println("\tenabled : ", r.Enabled)
	fmt.Println("\tchart : ", r.Chart)
	fmt.Println("\tversion : ", r.Version)
	fmt.Println("\tvaluesFile : ", r.ValuesFile)
	fmt.Println("\tpurge : ", r.Purge)
	fmt.Println("\ttest : ", r.Test)
	fmt.Println("------------------- ")
}
