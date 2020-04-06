package app

import (
	"flag"
	"fmt"
	"os"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/imdario/mergo"
	"github.com/joho/godotenv"
)

const (
	banner = "\n" +
		" _          _ \n" +
		"| |        | | \n" +
		"| |__   ___| |_ __ ___  ___ _ __ ___   __ _ _ __\n" +
		"| '_ \\ / _ \\ | '_ ` _ \\/ __| '_ ` _ \\ / _` | '_ \\ \n" +
		"| | | |  __/ | | | | | \\__ \\ | | | | | (_| | | | | \n" +
		"|_| |_|\\___|_|_| |_| |_|___/_| |_| |_|\\__,_|_| |_|"
	slogan = "A Helm-Charts-as-Code tool.\n\n"
)

// Allow parsing of multiple string command line options into an array of strings
type stringArray []string

func (i *stringArray) String() string {
	return strings.Join(*i, " ")
}

func (i *stringArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type cli struct {
	debug                 bool
	files                 stringArray
	envFiles              stringArray
	target                stringArray
	group                 stringArray
	kubeconfig            string
	apply                 bool
	destroy               bool
	dryRun                bool
	verbose               bool
	noBanner              bool
	noColors              bool
	noFancy               bool
	noNs                  bool
	nsOverride            string
	contextOverride       string
	skipValidation        bool
	keepUntrackedReleases bool
	showDiff              bool
	diffContext           int
	noEnvSubst            bool
	substEnvValues        bool
	noSSMSubst            bool
	substSSMValues        bool
	updateDeps            bool
	forceUpgrades         bool
	version               bool
	noCleanup             bool
	migrateContext        bool
	parallel              int
}

func printUsage() {
	fmt.Print(banner)
	fmt.Printf("Helmsman version: " + appVersion + "\n")
	fmt.Printf("Helmsman is a Helm Charts as Code tool which allows you to automate the deployment/management of your Helm charts.")
	fmt.Printf("")
	fmt.Printf("Usage: helmsman [options]\n")
	flag.PrintDefaults()
}

// Cli parses cmd flags, validates them and performs some initializations
func (c *cli) parse() {
	//parsing command line flags
	flag.Var(&c.files, "f", "desired state file name(s), may be supplied more than once to merge state files")
	flag.Var(&c.envFiles, "e", "file(s) to load environment variables from (default .env), may be supplied more than once")
	flag.Var(&c.target, "target", "limit execution to specific app.")
	flag.Var(&c.group, "group", "limit execution to specific group of apps.")
	flag.IntVar(&c.diffContext, "diff-context", -1, "number of lines of context to show around changes in helm diff output")
	flag.IntVar(&c.parallel, "p", 1, "max number of concurrent helm releases to run")
	flag.StringVar(&c.kubeconfig, "kubeconfig", "", "path to the kubeconfig file to use for CLI requests")
	flag.StringVar(&c.nsOverride, "ns-override", "", "override defined namespaces with this one")
	flag.StringVar(&c.contextOverride, "context-override", "", "override releases context defined in release state with this one")
	flag.BoolVar(&c.apply, "apply", false, "apply the plan directly")
	flag.BoolVar(&c.dryRun, "dry-run", false, "apply the dry-run option for helm commands.")
	flag.BoolVar(&c.destroy, "destroy", false, "delete all deployed releases.")
	flag.BoolVar(&c.version, "v", false, "show the version")
	flag.BoolVar(&c.debug, "debug", false, "show the debug execution logs and actual helm/kubectl commands. This can log secrets and should only be used for debugging purposes.")
	flag.BoolVar(&c.verbose, "verbose", false, "show verbose execution logs.")
	flag.BoolVar(&c.noBanner, "no-banner", false, "don't show the banner")
	flag.BoolVar(&c.noColors, "no-color", false, "don't use colors")
	flag.BoolVar(&c.noFancy, "no-fancy", false, "don't display the banner and don't use colors")
	flag.BoolVar(&c.noNs, "no-ns", false, "don't create namespaces")
	flag.BoolVar(&c.skipValidation, "skip-validation", false, "skip desired state validation")
	flag.BoolVar(&c.keepUntrackedReleases, "keep-untracked-releases", false, "keep releases that are managed by Helmsman from the used DSFs in the command, and are no longer tracked in your desired state.")
	flag.BoolVar(&c.showDiff, "show-diff", false, "show helm diff results. Can expose sensitive information.")
	flag.BoolVar(&c.noEnvSubst, "no-env-subst", false, "turn off environment substitution globally")
	flag.BoolVar(&c.substEnvValues, "subst-env-values", false, "turn on environment substitution in values files.")
	flag.BoolVar(&c.noSSMSubst, "no-ssm-subst", false, "turn off SSM parameter substitution globally")
	flag.BoolVar(&c.substSSMValues, "subst-ssm-values", false, "turn on SSM parameter substitution in values files.")
	flag.BoolVar(&c.updateDeps, "update-deps", false, "run 'helm dep up' for local chart")
	flag.BoolVar(&c.forceUpgrades, "force-upgrades", false, "use --force when upgrading helm releases. May cause resources to be recreated.")
	flag.BoolVar(&c.noCleanup, "no-cleanup", false, "keeps any credentials files that has been downloaded on the host where helmsman runs.")
	flag.BoolVar(&c.migrateContext, "migrate-context", false, "Updates the context name for all apps defined in the DSF and applies Helmsman labels. Using this flag is required if you want to change context name after it has been set.")
	flag.Usage = printUsage
	flag.Parse()

	if c.version {
		fmt.Println("Helmsman version: " + appVersion)
		os.Exit(0)
	}

	if c.noFancy {
		c.noColors = true
		c.noBanner = true
	}
	verbose := c.verbose || c.debug
	initLogs(verbose, c.noColors)

	if !c.noBanner {
		fmt.Printf("%s version: %s\n%s", banner, appVersion, slogan)
	}

	if c.dryRun && c.apply {
		log.Fatal("--apply and --dry-run can't be used together.")
	}

	if c.destroy && c.apply {
		log.Fatal("--destroy and --apply can't be used together.")
	}

	if len(c.target) > 0 && len(c.group) > 0 {
		log.Fatal("--target and --group can't be used together.")
	}

	if c.parallel < 1 {
		c.parallel = 1
	}

	helmVersion := strings.TrimSpace(getHelmVersion())
	extractedHelmVersion := helmVersion
	if !strings.HasPrefix(helmVersion, "v") {
		extractedHelmVersion = strings.TrimSpace(strings.Split(helmVersion, ":")[1])
	}
	log.Verbose("Helm client version: " + extractedHelmVersion)
	v1, _ := version.NewVersion(extractedHelmVersion)
	jsonConstraint, _ := version.NewConstraint(">=3.0.0")
	if !jsonConstraint.Check(v1) {
		log.Fatal("this version of Helmsman does not work with helm releases older than 3.0.0")
	}

	kubectlVersion := strings.TrimSpace(strings.SplitN(getKubectlClientVersion(), ": ", 2)[1])
	log.Verbose("kubectl client version: " + kubectlVersion)

	if len(c.files) == 0 {
		log.Info("No desired state files provided.")
		os.Exit(0)
	}

	if c.kubeconfig != "" {
		os.Setenv("KUBECONFIG", c.kubeconfig)
	}

	if !toolExists("kubectl") {
		log.Fatal("kubectl is not installed/configured correctly. Aborting!")
	}

	if !toolExists(helmBin) {
		log.Fatal("" + helmBin + " is not installed/configured correctly. Aborting!")
	}

	if !helmPluginExists("diff") {
		log.Fatal("helm diff plugin is not installed/configured correctly. Aborting!")
	}

	if !c.noEnvSubst {
		log.Verbose("Substitution of env variables enabled")
		if c.substEnvValues {
			log.Verbose("Substitution of env variables in values enabled")
		}
	}
	if !c.noSSMSubst {
		log.Verbose("Substitution of SSM variables enabled")
		if c.substSSMValues {
			log.Verbose("Substitution of SSM variables in values enabled")
		}
	}
}

// readState gets the desired state from files
func (c *cli) readState(s *state) {
	// read the env file
	if len(c.envFiles) == 0 {
		if _, err := os.Stat(".env"); err == nil {
			err = godotenv.Load()
			if err != nil {
				log.Fatal("Error loading .env file")
			}
		}
	}

	for _, e := range c.envFiles {
		err := godotenv.Load(e)
		if err != nil {
			log.Fatal("Error loading " + e + " env file")
		}
	}

	// wipe & create a temporary directory
	os.RemoveAll(tempFilesDir)
	_ = os.MkdirAll(tempFilesDir, 0755)

	// read the TOML/YAML desired state file
	var fileState state
	for _, f := range c.files {

		result, msg := fileState.fromFile(f)
		if result {
			log.Info(msg)
		} else {
			log.Fatal(msg)
		}
		// Merge Apps that already existed in the state
		for appName, app := range fileState.Apps {
			if _, ok := s.Apps[appName]; ok {
				if err := mergo.Merge(s.Apps[appName], app, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
					log.Fatal("Failed to merge " + appName + " from desired state file" + f)
				}
			}
		}

		// Merge the remaining Apps
		if err := mergo.Merge(&s.Apps, &fileState.Apps); err != nil {
			log.Fatal("Failed to merge desired state file" + f)
		}
		// All the apps are already merged, make fileState.Apps empty to avoid conflicts in the final merge
		fileState.Apps = make(map[string]*release)

		if err := mergo.Merge(s, &fileState, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
			log.Fatal("Failed to merge desired state file" + f)
		}
	}

	if len(c.target) > 0 {
		s.TargetMap = map[string]bool{}
		for _, v := range c.target {
			s.TargetMap[v] = true
		}
	}

	if len(c.group) > 0 {
		s.GroupMap = map[string]bool{}
		for _, v := range c.group {
			s.GroupMap[v] = true
		}
	}

	if c.debug {
		s.print()
	}

	if !c.skipValidation {
		// validate the desired state content
		if len(c.files) > 0 {
			log.Info("Validating desired state definition...")
			if err := s.validate(); err != nil { // syntax validation
				log.Fatal(err.Error())
			}
		}
	} else {
		log.Info("Desired state validation is skipped.")
	}

	if s.Settings.StorageBackend != "" {
		os.Setenv("HELM_DRIVER", s.Settings.StorageBackend)
	} else {
		// set default storage background to secret if not set by user
		s.Settings.StorageBackend = "secret"
	}

	// if there is no user-defined context name in the DSF(s), use the default context name
	if s.Context == "" {
		s.Context = defaultContextName
	}
}

// getDryRunFlags returns dry-run flag
func (c *cli) getDryRunFlags() []string {
	if c.dryRun {
		return []string{"--dry-run", "--debug"}
	}
	return []string{}
}
