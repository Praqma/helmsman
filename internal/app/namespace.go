package app

import (
	"fmt"
)

// resources type
type resources struct {
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

// custom resource type
type customResource struct {
	Name  string `yaml:"name,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// limits type
type limits []struct {
	Max                  resources `yaml:"max,omitempty"`
	Min                  resources `yaml:"min,omitempty"`
	Default              resources `yaml:"default,omitempty"`
	DefaultRequest       resources `yaml:"defaultRequest,omitempty"`
	MaxLimitRequestRatio resources `yaml:"maxLimitRequestRatio,omitempty"`
	Type                 string    `yaml:"type"`
}

// quota type
type quotas struct {
	Pods           string           `yaml:"pods,omitempty"`
	CPULimits      string           `yaml:"limits.cpu,omitempty"`
	CPURequests    string           `yaml:"requests.cpu,omitempty"`
	MemoryLimits   string           `yaml:"limits.memory,omitempty"`
	MemoryRequests string           `yaml:"requests.memory,omitempty"`
	CustomQuotas   []customResource `yaml:"customQuotas,omitempty"`
}

// namespace type represents the fields of a namespace
type namespace struct {
	Protected   bool              `yaml:"protected"`
	Limits      limits            `yaml:"limits,omitempty"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
	Quotas      *quotas           `yaml:"quotas,omitempty"`
	disabled    bool
}

func (n *namespace) Disable() {
	n.disabled = true
}

// print prints the namespace
func (n *namespace) print() {
	fmt.Println("")
	fmt.Println("\tprotected : ", n.Protected)
	fmt.Println("\tlabels : ")
	printMap(n.Labels, 2)
	fmt.Println("------------------- ")
}
