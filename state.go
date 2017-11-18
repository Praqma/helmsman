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
	Metadata     map[string]string
	Certificates map[string]string
	Settings     map[string]string
	Namespaces   map[string]string
	HelmRepos    map[string]string
	Apps         map[string]release
}

// validate validates that the values specified in the desired state are valid according to the desired state spec.
// check https://github.com/Praqma/Helmsman/docs/desired_state_spec.md for the detailed specification
func (s state) validate() (bool, string) {

	// settings
	if s.Settings == nil {
		return false, "ERROR: settings validation failed -- no settings table provided in TOML."
	} else if s.Settings["kubeContext"] == "" {
		return false, "ERROR: settings validation failed -- you have not provided a " +
			"kubeContext to use. Can't work without it. Sorry!"
	} else if len(s.Settings) > 1 {
		if s.Settings["password"] == "" || !strings.HasPrefix(s.Settings["password"], "$") {
			return false, "ERROR: settings validation failed -- password should be an env variable starting with $ "
		} else if _, err := url.ParseRequestURI(s.Settings["clusterURI"]); err != nil {
			return false, "ERROR: settings validation failed -- clusterURI must have a valid URL."
		}
	}

	// certificates
	if s.Certificates != nil {
		if len(s.Settings) > 1 && len(s.Certificates) != 2 {
			return false, "ERROR: certifications validation failed -- You want me to connect to your cluster for you " +
				"but have not given me the keys to do so. Please add [caCrt] and [caKey] under Certifications."
		}
		for key, value := range s.Certificates {
			_, err := url.ParseRequestURI(value)
			if err != nil || !strings.HasPrefix(value, "s3://") {
				return false, "ERROR: certifications validation failed -- [ " + key + " ] must be a valid S3 bucket URL."
			}
		}

	} else {
		if len(s.Settings) > 1 {
			return false, "ERROR: certifications validation failed -- You want me to connect to your cluster for you " +
				"but have not given me the keys to do so. Please add [caCrt] and [caKey] under Certifications."
		}
	}

	// namespaces
	if s.Namespaces == nil || len(s.Namespaces) < 1 {
		return false, "ERROR: namespaces validation failed -- I need at least one namespace " +
			"to work with!"
	}
	for k, v := range s.Namespaces {
		if v == "" {
			return false, "ERROR: namespaces validation failed -- namespace [" + k + " ] " +
				"must have a value or be removed/commented."
		}
	}

	// repos
	if s.HelmRepos == nil || len(s.HelmRepos) < 1 {
		return false, "ERROR: repos validation failed -- I need at least one helm repo " +
			"to work with!"
	}
	for k, v := range s.HelmRepos {
		_, err := url.ParseRequestURI(v)
		if err != nil {
			return false, "ERROR: repos validation failed -- repo [" + k + " ] " +
				"must have a valid URL."
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
			return false, "ERROR: apps validation failed -- for app [" + appLabel + " ]. " + errMsg
		}
	}

	return true, ""
}

// print prints the desired state
func (s state) print() {

	fmt.Println("\nMetadata: ")
	fmt.Println("--------- ")
	printMap(s.Metadata)
	fmt.Println("\nCertificates: ")
	fmt.Println("--------- ")
	printMap(s.Certificates)
	fmt.Println("\nSettings: ")
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
