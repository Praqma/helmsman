package main

import (
	"log"
)

// validateServiceAccount checks if k8s service account exists in a given namespace
// if the provided namespace is empty, it checks in the "default" namespace
// if the serviceaccount does not exist, it will be created
func validateServiceAccount(sa string, namespace string) (bool, string) {
	log.Println("INFO: validating if service account [" + sa + "] exists in namespace [" + namespace + "]")
	if namespace == "" {
		namespace = "default"
	}
	ns := " -n " + namespace

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl get serviceaccount " + sa + ns},
		Description: "validating that serviceaccount [ " + sa + " ] exists in namespace [ " + namespace + " ].",
	}

	if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
		return false, err
		// if strings.Contains(err, "NotFound") || strings.Contains(err, "not found") {
		// 	log.Println("INFO: service account [ " + sa + " ] does not exist in namespace [ " + namespace + " ] .. attempting to create it ... ")

		// 	if _, rbacErr := createRBAC(sa, namespace); rbacErr != "" {
		// 		return false, rbacErr
		// 	}
		// 	return true, ""

		// }
		// return false, err
	}
	return true, ""
}

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
	log.Println("INFO: creating service account [ " + saName + " ] in namespace [ " + namespace + " ]")
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

	log.Println("INFO: creating " + resource + " for service account [ " + saName + " ] in namespace [ " + namespace + " ] with role: " + role)

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl create " + resource + " " + saName + "-binding " + bindingOption + " --serviceaccount " + namespace + ":" + saName + " -n " + namespace},
		Description: "creating " + resource + " for [ " + saName + " ] in namespace [ " + namespace + " ]",
	}

	exitCode, err := cmd.exec(debug, verbose)

	if exitCode != 0 {
		//logError("ERROR: failed to bind service account " + saName + " in namespace [ " + namespace + " ] to role " + role + " : " + err)
		return false, err
	}

	return true, ""
}

func createRole(namespace string) (bool, string) {

	// load static resource
	resource, e := Asset("data/role.yaml")
	if e != nil {
		logError(e.Error())
	}
	replaceStringInFile(resource, "temp-modified-role.yaml", map[string]string{"<<namespace>>": namespace})

	log.Println("INFO: creating role [helmsman-tiller] in namespace [ " + namespace + " ] ")

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
