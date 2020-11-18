package app

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
)

// config type represents the settings fields
type config struct {
	KubeContext         string                 `yaml:"kubeContext"`
	Username            string                 `yaml:"username"`
	Password            string                 `yaml:"password"`
	ClusterURI          string                 `yaml:"clusterURI"`
	ServiceAccount      string                 `yaml:"serviceAccount"`
	StorageBackend      string                 `yaml:"storageBackend"`
	SlackWebhook        string                 `yaml:"slackWebhook"`
	ReverseDelete       bool                   `yaml:"reverseDelete"`
	BearerToken         bool                   `yaml:"bearerToken"`
	BearerTokenPath     string                 `yaml:"bearerTokenPath"`
	EyamlEnabled        bool                   `yaml:"eyamlEnabled"`
	EyamlPrivateKeyPath string                 `yaml:"eyamlPrivateKeyPath"`
	EyamlPublicKeyPath  string                 `yaml:"eyamlPublicKeyPath"`
	GlobalHooks         map[string]interface{} `yaml:"globalHooks"`
	GlobalMaxHistory    int                    `yaml:"globalMaxHistory"`
}

// state type represents the desired state of applications on a k8s cluster.
type state struct {
	Metadata               map[string]string     `yaml:"metadata"`
	Certificates           map[string]string     `yaml:"certificates"`
	Settings               config                `yaml:"settings"`
	Context                string                `yaml:"context"`
	Namespaces             map[string]*namespace `yaml:"namespaces"`
	HelmRepos              map[string]string     `yaml:"helmRepos"`
	PreconfiguredHelmRepos []string              `yaml:"preconfiguredHelmRepos"`
	Apps                   map[string]*release   `yaml:"apps"`
	AppsTemplates          map[string]*release   `yaml:"appsTemplates,omitempty"`
	TargetMap              map[string]bool
	ChartInfo              map[string]map[string]*chartInfo
}

func (s *state) setDefaults() {
	if s.Settings.StorageBackend != "" {
		os.Setenv("HELM_DRIVER", s.Settings.StorageBackend)
	} else {
		// set default storage background to secret if not set by user
		s.Settings.StorageBackend = "secret"
	}

	// if there is no user-defined context name in the DSF(s), use the default context name
	if s.Context == "" {
		s.Context = defaultContextName
	}

	for name, r := range s.Apps {
		// Default app.Name to state name when unset
		if r.Name == "" {
			r.Name = name
		}
		// inherit globalHooks if local ones are not set
		r.inheritHooks(s)
		r.inheritMaxHistory(s)
	}
}

// validate validates that the values specified in the desired state are valid according to the desired state spec.
// check https://github.com/Praqma/helmsman/blob/master/docs/desired_state_specification.md for the detailed specification
func (s *state) validate() error {

	// apps
	if s.Apps == nil {
		log.Info("No apps specified. Nothing to be executed.")
		os.Exit(0)
	}

	// settings
	// use reflect.DeepEqual to compare Settings are empty, since it contains a map
	if (reflect.DeepEqual(s.Settings, config{}) || s.Settings.KubeContext == "") && !getKubeContext() {
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

	// lifecycle hooks validation
	if len(s.Settings.GlobalHooks) != 0 {
		if err := validateHooks(s.Settings.GlobalHooks); err != nil {
			return err
		}
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
			if !isValidCert(value) {
				return errors.New("certifications validation failed -- [ " + key + " ] must be a valid S3, GCS, AZ bucket/container URL or a valid relative file path")
			}
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

// getReleaseChartsInfo retrieves valid chart information.
// Valid charts are the ones that can be found in the defined repos.
func (s *state) getReleaseChartsInfo() error {
	var fail bool
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	sem := make(chan struct{}, resourcePool)
	chartErrors := make(chan error, len(s.Apps))

	charts := make(map[string]map[string][]string)
	s.ChartInfo = make(map[string]map[string]*chartInfo)

	for app, r := range s.Apps {
		if !r.isConsideredToRun() {
			continue
		}

		if charts[r.Chart] == nil {
			charts[r.Chart] = make(map[string][]string)
		}

		if charts[r.Chart][r.Version] == nil {
			charts[r.Chart][r.Version] = make([]string, 0)
		}

		if s.ChartInfo[r.Chart] == nil {
			s.ChartInfo[r.Chart] = make(map[string]*chartInfo)
		}

		charts[r.Chart][r.Version] = append(charts[r.Chart][r.Version], app)
	}

	for chart, versions := range charts {
		for version, apps := range versions {
			concattedApps := strings.Join(apps, ", ")

			sem <- struct{}{}
			wg.Add(1)
			go func(apps, chart, version string) {
				defer func() {
					wg.Done()
					<-sem
				}()

				info, err := getChartInfo(chart, version)
				if err != nil {
					chartErrors <- err
				} else {
					log.Verbose(fmt.Sprintf("Extracted chart information from chart [ %s ] with version [ %s ]: %s %s", chart, version, info.Name, info.Version))
					mutex.Lock()
					s.ChartInfo[chart][version] = info
					mutex.Unlock()
				}

				//validateChart(concattedApps, chart, version, chartErrors)
			}(concattedApps, chart, version)
		}
	}

	wg.Wait()
	close(chartErrors)
	for err := range chartErrors {
		if err != nil {
			fail = true
			log.Error(err.Error())
		}
	}
	if fail {
		return errors.New("chart validation failed")
	}
	return nil
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

// get only those Apps that exist in TargetMap
func (s *state) disableUntargetedApps(groups, targets []string) {
	if s.TargetMap == nil {
		s.TargetMap = make(map[string]bool)
	}
	if len(targets) == 0 && len(groups) == 0 {
		return
	}
	for _, t := range targets {
		s.TargetMap[t] = true
	}
	groupMap := make(map[string]struct{})
	namespaces := make(map[string]struct{})
	for _, g := range groups {
		groupMap[g] = struct{}{}
	}
	for appName, app := range s.Apps {
		if _, ok := s.TargetMap[appName]; ok {
			namespaces[app.Namespace] = struct{}{}
			continue
		}
		if _, ok := groupMap[app.Group]; ok {
			s.TargetMap[appName] = true
			namespaces[app.Namespace] = struct{}{}
		} else {
			app.Disable()
		}
	}

	for nsName, ns := range s.Namespaces {
		if _, ok := namespaces[nsName]; !ok {
			ns.Disable()
		}
	}
}

// updateContextLabels applies Helmsman labels including overriding any previously-set context with the one found in the DSF
func (s *state) updateContextLabels() {
	for _, r := range s.Apps {
		if r.isConsideredToRun() {
			log.Info("Updating context and reapplying Helmsman labels for release [ " + r.Name + " ]")
			r.mark(s.Settings.StorageBackend)
		} else {
			log.Warning(r.Name + " is not in the target group and therefore context and labels are not changed.")
		}
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
}
