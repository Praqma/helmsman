package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

// namespace type represents the fields of a namespace
type namespace struct {
	Protected bool
}

// state type represents the desired state of applications on a k8s cluster.
type state struct {
	Metadata     map[string]string
	Certificates map[string]string
	Settings     map[string]string
	Namespaces   map[string]namespace
	HelmRepos    map[string]string
	Apps         map[string]*release
}

// validate validates that the values specified in the desired state are valid according to the desired state spec.
// check https://github.com/Praqma/Helmsman/docs/desired_state_spec.md for the detailed specification
func (s state) validate() (bool, string) {

	// settings
	if s.Settings == nil || len(s.Settings) == 0 {
		return false, "ERROR: settings validation failed -- no settings table provided in TOML."
	} else if value, ok := s.Settings["kubeContext"]; !ok || value == "" {
		return false, "ERROR: settings validation failed -- you have not provided a " +
			"kubeContext to use. Can't work without it. Sorry!"
	} else if value, ok = s.Settings["clusterURI"]; ok {

		s.Settings["clusterURI"] = subsituteEnv(value)
		if _, err := url.ParseRequestURI(s.Settings["clusterURI"]); err != nil {
			return false, "ERROR: settings validation failed -- clusterURI must have a valid URL set in an env varibale or passed directly. Either the env var is missing/empty or the URL is invalid."
		}

		if _, ok = s.Settings["username"]; !ok {
			return false, "ERROR: settings validation failed -- username must be provided if clusterURI is defined."
		}
		if value, ok = s.Settings["password"]; ok {
			s.Settings["password"] = subsituteEnv(value)
		} else {
			return false, "ERROR: settings validation failed -- password must be provided if clusterURI is defined."
		}

		if s.Settings["password"] == "" {
			return false, "ERROR: settings validation failed -- password should be set as an env variable. It is currently missing or empty. "
		}
	}

	// certificates
	if s.Certificates != nil && len(s.Certificates) != 0 {
		_, ok1 := s.Settings["clusterURI"]
		_, ok2 := s.Certificates["caCrt"]
		_, ok3 := s.Certificates["caKey"]
		if ok1 && (!ok2 || !ok3) {
			return false, "ERROR: certifications validation failed -- You want me to connect to your cluster for you " +
				"but have not given me the cert/key to do so. Please add [caCrt] and [caKey] under Certifications. You might also need to provide [clientCrt]."
		} else if ok1 {
			for key, value := range s.Certificates {
				tmp := subsituteEnv(value)
				_, err1 := url.ParseRequestURI(tmp)
				_, err2 := os.Stat(tmp)
				if (err1 != nil || (!strings.HasPrefix(tmp, "s3://") && !strings.HasPrefix(tmp, "gs://"))) && err2 != nil {
					return false, "ERROR: certifications validation failed -- [ " + key + " ] must be a valid S3 or GCS bucket URL or a valid relative file path."
				}
				s.Certificates[key] = tmp
			}
		} else {
			log.Println("INFO: certificates provided but not needed. Skipping certificates validation.")
		}

	} else {
		if _, ok := s.Settings["clusterURI"]; ok {
			return false, "ERROR: certifications validation failed -- You want me to connect to your cluster for you " +
				"but have not given me the cert/key to do so. Please add [caCrt] and [caKey] under Certifications. You might also need to provide [clientCrt]."
		}
	}

	// namespaces
	if nsOverride == "" {
		if s.Namespaces == nil || len(s.Namespaces) == 0 {
			return false, "ERROR: namespaces validation failed -- I need at least one namespace " +
				"to work with!"
		}
	} else {
		log.Println("INFO: ns-override is used. Overriding all namespaces with [ " + nsOverride + " ] Skipping defined namespaces validation.")
	}

	// repos
	if s.HelmRepos == nil || len(s.HelmRepos) == 0 {
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
		result, errMsg := validateRelease(r, names, s)
		if !result {
			return false, "ERROR: apps validation failed -- for app [" + appLabel + " ]. " + errMsg
		}
	}

	return true, ""
}

// substitueEnv checks if a string is an env variable (starts with '$'), then it returns its value
// if the env variable is empty or unset, an empty string is returned
// if the string does not start with '$', it is returned as is.
func subsituteEnv(name string) string {
	if strings.HasPrefix(name, "$") {
		return os.Getenv(strings.SplitAfterN(name, "$", 2)[1])
	}
	return name
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
	printNamespacesMap(s.Namespaces)
	fmt.Println("\nRepositories: ")
	fmt.Println("------------- ")
	printMap(s.HelmRepos)
	fmt.Println("\nApplications: ")
	fmt.Println("--------------- ")
	for _, r := range s.Apps {
		r.print()
	}
}
