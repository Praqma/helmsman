package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/Praqma/helmsman/gcs"
)

var s state
var debug bool
var file string
var apply bool
var help bool
var v bool
var verbose bool
var nsOverride string
var checkCleanup bool
var skipValidation bool
var version = "v1.2.0-rc"

func main() {

	// set the kubecontext to be used Or create it if it does not exist
	if !setKubeContext(s.Settings["kubeContext"]) {
		if r, msg := createContext(); !r {
			log.Fatal(msg)
		}
		checkCleanup = true
	}

	// add/validate namespaces
	addNamespaces(s.Namespaces)

	if r, msg := initHelm(); !r {
		log.Fatal(msg)
	}

	// check if helm Tiller is ready
	for k, v := range s.Namespaces {
		if v.InstallTiller {
			waitForTiller(k)
		}
	}

	waitForTiller("kube-system")

	if verbose {
		logVersions()
	}

	// add repos -- fails if they are not valid
	if r, msg := addHelmRepos(s.HelmRepos); !r {
		log.Fatal(msg)
	}

	if !skipValidation {
		// validate charts-versions exist in defined repos
		if r, msg := validateReleaseCharts(s.Apps); !r {
			log.Fatal(msg)
		}
	} else {
		log.Println("INFO: charts validation is skipped.")
	}

	log.Println("INFO: checking what I need to do for your charts ... ")
	p := makePlan(&s)

	p.sortPlan()
	p.printPlan()

	if apply {
		p.execPlan()
	}

	if checkCleanup {
		cleanup()
	}

	log.Println("INFO: completed successfully!")
}

// setKubeContext sets your kubectl context to the one specified in the desired state file.
// It returns false if it fails to set the context. This means the context does not exist.
func setKubeContext(context string) bool {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl config use-context " + context},
		Description: "setting kubectl context to [ " + context + " ]",
	}

	exitCode, _ := cmd.exec(debug, verbose)

	if exitCode != 0 {
		log.Println("INFO: KubeContext: " + context + " does not exist. I will try to create it.")
		return false
	}

	return true
}

// deployTiller deploys Helm's Tiller in a specific namespace with a serviceAccount
// If serviceAccount is not provided (empty string), the defaultServiceAccount is used.
// If no defaultServiceAccount is provided, Tiller is deployed with the namespace default service account
// If no namespace is provided, Tiller is deployed to kube-system
func deployTiller(namespace string, serviceAccount string, defaultServiceAccount string) (bool, string) {
	sa := ""
	if serviceAccount != "" {
		if ok, err := validateServiceAccount(serviceAccount, namespace); ok {
			sa = "--service-account " + serviceAccount
		} else {
			return false, "ERROR: while deploying Helm Tiller in namespace [" + namespace + "]: " + err
		}
	} else if defaultServiceAccount != "" {
		sa = "--service-account " + defaultServiceAccount
	}

	if namespace == "" {
		namespace = "kube-system"
	}
	tillerNameSpace := " --tiller-namespace " + namespace

	tls := ""
	if tillerTLSEnabled(namespace) {
		tillerCert := downloadFile(s.Namespaces[namespace].TillerCert, namespace+"-tiller.cert")
		tillerKey := downloadFile(s.Namespaces[namespace].TillerKey, namespace+"-tiller.key")
		caCert := downloadFile(s.Namespaces[namespace].CaCert, namespace+"-ca.cert")
		// client cert and key
		downloadFile(s.Namespaces[namespace].ClientCert, namespace+"-client.cert")
		downloadFile(s.Namespaces[namespace].ClientKey, namespace+"-client.key")
		tls = " --tiller-tls --tiller-tls-cert " + tillerCert + " --tiller-tls-key " + tillerKey + " --tiller-tls-verify --tls-ca-cert " + caCert
	}

	storageBackend := ""
	if v, ok := s.Settings["storageBackend"]; ok && v == "secret" {
		storageBackend = " --override 'spec.template.spec.containers[0].command'='{/tiller,--storage=secret}'"
	}
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm init --upgrade " + sa + tillerNameSpace + tls + storageBackend},
		Description: "initializing helm on the current context and upgrading Tiller on namespace [ " + namespace + " ].",
	}

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "ERROR: while deploying Helm Tiller in namespace [" + namespace + "]: " + err
	}
	return true, ""
}

// initHelm initializes helm on a k8s cluster and deploys Tiller in one or more namespaces
func initHelm() (bool, string) {

	defaultSA := ""
	if value, ok := s.Settings["serviceAccount"]; ok {
		if ok, err := validateServiceAccount(value, "kube-system"); ok {
			defaultSA = value
		} else {
			return false, "ERROR: while validating service account: " + err
		}
	}

	log.Println("INFO: deploying shared Tiller on namespace [ kube-system ].")
	if v, ok := s.Namespaces["kube-system"]; ok {
		if ok, err := deployTiller("kube-system", v.TillerServiceAccount, defaultSA); !ok {
			return false, err
		}
	} else {
		if ok, err := deployTiller("kube-system", "", defaultSA); !ok {
			return false, err
		}
	}

	for k, ns := range s.Namespaces {
		if ns.InstallTiller && k != "kube-system" {
			log.Println("INFO: deploying Tiller on namespace [ " + k + " ].")
			if ok, err := deployTiller(k, ns.TillerServiceAccount, defaultSA); !ok {
				return false, err
			}
		}
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

		if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
			return false, "ERROR: while adding repo [" + repoName + "]: " + err
		}

	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm repo update "},
		Description: "updating helm repos",
	}

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "ERROR: while updating helm repos : " + err
	}

	return true, ""
}

// validateReleaseCharts validates if the charts defined in a release are valid.
// Valid charts are the ones that can be found in the defined repos.
// This function uses Helm search to verify if the chart can be found or not.
func validateReleaseCharts(apps map[string]*release) (bool, string) {

	for app, r := range apps {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm search " + r.Chart + " --version " + r.Version + " -l"},
			Description: "validating if chart " + r.Chart + "-" + r.Version + " is available in the defined repos.",
		}

		if exitCode, result := cmd.exec(debug, verbose); exitCode != 0 || strings.Contains(result, "No results found") {
			return false, "ERROR: chart " + r.Chart + "-" + r.Version + " is specified for " +
				"app [" + app + "] but is not found in the defined repos."
		}
	}
	return true, ""
}

// addNamespaces creates a set of namespaces in your k8s cluster.
// If a namespace with the same name exsts, it will skip it.
// If --ns-override flag is used, it only creates the provided namespace in that flag
func addNamespaces(namespaces map[string]namespace) {
	if nsOverride == "" {
		for ns := range namespaces {
			createNamespace(ns)
		}
	} else {
		createNamespace(nsOverride)
		overrideAppsNamespace(nsOverride)
	}
}

// overrideAppsNamespace replaces all apps namespaces with one specific namespace
func overrideAppsNamespace(newNs string) {
	log.Println("INFO: overriding apps namespaces with [ " + newNs + " ] ...")
	for _, r := range s.Apps {
		overrideNamespace(r, newNs)
	}
}

// createNamespace creates a namespace in the k8s cluster
func createNamespace(ns string) {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl create namespace " + ns},
		Description: "creating namespace  " + ns,
	}

	if exitCode, _ := cmd.exec(debug, verbose); exitCode != 0 {
		log.Println("WARN: I could not create namespace [" +
			ns + " ]. It already exists. I am skipping this.")
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

		caCrt = downloadFile(caCrt, "ca.crt")

	}

	// CA key
	if caKey != "" {
		caKey = downloadFile(caKey, "ca.key")

	}

	// client certificate
	if caClient != "" {

		caClient = downloadFile(caClient, "client.crt")

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

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ]:  " + err
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-cluster " + s.Settings["kubeContext"] + " --server=" + s.Settings["clusterURI"] +
			" --certificate-authority=" + caCrt},
		Description: "creating kubectl context - setting cluster.",
	}

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ]: " + err
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-context " + s.Settings["kubeContext"] + " --cluster=" + s.Settings["kubeContext"] +
			" --user=" + s.Settings["username"]},
		Description: "creating kubectl context - setting context.",
	}

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ]: " + err
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

// waitForTiller keeps checking if the helm Tiller is ready or not by executing helm list and checking its error (if any)
// waits for 5 seconds before each new attempt and eventually terminates after 10 failed attempts.
func waitForTiller(namespace string) {

	attempt := 0

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm list --tiller-namespace " + namespace + getNSTLSFlags(namespace)},
		Description: "checking if helm Tiller is ready in namespace [ " + namespace + " ].",
	}

	exitCode, err := cmd.exec(debug, verbose)

	for attempt < 10 {
		if exitCode == 0 {
			return
		} else if strings.Contains(err, "could not find a ready tiller pod") {
			log.Println("INFO: waiting for helm Tiller to be ready ...")
			time.Sleep(5 * time.Second)
			exitCode, err = cmd.exec(debug, verbose)
		} else {
			log.Fatal("ERROR: while waiting for helm Tiller to be ready in namespace [ " + namespace + " ] : " + err)
		}
		attempt = attempt + 1
	}
	log.Fatal("ERROR: timeout reached while waiting for helm Tiller to be ready in namespace [ " + namespace + " ]. Aborting!")
}

// cleanup deletes the k8s certificates and keys files
// It also deletes any Tiller TLS certs and keys
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
}
