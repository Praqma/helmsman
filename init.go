package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"github.com/joho/godotenv"
	"github.com/logrusorgru/aurora"
)

// colorizer
var style aurora.Aurora

const (
	banner = " _          _ \n" +
		"| |        | | \n" +
		"| |__   ___| |_ __ ___  ___ _ __ ___   __ _ _ __\n" +
		"| '_ \\ / _ \\ | '_ ` _ \\/ __| '_ ` _ \\ / _` | '_ \\ \n" +
		"| | | |  __/ | | | | | \\__ \\ | | | | | (_| | | | | \n" +
		"|_| |_|\\___|_|_| |_| |_|___/_| |_| |_|\\__,_|_| |_|"
	slogan = "A Helm-Charts-as-Code tool.\n\n"
)

func printUsage() {
	log.Println(banner + "\n")
	log.Println("Helmsman version: " + appVersion)
	log.Println("Helmsman is a Helm Charts as Code tool which allows you to automate the deployment/management of your Helm charts.")
	log.Println()
	log.Println("Usage: helmsman [options]")
	flag.PrintDefaults()
}

// init is executed after all package vars are initialized [before the main() func in this case].
// It checks if Helm and Kubectl exist and configures: the connection to the k8s cluster, helm repos, namespaces, etc.
func init() {
	//parsing command line flags
	flag.Var(&files, "f", "desired state file name(s), may be supplied more than once to merge state files")
	flag.Var(&envFiles, "e", "file(s) to load environment variables from (default .env), may be supplied more than once")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file to use for CLI requests")
	flag.BoolVar(&apply, "apply", false, "apply the plan directly")
	flag.BoolVar(&debug, "debug", false, "show the execution logs")
	flag.BoolVar(&dryRun, "dry-run", false, "apply the dry-run option for helm commands.")
	flag.Var(&target, "target", "limit execution to specific app.")
	flag.BoolVar(&destroy, "destroy", false, "delete all deployed releases. Purge delete is used if the purge option is set to true for the releases.")
	flag.BoolVar(&v, "v", false, "show the version")
	flag.BoolVar(&verbose, "verbose", false, "show verbose execution logs")
	flag.BoolVar(&noBanner, "no-banner", false, "don't show the banner")
	flag.BoolVar(&noColors, "no-color", false, "don't use colors")
	flag.BoolVar(&noFancy, "no-fancy", false, "don't display the banner and don't use colors")
	flag.BoolVar(&noNs, "no-ns", false, "don't create namespaces")
	flag.StringVar(&nsOverride, "ns-override", "", "override defined namespaces with this one")
	flag.BoolVar(&skipValidation, "skip-validation", false, "skip desired state validation")
	flag.BoolVar(&applyLabels, "apply-labels", false, "apply Helmsman labels to Helm state for all defined apps.")
	flag.BoolVar(&keepUntrackedReleases, "keep-untracked-releases", false, "keep releases that are managed by Helmsman and are no longer tracked in your desired state.")
	flag.BoolVar(&showDiff, "show-diff", false, "show helm diff results. Can expose sensitive information.")
	flag.BoolVar(&suppressDiffSecrets, "suppress-diff-secrets", false, "don't show secrets in helm diff output.")
	flag.BoolVar(&noEnvSubst, "no-env-subst", false, "turn off environment substitution globally")

	log.SetOutput(os.Stdout)

	flag.Usage = printUsage
	flag.Parse()

	if noFancy {
		noColors = true
		noBanner = true
	}

	style = aurora.NewAurora(!noColors)

	if !noBanner {
		fmt.Println(banner + " version: " + appVersion + "\n" + slogan)
	}

	if dryRun && apply {
		logError("ERROR: --apply and --dry-run can't be used together.")
	}

	if destroy && apply {
		logError("ERROR: --destroy and --apply can't be used together.")
	}

	helmVersion = strings.TrimSpace(strings.SplitN(getHelmClientVersion(), ": ", 2)[1])
	kubectlVersion = strings.TrimSpace(strings.SplitN(getKubectlClientVersion(), ": ", 2)[1])

	if verbose {
		logVersions()
	}

	if v {
		fmt.Println("Helmsman version: " + appVersion)
		os.Exit(0)
	}

	if len(files) == 0 {
		log.Println("INFO: No desired state files provided.")
		os.Exit(0)
	}

	if kubeconfig != "" {
		os.Setenv("KUBECONFIG", kubeconfig)
	}

	if !toolExists("kubectl") {
		logError("ERROR: kubectl is not installed/configured correctly. Aborting!")
	}

	if !toolExists("helm") {
		logError("ERROR: helm is not installed/configured correctly. Aborting!")
	}

	if !helmPluginExists("diff") {
		logError("ERROR: helm diff plugin is not installed/configured correctly. Aborting!")
	}

	// read the env file
	if len(envFiles) == 0 {
		if _, err := os.Stat(".env"); err == nil {
			err = godotenv.Load()
			if err != nil {
				logError("Error loading .env file")
			}
		}
	}

	for _, e := range envFiles {
		err := godotenv.Load(e)
		if err != nil {
			logError("Error loading " + e + " env file")
		}
	}

	// wipe & create a temporary directory
	os.RemoveAll(tempFilesDir)
	_ = os.MkdirAll(tempFilesDir, 0755)

	// read the TOML/YAML desired state file
	var fileState state
	for _, f := range files {
		result, msg := fromFile(f, &fileState)
		if result {
			log.Printf(msg)
		} else {
			logError(msg)
		}

		// Merge Apps that already existed in the state
		for appName, app := range fileState.Apps {
			if _, ok := s.Apps[appName]; ok {
				if err := mergo.Merge(s.Apps[appName], app, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
					logError("Failed to merge " + appName + " from desired state file" + f)
				}
			}
		}

		// Merge the remaining Apps
		if err := mergo.Merge(&s.Apps, &fileState.Apps); err != nil {
			logError("Failed to merge desired state file" + f)
		}
		// All the apps are already merged, make fileState.Apps empty to avoid conflicts in the final merge
		fileState.Apps = make(map[string]*release)

		if err := mergo.Merge(&s, &fileState, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
			logError("Failed to merge desired state file" + f)
		}
	}

	if debug {
		s.print()
	}

	if !skipValidation {
		// validate the desired state content
		if len(files) > 0 {
			if result, msg := s.validate(); !result { // syntax validation
				logError(msg)
			}
		}
	} else {
		log.Println("INFO: desired state validation is skipped.")
	}

	if applyLabels {
		for _, r := range s.Apps {
			labelResource(r)
		}
	}

	if len(target) > 0 {
		targetMap = map[string]bool{}
		for _, v := range target {
			targetMap[v] = true
		}
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

	exitCode, _ := cmd.exec(debug, false)

	if exitCode != 0 {
		return false
	}

	return true
}

// helmPluginExists returns true if the plugin is present in the environment and false otherwise.
// It takes as input the plugin's name to check if it is recognizable or not. e.g. diff
func helmPluginExists(plugin string) bool {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm plugin list"},
		Description: "validating that " + plugin + " is installed.",
	}

	exitCode, result := cmd.exec(debug, false)

	if exitCode != 0 {
		return false
	}

	return strings.Contains(result, plugin)
}
