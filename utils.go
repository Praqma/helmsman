package main

import (
	"bytes"
	"fmt"
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
// isisOfType is case insensitive.
func isOfType(filename string, filetype string) bool {
	return filepath.Ext(strings.ToLower(filename)) == filetype
}
