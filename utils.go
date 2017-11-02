package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

func printMap(m map[string]string) {
	for key, value := range m {
		fmt.Println(key, " : ", value)
	}
}

func fromTOML(file string, s *state) {

	if _, err := toml.DecodeFile(file, s); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Printf("Parsed [[ %s ]] successfully and found [%v] apps", file, len(s.Apps))

}

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
