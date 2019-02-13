package main

import (
	"fmt"
)

// resources type
type resources struct {
	Cpu    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

// limits type
type limits struct {
	Max            resources `yaml:"max,omitempty"`
	Min            resources `yaml:"min,omitempty"`
	Default        resources `yaml:"default,omitempty"`
	DefaultRequest resources `yaml:"defaultRequest,omitempty"`
}

// namespace type represents the fields of a namespace
type namespace struct {
	Protected            bool              `yaml:"protected"`
	InstallTiller        bool              `yaml:"installTiller"`
	UseTiller            bool              `yaml:"useTiller"`
	TillerServiceAccount string            `yaml:"tillerServiceAccount"`
	TillerRole           string            `yaml:"tillerRole"`
	CaCert               string            `yaml:"caCert"`
	TillerCert           string            `yaml:"tillerCert"`
	TillerKey            string            `yaml:"tillerKey"`
	ClientCert           string            `yaml:"clientCert"`
	ClientKey            string            `yaml:"clientKey"`
	Limits               limits            `yaml:"limits,omitempty"`
	Labels               map[string]string `yaml:"labels"`
	Annotations          map[string]string `yaml:"annotations"`
}

// checkNamespaceDefined checks if a given namespace is defined in the namespaces section of the desired state file
func checkNamespaceDefined(ns string, s state) bool {
	_, ok := s.Namespaces[ns]
	if !ok {
		return false
	}
	return true
}

// print prints the namespace
func (n namespace) print() {
	fmt.Println("")
	fmt.Println("\tprotected : ", n.Protected)
	fmt.Println("\tinstallTiller : ", n.InstallTiller)
	fmt.Println("\tuseTiller : ", n.UseTiller)
	fmt.Println("\ttillerServiceAccount : ", n.TillerServiceAccount)
	fmt.Println("\ttillerRole: ", n.TillerRole)
	fmt.Println("\tcaCert : ", n.CaCert)
	fmt.Println("\ttillerCert : ", n.TillerCert)
	fmt.Println("\ttillerKey : ", n.TillerKey)
	fmt.Println("\tclientCert : ", n.ClientCert)
	fmt.Println("\tclientKey : ", n.ClientKey)
	fmt.Println("\tlabels : ")
	printMap(n.Labels, 2)
	fmt.Println("------------------- ")
}
