package main

import (
	"fmt"
)

type release struct {
	Name        string
	Description string
	Env         string
	Enabled     bool
	Chart       string
	Version     string
	ValuesFile  string
	Purge       bool
}

func (r release) print() {
	fmt.Println("")
	fmt.Println("\tname : ", r.Name)
	fmt.Println("\tdescription : ", r.Description)
	fmt.Println("\tenv : ", r.Env)
	fmt.Println("\tenabled : ", r.Enabled)
	fmt.Println("\tchart : ", r.Chart)
	fmt.Println("\tversion : ", r.Version)
	fmt.Println("\tvaluesFile : ", r.ValuesFile)
	fmt.Println("\tpurge : ", r.Purge)
	fmt.Println("------------------- ")
}
