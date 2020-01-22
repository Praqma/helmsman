package app

import (
	"os"
)

const (
	helmBin            = "helm"
	appVersion         = "v3.0.1"
	tempFilesDir       = ".helmsman-tmp"
	defaultContextName = "default"
	resourcePool       = 10
)

var (
	flags      cli
	settings   config
	curContext string
)

func init() {
	// Parse cli flags and read config files
	flags.parse()
}

// Main is the app main function
func Main() {
	var s state

	// delete temp files with substituted env vars when the program terminates
	defer os.RemoveAll(tempFilesDir)
	defer s.cleanup()

	flags.readState(&s)
	settings = s.Settings
	curContext = s.Context

	// set the kubecontext to be used Or create it if it does not exist
	log.Info("Setting up kubectl...")
	if !setKubeContext(settings.KubeContext) {
		if err := createContext(&s); err != nil {
			log.Fatal(err.Error())
		}
	}

	// add repos -- fails if they are not valid
	if !flags.destroy {
		log.Info("Setting up helm...")
		if err := addHelmRepos(s.HelmRepos); err != nil {
			log.Fatal(err.Error())
		}
	}

	if flags.apply || flags.dryRun || flags.destroy {
		// add/validate namespaces
		if !flags.noNs {
			log.Info("Setting up namespaces...")
			if flags.nsOverride == "" {
				addNamespaces(s.Namespaces)
			} else {
				createNamespace(flags.nsOverride)
				s.overrideAppsNamespace(flags.nsOverride)
			}
		}
	}

	if !flags.skipValidation {
		log.Info("Validating charts...")
		// validate charts-versions exist in defined repos
		if err := validateReleaseCharts(&s); err != nil {
			log.Fatal(err.Error())
		}
	} else {
		log.Info("Skipping charts' validation.")
	}

	log.Info("Preparing plan...")
	if flags.destroy {
		log.Warning("Destroy flag is enabled. Your releases will be deleted!")
	}

	cs := buildState(&s)
	p := cs.makePlan(&s)
	if !flags.keepUntrackedReleases {
		cs.cleanUntrackedReleases(&s, p)
	}

	p.sort()
	p.print()
	if flags.debug {
		p.printCmds()
	}
	p.sendToSlack()

	if flags.apply || flags.dryRun || flags.destroy {
		p.exec()
	}
}

// cleanup deletes the k8s certificates and keys files
// It also deletes any Tiller TLS certs and keys
// and secret files
func (s *state) cleanup() {
	log.Verbose("Cleaning up sensitive and temp files")
	if _, err := os.Stat("ca.crt"); err == nil {
		deleteFile("ca.crt")
	}

	if _, err := os.Stat("ca.key"); err == nil {
		deleteFile("ca.key")
	}

	if _, err := os.Stat("client.crt"); err == nil {
		deleteFile("client.crt")
	}

	if _, err := os.Stat("bearer.token"); err == nil {
		deleteFile("bearer.token")
	}

	for _, app := range s.Apps {
		if _, err := os.Stat(app.SecretsFile + ".dec"); err == nil {
			deleteFile(app.SecretsFile + ".dec")
		}
		for _, secret := range app.SecretsFiles {
			if _, err := os.Stat(secret + ".dec"); err == nil {
				deleteFile(secret + ".dec")
			}
		}
	}
}
