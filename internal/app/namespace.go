package app

import (
	"fmt"
)

// Resources type
type Resources struct {
	// CPU is the number of CPU cores
	CPU string `yaml:"cpu,omitempty"`
	// Memory is the amount of memory
	Memory string `yaml:"memory,omitempty"`
}

// custom resource type
type CustomResource struct {
	// Name of the custom resource
	Name string `yaml:"name,omitempty"`
	// Value of the custom resource
	Value string `yaml:"value,omitempty"`
}

// Limit represents a resource limit
type Limit struct {
	// Max defines the resource limits
	Max Resources `yaml:"max,omitempty"`
	// Min defines the resource request
	Min Resources `yaml:"min,omitempty"`
	// Default stes resource limits to pods without defined resource limits
	Default Resources `yaml:"default,omitempty"`
	// DefaultRequest sets the resource requests for pods without defined resource requests
	DefaultRequest Resources `yaml:"defaultRequest,omitempty"`
	// MaxLimitRequestRatio set the max limit request ratio
	MaxLimitRequestRatio Resources `yaml:"maxLimitRequestRatio,omitempty"`
	Type                 string    `yaml:"type"`
}

// Limits type
type Limits []Limit

// quota type
type Quotas struct {
	// Pods is the pods quota
	Pods string `yaml:"pods,omitempty"`
	// CPULimits is the CPU quota
	CPULimits string `yaml:"limits.cpu,omitempty"`
	// CPURequests is the CPU requests quota
	CPURequests string `yaml:"requests.cpu,omitempty"`
	// MemoryLimits is the memory quota
	MemoryLimits string `yaml:"limits.memory,omitempty"`
	// MemoryRequests is the memory requests quota
	MemoryRequests string `yaml:"requests.memory,omitempty"`
	// CustomResource is a list of custom resource quotas
	CustomQuotas []CustomResource `yaml:"customQuotas,omitempty"`
}

// Namespace type represents the fields of a Namespace
type Namespace struct {
	// Protected if set to true no changes can be applied to the namespace
	Protected bool `yaml:"protected"`
	// Limits to set on the namespace
	Limits Limits `yaml:"limits,omitempty"`
	// Labels to set to the namespace
	Labels map[string]string `yaml:"labels,omitempty"`
	// Annotations to set on the namespace
	Annotations map[string]string `yaml:"annotations,omitempty"`
	// Quotas to set on the namespace
	Quotas   *Quotas `yaml:"quotas,omitempty"`
	disabled bool
}

func (n *Namespace) Disable() {
	n.disabled = true
}

// print prints the namespace
func (n *Namespace) print() {
	fmt.Println("\tprotected: ", n.Protected)
	fmt.Println("\tdisabled: ", n.disabled)
	fmt.Println("\tlabels:")
	printMap(n.Labels, 2)
	fmt.Println("\tannotations:")
	printMap(n.Annotations, 2)
	fmt.Println("-------------------")
}
