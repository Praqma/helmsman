package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

// invokes either yaml or toml parser considering file extension
func (s *state) fromFile(file string) error {
	if isOfType(file, []string{".toml", ".tml"}) {
		return s.fromTOML(file)
	} else if isOfType(file, []string{".yaml", ".yml"}) {
		return s.fromYAML(file)
	} else {
		return fmt.Errorf("state file does not have a valid extension")
	}
}

func (s *state) toFile(file string) {
	if isOfType(file, []string{".toml"}) {
		s.toTOML(file)
	} else if isOfType(file, []string{".yaml", ".yml"}) {
		s.toYAML(file)
	} else {
		log.Fatal("State file does not have toml/yaml extension.")
	}
}

// fromTOML reads a toml file and decodes it to a state type.
// It uses the BurntSuchi TOML parser which throws an error if the TOML file is not valid.
func (s *state) fromTOML(file string) error {
	rawTomlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	tomlFile := string(rawTomlFile)
	if !flags.noEnvSubst {
		if err := validateEnvVars(tomlFile, file); err != nil {
			return err
		}
		tomlFile = substituteEnv(tomlFile)
	}
	if !flags.noSSMSubst {
		tomlFile = substituteSSM(tomlFile)
	}
	if _, err := toml.Decode(tomlFile, s); err != nil {
		return err
	}
	s.expand(file)

	return nil
}

// toTOML encodes a state type into a TOML file.
// It uses the BurntSuchi TOML parser.
func (s *state) toTOML(file string) {
	log.Info("Printing generated toml ... ")
	var buff bytes.Buffer
	var (
		newFile *os.File
		err     error
	)

	if err := toml.NewEncoder(&buff).Encode(s); err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	newFile, err = os.Create(file)
	if err != nil {
		log.Fatal(err.Error())
	}
	bytesWritten, err := newFile.Write(buff.Bytes())
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(fmt.Sprintf("Wrote %d bytes.\n", bytesWritten))
	newFile.Close()
}

// fromYAML reads a yaml file and decodes it to a state type.
// parser which throws an error if the YAML file is not valid.
func (s *state) fromYAML(file string) error {
	rawYamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	yamlFile := string(rawYamlFile)
	if !flags.noEnvSubst {
		if err := validateEnvVars(yamlFile, file); err != nil {
			return err
		}
		yamlFile = substituteEnv(yamlFile)
	}
	if !flags.noSSMSubst {
		yamlFile = substituteSSM(yamlFile)
	}

	if err = yaml.UnmarshalStrict([]byte(yamlFile), s); err != nil {
		return err
	}
	s.expand(file)

	return nil
}

// toYaml encodes a state type into a YAML file
func (s *state) toYAML(file string) {
	log.Info("Printing generated yaml ... ")
	var buff bytes.Buffer
	var (
		newFile *os.File
		err     error
	)

	if err := yaml.NewEncoder(&buff).Encode(s); err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	newFile, err = os.Create(file)
	if err != nil {
		log.Fatal(err.Error())
	}
	bytesWritten, err := newFile.Write(buff.Bytes())
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(fmt.Sprintf("Wrote %d bytes.\n", bytesWritten))
	newFile.Close()
}

func (s *state) build(files fileOptionArray) error {
	for _, f := range files {
		fileState := state{mergedState: s}

		if err := fileState.fromFile(f.name); err != nil {
			return err
		}

		log.Infof("Parsed [[ %s ]] successfully and found [ %d ] apps", f.name, len(fileState.Apps))
		// Merge Apps that already existed in the state
		for appName, app := range fileState.Apps {
			if _, ok := s.Apps[appName]; ok {
				if err := mergo.Merge(s.Apps[appName], app, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
					return fmt.Errorf("failed to merge %s from desired state file %s: %w", appName, f.name, err)
				}
			}
		}

		// Merge the remaining Apps
		if err := mergo.Merge(&s.Apps, &fileState.Apps); err != nil {
			return fmt.Errorf("failed to merge desired state file %s: %w", f.name, err)
		}
		// All the apps are already merged, make fileState.Apps empty to avoid conflicts in the final merge
		fileState.Apps = make(map[string]*release)

		if err := mergo.Merge(s, &fileState, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
			return fmt.Errorf("failed to merge desired state file %s: %w", f.name, err)
		}
	}

	s.init() // Set defaults
	return nil
}

// expand resolves relative paths of certs/keys/chart/value file/secret files/etc and replace them with a absolute paths
// it also loops through the values/secrets files and substitutes variables into them.
func (s *state) expand(relativeToFile string) {
	dir := filepath.Dir(relativeToFile)
	downloadDest, _ := filepath.Abs(createTempDir(tempFilesDir, "tmp"))
	validProtocols := []string{"http", "https"}
	if checkHelmVersion(">=3.8.0") {
		validProtocols = append(validProtocols, "oci")
	}
	for _, r := range s.Apps {
		// resolve paths for all release files (values, secrets, hooks, etc...)
		r.resolvePaths(dir, downloadDest)

		// resolve paths for local charts
		if r.Chart != "" {
			var download bool
			// support env vars in path
			r.Chart = os.Expand(r.Chart, getEnv)
			// if there is no repo for the chart, we assume it's intended to be a local path or url
			if !s.isChartFromRepo(r.Chart) {
				// unless explicitly requested by the user, we don't need to download if the protocol is natively supported by helm
				download = flags.downloadCharts || !isSupportedProtocol(r.Chart, validProtocols)
			}
			if download {
				if strings.HasPrefix(r.Chart, "oci://") && !strings.HasSuffix(r.Chart, r.Version) {
					r.Chart = fmt.Sprintf("%s:%s", r.Chart, r.Version)
				}
				r.Chart, _ = resolveOnePath(r.Chart, dir, downloadDest)
			}
		}
		// expand env variables for all release files
		r.substituteVarsInStaticFiles()
	}
	// resolve paths and expand env variables for global hook files
	for key, val := range s.Settings.GlobalHooks {
		if key != "deleteOnSuccess" && key != "successTimeout" && key != "successCondition" {
			file := val.(string)
			hook := strings.Fields(file)
			if isOfType(hook[0], validHookFiles) && !ToolExists(hook[0]) {
				hook[0], _ = resolveOnePath(hook[0], dir, downloadDest)
				file = strings.Join(hook, " ")
				s.Settings.GlobalHooks[key] = file
			}
			if isOfType(file, []string{".yaml", ".yml"}) {
				s.Settings.GlobalHooks[key] = substituteVarsInYaml(file)
			}
		}
	}
	// resolving paths for Bearer Token path in settings
	if s.Settings.BearerTokenPath != "" {
		s.Settings.BearerTokenPath, _ = resolveOnePath(s.Settings.BearerTokenPath, dir, downloadDest)
	}
	// resolving paths for k8s certificate files
	for k := range s.Certificates {
		s.Certificates[k], _ = resolveOnePath(s.Certificates[k], "", downloadDest)
	}
}

func (s *state) isChartFromRepo(chart string) bool {
	repoName := strings.Split(chart, "/")[0]
	if _, isRepo := s.HelmRepos[repoName]; isRepo {
		return true
	}
	if _, isRepo := s.mergedState.HelmRepos[repoName]; isRepo {
		return true
	}
	return stringInSlice(repoName, s.PreconfiguredHelmRepos)
}

// cleanup deletes the k8s certificates and keys files
// It also deletes any Tiller TLS certs and keys
// and secret files
func (s *state) cleanup() {
	log.Verbose("Cleaning up sensitive and temp files")
	if _, err := os.Stat("ca.crt"); err == nil {
		deleteFile("ca.crt")
	}

	if _, err := os.Stat("ca.key"); err == nil {
		deleteFile("ca.key")
	}

	if _, err := os.Stat("client.crt"); err == nil {
		deleteFile("client.crt")
	}

	if _, err := os.Stat("bearer.token"); err == nil {
		deleteFile("bearer.token")
	}

	for _, app := range s.Apps {
		if _, err := os.Stat(app.SecretsFile + ".dec"); err == nil {
			deleteFile(app.SecretsFile + ".dec")
		}
		for _, secret := range app.SecretsFiles {
			if _, err := os.Stat(secret + ".dec"); err == nil {
				deleteFile(secret + ".dec")
			}
		}
	}
}
