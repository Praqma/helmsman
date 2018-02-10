package main

import (
	"log"
	"strings"

	"github.com/Praqma/helmsman/aws"
	"github.com/Praqma/helmsman/gcs"
)

var s state
var debug bool
var file string
var apply bool
var help bool
var v bool
var version = "master"

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
// It returns false if it fails to set the context. This means the context does not exist.
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
		// check if repo is in GCS, then perform GCS auth -- needed for private GCS helm repos
		// failed auth would not throw an error here, as it is possible that the repo is public and does not need authentication
		if strings.HasPrefix(url, "gs://") {
			gcs.Auth()
		}
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

	if s.Settings["password"] == "" || s.Settings["username"] == "" || s.Settings["clusterURI"] == "" {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ] " +
			"as you did not specify enough information in the Settings section of your desired state file."
	} else if s.Certificates == nil || s.Certificates["caCrt"] == "" || s.Certificates["caKey"] == "" {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ] " +
			"as you did not provide Certifications to use in your desired state file."
	}

	// set certs locations (relative filepath, GCS bucket, AWS bucket)
	caCrt := s.Certificates["caCrt"]
	caKey := s.Certificates["caKey"]
	caClient := s.Certificates["caClient"]

	// download certs and keys
	// GCS bucket+file format should be: gs://bucket-name/dir.../filename.ext
	// S3 bucket+file format should be: s3://bucket-name/dir.../filename.ext

	// CA cert
	if caCrt != "" {

		tmp := getBucketElements(caCrt)
		if strings.HasPrefix(caCrt, "s3") {

			aws.ReadFile(tmp["bucketName"], tmp["filePath"], "ca.crt")
			caCrt = "ca.crt"

		} else if strings.HasPrefix(caCrt, "gs") {

			gcs.ReadFile(tmp["bucketName"], tmp["filePath"], "ca.crt")
			caCrt = "ca.crt"

		} else {
			log.Println("INFO: CA certificate will be used from local file system.")
		}

	}

	// CA key
	if caKey != "" {

		tmp := getBucketElements(caKey)
		if strings.HasPrefix(caKey, "s3") {

			aws.ReadFile(tmp["bucketName"], tmp["filePath"], "ca.key")
			caKey = "ca.key"

		} else if strings.HasPrefix(caKey, "gs") {

			gcs.ReadFile(tmp["bucketName"], tmp["filePath"], "ca.key")
			caKey = "ca.key"

		} else {
			log.Println("INFO: CA key will be used from local file system.")
		}
	}

	// client certificate
	if caClient != "" {

		tmp := getBucketElements(caClient)
		if strings.HasPrefix(caClient, "s3") {

			aws.ReadFile(tmp["bucketName"], tmp["filePath"], "client.crt")
			caClient = "client.crt"

		} else if strings.HasPrefix(caClient, "gs") {

			gcs.ReadFile(tmp["bucketName"], tmp["filePath"], "client.crt")
			caClient = "client.crt"

		} else {
			log.Println("INFO: CA client key will be used from local file system.")
		}

	}

	// connecting to the cluster
	setCredentialsCmd := "kubectl config set-credentials " + s.Settings["username"] + " --username=" + s.Settings["username"] +
		" --password=" + s.Settings["password"] + " --client-key=" + caKey
	if caClient != "" {
		setCredentialsCmd = setCredentialsCmd + " --client-certificate=" + caClient
	}
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", setCredentialsCmd},
		Description: "creating kubectl context - setting credentials.",
	}

	if exitCode, _ := cmd.exec(debug); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ]. "
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-cluster " + s.Settings["kubeContext"] + " --server=" + s.Settings["clusterURI"] +
			" --certificate-authority=" + caCrt},
		Description: "creating kubectl context - setting cluster.",
	}

	if exitCode, _ := cmd.exec(debug); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ]."
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-context " + s.Settings["kubeContext"] + " --cluster=" + s.Settings["kubeContext"] +
			" --user=" + s.Settings["username"]},
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

// getBucketElements returns a map containing the bucket name and the file path inside the bucket
// this func works for S3 and GCS bucket links of the format:
// s3 or gs://bucketname/dir.../file.ext
func getBucketElements(link string) map[string]string {

	tmp := strings.SplitAfterN(link, "//", 2)[1]
	m := make(map[string]string)
	m["bucketName"] = strings.SplitN(tmp, "/", 2)[0]
	m["filePath"] = strings.SplitN(tmp, "/", 2)[1]
	return m
}
