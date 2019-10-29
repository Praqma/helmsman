package main

import (
	"flag"
	"fmt"
	"github.com/apsdehal/go-logger"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"github.com/joho/godotenv"
)

var logs *logger.Logger

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
	logs.Info(banner + "\n")
	logs.Info("Helmsman version: " + appVersion)
	logs.Info("Helmsman is a Helm Charts as Code tool which allows you to automate the deployment/management of your Helm charts.")
	logs.Info("")
	logs.Info("Usage: helmsman [options]")
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
	flag.Var(&group, "group", "limit execution to specific group of apps.")
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
	flag.IntVar(&diffContext, "diff-context", -1, "number of lines of context to show around changes in helm diff output")
	flag.BoolVar(&noEnvSubst, "no-env-subst", false, "turn off environment substitution globally")
	flag.BoolVar(&noEnvValuesSubst, "no-env-values-subst", true, "turn off environment substitution in values files only")
	flag.BoolVar(&noSSMSubst, "no-ssm-subst", false, "turn off SSM parameter substitution globally")
	flag.BoolVar(&noSSMValuesSubst, "no-ssm-values-subst", true, "turn off SSM parameter substitution in values files only")
	flag.BoolVar(&updateDeps, "update-deps", false, "run 'helm dep up' for local chart")
	flag.BoolVar(&forceUpgrades, "force-upgrades", false, "use --force when upgrading helm releases. May cause resources to be recreated.")
	flag.BoolVar(&noDefaultRepos, "no-default-repos", false, "don't set default Helm repos from Google for 'stable' and 'incubator'")

	flag.Usage = printUsage
	flag.Parse()

	logger.SetDefaultFormat("%{time:2006-01-02 15:04:05} %{level}: %{message}")
	var logLevel = logger.InfoLevel
	if verbose {
		logLevel = logger.DebugLevel
	}

	if noFancy {
		noColors = true
		noBanner = true
	}

	var logColors = 1
	if noColors {
		logColors = 0
	}
	logs, _ = logger.New("logger", logColors, os.Stdout, logLevel)

	if !noBanner {
		fmt.Println(banner + " version: " + appVersion + "\n" + slogan)
	}

	if dryRun && apply {
		logError("--apply and --dry-run can't be used together.")
	}

	if destroy && apply {
		logError("--destroy and --apply can't be used together.")
	}

	if len(target) > 0 && len(group) > 0 {
		logError("--target and --group can't be used together.")
	}

	if (settings.EyamlPrivateKeyPath != "" && settings.EyamlPublicKeyPath == "") || (settings.EyamlPrivateKeyPath == "" && settings.EyamlPublicKeyPath != "") {
		logError("both EyamlPrivateKeyPath and EyamlPublicKeyPath are required")
	}

	helmVersion = strings.TrimSpace(getHelmVersion())
	kubectlVersion = strings.TrimSpace(strings.SplitN(getKubectlClientVersion(), ": ", 2)[1])

	if verbose {
		logVersions()
	}

	if v {
		fmt.Println("Helmsman version: " + appVersion)
		os.Exit(0)
	}

	if len(files) == 0 {
		logs.Info("No desired state files provided.")
		os.Exit(0)
	}

	if kubeconfig != "" {
		os.Setenv("KUBECONFIG", kubeconfig)
	}

	if !toolExists("kubectl") {
		logError("kubectl is not installed/configured correctly. Aborting!")
	}

	if !toolExists(helmBin) {
		logError("" + helmBin + " is not installed/configured correctly. Aborting!")
	}

	if !helmPluginExists("diff") {
		logError("helm diff plugin is not installed/configured correctly. Aborting!")
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
			logs.Info(msg)
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
			if err := s.validate(); err != nil { // syntax validation
				logs.Error(err.Error())
			}
		}
	} else {
		logs.Info("Desired state validation is skipped.")
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

	if len(group) > 0 {
		groupMap = map[string]bool{}
		for _, v := range group {
			groupMap[v] = true
		}
	}

}

// toolExists returns true if the tool is present in the environment and false otherwise.
// It takes as input the tool's command to check if it is recognizable or not. e.g. helm or kubectl
func toolExists(tool string) bool {
	cmd := command{
		Cmd:         tool,
		Args:        []string{},
		Description: "validating that " + tool + " is installed.",
	}

	exitCode, _, _ := cmd.exec(debug, false)

	if exitCode != 0 {
		return false
	}

	return true
}

// helmPluginExists returns true if the plugin is present in the environment and false otherwise.
// It takes as input the plugin's name to check if it is recognizable or not. e.g. diff
func helmPluginExists(plugin string) bool {
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"plugin", "list"},
		Description: "validating that " + plugin + " is installed.",
	}

	exitCode, result, _ := cmd.exec(debug, false)

	if exitCode != 0 {
		return false
	}

	return strings.Contains(result, plugin)
}
