package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

type state struct {
	Settings   map[string]string
	Metadata   map[string]string
	Namespaces map[string]string
	HelmRepos  map[string]string
	Apps       map[string]release
}

func (s state) validate() bool {

	// settings
	if s.Settings == nil {
		log.Fatal("ERROR: settings validation failed -- no settings table provided in TOML.")
		return false
	} else if s.Settings["kubeContext"] == "" {
		log.Fatal("ERROR: settings validation failed -- you have not provided a ",
			"kubeContext to use. Can't work without it. Sorry!")
		return false
	}

	// namespaces
	if s.Namespaces == nil || len(s.Namespaces) < 1 {
		log.Fatal("ERROR: namespaces validation failed -- I need at least one namespace ",
			"to work with!")
		return false
	}
	for k, v := range s.Namespaces {
		if v == "" {
			log.Fatal("ERROR: namespaces validation failed -- namespace ["+k+" ] ",
				"must have a value or be removed/commented.")
			return false
		}
	}

	// repos
	if s.HelmRepos == nil || len(s.HelmRepos) < 1 {
		log.Fatal("ERROR: repos validation failed -- I need at least one helm repo ",
			"to work with!")
		return false
	}
	for k, v := range s.HelmRepos {
		_, err := url.ParseRequestURI(v)
		if err != nil {
			log.Fatal("ERROR: repos validation failed -- repo ["+k+" ] ",
				"must have a valid URL.")
			return false
		}

		continue

	}

	// apps
	if s.Apps == nil {
		log.Println("INFO: You have not specified any apps. I have nothing to do. ",
			"Horraayyy!.")
		os.Exit(0)
	}

	names := make(map[string]bool)
	for appLabel, r := range s.Apps {
		result, errMsg := validateRelease(r, names)
		if !result {
			log.Fatal("ERROR: apps validation failed -- for app ["+appLabel+" ]. ",
				errMsg)
			return false
		}
	}

	return true
}

// TODO: make a smarter validation beyond checking for empty text
func validateRelease(r release, names map[string]bool) (bool, string) {
	_, err := os.Stat(r.ValuesFile)
	if r.Name == "" || names[r.Name] {
		return false, "release name can't be empty and must be unique."
	} else if r.Env == "" {
		return false, "env can't be empty."
	} else if r.Chart == "" || !strings.ContainsAny(r.Chart, "/") {
		return false, "chart can't be empty and must be of the format: repo/chart."
	} else if r.Version == "" {
		return false, "version can't be empty."
	} else if r.ValuesFile == "" || err != nil {
		return false, "valuesFile can't be empty and must be a valid file path."
	}

	names[r.Name] = true
	return true, ""

}

func (s state) print() {

	fmt.Println("Settings: ")
	fmt.Println("--------- ")
	printMap(s.Settings)
	fmt.Println("\nMetadata: ")
	fmt.Println("--------- ")
	printMap(s.Metadata)
	fmt.Println("\nNamespaces: ")
	fmt.Println("------------- ")
	printMap(s.Namespaces)
	fmt.Println("\nRepositories: ")
	fmt.Println("------------- ")
	printMap(s.HelmRepos)
	fmt.Println("\nApplications: ")
	fmt.Println("--------------- ")
	for _, r := range s.Apps {
		r.print()
	}
}
