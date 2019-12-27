package app

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// addNamespaces creates a set of namespaces in your k8s cluster.
// If a namespace with the same name exists, it will skip it.
// If --ns-override flag is used, it only creates the provided namespace in that flag
func addNamespaces(namespaces map[string]namespace) {
	if nsOverride == "" {
		for nsName, ns := range namespaces {
			createNamespace(nsName)
			labelNamespace(nsName, ns.Labels)
			annotateNamespace(nsName, ns.Annotations)
			setLimits(nsName, ns.Limits)
		}
	} else {
		createNamespace(nsOverride)
		s.overrideAppsNamespace(nsOverride)
	}
}

// kubectl prepares a kubectl command to be executed
func kubectl(args []string, desc string) command {
	return command{
		Cmd:         "kubectl",
		Args:        args,
		Description: desc,
	}
}

// createNamespace creates a namespace in the k8s cluster
func createNamespace(ns string) {
	checkCmd := kubectl([]string{"get", "namespace", ns}, "Looking for namespace [ "+ns+" ]")
	checkExitCode, _, _ := checkCmd.exec(debug, verbose)
	if checkExitCode == 0 {
		log.Verbose("Namespace [ " + ns + " ] exists")
		return
	}
	cmd := kubectl([]string{"create", "namespace", ns}, "Creating namespace [ "+ns+" ]")
	exitCode, err, _ := cmd.exec(debug, verbose)
	if exitCode == 0 {
		log.Info("Namespace [ " + ns + " ] created")
	} else {
		log.Fatal("Failed creating namespace [ " + ns + " ] with error: " + err)
	}
}

// labelNamespace labels a namespace with provided labels
func labelNamespace(ns string, labels map[string]string) {
	if len(labels) == 0 {
		return
	}

	var labelSlice []string
	for k, v := range labels {
		labelSlice = append(labelSlice, k+"="+v)
	}

	args := []string{"label", "--overwrite", "namespace/" + ns}
	args = append(args, labelSlice...)

	cmd := kubectl(args, "Labeling namespace [ "+ns+" ]")

	exitCode, errMsg, _ := cmd.exec(debug, verbose)
	if exitCode != 0 && verbose {
		log.Warning(fmt.Sprintf("Could not label namespace [ %s with %v ]. Error message: %s", ns, labelSlice, errMsg))
	}
}

// annotateNamespace annotates a namespace with provided annotations
func annotateNamespace(ns string, annotations map[string]string) {
	if len(annotations) == 0 {
		return
	}

	var annotationSlice []string
	for k, v := range annotations {
		annotationSlice = append(annotationSlice, k+"="+v)
	}
	args := []string{"annotate", "--overwrite", "namespace/" + ns}
	args = append(args, annotationSlice...)
	cmd := kubectl(args, "Annotating namespace [ "+ns+" ]")

	exitCode, errMsg, _ := cmd.exec(debug, verbose)
	if exitCode != 0 && verbose {
		log.Info(fmt.Sprintf("Could not annotate namespace [ %s with %v ]. Error message: %s", ns, annotationSlice, errMsg))
	}
}

// setLimits creates a LimitRange resource in the provided Namespace
func setLimits(ns string, lims limits) {

	if len(lims) == 0 {
		return
	}

	definition := `
---
apiVersion: v1
kind: LimitRange
metadata:
  name: limit-range
spec:
  limits:
`
	d, err := yaml.Marshal(&lims)
	if err != nil {
		log.Fatal(err.Error())
	}

	definition = definition + Indent(string(d), strings.Repeat(" ", 4))

	if err := ioutil.WriteFile("temp-LimitRange.yaml", []byte(definition), 0666); err != nil {
		log.Fatal(err.Error())
	}

	cmd := kubectl([]string{"apply", "-f", "temp-LimitRange.yaml", "-n", ns}, "Creating LimitRange in namespace [ "+ns+" ]")
	exitCode, e, _ := cmd.exec(debug, verbose)

	if exitCode != 0 {
		log.Fatal("Failed to create LimitRange in namespace [ " + ns + " ] with error: " + e)
	}

	deleteFile("temp-LimitRange.yaml")

}

// createContext creates a context -connecting to a k8s cluster- in kubectl config.
// It returns true if successful, false otherwise
func createContext() (bool, string) {
	if s.Settings.BearerToken && s.Settings.BearerTokenPath == "" {
		log.Info("Creating kube context with bearer token from K8S service account.")
		s.Settings.BearerTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	} else if s.Settings.BearerToken && s.Settings.BearerTokenPath != "" {
		log.Info("Creating kube context with bearer token from " + s.Settings.BearerTokenPath)
	} else if s.Settings.Password == "" || s.Settings.Username == "" || s.Settings.ClusterURI == "" {
		return false, "missing information to create context [ " + s.Settings.KubeContext + " ] " +
			"you are either missing PASSWORD, USERNAME or CLUSTERURI in the Settings section of your desired state file."
	} else if !s.Settings.BearerToken && (s.Certificates == nil || s.Certificates["caCrt"] == "" || s.Certificates["caKey"] == "") {
		return false, "missing information to create context [ " + s.Settings.KubeContext + " ] " +
			"you are either missing caCrt or caKey or both in the Certifications section of your desired state file."
	} else if s.Settings.BearerToken && (s.Certificates == nil || s.Certificates["caCrt"] == "") {
		return false, "missing information to create context [ " + s.Settings.KubeContext + " ] " +
			"caCrt is missing in the Certifications section of your desired state file."
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

	// bearer token
	tokenPath := "bearer.token"
	if s.Settings.BearerToken && s.Settings.BearerTokenPath != "" {
		downloadFile(s.Settings.BearerTokenPath, tokenPath)
	}

	// connecting to the cluster
	setCredentialsCmdArgs := []string{}
	if s.Settings.BearerToken {
		token := readFile(tokenPath)
		if s.Settings.Username == "" {
			s.Settings.Username = "helmsman"
		}
		setCredentialsCmdArgs = append(setCredentialsCmdArgs, "config", "set-credentials", s.Settings.Username, "--token="+token)
	} else {
		setCredentialsCmdArgs = append(setCredentialsCmdArgs, "config", "set-credentials", s.Settings.Username, "--username="+s.Settings.Username,
			"--password="+s.Settings.Password, "--client-key="+caKey)
		if caClient != "" {
			setCredentialsCmdArgs = append(setCredentialsCmdArgs, "--client-certificate="+caClient)
		}
	}
	cmd := kubectl(setCredentialsCmdArgs, "Creating kubectl context - setting credentials")

	if exitCode, err, _ := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "failed to create context [ " + s.Settings.KubeContext + " ]:  " + err
	}

	cmd = kubectl([]string{"config", "set-cluster", s.Settings.KubeContext, "--server=" + s.Settings.ClusterURI, "--certificate-authority=" + caCrt}, "Creating kubectl context - setting cluster")

	if exitCode, err, _ := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "failed to create context [ " + s.Settings.KubeContext + " ]: " + err
	}

	cmd = kubectl([]string{"config", "set-context", s.Settings.KubeContext, "--cluster=" + s.Settings.KubeContext, "--user=" + s.Settings.Username}, "Creating kubectl context - setting context")

	if exitCode, err, _ := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "failed to create context [ " + s.Settings.KubeContext + " ]: " + err
	}

	if setKubeContext(s.Settings.KubeContext) {
		return true, ""
	}

	return false, "something went wrong while setting the kube context to the newly created one."
}

// setKubeContext sets your kubectl context to the one specified in the desired state file.
// It returns false if it fails to set the context. This means the context does not exist.
func setKubeContext(context string) bool {
	if context == "" {
		return getKubeContext()
	}

	cmd := kubectl([]string{"config", "use-context", context}, "Setting kube context to [ "+context+" ]")

	exitCode, _, _ := cmd.exec(debug, verbose)

	if exitCode != 0 {
		log.Info("Kubectl context [ " + context + " ] does not exist. Attempting to create it...")
		return false
	}

	return true
}

// getKubeContext gets your kubectl context.
// It returns false if no context is set.
func getKubeContext() bool {
	cmd := kubectl([]string{"config", "current-context"}, "Getting kubectl context")

	exitCode, result, _ := cmd.exec(debug, verbose)

	if exitCode != 0 || result == "" {
		log.Info("Kubectl context is not set")
		return false
	}

	return true
}

// labelResource applies Helmsman specific labels to Helm's state resources (secrets/configmaps)
func labelResource(r *release) {
	if r.Enabled {
		storageBackend := s.Settings.StorageBackend

		cmd := kubectl([]string{"label", storageBackend, "-n", r.Namespace, "-l", "owner=helm,name=" + r.Name, "MANAGED-BY=HELMSMAN", "NAMESPACE=" + r.Namespace, "HELMSMAN_CONTEXT=" + s.Context, "--overwrite"}, "Applying Helmsman labels to [ "+r.Name+" ] release")

		exitCode, err, _ := cmd.exec(debug, verbose)

		if exitCode != 0 {
			log.Fatal(err)
		}
	}
}

// getReleaseContext extracts the Helmsman release context from the helm storage driver objects (secret or configmap) labels
func getReleaseContext(releaseName string, namespace string) string {
	storageBackend := s.Settings.StorageBackend
	// kubectl get secrets -n test1 -l MANAGED-BY=HELMSMAN -o=jsonpath='{.items[0].metadata.labels.HELMSMAN_CONTEXT}'
	// kubectl get secret sh.helm.release.v1.argo.v1  -n test1  -o=jsonpath='{.metadata.labels.HELMSMAN_CONTEXT}'
	// kubectl get secret -l owner=helm,name=argo -n test1 -o=jsonpath='{.items[-1].metadata.labels.HELMSMAN_CONTEXT}'
	cmd := kubectl([]string{"get", storageBackend, "-n", namespace, "-l", "owner=helm", "-l", "name=" + releaseName, "-o", "jsonpath='{.items[-1].metadata.labels.HELMSMAN_CONTEXT}'"}, "Getting Helmsman context for [ "+releaseName+" ] release")

	exitCode, out, _ := cmd.exec(debug, verbose)
	if exitCode != 0 {
		log.Fatal(out)
	}
	return strings.Trim(out, `"' `)
}

// getKubectlClientVersion returns kubectl client version
func getKubectlClientVersion() string {
	cmd := kubectl([]string{"version", "--client", "--short"}, "Checking kubectl version")

	exitCode, result, _ := cmd.exec(debug, false)
	if exitCode != 0 {
		log.Fatal("While checking kubectl version: " + result)
	}
	return result
}
