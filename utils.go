package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/BurntSushi/toml"
	"github.com/Praqma/helmsman/aws"
	"github.com/Praqma/helmsman/gcs"
)

// printMap prints to the console any map of string keys and values.
func printMap(m map[string]string, indent int) {
	for key, value := range m {
		fmt.Println(strings.Repeat("\t", indent)+key, " : ", value)
	}
}

// printObjectMap prints to the console any map of string keys and object values.
func printNamespacesMap(m map[string]namespace) {
	for key, value := range m {
		fmt.Println(key, " : protected = ", value)
	}
}

// fromTOML reads a toml file and decodes it to a state type.
// It uses the BurntSuchi TOML parser which throws an error if the TOML file is not valid.
func fromTOML(file string, s *state) (bool, string) {
	rawTomlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return false, err.Error()
	}
	tomlFile := substituteEnv(string(rawTomlFile))
	if _, err := toml.Decode(tomlFile, s); err != nil {
		return false, err.Error()
	}

	resolvePaths(file, s)

	return true, "INFO: Parsed TOML [[ " + file + " ]] successfully and found [ " + strconv.Itoa(len(s.Apps)) + " ] apps."
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
		logError(err.Error())
		os.Exit(1)
	}
	newFile, err = os.Create(file)
	if err != nil {
		logError(err.Error())
	}
	bytesWritten, err := newFile.Write(buff.Bytes())
	if err != nil {
		logError(err.Error())
	}
	log.Printf("Wrote %d bytes.\n", bytesWritten)
	newFile.Close()
}

// fromYAML reads a yaml file and decodes it to a state type.
// parser which throws an error if the YAML file is not valid.
func fromYAML(file string, s *state) (bool, string) {
	rawYamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return false, err.Error()
	}
	yamlFile := []byte(substituteEnv(string(rawYamlFile)))
	if err = yaml.UnmarshalStrict(yamlFile, s); err != nil {
		return false, err.Error()
	}

	resolvePaths(file, s)

	return true, "INFO: Parsed YAML [[ " + file + " ]] successfully and found [ " + strconv.Itoa(len(s.Apps)) + " ] apps."
}

// toYaml encodes a state type into a YAML file
func toYAML(file string, s *state) {
	log.Println("printing generated yaml ... ")
	var buff bytes.Buffer
	var (
		newFile *os.File
		err     error
	)

	if err := yaml.NewEncoder(&buff).Encode(s); err != nil {
		logError(err.Error())
		os.Exit(1)
	}
	newFile, err = os.Create(file)
	if err != nil {
		logError(err.Error())
	}
	bytesWritten, err := newFile.Write(buff.Bytes())
	if err != nil {
		logError(err.Error())
	}
	log.Printf("Wrote %d bytes.\n", bytesWritten)
	newFile.Close()
}

// invokes either yaml or toml parser considering file extension
func fromFile(file string, s *state) (bool, string) {
	if isOfType(file, ".toml") {
		return fromTOML(file, s)
	} else if isOfType(file, ".yaml") {
		return fromYAML(file, s)
	} else {
		return false, "State file does not have toml/yaml extension."
	}
}

func toFile(file string, s *state) {
	if isOfType(file, ".toml") {
		toTOML(file, s)
	} else if isOfType(file, ".yaml") {
		fromYAML(file, s)
	} else {
		logError("ERROR: State file does not have toml/yaml extension.")
	}
}

func resolvePaths(relativeToFile string, s *state) {
	dir := filepath.Dir(relativeToFile)
	for k, v := range s.Apps {
		if v.ValuesFile != "" {
			v.ValuesFile, _ = filepath.Abs(filepath.Join(dir, v.ValuesFile))
		}
		if v.SecretsFile != "" {
			v.SecretsFile, _ = filepath.Abs(filepath.Join(dir, v.SecretsFile))
		}
		for i, f := range v.ValuesFiles {
			v.ValuesFiles[i], _ = filepath.Abs(filepath.Join(dir, f))
		}
		for i, f := range v.SecretsFiles {
			v.SecretsFiles[i], _ = filepath.Abs(filepath.Join(dir, f))
		}
		s.Apps[k] = v
	}
	//resolving paths for k8s certificate files
	for k, v := range s.Certificates {
		if _, err := url.ParseRequestURI(v); err != nil {
			v, _ = filepath.Abs(filepath.Join(dir, v))
		}
		s.Certificates[k] = v
	}
	// resolving paths for helm certificate files
	for k, v := range s.Namespaces {
		if tillerTLSEnabled(k) {
			if _, err := url.ParseRequestURI(v.CaCert); err != nil {
				v.CaCert, _ = filepath.Abs(filepath.Join(dir, v.CaCert))
			}
			if _, err := url.ParseRequestURI(v.ClientCert); err != nil {
				v.ClientCert, _ = filepath.Abs(filepath.Join(dir, v.ClientCert))
			}
			if _, err := url.ParseRequestURI(v.ClientKey); err != nil {
				v.ClientKey, _ = filepath.Abs(filepath.Join(dir, v.ClientKey))
			}
			if _, err := url.ParseRequestURI(v.TillerCert); err != nil {
				v.TillerCert, _ = filepath.Abs(filepath.Join(dir, v.TillerCert))
			}
			if _, err := url.ParseRequestURI(v.TillerKey); err != nil {
				v.TillerKey, _ = filepath.Abs(filepath.Join(dir, v.TillerKey))
			}
		}
		s.Namespaces[k] = v
	}

}

// isOfType checks if the file extension of a filename/path is the same as "filetype".
// isisOfType is case insensitive. filetype should contain the "." e.g. ".yaml"
func isOfType(filename string, filetype string) bool {
	return filepath.Ext(strings.ToLower(filename)) == strings.ToLower(filetype)
}

// readFile returns the content of a file as a string.
// takes a file path as input. It throws an error and breaks the program execution if it fails to read the file.
func readFile(filepath string) string {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		logError("ERROR: failed to read [ " + filepath + " ] file content: " + err.Error())
	}
	return string(data)
}

// logVersions prints the versions of kubectl and helm to the logs
func logVersions() {
	log.Println("VERBOSE: kubectl client version: " + kubectlVersion)
	log.Println("VERBOSE: Helm client version: " + helmVersion)
}

// substituteEnv checks if a string has an env variable (contains '$'), then it returns its value
// if the env variable is empty or unset, an empty string is returned
// if the string does not contain '$', it is returned as is.
func substituteEnv(name string) string {
	if strings.Contains(name, "$") {
		// add $$ escaping for $ strings
		os.Setenv("HELMSMAN_DOLLAR", "$")
		return os.ExpandEnv(strings.Replace(name, "$$", "${HELMSMAN_DOLLAR}", -1))
	}
	return name
}

// sliceContains checks if a string slice contains a given string
func sliceContains(slice []string, s string) bool {
	for _, a := range slice {
		if strings.TrimSpace(a) == s {
			return true
		}
	}
	return false
}

// downloadFile downloads a file from GCS or AWS buckets and name it with a given outfile
// if downloaded, returns the outfile name. If the file path is local file system path, it is returned as is.
func downloadFile(path string, outfile string) string {
	if strings.HasPrefix(path, "s3") {

		tmp := getBucketElements(path)
		aws.ReadFile(tmp["bucketName"], tmp["filePath"], outfile, noColors)

	} else if strings.HasPrefix(path, "gs") {

		tmp := getBucketElements(path)
		gcs.ReadFile(tmp["bucketName"], tmp["filePath"], outfile, noColors)

	} else {

		log.Println("INFO: " + outfile + " will be used from local file system.")
		copyFile(path, outfile)
	}
	return outfile
}

// copyFile copies a file from source to destination
func copyFile(source string, destination string) {
	from, err := os.Open(source)
	if err != nil {
		logError("ERROR: while copying " + source + " to " + destination + " : " + err.Error())
	}
	defer from.Close()

	to, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		logError("ERROR: while copying " + source + " to " + destination + " : " + err.Error())
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		logError("ERROR: while copying " + source + " to " + destination + " : " + err.Error())
	}
}

// deleteFile deletes a file
func deleteFile(path string) {
	log.Println("INFO: cleaning up ... deleting " + path)
	if err := os.Remove(path); err != nil {
		logError("ERROR: could not delete file: " + path)
	}
}

// notifySlack sends a JSON formatted message to Slack over a webhook url
// It takes the content of the message (what changes helmsman is going to do or have done separated by \n)
// and the webhook URL as well as a flag specifying if this is a failure message or not
// It returns true if the sending of the message is successful, otherwise returns false
func notifySlack(content string, url string, failure bool, executing bool) bool {
	log.Println("INFO: posting notifications to slack ... ")

	color := "#36a64f" // green
	if failure {
		color = "#FF0000" // red
	}

	var pretext string
	if content == "" {
		pretext = "No actions to perform!"
	} else if failure {
		pretext = "Failed to generate/execute a plan: "
	} else if executing && !failure {
		pretext = "Here is what I have done: "
	} else {
		pretext = "Here is what I am going to do:"
	}

	t := time.Now().UTC()

	var jsonStr = []byte(`{
		"attachments": [
			{
				"fallback": "Helmsman results.",
				"color": "` + color + `" ,
				"pretext": "` + pretext + `",
				"title": "` + content + `",
				"footer": "Helmsman ` + appVersion + `",
				"ts": ` + strconv.FormatInt(t.Unix(), 10) + `
			}
		]
	}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logError("ERROR: while sending notifications to slack" + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true
	}
	return false
}

// logError sends a notification on slack if a webhook URL is provided and logs the error before terminating.
func logError(msg string) {
	if _, err := url.ParseRequestURI(s.Settings.SlackWebhook); err == nil {
		notifySlack(msg, s.Settings.SlackWebhook, true, apply)
	}
	log.Fatal(style.Bold(style.Red(msg)))
}

// getBucketElements returns a map containing the bucket name and the file path inside the bucket
// this func works for S3 and GCS bucket links of the format:
// s3 or gs://bucketname/dir.../file.ext
func getBucketElements(link string) map[string]string {

	tmp := strings.SplitAfterN(link, "//", 2)[1]
	m := make(map[string]string)
	m["bucketName"] = strings.SplitN(tmp, "/", 2)[0]
	m["filePath"] = strings.SplitN(tmp, "/", 2)[1]
	return m
}

// replaceStringInFile takes a map of keys and values and replaces the keys with values within a given file.
// It saves the modified content in a new file
func replaceStringInFile(input []byte, outfile string, replacements map[string]string) {
	output := input
	for k, v := range replacements {
		output = bytes.Replace(output, []byte(k), []byte(v), -1)
	}

	if err := ioutil.WriteFile(outfile, output, 0666); err != nil {
		logError(err.Error())
	}
}
