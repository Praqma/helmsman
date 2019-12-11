package app

import (
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
var kubeconfig string
var apply bool
var v bool
var verbose bool
var noBanner bool
var noColors bool
var noFancy bool
var noNs bool
var nsOverride string
var skipValidation bool
var applyLabels bool
var keepUntrackedReleases bool
var appVersion = "v3.0.0-beta1"
var helmBin = "helm"
var helmVersion string
var kubectlVersion string
var dryRun bool
var target stringArray
var group stringArray
var targetMap map[string]bool
var groupMap map[string]bool
var destroy bool
var showDiff bool
var suppressDiffSecrets bool
var diffContext int
var noEnvSubst bool
var noEnvValuesSubst bool
var noSSMSubst bool
var noSSMValuesSubst bool
var updateDeps bool
var forceUpgrades bool
var noDefaultRepos bool

const tempFilesDir = ".helmsman-tmp"
const stableHelmRepo = "https://kubernetes-charts.storage.googleapis.com"
const incubatorHelmRepo = "http://storage.googleapis.com/kubernetes-charts-incubator"

func init() {
	Cli()
}

func Main() {
	// delete temp files with substituted env vars when the program terminates
	defer os.RemoveAll(tempFilesDir)
	defer cleanup()

	// set the kubecontext to be used Or create it if it does not exist
	if !setKubeContext(s.Settings.KubeContext) {
		if r, msg := createContext(); !r {
			log.Fatal(msg)
		}
	}

	// add repos -- fails if they are not valid
	if r, msg := addHelmRepos(s.HelmRepos); !r {
		log.Fatal(msg)
	}

	if apply || dryRun || destroy {
		// add/validate namespaces
		if !noNs {
			addNamespaces(s.Namespaces)
		}
	}

	if !skipValidation {
		// validate charts-versions exist in defined repos
		if r, msg := validateReleaseCharts(s.Apps); !r {
			log.Fatal(msg)
		}
	} else {
		log.Info("Skipping charts' validation.")
	}

	log.Info("Preparing plan...")
	if destroy {
		log.Info("--destroy is enabled. Your releases will be deleted!")
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
}

// cleanup deletes the k8s certificates and keys files
// It also deletes any Tiller TLS certs and keys
// and secret files
func cleanup() {
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
