package main

import (
	"log"
	"os"
)

// Allow parsing of multiple string command line options into an array of strings
type stringArray []string

func (i *stringArray) String() string {
	return "my string representation"
}

func (i *stringArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var s state
var debug bool
var files stringArray
var envFiles stringArray
var apply bool
var v bool
var verbose bool
var noBanner bool
var noColors bool
var noFancy bool
var noNs bool
var nsOverride string
var checkCleanup bool
var skipValidation bool
var applyLabels bool
var keepUntrackedReleases bool
var appVersion = "v1.7.1"
var helmVersion string
var kubectlVersion string
var dryRun bool
var destroy bool
var showDiff bool
var suppressDiffSecrets bool

func main() {
	// set the kubecontext to be used Or create it if it does not exist
	if !setKubeContext(s.Settings.KubeContext) {
		if r, msg := createContext(); !r {
			logError(msg)
		}
		checkCleanup = true
	}

	if apply {
		// add/validate namespaces
		if !noNs {
			addNamespaces(s.Namespaces)
		}

		if r, msg := initHelm(); !r {
			logError(msg)
		}

		// check if helm Tiller is ready
		for k, ns := range s.Namespaces {
			if ns.InstallTiller || ns.UseTiller {
				waitForTiller(k)
			}
		}

		if _, ok := s.Namespaces["kube-system"]; !ok {
			waitForTiller("kube-system")
		}
	}

	// add repos -- fails if they are not valid
	if r, msg := addHelmRepos(s.HelmRepos); !r {
		logError(msg)
	}

	if !skipValidation {
		// validate charts-versions exist in defined repos
		if r, msg := validateReleaseCharts(s.Apps); !r {
			logError(msg)
		}
	} else {
		log.Println("INFO: charts validation is skipped.")
	}

	log.Println("INFO: checking what I need to do for your charts ... ")
	if destroy {
		log.Println("WARN: --destroy is enabled. Your releases will be deleted!")
	}

	p := makePlan(&s)
	if !keepUntrackedReleases {
		cleanUntrackedReleases()
	}

	p.sortPlan()
	p.printPlan()
	p.sendPlanToSlack()

	if apply || dryRun || destroy {
		p.execPlan()
	}

	if checkCleanup {
		cleanup()
	}

	log.Println("INFO: completed successfully!")
}

// cleanup deletes the k8s certificates and keys files
// It also deletes any Tiller TLS certs and keys
// and secret files
func cleanup() {
	if _, err := os.Stat("ca.crt"); err == nil {
		deleteFile("ca.crt")
	}

	if _, err := os.Stat("ca.key"); err == nil {
		deleteFile("ca.key")
	}

	if _, err := os.Stat("client.crt"); err == nil {
		deleteFile("client.crt")
	}

	for k := range s.Namespaces {
		if _, err := os.Stat(k + "-tiller.cert"); err == nil {
			deleteFile(k + "-tiller.cert")
		}
		if _, err := os.Stat(k + "-tiller.key"); err == nil {
			deleteFile(k + "-tiller.key")
		}
		if _, err := os.Stat(k + "-ca.cert"); err == nil {
			deleteFile(k + "-ca.cert")
		}
		if _, err := os.Stat(k + "-client.cert"); err == nil {
			deleteFile(k + "-client.cert")
		}
		if _, err := os.Stat(k + "-client.key"); err == nil {
			deleteFile(k + "-client.key")
		}
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
