package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// printMap prints to the console any map of string keys and values.
func printMap(m map[string]string) {
	for key, value := range m {
		fmt.Println(key, " : ", value)
	}
}

// fromTOML reads a toml file and decodes it to a state type.
// It uses the BurntSuchi TOML parser which throws an error if the TOML file is not valid.
func fromTOML(file string, s *state) {

	if _, err := toml.DecodeFile(file, s); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Printf("Parsed [[ %s ]] successfully and found [%v] apps", file, len(s.Apps))

}

// toTOML encodes a state type into a TOML file.
// It uses the BurntSuchi TOML parser.
func toTOML(file string, s *state) {
	log.Println("printing generated toml ... ")
	var buff bytes.Buffer
	var (
		newFile *os.File
		err     error
	)

	if err := toml.NewEncoder(&buff).Encode(s); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	newFile, err = os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	bytesWritten, err := newFile.Write(buff.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote %d bytes.\n", bytesWritten)
	newFile.Close()
}

// isOfType checks if the file extension of a filename/path is the same as "filetype".
// isisOfType is case insensitive. filetype should contain the "." e.g. ".yaml"
func isOfType(filename string, filetype string) bool {
	return filepath.Ext(strings.ToLower(filename)) == strings.ToLower(filetype)
}

// readFile returns the content of a file as a string.
// takes a file path as input. It throws an error and breaks the program execution if it failes to read the file.
func readFile(filepath string) string {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("ERROR: failed to read password file content: " + err.Error())
	}
	return string(data)
}

// printHelp prints Helmsman commands
func printHelp() {
	fmt.Println("Helmsman is a Helm Charts as Code tool which allows you to automate the deployment/management of your Helm charts.")
	fmt.Println(" Usage: helmsman [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("-f     specifies the desired state TOML file.")
	fmt.Println("-debug prints all the logs during execution.")
	fmt.Println("-apply generates and applies an action plan.")
	fmt.Println("-help  prints Helmsman help.")

}
