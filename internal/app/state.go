package app

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// config type represents the settings fields
type config struct {
	KubeContext         string `yaml:"kubeContext"`
	Username            string `yaml:"username"`
	Password            string `yaml:"password"`
	ClusterURI          string `yaml:"clusterURI"`
	ServiceAccount      string `yaml:"serviceAccount"`
	StorageBackend      string `yaml:"storageBackend"`
	SlackWebhook        string `yaml:"slackWebhook"`
	ReverseDelete       bool   `yaml:"reverseDelete"`
	BearerToken         bool   `yaml:"bearerToken"`
	BearerTokenPath     string `yaml:"bearerTokenPath"`
	EyamlEnabled        bool   `yaml:"eyamlEnabled"`
	EyamlPrivateKeyPath string `yaml:"eyamlPrivateKeyPath"`
	EyamlPublicKeyPath  string `yaml:"eyamlPublicKeyPath"`
}

// state type represents the desired state of applications on a k8s cluster.
type state struct {
	Metadata               map[string]string    `yaml:"metadata"`
	Certificates           map[string]string    `yaml:"certificates"`
	Settings               config               `yaml:"settings"`
	Context                string               `yaml:"context"`
	Namespaces             map[string]namespace `yaml:"namespaces"`
	HelmRepos              map[string]string    `yaml:"helmRepos"`
	PreconfiguredHelmRepos []string             `yaml:"preconfiguredHelmRepos"`
	Apps                   map[string]*release  `yaml:"apps"`
	AppsTemplates          map[string]*release  `yaml:"appsTemplates,omitempty"`
	TargetMap              map[string]bool
	GroupMap               map[string]bool
}

// invokes either yaml or toml parser considering file extension
func (s *state) fromFile(file string) (bool, string) {
	if isOfType(file, []string{".toml"}) {
		return fromTOML(file, s)
	} else if isOfType(file, []string{".yaml", ".yml"}) {
		return fromYAML(file, s)
	} else {
		return false, "State file does not have toml/yaml extension."
	}
}

func (s *state) toFile(file string) {
	if isOfType(file, []string{".toml"}) {
		toTOML(file, s)
	} else if isOfType(file, []string{".yaml", ".yml"}) {
		toYAML(file, s)
	} else {
		log.Fatal("State file does not have toml/yaml extension.")
	}
}

// validate validates that the values specified in the desired state are valid according to the desired state spec.
// check https://github.com/Praqma/Helmsman/docs/desired_state_spec.md for the detailed specification
func (s *state) validate() error {

	// apps
	if s.Apps == nil {
		log.Info("No apps specified. Nothing to be executed.")
		os.Exit(0)
	}

	// settings
	if (s.Settings == (config{}) || s.Settings.KubeContext == "") && !getKubeContext() {
		return errors.New("settings validation failed -- you have not defined a " +
			"kubeContext to use. Either define it in the desired state file or pass a kubeconfig with --kubeconfig to use an existing context")
	}

	if s.Settings.ClusterURI != "" {
		if _, err := url.ParseRequestURI(s.Settings.ClusterURI); err != nil {
			return errors.New("settings validation failed -- clusterURI must have a valid URL set in an env variable or passed directly. Either the env var is missing/empty or the URL is invalid")
		}
		if s.Settings.KubeContext == "" {
			return errors.New("settings validation failed -- KubeContext needs to be provided in the settings stanza")
		}
		if !s.Settings.BearerToken && s.Settings.Username == "" {
			return errors.New("settings validation failed -- username needs to be provided in the settings stanza")
		}
		if !s.Settings.BearerToken && s.Settings.Password == "" {
			return errors.New("settings validation failed -- password needs to be provided (directly or from env var) in the settings stanza")
		}
		if s.Settings.BearerToken && s.Settings.BearerTokenPath != "" {
			if _, err := os.Stat(s.Settings.BearerTokenPath); err != nil {
				return errors.New("settings validation failed -- bearer token path " + s.Settings.BearerTokenPath + " is not found. The path has to be relative to the desired state file")
			}
		}
	} else if s.Settings.BearerToken && s.Settings.ClusterURI == "" {
		return errors.New("settings validation failed -- bearer token is enabled but no cluster URI provided")
	}

	// slack webhook validation (if provided)
	if s.Settings.SlackWebhook != "" {
		if _, err := url.ParseRequestURI(s.Settings.SlackWebhook); err != nil {
			return errors.New("settings validation failed -- slackWebhook must be a valid URL")
		}
	}

	// certificates
	if s.Certificates != nil && len(s.Certificates) != 0 {

		for key, value := range s.Certificates {
			r, path := isValidCert(value)
			if !r {
				return errors.New("certifications validation failed -- [ " + key + " ] must be a valid S3, GCS, AZ bucket/container URL or a valid relative file path")
			}
			s.Certificates[key] = path
		}

		_, caCrt := s.Certificates["caCrt"]
		_, caKey := s.Certificates["caKey"]

		if s.Settings.ClusterURI != "" && !s.Settings.BearerToken {
			if !caCrt || !caKey {
				return errors.New("certificates validation failed -- connection to cluster is required " +
					"but no cert/key was given. Please add [caCrt] and [caKey] under Certifications. You might also need to provide [clientCrt]")
			}

		} else if s.Settings.ClusterURI != "" && s.Settings.BearerToken {
			if !caCrt {
				return errors.New("certificates validation failed -- cluster connection with bearer token is enabled but " +
					"[caCrt] is missing. Please provide [caCrt] in the Certifications stanza")
			}
		}

	} else {
		if s.Settings.ClusterURI != "" {
			return errors.New("certificates validation failed -- kube context setup is required but no certificates stanza provided")
		}
	}

	if (s.Settings.EyamlPrivateKeyPath != "" && s.Settings.EyamlPublicKeyPath == "") || (s.Settings.EyamlPrivateKeyPath == "" && s.Settings.EyamlPublicKeyPath != "") {
		return errors.New("both EyamlPrivateKeyPath and EyamlPublicKeyPath are required")
	}

	// namespaces
	if flags.nsOverride == "" {
		if s.Namespaces == nil || len(s.Namespaces) == 0 {
			return errors.New("namespaces validation failed -- at least one namespace is required")
		}
	} else {
		log.Info("ns-override is used to override all namespaces with [ " + flags.nsOverride + " ] Skipping defined namespaces validation.")
	}

	// repos
	for k, v := range s.HelmRepos {
		_, err := url.ParseRequestURI(v)
		if err != nil {
			return errors.New("repos validation failed -- repo [" + k + " ] " +
				"must have a valid URL")
		}
	}

	names := make(map[string]map[string]bool)
	for appLabel, r := range s.Apps {
		if err := r.validate(appLabel, names, s); err != nil {
			return fmt.Errorf("apps validation failed -- for app ["+appLabel+" ]. %w", err)
		}
	}

	return nil
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

// isNamespaceDefined checks if a given namespace is defined in the namespaces section of the desired state file
func (s *state) isNamespaceDefined(ns string) bool {
	_, ok := s.Namespaces[ns]
	return ok
}

// overrideAppsNamespace replaces all apps namespaces with one specific namespace
func (s *state) overrideAppsNamespace(newNs string) {
	log.Info("Overriding apps namespaces with [ " + newNs + " ] ...")
	for _, r := range s.Apps {
		r.overrideNamespace(newNs)
	}
}

// print prints the desired state
func (s *state) print() {

	fmt.Println("\nMetadata: ")
	fmt.Println("--------- ")
	printMap(s.Metadata, 0)
	fmt.Println("\nContext: ")
	fmt.Println("--------- ")
	fmt.Println(s.Context)
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
	fmt.Println("\nTargets: ")
	fmt.Println("--------------- ")
	for t := range s.TargetMap {
		fmt.Println(t)
	}
	fmt.Println("\nGroups: ")
	fmt.Println("--------------- ")
	for g := range s.GroupMap {
		fmt.Println(g)
	}
}
