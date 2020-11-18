// Package app contains the main logic for the application.
package app

import (
	"os"
)

const (
	helmBin            = "helm"
	appVersion         = "v3.5.1"
	tempFilesDir       = ".helmsman-tmp"
	defaultContextName = "default"
	resourcePool       = 10
)

var (
	flags      cli
	settings   *config
	curContext string
	log        = &Logger{}
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
	if !flags.noCleanup {
		defer s.cleanup()
	}

	flags.readState(&s)
	log.SlackWebhook = s.Settings.SlackWebhook

	settings = &s.Settings
	curContext = s.Context

	// set the kubecontext to be used Or create it if it does not exist
	log.Info("Setting up kubectl")
	if !setKubeContext(s.Settings.KubeContext) {
		if err := createContext(&s); err != nil {
			log.Fatal(err.Error())
		}
	}

	// add repos -- fails if they are not valid
	log.Info("Setting up helm")
	if err := addHelmRepos(s.HelmRepos); err != nil && !flags.destroy {
		log.Fatal(err.Error())
	}

	if flags.apply || flags.dryRun || flags.destroy {
		// add/validate namespaces
		if !flags.noNs {
			log.Info("Setting up namespaces")
			if flags.nsOverride == "" {
				addNamespaces(&s)
			} else {
				createNamespace(flags.nsOverride)
				s.overrideAppsNamespace(flags.nsOverride)
			}
		}
	}

	log.Info("Getting chart information")

	err := s.getReleaseChartsInfo()
	if flags.skipValidation {
		log.Info("Skipping charts' validation.")
	} else if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Info("Charts validated.")
	}

	if flags.destroy {
		log.Warning("Destroy flag is enabled. Your releases will be deleted!")
	}

	if flags.migrateContext {
		log.Warning("migrate-context flag is enabled. Context will be changed to [ " + s.Context + " ] and Helmsman labels will be applied.")
		s.updateContextLabels()
	}

	log.Info("Preparing plan")
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
