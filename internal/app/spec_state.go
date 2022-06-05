package app

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type StatePath struct {
	Path string `yaml:"path"`
}

type StateFiles struct {
	StateFiles []StatePath `yaml:"stateFiles"`
}

// specFromYAML reads a yaml file and decodes it to a state type.
// parser which throws an error if the YAML file is not valid.
func (pc *StateFiles) specFromYAML(file string) error {
	rawYamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errorf("specFromYaml %v %v", file, err)
		return err
	}

	yamlFile := string(rawYamlFile)

	return yaml.UnmarshalStrict([]byte(yamlFile), pc)
}
