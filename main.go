package main

import (
	"log"
)

var s state
var debug bool
var file string
var apply bool
var help bool

func main() {

	// set the kubecontext to be used Or create it if it does not exist
	if !setKubeContext(s.Settings["kubeContext"]) {
		if r, msg := createContext(); !r {
			log.Fatal(msg)
		}
	}

	if r, msg := initHelm(); !r {
		log.Fatal(msg)
	}

	// add repos -- fails if they are not valid
	if r, msg := addHelmRepos(s.HelmRepos); !r {
		log.Fatal(msg)
	}

	// validate charts-versions exist in supllied repos
	if r, msg := validateReleaseCharts(s.Apps); !r {
		log.Fatal(msg)
	}

	// add/validate namespaces
	addNamespaces(s.Namespaces)

	p := makePlan(&s)

	if !apply {
		p.printPlan()
	} else {
		p.execPlan()
	}

}

// setKubeContext sets your kubectl context to the one specified in the desired state file.
// It returns false if it fails to set the context. This means the context deos not exist.
func setKubeContext(context string) bool {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl config use-context " + context},
		Description: "setting kubectl context to [ " + context + " ]",
	}

	exitCode, _ := cmd.exec(debug)

	if exitCode != 0 {
		log.Println("INFO: KubeContext: " + context + " does not exist. I will try to create it.")
		return false
	}

	return true
}

// initHelm initialize helm on a k8s cluster
func initHelm() (bool, string) {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm init --upgrade"},
		Description: "initializing helm on the current context and upgrade Tiller.",
	}

	if exitCode, _ := cmd.exec(debug); exitCode != 0 {
		return false, "ERROR: there was a problem while initializing helm. "
	}
	return true, ""
}

// addHelmRepos adds repositories to Helm if they don't exist already.
// Helm does not mind if a repo with the same name exists. It treats it as an update.
func addHelmRepos(repos map[string]string) (bool, string) {

	for repoName, url := range repos {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm repo add " + repoName + " " + url},
			Description: "adding repo " + repoName,
		}

		if exitCode, _ := cmd.exec(debug); exitCode != 0 {
			return false, "ERROR: there was a problem while adding repo [" + repoName + "]."
		}

	}

	return true, ""
}

// validateReleaseCharts validates if the charts defined in a release are valid.
// Valid charts are the ones that can be found in the defined repos.
// This function uses Helm search to verify if the chart can be found or not.
func validateReleaseCharts(apps map[string]release) (bool, string) {

	for app, r := range apps {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm search " + r.Chart + " --version " + r.Version},
			Description: "validating chart " + r.Chart + "-" + r.Version + " is available in the used repos.",
		}

		if exitCode, _ := cmd.exec(debug); exitCode != 0 {
			return false, "ERROR: chart " + r.Chart + "-" + r.Version + " is specified for " +
				"app [" + app + "] but is not found in the provided repos."
		}
	}
	return true, ""
}

// addNamespaces creates a set of namespaces in your k8s cluster.
// If a namespace with the same name exsts, it will skip it.
func addNamespaces(namespaces map[string]string) {
	for _, namespace := range namespaces {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "kubectl create namespace " + namespace},
			Description: "creating namespace  " + namespace,
		}

		if exitCode, _ := cmd.exec(debug); exitCode != 0 {
			log.Println("WARN: I could not create namespace [" +
				namespace + " ]. It already exists. I am skipping this.")
		}
	}
}

// createContext creates a context -connecting to a k8s cluster- in kubectl config.
// It returns true if successful, false otherwise
func createContext() (bool, string) {

	var password string
	var ok bool

	if s.Settings["password"] == "" || s.Settings["username"] == "" || s.Settings["clusterURI"] == "" {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ] " +
			"as you did not specify enough information in the Settings section of your desired state file."
	} else if s.Certificates == nil || s.Certificates["caCrt"] == "" || s.Certificates["caKey"] == "" {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ] " +
			"as you did not provide Certifications to use in your desired state file."
	} else {
		ok, password = envVarExists(s.Settings["password"])
		if !ok {
			return false, "ERROR: env variable [ " + s.Settings["password"] + " ] does not exist in the environment."
		}
	}

	// download certs using AWS cli
	if !toolExists("aws help") {
		return false, "ERROR: aws is not installed/configured correctly. It is needed for downloading certs. Aborting!"
	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "aws s3 cp " + s.Certificates["caCrt"] + " ca.crt"},
		Description: "downloading ca.crt from S3.",
	}

	if exitCode, _ := cmd.exec(debug); exitCode != 0 {
		return false, "ERROR: failed to download caCrt."
	}

	cmd = command{
		Cmd:         "bash",
		Args:        []string{"-c", "aws s3 cp " + s.Certificates["caKey"] + " ca.key"},
		Description: "downloading ca.key from S3.",
	}

	if exitCode, _ := cmd.exec(debug); exitCode != 0 {
		return false, "ERROR: failed to download caKey."
	}

	// connecting to the cluster
	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-credentials " + s.Settings["username"] + " --username=" + s.Settings["username"] +
			" --password=" + password + " --client-key=ca.key"},
		Description: "creating kubectl context - setting credentials.",
	}

	if exitCode, _ := cmd.exec(debug); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ]. "
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-cluster " + s.Settings["kubeContext"] + " --server=" + s.Settings["clusterURI"] +
			" --certificate-authority=ca.crt"},
		Description: "creating kubectl context - setting cluster.",
	}

	if exitCode, _ := cmd.exec(debug); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ]."
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-context " + s.Settings["kubeContext"] + " --cluster=" + s.Settings["kubeContext"] +
			" --user=" + s.Settings["username"] + " --password=" + password},
		Description: "creating kubectl context - setting context.",
	}

	if exitCode, _ := cmd.exec(debug); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ]."
	}

	if setKubeContext(s.Settings["kubeContext"]) {
		return true, ""
	}

	return false, "ERROR: something went wrong while setting the kube context to the newly created one."
}
