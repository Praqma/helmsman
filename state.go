package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

// config type represents the settings fields
type config struct {
	KubeContext     string `yaml:"kubeContext"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	ClusterURI      string `yaml:"clusterURI"`
	ServiceAccount  string `yaml:"serviceAccount"`
	StorageBackend  string `yaml:"storageBackend"`
	SlackWebhook    string `yaml:"slackWebhook"`
	ReverseDelete   bool   `yaml:"reverseDelete"`
	BearerToken     bool   `yaml:"bearerToken"`
	BearerTokenPath string `yaml:"bearerTokenPath"`
}

// state type represents the desired state of applications on a k8s cluster.
type state struct {
	Metadata               map[string]string    `yaml:"metadata"`
	Certificates           map[string]string    `yaml:"certificates"`
	Settings               config               `yaml:"settings"`
	Namespaces             map[string]namespace `yaml:"namespaces"`
	HelmRepos              map[string]string    `yaml:"helmRepos"`
	PreconfiguredHelmRepos []string             `yaml:"preconfiguredHelmRepos"`
	Apps                   map[string]*release  `yaml:"apps"`
}

// validate validates that the values specified in the desired state are valid according to the desired state spec.
// check https://github.com/Praqma/Helmsman/docs/desired_state_spec.md for the detailed specification
func (s state) validate() (bool, string) {

	// settings
	if (s.Settings == (config{}) || s.Settings.KubeContext == "") && !getKubeContext() {
		return false, "ERROR: settings validation failed -- you have not defined a " +
			"kubeContext to use. Either define it in the desired state file or pass a kubeconfig with --kubeconfig to use an existing context."
	} else if s.Settings.ClusterURI != "" {

		if _, err := url.ParseRequestURI(s.Settings.ClusterURI); err != nil {
			return false, "ERROR: settings validation failed -- clusterURI must have a valid URL set in an env variable or passed directly. Either the env var is missing/empty or the URL is invalid."
		}
		if s.Settings.KubeContext == "" {
			return false, "ERROR: settings validation failed -- KubeContext needs to be provided in the settings stanza."
		}
		if !s.Settings.BearerToken && s.Settings.Username == "" {
			return false, "ERROR: settings validation failed -- username needs to be provided in the settings stanza."
		}
		if !s.Settings.BearerToken && s.Settings.Password == "" {
			return false, "ERROR: settings validation failed -- password needs to be provided (directly or from env var) in the settings stanza."
		}
		if s.Settings.BearerToken && s.Settings.BearerTokenPath != "" {
			if _, err := os.Stat(s.Settings.BearerTokenPath); err != nil {
				return false, "ERROR: settings validation failed -- bearer token path " + s.Settings.BearerTokenPath + " is not found. The path has to be relative to the desired state file."
			}
		}
	} else if s.Settings.BearerToken && s.Settings.ClusterURI == "" {
		return false, "ERROR: settings validation failed -- bearer token is enabled but no cluster URI provided."
	}

	// slack webhook validation (if provided)
	if s.Settings.SlackWebhook != "" {
		if _, err := url.ParseRequestURI(s.Settings.SlackWebhook); err != nil {
			return false, "ERROR: settings validation failed -- slackWebhook must be a valid URL."
		}
	}

	// certificates
	if s.Certificates != nil && len(s.Certificates) != 0 {

		for key, value := range s.Certificates {
			r, path := isValidCert(value)
			if !r {
				return false, "ERROR: certifications validation failed -- [ " + key + " ] must be a valid S3, GCS, AZ bucket/container URL or a valid relative file path."
			}
			s.Certificates[key] = path
		}

		_, caCrt := s.Certificates["caCrt"]
		_, caKey := s.Certificates["caKey"]

		if s.Settings.ClusterURI != "" && !s.Settings.BearerToken {
			if !caCrt || !caKey {
				return false, "ERROR: certificates validation failed -- You want me to connect to your cluster for you " +
					"but have not given me the cert/key to do so. Please add [caCrt] and [caKey] under Certifications. You might also need to provide [clientCrt]."
			}

		} else if s.Settings.ClusterURI != "" && s.Settings.BearerToken {
			if !caCrt {
				return false, "ERROR: certificates validation failed -- cluster connection with bearer token is enabled but " +
					"[caCrt] is missing. Please provide [caCrt] in the Certifications stanza."
			}
		}

	} else {
		if s.Settings.ClusterURI != "" {
			return false, "ERROR: certificates validation failed -- kube context setup is required but no certificates stanza provided."
		}
	}

	// namespaces
	if nsOverride == "" {
		if s.Namespaces == nil || len(s.Namespaces) == 0 {
			return false, "ERROR: namespaces validation failed -- I need at least one namespace " +
				"to work with!"
		}

		for k, ns := range s.Namespaces {
			if ns.InstallTiller && ns.UseTiller {
				return false, "ERROR: namespaces validation failed -- installTiller and useTiller can't be used together for namespace [ " + k + " ]"
			}
			if ns.UseTiller {
				log.Println("INFO: namespace validation -- a pre-installed Tiller is desired to be used in namespace [ " + k + " ].")
			} else if !ns.InstallTiller {
				log.Println("INFO: namespace validation -- Tiller is NOT desired to be deployed in namespace [ " + k + " ].")
			}

			if ns.UseTiller || ns.InstallTiller {
				// validating the TLS certs and keys for Tiller
				// if they are valid, their values (if they are env vars) are substituted
				var ok1, ok2, ok3, ok4, ok5 bool
				ok1, ns.CaCert = isValidCert(ns.CaCert)
				ok2, ns.ClientCert = isValidCert(ns.ClientCert)
				ok3, ns.ClientKey = isValidCert(ns.ClientKey)
				ok4, ns.TillerCert = isValidCert(ns.TillerCert)
				ok5, ns.TillerKey = isValidCert(ns.TillerKey)

				if ns.InstallTiller {
					if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
						log.Println("INFO: namespace validation -- Either no or invalid certs/keys provided for DEPLOYING Tiller with TLS in namespace [ " + k + " ].")
					} else {
						log.Println("INFO: namespace validation -- Tiller is desired to be DEPLOYED with TLS in namespace [ " + k + " ]. ")
					}
				} else if ns.UseTiller {
					if !ok1 || !ok2 || !ok3 {
						log.Println("INFO: namespace validation -- Either no or invalid certs/keys provided for USING Tiller with TLS in namespace [ " + k + " ].")
					} else {
						log.Println("INFO: namespace validation -- Tiller is desired to be USED with TLS in namespace [ " + k + " ]. ")
					}
				}
			}
		}
	} else {
		log.Println("INFO: ns-override is used to override all namespaces with [ " + nsOverride + " ] Skipping defined namespaces validation.")
	}

	// repos
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

// isValidCert checks if a certificate/key path/URI is valid
func isValidCert(value string) (bool, string) {
	_, err1 := url.ParseRequestURI(value)
	_, err2 := os.Stat(value)
	if err2 != nil && (err1 != nil || (!strings.HasPrefix(value, "s3://") && !strings.HasPrefix(value, "gs://") && !strings.HasPrefix(value, "az://"))) {
		return false, ""
	}
	return true, value
}

// tillerTLSEnabled checks if Tiller is desired to be deployed with TLS enabled for a given namespace or
// if helmsman is supposed to use an existing Tiller which is secured with TLS.
// For deploying Tiller, TLS is considered desired ONLY if all certs and keys for both Tiller and the Helm client are provided.
// For using an existing Tiller, TLS is considered desired ONLY if "CaCert" & "ClientCert" & "ClientKey" are provided.
func tillerTLSEnabled(ns namespace) bool {
	if ns.UseTiller {
		if ns.CaCert != "" && ns.ClientCert != "" && ns.ClientKey != "" {
			return true
		}
	} else if ns.InstallTiller {
		if ns.CaCert != "" && ns.TillerCert != "" && ns.TillerKey != "" && ns.ClientCert != "" && ns.ClientKey != "" {
			return true
		}
	}
	return false
}

// print prints the desired state
func (s state) print() {

	fmt.Println("\nMetadata: ")
	fmt.Println("--------- ")
	printMap(s.Metadata, 0)
	fmt.Println("\nCertificates: ")
	fmt.Println("--------- ")
	printMap(s.Certificates, 0)
	fmt.Println("\nSettings: ")
	fmt.Println("--------- ")
	fmt.Printf("%+v\n", s.Settings)
	fmt.Println("\nNamespaces: ")
	fmt.Println("------------- ")
	printNamespacesMap(s.Namespaces)
	fmt.Println("\nRepositories: ")
	fmt.Println("------------- ")
	printMap(s.HelmRepos, 0)
	fmt.Println("\nApplications: ")
	fmt.Println("--------------- ")
	for _, r := range s.Apps {
		r.print()
	}
}
