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
	Protected            bool   `yaml:"protected"`
	InstallTiller        bool   `yaml:"installTiller"`
	TillerServiceAccount string `yaml:"tillerServiceAccount"`
	CaCert               string `yaml:"caCert"`
	TillerCert           string `yaml:"tillerCert"`
	TillerKey            string `yaml:"tillerKey"`
	ClientCert           string `yaml:"clientCert"`
	ClientKey            string `yaml:"clientKey"`
}

// state type represents the desired state of applications on a k8s cluster.
type state struct {
	Metadata     map[string]string    `yaml:"metadata"`
	Certificates map[string]string    `yaml:"certificates"`
	Settings     map[string]string    `yaml:"settings"`
	Namespaces   map[string]namespace `yaml:"namespaces"`
	HelmRepos    map[string]string    `yaml:"helmRepos"`
	Apps         map[string]*release  `yaml:"apps"`
}

// validate validates that the values specified in the desired state are valid according to the desired state spec.
// check https://github.com/Praqma/Helmsman/docs/desired_state_spec.md for the detailed specification
func (s state) validate() (bool, string) {

	// settings
	if s.Settings == nil || len(s.Settings) == 0 {
		return false, "ERROR: settings validation failed -- no settings table provided in state file."
	} else if value, ok := s.Settings["kubeContext"]; !ok || value == "" {
		return false, "ERROR: settings validation failed -- you have not provided a " +
			"kubeContext to use. Can't work without it. Sorry!"
	} else if value, ok = s.Settings["clusterURI"]; ok {

		s.Settings["clusterURI"] = subsituteEnv(value)
		if _, err := url.ParseRequestURI(s.Settings["clusterURI"]); err != nil {
			return false, "ERROR: settings validation failed -- clusterURI must have a valid URL set in an env variable or passed directly. Either the env var is missing/empty or the URL is invalid."
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

	// slack webhook validation (if provided)
	if value, ok := s.Settings["slackWebhook"]; ok {
		s.Settings["slackWebhook"] = subsituteEnv(value)
		if _, err := url.ParseRequestURI(s.Settings["slackWebhook"]); err != nil {
			return false, "ERROR: settings validation failed -- slackWebhook must be a valid URL."
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
				r, path := isValidCert(value)
				if !r {
					return false, "ERROR: certifications validation failed -- [ " + key + " ] must be a valid S3 or GCS bucket URL or a valid relative file path."
				}
				s.Certificates[key] = path
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

		for k, v := range s.Namespaces {
			if !v.InstallTiller {
				log.Println("INFO: namespace validation -- Tiller is NOT desired to be deployed in namespace [ " + k + " ].")
			} else {
				if tillerTLSEnabled(k) {
					// validating the TLS certs and keys for Tiller
					// if they are valid, their values (if they are env vars) are substituted
					var ok1, ok2, ok3, ok4, ok5 bool
					ok1, v.CaCert = isValidCert(v.CaCert)
					ok2, v.ClientCert = isValidCert(v.ClientCert)
					ok3, v.ClientKey = isValidCert(v.ClientKey)
					ok4, v.TillerCert = isValidCert(v.TillerCert)
					ok5, v.TillerKey = isValidCert(v.TillerKey)
					if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
						return false, "ERROR: namespaces validation failed  -- some certs/keys are not valid for Tiller TLS in namespace [ " + k + " ]."
					}
					log.Println("INFO: namespace validation -- Tiller is desired to be deployed with TLS in namespace [ " + k + " ]. ")
				} else {
					log.Println("INFO: namespace validation -- Tiller is desired to be deployed WITHOUT TLS in namespace [ " + k + " ]. ")
				}
			}
		}
	} else {
		log.Println("INFO: ns-override is used to override all namespaces with [ " + nsOverride + " ] Skipping defined namespaces validation.")
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

	names := make(map[string]map[string]bool)
	for appLabel, r := range s.Apps {
		result, errMsg := validateRelease(appLabel, r, names, s)
		if !result {
			return false, "ERROR: apps validation failed -- for app [" + appLabel + " ]. " + errMsg
		}
	}

	return true, ""
}

// subsituteEnv checks if a string is an env variable (contains '$'), then it returns its value
// if the env variable is empty or unset, an empty string is returned
// if the string does not contain '$', it is returned as is.
func subsituteEnv(name string) string {
	if strings.Contains(name, "$") {
		return os.ExpandEnv(name)
	}
	return name
}

// isValidCert checks if a certificate/key path/URI is valid
func isValidCert(value string) (bool, string) {
	tmp := subsituteEnv(value)
	_, err1 := url.ParseRequestURI(tmp)
	_, err2 := os.Stat(tmp)
	if err2 != nil && (err1 != nil || (!strings.HasPrefix(tmp, "s3://") && !strings.HasPrefix(tmp, "gs://"))) {
		return false, ""
	}
	return true, tmp
}

// tillerTLSEnabled checks if Tiller is desired to be deployed with TLS enabled for a given namespace
// TLS is considered desired ONLY if all certs and keys for both Tiller and the Helm client are defined.
func tillerTLSEnabled(namespace string) bool {

	ns := s.Namespaces[namespace]
	if ns.CaCert != "" && ns.TillerCert != "" && ns.TillerKey != "" && ns.ClientCert != "" && ns.ClientKey != "" {
		return true
	}
	return false
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
