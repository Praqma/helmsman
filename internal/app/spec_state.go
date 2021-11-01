package app

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type SpecConfig struct {
	Path     string `yaml:"path"`
	Priority int    `yaml:"priority"`
}

type StateFiles struct {
	StateFiles []SpecConfig `yaml:"stateFiles"`
}

// fromYAML reads a yaml file and decodes it to a state type.
// parser which throws an error if the YAML file is not valid.
func (pc *StateFiles) specFromYAML(file string) error {
	rawYamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errorf("specFromYaml %v %v", file, err)
		return err
	}

	yamlFile := string(rawYamlFile)

	if err = yaml.UnmarshalStrict([]byte(yamlFile), pc); err != nil {
		return err
	}

	return nil
}

func checkSpecValid(f string) error {
	if err := isValidFile(f, []string{"yaml", "yml"}); err != nil {
		// exit
		return err
	}

	return nil
}

func (s *state) patchPriority(priority int) error {
	for app, _ := range s.Apps {
		if s.Apps[app].Priority > priority {
			s.Apps[app].Priority = priority
		}
	}
	return nil
}
