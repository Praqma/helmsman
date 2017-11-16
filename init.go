package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

// init is executed after all package vars are initialized [before the main() func in this case].
// It checks if Helm and Kubectl exist and configures: the connection to the k8s cluster, helm repos, namespaces, etc.
func init() {
	//parsing command line flags
	flag.StringVar(&file, "f", "example.toml", "desired state file name")
	flag.BoolVar(&apply, "apply", false, "apply the plan directly")
	flag.BoolVar(&debug, "debug", false, "show the execution logs")
	flag.BoolVar(&help, "help", false, "show Helmsman help")

	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
	}

	if !toolExists("helm") {
		log.Fatal("ERROR: helm is not installed/configured correctly. Aborting!")
		os.Exit(1)
	}

	if !toolExists("kubectl") {
		log.Fatal("ERROR: kubectl is not installed/configured correctly. Aborting!")
		os.Exit(1)
	}

	// after the init() func is run, read the TOML desired state file
	result, msg := fromTOML(file, &s)
	if result {
		log.Printf(msg)
	} else {
		log.Fatal(msg)
	}

	// validate the desired state content
	if result, msg := s.validate(); !result { // syntax validation
		log.Fatal(msg)
	}

}

// toolExists returns true if the tool is present in the environment and false otherwise.
// It takes as input the tool's command to check if it is recognizable or not. e.g. helm or kubectl
func toolExists(tool string) bool {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", tool},
		Description: "validating that " + tool + " is installed.",
	}

	exitCode, _ := cmd.exec(debug)

	if exitCode != 0 {
		return false
	}

	return true
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

		exitCode, _ := cmd.exec(debug)

		if exitCode != 0 {
			log.Println("WARN: I could not create namespace [" +
				namespace + " ]. It already exists. I am skipping this.")
		}
	}
}

// validateReleaseCharts validates if the charts defined in a release are valid.
// Valid charts are the ones that can be found in the defined repos.
// This function uses Helm search to verify if the chart can be found or not.
func validateReleaseCharts(apps map[string]release) bool {

	for app, r := range apps {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm search " + r.Chart + " --version " + r.Version},
			Description: "validating chart " + r.Chart + "-" + r.Version + " is available in the used repos.",
		}

		exitCode, _ := cmd.exec(debug)

		if exitCode != 0 {
			log.Fatal("ERROR: chart "+r.Chart+"-"+r.Version+" is specified for ",
				"app ["+app+"] but is not found in the provided repos.")
			return false
		}
	}
	return true
}

// addHelmRepos adds repositories to Helm if they don't exist already.
// Helm does not mind if a repo with the same name exists. It treats it as an update.
func addHelmRepos(repos map[string]string) bool {

	for repoName, url := range repos {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm repo add " + repoName + " " + url},
			Description: "adding repo " + repoName,
		}

		exitCode, _ := cmd.exec(debug)

		if exitCode != 0 {
			log.Fatal("ERROR: there has been a problem while adding repo [" +
				repoName + "].")
			return false
		}

	}

	return true
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

// createContext creates a context -connecting to a k8s cluster- in kubectl config.
// It returns true if successful, false otherwise
func createContext() bool {

	var password string
	if s.Settings["password"] == "" || s.Settings["username"] == "" || s.Settings["clusterURI"] == "" {
		log.Fatal("ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ] " +
			"as you did not specify enough information in the Settings section of your desired state file.")
		return false
	} else if s.Certificates == nil || s.Certificates["caCrt"] == "" || s.Certificates["caKey"] == "" {
		log.Fatal("ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ] " +
			"as you did not provide Certifications to use in your desired state file.")
		return false
	} else {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "echo " + s.Settings["password"]},
			Description: "reading the password from env variable.",
		}

		var exitCode int
		exitCode, password = cmd.exec(debug)

		password = strings.TrimSpace(password)
		if exitCode != 0 || password == "" {
			log.Fatal("ERROR: failed to read password from env variable.")
			return false
		}
	}

	// download certs using AWS cli
	if !toolExists("aws help") {
		log.Fatal("ERROR: aws is not installed/configured correctly. It is needed for downloading certs. Aborting!")
		return false
	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "aws s3 cp " + s.Certificates["caCrt"] + " ca.crt"},
		Description: "downloading ca.crt from S3.",
	}

	exitCode, _ := cmd.exec(debug)

	if exitCode != 0 {
		log.Fatal("ERROR: failed to download caCrt.")
		return false
	}

	cmd = command{
		Cmd:         "bash",
		Args:        []string{"-c", "aws s3 cp " + s.Certificates["caKey"] + " ca.key"},
		Description: "downloading ca.key from S3.",
	}

	exitCode, _ = cmd.exec(debug)

	if exitCode != 0 {
		log.Fatal("ERROR: failed to download caKey.")
		return false
	}

	// connecting to the cluster
	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-credentials " + s.Settings["username"] + " --username=" + s.Settings["username"] +
			" --password=" + password + " --client-key=ca.key"},
		Description: "creating kubectl context - setting credentials.",
	}

	exitCode, _ = cmd.exec(debug)

	if exitCode != 0 {
		log.Fatal("ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ].")
		return false
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-cluster " + s.Settings["kubeContext"] + " --server=" + s.Settings["clusterURI"] +
			" --certificate-authority=ca.crt"},
		Description: "creating kubectl context - setting cluster.",
	}

	exitCode, _ = cmd.exec(debug)

	if exitCode != 0 {
		log.Fatal("ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ].")
		return false
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-context " + s.Settings["kubeContext"] + " --cluster=" + s.Settings["kubeContext"] +
			" --user=" + s.Settings["username"] + " --password=" + password},
		Description: "creating kubectl context - setting context.",
	}

	exitCode, _ = cmd.exec(debug)

	if exitCode != 0 {
		log.Fatal("ERROR: failed to create context [ " + s.Settings["kubeContext"] + " ].")
		return false
	}

	return setKubeContext(s.Settings["kubeContext"])
}
