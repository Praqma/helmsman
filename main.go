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
var nsOverride string
var checkCleanup bool
var skipValidation bool
var applyLabels bool
var keepUntrackedReleases bool
var appVersion = "v1.6.0"
var helmVersion string
var kubectlVersion string
var pwd string
var relativeDir string
var dryRun bool

func main() {
	// set the kubecontext to be used Or create it if it does not exist
	if !setKubeContext(s.Settings.KubeContext) {
		if r, msg := createContext(); !r {
			logError(msg)
		}
		checkCleanup = true
	}

	// add/validate namespaces
	addNamespaces(s.Namespaces)

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

	p := makePlan(&s)
	if !keepUntrackedReleases {
		cleanUntrackedReleases()
	}

	p.sortPlan()
	p.printPlan()
	p.sendPlanToSlack()

	if apply || dryRun {
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
		if _, err := os.Stat(app.SecretFile + ".dec"); err == nil {
			deleteFile(app.SecretFile + ".dec")
		}
		for _, secret := range app.SecretFiles {
			if _, err := os.Stat(secret + ".dec"); err == nil {
				deleteFile(secret + ".dec")
			}
		}
	}
}
