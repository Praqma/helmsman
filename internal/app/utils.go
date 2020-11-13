package app

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Praqma/helmsman/internal/aws"
	"github.com/Praqma/helmsman/internal/azure"
	"github.com/Praqma/helmsman/internal/gcs"
)

// printMap prints to the console any map of string keys and values.
func printMap(m map[string]string, indent int) {
	for key, value := range m {
		fmt.Println(strings.Repeat("\t", indent)+key, ": ", value)
	}
}

// printObjectMap prints to the console any map of string keys and object values.
func printNamespacesMap(m map[string]*namespace) {
	for key, value := range m {
		fmt.Println(key, " : protected = ", value)
	}
}

// substituteVarsInYaml substitutes variables in a Yaml file and creates a temp file with these values.
// Returns the path for the temp file
func substituteVarsInYaml(file string) string {
	rawYamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err.Error())
	}

	yamlFile := string(rawYamlFile)
	if !flags.noEnvSubst && flags.substEnvValues {
		if err := validateEnvVars(yamlFile, file); err != nil {
			log.Critical(err.Error())
		}
		yamlFile = substituteEnv(yamlFile)
	}
	if !flags.noSSMSubst && flags.substSSMValues {
		yamlFile = substituteSSM(yamlFile)
	}

	dir := createTempDir(tempFilesDir, "tmp")

	// output file contents with env variables substituted into temp files
	outFile := path.Join(dir, filepath.Base(file))
	err = ioutil.WriteFile(outFile, []byte(yamlFile), 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
	return outFile
}

// func stringInSlice checks if a string is in a slice
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// resolveOnePath takes the input file (URL, cloud bucket, or local file relative path),
// the directory containing the DSF and the temp directory where files will be fetched to
// and downloads/fetches the file locally into helmsman temp directory and returns
// its absolute path
func resolveOnePath(file string, dir string, downloadDest string) (string, error) {
	if destFile, err := ioutil.TempFile(downloadDest, fmt.Sprintf("*%s", path.Base(file))); err != nil {
		return "", err
	} else {
		_ = destFile.Close()
		return filepath.Abs(downloadFile(file, dir, destFile.Name()))
	}
}

// createTempDir creates a temp directory in a specific location with a pattern
func createTempDir(parent string, pattern string) string {
	dir, err := ioutil.TempDir(parent, pattern)
	if err != nil {
		log.Fatal(err.Error())
	}
	return dir
}

// isOfType checks if the file extension of a filename/path is the same as "filetype".
// isisOfType is case insensitive. filetype should contain the "." e.g. ".yaml"
func isOfType(filename string, filetypes []string) bool {
	lowerMap := make(map[string]struct{})
	for _, v := range filetypes {
		lowerMap[strings.ToLower(v)] = struct{}{}
	}
	_, result := lowerMap[filepath.Ext(strings.ToLower(filename))]
	return result
}

// readFile returns the content of a file as a string.
// takes a file path as input. It throws an error and breaks the program execution if it fails to read the file.
func readFile(filepath string) string {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("failed to read [ " + filepath + " ] file content: " + err.Error())
	}
	return string(data)
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

// validateEnvVars parses a string line-by-line and detect env variables in
// non-comment lines. It then checks that each env var found has a value.
func validateEnvVars(s string, filename string) error {
	if !flags.skipValidation && strings.Contains(s, "$") {
		log.Info("validating environment variables in " + filename)
		var key string
		comment, _ := regexp.Compile("#(.*)$")
		envVar, _ := regexp.Compile(`\${([a-zA-Z_][a-zA-Z0-9_-]*)}|\$([a-zA-Z_][a-zA-Z0-9_-]*)`)
		scanner := bufio.NewScanner(strings.NewReader(s))
		for scanner.Scan() {
			// remove spaces from the single line, then replace $$ with !? to prevent it from matching the regex,
			// then remove new line and inline comments from the text
			text := comment.ReplaceAllString(strings.ReplaceAll(strings.TrimSpace(scanner.Text()), "$$", "!?"), "")
			for _, v := range envVar.FindAllStringSubmatch(text, -1) {
				// FindAllStringSubmatch may match the first (${MY_VAR}) or the second ($MY_VAR) group from the regex
				if v[1] != "" {
					key = v[1]
				} else {
					key = v[2]
				}
				if _, ok := os.LookupEnv(key); !ok {
					return fmt.Errorf("%s is used as an env variable but is currently unset. Either set it or escape it like so: $%s", v[0], v[0])
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Critical(err.Error())
		}
	}
	return nil
}

// substituteSSM checks if a string has an SSM parameter variable (contains '{{ssm: '), then it returns its value
// if the env variable is empty or unset, an empty string is returned
// if the string does not contain '$', it is returned as is.
func substituteSSM(name string) string {
	if strings.Contains(name, "{{ssm: ") {
		re := regexp.MustCompile(`{{ssm: ([^~}]+)(~(true))?}}`)
		matches := re.FindAllSubmatch([]byte(name), -1)
		for _, match := range matches {
			placeholder := string(match[0])
			paramPath := string(match[1])
			withDecryption, err := strconv.ParseBool(string(match[3]))
			if err != nil {
				fmt.Printf("Invalid decryption argument %T \n", string(match[3]))
			}
			value := aws.ReadSSMParam(paramPath, withDecryption, flags.noColors)
			name = strings.ReplaceAll(name, placeholder, value)
		}
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

// downloadFile downloads a file from a URL, GCS, Azure or AWS buckets or local file system
// and saves it with a given outfile name and in a given dir
// if downloaded, returns the outfile name. If the file path is local file system path, it is copied to current directory.
func downloadFile(file string, dir string, outfile string) string {
	if strings.HasPrefix(file, "http") {
		if err := downloadFileFromURL(file, outfile); err != nil {
			log.Fatal(err.Error())
		}
	} else if strings.HasPrefix(file, "s3") {

		tmp := getBucketElements(file)
		aws.ReadFile(tmp["bucketName"], tmp["filePath"], outfile, flags.noColors)

	} else if strings.HasPrefix(file, "gs") {

		tmp := getBucketElements(file)
		msg, err := gcs.ReadFile(tmp["bucketName"], tmp["filePath"], outfile, flags.noColors)
		if err != nil {
			log.Fatal(msg)
		}

	} else if strings.HasPrefix(file, "az") {

		tmp := getBucketElements(file)
		azure.ReadFile(tmp["bucketName"], tmp["filePath"], outfile, flags.noColors)

	} else {

		log.Verbose("" + file + " will be used from local file system.")
		toCopy := file
		if !filepath.IsAbs(file) {
			toCopy, _ = filepath.Abs(filepath.Join(dir, file))
		}
		copyFile(toCopy, outfile)
	}
	return outfile
}

// downloadFileFromURL will download a url to a local file. It writes to the file as it downloads
func downloadFileFromURL(url string, filepath string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// copyFile copies a file from source to destination
func copyFile(source string, destination string) {
	from, err := os.Open(source)
	if err != nil {
		log.Fatal("while copying " + source + " to " + destination + " : " + err.Error())
	}
	defer from.Close()

	to, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal("while copying " + source + " to " + destination + " : " + err.Error())
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal("while copying " + source + " to " + destination + " : " + err.Error())
	}
}

// deleteFile deletes a file
func deleteFile(path string) {
	log.Info("Cleaning up... deleting " + path)
	if err := os.Remove(path); err != nil {
		log.Fatal("Could not delete file: " + path)
	}
}

// notifySlack sends a JSON formatted message to Slack over a webhook url
// It takes the content of the message (what changes helmsman is going to do or have done separated by \n)
// and the webhook URL as well as a flag specifying if this is a failure message or not
// It returns true if the sending of the message is successful, otherwise returns false
func notifySlack(content string, url string, failure bool, executing bool) bool {
	log.Info("Posting notifications to Slack ... ")

	color := "#36a64f" // green
	if failure {
		color = "#FF0000" // red
	}

	var contentBold string
	var pretext string

	if content == "" {
		pretext = "*No actions to perform!*"
	} else if failure {
		pretext = "*Failed to generate/execute a plan: *"
		contentTrimmed := strings.TrimSuffix(content, "\n")
		contentBold = "*" + contentTrimmed + "*"
	} else if executing && !failure {
		pretext = "*Here is what I have done: *"
		contentBold = "*" + content + "*"
	} else {
		pretext = "*Here is what I am going to do: *"
		contentSplit := strings.Split(content, "\n")
		for i := range contentSplit {
			contentSplit[i] = "* *" + contentSplit[i] + "*"
		}
		contentBold = strings.Join(contentSplit, "\n")
	}

	t := time.Now().UTC()

	var jsonStr = []byte(`{
		"attachments": [
			{
				"fallback": "Helmsman results.",
				"color": "` + color + `" ,
				"pretext": "` + pretext + `",
				"text": "` + contentBold + `",
				"footer": "Helmsman ` + appVersion + `",
				"ts": ` + strconv.FormatInt(t.Unix(), 10) + `,
				"mrkdwn_in": ["text","pretext"]
			}
		]
	}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Errorf("Failed to send slack message: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Failed to send notification to slack: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// getBucketElements returns a map containing the bucket name and the file path inside the bucket
// this func works for S3, Azure and GCS bucket links of the format:
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
		log.Fatal(err.Error())
	}
}

// Indent inserts prefix at the beginning of each non-empty line of s. The
// end-of-line marker is NL.
func Indent(s, prefix string) string {
	var res []byte
	bol := true
	for _, c := range []byte(s) {
		if bol && c != '\n' {
			res = append(res, []byte(prefix)...)
		}
		res = append(res, c)
		bol = c == '\n'
	}
	return string(res)
}

// concat appends all slices to a single slice
func concat(slices ...[]string) []string {
	slice := []string{}
	for _, item := range slices {
		slice = append(slice, item...)
	}
	return slice
}

func writeStringToFile(filename string, data string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}

// decrypt a eyaml or helm secret file
func decryptSecret(name string) error {
	cmd := helmBin
	args := []string{"secrets", "dec", name}

	if settings.EyamlEnabled {
		cmd = "eyaml"
		args = []string{"decrypt", "-f", name}
		if settings.EyamlPrivateKeyPath != "" && settings.EyamlPublicKeyPath != "" {
			args = append(args, []string{"--pkcs7-private-key", settings.EyamlPrivateKeyPath, "--pkcs7-public-key", settings.EyamlPublicKeyPath}...)
		}
	}

	command := Command{
		Cmd:         cmd,
		Args:        args,
		Description: "Decrypting " + name,
	}

	result := command.Exec()
	if !settings.EyamlEnabled {
		_, fileNotFound := os.Stat(name + ".dec")
		if fileNotFound != nil && !isOfType(name, []string{".dec"}) {
			return errors.New(result.errors)
		}
	}

	if result.code != 0 {
		return errors.New(result.errors)
	} else if result.errors != "" {
		return errors.New(result.errors)
	}

	if settings.EyamlEnabled {
		var outfile string
		if isOfType(name, []string{".dec"}) {
			outfile = name
		} else {
			outfile = name + ".dec"
		}
		err := writeStringToFile(outfile, result.output)
		if err != nil {
			log.Fatal("Can't write [ " + outfile + " ] file")
		}
	}
	return nil
}

// isLocalChart checks if a chart specified in the DSF is a local directory or not
func isLocalChart(chart string) bool {
	_, err := os.Stat(chart)
	return err == nil
}

// isValidCert checks if a certificate/key path/URI is valid
func isValidCert(value string) bool {
	if _, err := os.Stat(value); err != nil {
		_, err1 := url.ParseRequestURI(value)
		if err1 != nil || (!strings.HasPrefix(value, "s3://") && !strings.HasPrefix(value, "gs://") && !strings.HasPrefix(value, "az://")) {
			return false
		}
	}
	return true
}

// isValidFile checks if the file exists in the given path or accessible via http and is of allowed file extension (e.g. yaml, json ...)
func isValidFile(filePath string, allowedFileTypes []string) error {
	if strings.HasPrefix(filePath, "http") {
		if _, err := url.ParseRequestURI(filePath); err != nil {
			return fmt.Errorf("%s must be valid URL path to a raw file", filePath)
		}
	} else if _, pathErr := os.Stat(filePath); pathErr != nil {
		return fmt.Errorf("%s must be valid relative (from dsf file) file path", filePath)
	} else if !isOfType(filePath, allowedFileTypes) {
		return fmt.Errorf("%s must be of one the following file formats: %s", filePath, strings.Join(allowedFileTypes, ", "))
	}
	return nil
}
