package main

import (
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

// validateServiceAccount checks if k8s service account exists in a given namespace
// if the provided namespace is empty, it checks in the "default" namespace
func validateServiceAccount(sa string, namespace string) (bool, string) {
	if namespace == "" {
		namespace = "default"
	}
	ns := " -n " + namespace

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl get serviceaccount " + sa + ns},
		Description: "validating if serviceaccount [ " + sa + " ] exists in namespace [ " + namespace + " ].",
	}

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, err
	}
	return true, ""
}

// createRBAC creates a k8s service account and bind it to a (Cluster)Role
// If sharedTiller is true , it binds the service account to cluster-admin role. Otherwise,
// It binds it to a new role called "helmsman-tiller"
func createRBAC(sa string, namespace string, sharedTiller bool) (bool, string) {
	var ok bool
	var err string
	if ok, err = createServiceAccount(sa, namespace); ok {
		if sharedTiller {
			if ok, err = createRoleBinding("cluster-admin", sa, namespace); ok {
				return true, ""
			}
			return false, err
		}
		if ok, err = createRole(namespace); ok {
			if ok, err = createRoleBinding("helmsman-tiller", sa, namespace); ok {
				return true, ""
			}
			return false, err
		}

		return false, err
	}
	return false, err
}

// addNamespaces creates a set of namespaces in your k8s cluster.
// If a namespace with the same name exists, it will skip it.
// If --ns-override flag is used, it only creates the provided namespace in that flag
func addNamespaces(namespaces map[string]namespace) {
	if nsOverride == "" {
		for nsName, ns := range namespaces {
			createNamespace(nsName)
			labelNamespace(nsName, ns.Labels)
			setLimits(nsName, ns.Limits)
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
		log.Println("WARN: I could not create namespace [ " +
			ns + " ]. It already exists. I am skipping this.")
	}
}

// labelNamespace labels a namespace with provided labels
func labelNamespace(ns string, labels map[string]string) {
	for k, v := range labels {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "kubectl label --overwrite namespace/" + ns + " " + k + "=" + v},
			Description: "labeling namespace  " + ns,
		}

		if exitCode, _ := cmd.exec(debug, verbose); exitCode != 0 {
			log.Println("WARN: I could not label namespace [ " + ns + " with " + k + "=" + v +
				" ]. It already exists. I am skipping this.")
		}
	}
}

// setLimits creates a LimitRange resource in the provided Namespace
func setLimits(ns string, lims limits) {

	if lims == (limits{}) {
		return
	}

	definition := `---
apiVersion: v1
kind: LimitRange
metadata:
  name: limit-range
spec:
  limits:
  - type: Container
`
	d, err := yaml.Marshal(&lims)
	if err != nil {
		logError(err.Error())
	}

	definition = definition + Indent(string(d), strings.Repeat(" ", 4))

	if err := ioutil.WriteFile("temp-LimitRange.yaml", []byte(definition), 0666); err != nil {
		logError(err.Error())
	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl apply -f temp-LimitRange.yaml -n " + ns},
		Description: "creating LimitRange in namespace [ " + ns + " ]",
	}

	exitCode, e := cmd.exec(debug, verbose)

	if exitCode != 0 {
		logError("ERROR: failed to create LimitRange in namespace [ " + ns + " ]: " + e)
	}

	deleteFile("temp-LimitRange.yaml")

}

// createContext creates a context -connecting to a k8s cluster- in kubectl config.
// It returns true if successful, false otherwise
func createContext() (bool, string) {

	if s.Settings.Password == "" || s.Settings.Username == "" || s.Settings.ClusterURI == "" {
		return false, "ERROR: failed to create context [ " + s.Settings.KubeContext + " ] " +
			"as you did not specify enough information in the Settings section of your desired state file."
	} else if s.Certificates == nil || s.Certificates["caCrt"] == "" || s.Certificates["caKey"] == "" {
		return false, "ERROR: failed to create context [ " + s.Settings.KubeContext + " ] " +
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
	setCredentialsCmd := "kubectl config set-credentials " + s.Settings.Username + " --username=" + s.Settings.Username +
		" --password=" + s.Settings.Password + " --client-key=" + caKey
	if caClient != "" {
		setCredentialsCmd = setCredentialsCmd + " --client-certificate=" + caClient
	}
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", setCredentialsCmd},
		Description: "creating kubectl context - setting credentials.",
	}

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings.KubeContext + " ]:  " + err
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-cluster " + s.Settings.KubeContext + " --server=" + s.Settings.ClusterURI +
			" --certificate-authority=" + caCrt},
		Description: "creating kubectl context - setting cluster.",
	}

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings.KubeContext + " ]: " + err
	}

	cmd = command{
		Cmd: "bash",
		Args: []string{"-c", "kubectl config set-context " + s.Settings.KubeContext + " --cluster=" + s.Settings.KubeContext +
			" --user=" + s.Settings.Username},
		Description: "creating kubectl context - setting context.",
	}

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "ERROR: failed to create context [ " + s.Settings.KubeContext + " ]: " + err
	}

	if setKubeContext(s.Settings.KubeContext) {
		return true, ""
	}

	return false, "ERROR: something went wrong while setting the kube context to the newly created one."
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

// createServiceAccount creates a service account in a given namespace and associates it with a cluster-admin role
func createServiceAccount(saName string, namespace string) (bool, string) {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl create serviceaccount -n " + namespace + " " + saName},
		Description: "creating service account [ " + saName + " ] in namespace [ " + namespace + " ]",
	}

	exitCode, err := cmd.exec(debug, verbose)

	if exitCode != 0 {
		//logError("ERROR: failed to create service account " + saName + " in namespace [ " + namespace + " ]: " + err)
		return false, err
	}

	return true, ""
}

// createRoleBinding creates a role binding in a given namespace for a service account with a cluster-role/role in the cluster.
func createRoleBinding(role string, saName string, namespace string) (bool, string) {
	clusterRole := false
	resource := "rolebinding"
	if role == "cluster-admin" {
		clusterRole = true
		resource = "clusterrolebinding"
	}

	bindingOption := "--role=" + role
	if clusterRole {
		bindingOption = "--clusterrole=" + role
	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl create " + resource + " " + saName + "-binding " + bindingOption + " --serviceaccount " + namespace + ":" + saName + " -n " + namespace},
		Description: "creating " + resource + " for service account [ " + saName + " ] in namespace [ " + namespace + " ] with role: " + role,
	}

	exitCode, err := cmd.exec(debug, verbose)

	if exitCode != 0 {
		return false, err
	}

	return true, ""
}

// createRole creates a k8s Role in a given namespace
func createRole(namespace string) (bool, string) {

	// load static resource
	resource, e := Asset("data/role.yaml")
	if e != nil {
		logError(e.Error())
	}
	replaceStringInFile(resource, "temp-modified-role.yaml", map[string]string{"<<namespace>>": namespace})

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl apply -f temp-modified-role.yaml "},
		Description: "creating role [helmsman-tiller] in namespace [ " + namespace + " ]",
	}

	exitCode, err := cmd.exec(debug, verbose)

	if exitCode != 0 {
		//logError("ERROR: failed to create Tiller role in namespace [ " + namespace + " ]: " + err)
		return false, err
	}

	deleteFile("temp-modified-role.yaml")

	return true, ""
}

// labelResource applies Helmsman specific labels to Helm's state resources (secrets/configmaps)
func labelResource(r *release) {
	if r.Enabled {
		log.Println("INFO: applying Helmsman labels to [ " + r.Name + " ] in namespace [ " + r.Namespace + " ] ")
		storageBackend := "configmap"

		if s.Settings.StorageBackend == "secret" {
			storageBackend = "secret"
		}

		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "kubectl label " + storageBackend + " -n " + getDesiredTillerNamespace(r) + " -l NAME=" + r.Name + " MANAGED-BY=HELMSMAN NAMESPACE=" + r.Namespace + " TILLER_NAMESPACE=" + getDesiredTillerNamespace(r) + "  --overwrite"},
			Description: "applying labels to Helm state in [ " + getDesiredTillerNamespace(r) + " ] for " + r.Name,
		}

		exitCode, err := cmd.exec(debug, verbose)

		if exitCode != 0 {
			logError(err)
		}
	}
}

// getHelmsmanReleases returns a map of all releases that are labeled with "MANAGED-BY=HELMSMAN"
// The releases are categorized by the namespaces in which their Tiller is running
// The returned map format is: map[<Tiller namespace>:map[<releases managed by Helmsman and deployed using this Tiller>:true]]
func getHelmsmanReleases() map[string]map[string]bool {
	var lines []string
	releases := make(map[string]map[string]bool)
	storageBackend := "configmap"

	if s.Settings.StorageBackend == "secret" {
		storageBackend = "secret"
	}

	namespaces := make([]string, len(s.Namespaces))
	i := 0
	for s, v := range s.Namespaces {
		if v.InstallTiller || v.UseTiller {
			namespaces[i] = s
			i++
		}
	}
	namespaces = namespaces[0:i]
	if v, ok := s.Namespaces["kube-system"]; !ok || (ok && (v.UseTiller || v.InstallTiller)) {
		namespaces = append(namespaces, "kube-system")
	}

	for _, ns := range namespaces {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "kubectl get " + storageBackend + " -n " + ns + " -l MANAGED-BY=HELMSMAN"},
			Description: "getting helm releases which are managed by Helmsman in namespace [[ " + ns + " ]].",
		}

		exitCode, output := cmd.exec(debug, verbose)

		if exitCode != 0 {
			logError(output)
		}
		if strings.ToUpper("No resources found.") != strings.ToUpper(strings.TrimSpace(output)) {
			lines = strings.Split(output, "\n")
		}

		for i := 0; i < len(lines); i++ {
			if lines[i] == "" || strings.HasSuffix(strings.TrimSpace(lines[i]), "AGE") {
				continue
			} else {
				fields := strings.Fields(lines[i])
				if _, ok := releases[ns]; !ok {
					releases[ns] = make(map[string]bool)
				}
				releases[ns][fields[0][0:strings.LastIndex(fields[0], ".v")]] = true
			}
		}
	}

	return releases
}

// getKubectlClientVersion returns kubectl client version
func getKubectlClientVersion() string {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl version --client --short"},
		Description: "checking kubectl version ",
	}

	exitCode, result := cmd.exec(debug, false)
	if exitCode != 0 {
		logError("ERROR: while checking kubectl version: " + result)
	}
	return result
}
