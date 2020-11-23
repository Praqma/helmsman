package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

// invokes either yaml or toml parser considering file extension
func (s *state) fromFile(file string) (bool, string) {
	if isOfType(file, []string{".toml"}) {
		return s.fromTOML(file)
	} else if isOfType(file, []string{".yaml", ".yml"}) {
		return s.fromYAML(file)
	} else {
		return false, "State file does not have toml/yaml extension."
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
func (s *state) fromTOML(file string) (bool, string) {
	rawTomlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return false, err.Error()
	}

	tomlFile := string(rawTomlFile)
	if !flags.noEnvSubst {
		if err := validateEnvVars(tomlFile, file); err != nil {
			return false, err.Error()
		}
		tomlFile = substituteEnv(tomlFile)
	}
	if !flags.noSSMSubst {
		tomlFile = substituteSSM(tomlFile)
	}
	if _, err := toml.Decode(tomlFile, s); err != nil {
		return false, err.Error()
	}
	s.expand(file)

	return true, "Parsed TOML [[ " + file + " ]] successfully and found [ " + strconv.Itoa(len(s.Apps)) + " ] apps"
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
func (s *state) fromYAML(file string) (bool, string) {
	rawYamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return false, err.Error()
	}

	yamlFile := string(rawYamlFile)
	if !flags.noEnvSubst {
		if err := validateEnvVars(yamlFile, file); err != nil {
			return false, err.Error()
		}
		yamlFile = substituteEnv(yamlFile)
	}
	if !flags.noSSMSubst {
		yamlFile = substituteSSM(yamlFile)
	}

	if err = yaml.UnmarshalStrict([]byte(yamlFile), s); err != nil {
		return false, err.Error()
	}
	s.expand(file)

	return true, "Parsed YAML [[ " + file + " ]] successfully and found [ " + strconv.Itoa(len(s.Apps)) + " ] apps"
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

// expand resolves relative paths of certs/keys/chart/value file/secret files/etc and replace them with a absolute paths
// it also loops through the values/secrets files and substitutes variables into them.
func (s *state) expand(relativeToFile string) {
	dir := filepath.Dir(relativeToFile)
	downloadDest, _ := filepath.Abs(createTempDir(tempFilesDir, "tmp"))
	for _, r := range s.Apps {

		r.resolvePaths(dir, downloadDest)
		if r.Chart != "" {
			var repoOrDir = filepath.Dir(r.Chart)
			_, isRepo := s.HelmRepos[repoOrDir]
			isRepo = isRepo || stringInSlice(repoOrDir, s.PreconfiguredHelmRepos)
			if !isRepo {
				// if there is no repo for the chart, we assume it's intended to be a local path

				// support env vars in path
				r.Chart = os.ExpandEnv(r.Chart)
				// respect absolute paths to charts but resolve relative paths
				if !filepath.IsAbs(r.Chart) {
					r.Chart, _ = filepath.Abs(filepath.Join(dir, r.Chart))
				}
			}
		}

		for key, val := range s.Settings.GlobalHooks {
			if key != "deleteOnSuccess" && key != "successTimeout" && key != "successCondition" {
				hook := val.(string)
				if err := isValidFile(hook, []string{".yaml", ".yml"}); err == nil {
					s.Settings.GlobalHooks[key] = substituteVarsInYaml(hook)
				}
			}
		}

		r.substituteVarsInStaticFiles()
	}
	// resolving paths for Bearer Token path in settings
	if s.Settings.BearerTokenPath != "" {
		s.Settings.BearerTokenPath, _ = resolveOnePath(s.Settings.BearerTokenPath, dir, downloadDest)
	}
	// resolve paths for global hooks
	for key, val := range s.Settings.GlobalHooks {
		if key != "deleteOnSuccess" && key != "successTimeout" && key != "successCondition" {
			hook := val.(string)
			if err := isValidFile(hook, []string{".yaml", ".yml", ".json"}); err == nil {
				s.Settings.GlobalHooks[key], _ = resolveOnePath(hook, dir, downloadDest)
			}
		}
	}
	// resolving paths for k8s certificate files
	for k := range s.Certificates {
		s.Certificates[k], _ = resolveOnePath(s.Certificates[k], "", downloadDest)
	}
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
