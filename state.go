package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

// state type represents the desired state of applications on a k8s cluster.
type state struct {
	Metadata       map[string]string
	Certifications map[string]string
	Settings       map[string]string
	Namespaces     map[string]string
	HelmRepos      map[string]string
	Apps           map[string]release
}

// validate validates that the values specified in the desired state are valid according to the desired state spec.
// check https://github.com/Praqma/Helmsman/docs/desired_state_spec.md for the detailed specification
func (s state) validate() bool {

	// settings
	if s.Settings == nil {
		log.Fatal("ERROR: settings validation failed -- no settings table provided in TOML.")
		return false
	} else if s.Settings["kubeContext"] == "" {
		log.Fatal("ERROR: settings validation failed -- you have not provided a ",
			"kubeContext to use. Can't work without it. Sorry!")
		return false
	} else if len(s.Settings) > 1 {
		_, err := os.Stat(s.Settings["password"])
		if s.Settings["password"] == "" || !isOfType(s.Settings["password"], ".passwd") || err != nil {
			log.Fatal("ERROR: settings validation failed -- the specified password file is not valid. ",
				"It should be a valid path and of type \".passwd\".")
			return false
		} else if _, err := url.ParseRequestURI(s.Settings["clusterURI"]); err != nil {
			log.Fatal("ERROR: settings validation failed -- clusterURI must have a valid URL.")
			return false
		}
	}

	// certifications
	if s.Certifications != nil {
		if len(s.Settings) > 1 && len(s.Certifications) < 2 {
			log.Fatal("ERROR: certifications validation failed -- You want me to connect to your cluster for you ",
				"but have not given me the keys to do so. Please add [caCrt] and [caKey] under Certifications.")
			return false
		}
		for key, value := range s.Certifications {
			_, err := url.ParseRequestURI(value)
			if err != nil || !strings.HasPrefix(value, "s3://") {
				log.Fatal("ERROR: certifications validation failed -- [ " + key + " ] must be a valid s3 bucket URL.")
				return false
			}
		}

	}

	// namespaces
	if s.Namespaces == nil || len(s.Namespaces) < 1 {
		log.Fatal("ERROR: namespaces validation failed -- I need at least one namespace ",
			"to work with!")
		return false
	}
	for k, v := range s.Namespaces {
		if v == "" {
			log.Fatal("ERROR: namespaces validation failed -- namespace ["+k+" ] ",
				"must have a value or be removed/commented.")
			return false
		}
	}

	// repos
	if s.HelmRepos == nil || len(s.HelmRepos) < 1 {
		log.Fatal("ERROR: repos validation failed -- I need at least one helm repo ",
			"to work with!")
		return false
	}
	for k, v := range s.HelmRepos {
		_, err := url.ParseRequestURI(v)
		if err != nil {
			log.Fatal("ERROR: repos validation failed -- repo ["+k+" ] ",
				"must have a valid URL.")
			return false
		}

		continue

	}

	// apps
	if s.Apps == nil {
		log.Println("INFO: You have not specified any apps. I have nothing to do. ",
			"Horraayyy!.")
		os.Exit(0)
	}

	names := make(map[string]bool)
	for appLabel, r := range s.Apps {
		result, errMsg := validateRelease(r, names)
		if !result {
			log.Fatal("ERROR: apps validation failed -- for app ["+appLabel+" ]. ",
				errMsg)
			return false
		}
	}

	return true
}

// print prints the desired state
func (s state) print() {

	fmt.Println("\nMetadata: ")
	fmt.Println("--------- ")
	printMap(s.Metadata)
	fmt.Println("\nCertifications: ")
	fmt.Println("--------- ")
	printMap(s.Certifications)
	fmt.Println("Settings: ")
	fmt.Println("--------- ")
	printMap(s.Settings)
	fmt.Println("\nNamespaces: ")
	fmt.Println("------------- ")
	printMap(s.Namespaces)
	fmt.Println("\nRepositories: ")
	fmt.Println("------------- ")
	printMap(s.HelmRepos)
	fmt.Println("\nApplications: ")
	fmt.Println("--------------- ")
	for _, r := range s.Apps {
		r.print()
	}
}
