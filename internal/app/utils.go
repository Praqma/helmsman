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

	"gopkg.in/yaml.v2"

	"github.com/BurntSushi/toml"
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

	tomlFile := string(rawTomlFile)
	if !flags.noEnvSubst {
		if ok, err := validateEnvVars(tomlFile, file); !ok {
			return false, err
		}
		tomlFile = substituteEnv(tomlFile)
	}
	if !flags.noSSMSubst {
		tomlFile = substituteSSM(tomlFile)
	}
	if _, err := toml.Decode(tomlFile, s); err != nil {
		return false, err.Error()
	}
	resolvePaths(file, s)
	substituteVarsInStaticFiles(s)

	return true, "Parsed TOML [[ " + file + " ]] successfully and found [ " + strconv.Itoa(len(s.Apps)) + " ] apps"
}

// toTOML encodes a state type into a TOML file.
// It uses the BurntSuchi TOML parser.
func toTOML(file string, s *state) {
	log.Info("Printing generated toml ... ")
	var buff bytes.Buffer
	var (
		newFile *os.File
		err     error
	)

	if err := toml.NewEncoder(&buff).Encode(s); err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	newFile, err = os.Create(file)
	if err != nil {
		log.Fatal(err.Error())
	}
	bytesWritten, err := newFile.Write(buff.Bytes())
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(fmt.Sprintf("Wrote %d bytes.\n", bytesWritten))
	newFile.Close()
}

// fromYAML reads a yaml file and decodes it to a state type.
// parser which throws an error if the YAML file is not valid.
func fromYAML(file string, s *state) (bool, string) {
	rawYamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		return false, err.Error()
	}

	yamlFile := string(rawYamlFile)
	if !flags.noEnvSubst {
		if ok, err := validateEnvVars(yamlFile, file); !ok {
			return false, err
		}
		yamlFile = substituteEnv(yamlFile)
	}
	if !flags.noSSMSubst {
		yamlFile = substituteSSM(yamlFile)
	}

	if err = yaml.UnmarshalStrict([]byte(yamlFile), s); err != nil {
		return false, err.Error()
	}
	resolvePaths(file, s)
	substituteVarsInStaticFiles(s)

	return true, "Parsed YAML [[ " + file + " ]] successfully and found [ " + strconv.Itoa(len(s.Apps)) + " ] apps"
}

// toYaml encodes a state type into a YAML file
func toYAML(file string, s *state) {
	log.Info("Printing generated yaml ... ")
	var buff bytes.Buffer
	var (
		newFile *os.File
		err     error
	)

	if err := yaml.NewEncoder(&buff).Encode(s); err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	newFile, err = os.Create(file)
	if err != nil {
		log.Fatal(err.Error())
	}
	bytesWritten, err := newFile.Write(buff.Bytes())
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Info(fmt.Sprintf("Wrote %d bytes.\n", bytesWritten))
	newFile.Close()
}

// substituteVarsInStaticFiles loops through the values/secrets files and substitutes variables into them.
func substituteVarsInStaticFiles(s *state) {
	for _, v := range s.Apps {
		if v.ValuesFile != "" {
			v.ValuesFile = substituteVarsInYaml(v.ValuesFile)
		}
		if v.SecretsFile != "" {
			v.SecretsFile = substituteVarsInYaml(v.SecretsFile)
		}

		if len(v.Hooks) != 0 {
			for key, val := range v.Hooks {
				if key != "deleteOnSuccess" && key != "successTimeout" && key != "successCondition" {
					v.Hooks[key] = substituteVarsInYaml(val.(string))
				}
			}
		}

		if len(s.Settings.GlobalHooks) != 0 {
			for key, val := range s.Settings.GlobalHooks {
				if key != "deleteOnSuccess" && key != "successTimeout" && key != "successCondition" {
					s.Settings.GlobalHooks[key] = substituteVarsInYaml(val.(string))
				}
			}
		}

		for i := range v.ValuesFiles {
			v.ValuesFiles[i] = substituteVarsInYaml(v.ValuesFiles[i])
		}
		for i := range v.SecretsFiles {
			v.SecretsFiles[i] = substituteVarsInYaml(v.SecretsFiles[i])
		}
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
		if ok, err := validateEnvVars(yamlFile, file); !ok {
			log.Critical(err)
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

// resolvePaths resolves relative paths of certs/keys/chart/value file/secret files/etc and replace them with a absolute paths
func resolvePaths(relativeToFile string, s *state) {
	dir := filepath.Dir(relativeToFile)
	downloadDest, _ := filepath.Abs(createTempDir(tempFilesDir, "tmp"))
	for k, v := range s.Apps {
		if v.ValuesFile != "" {
			v.ValuesFile, _ = resolveOnePath(v.ValuesFile, dir, downloadDest)
		}
		if v.SecretsFile != "" {
			v.SecretsFile, _ = resolveOnePath(v.SecretsFile, dir, downloadDest)
		}

		if len(v.Hooks) != 0 {
			for key, val := range v.Hooks {
				if key != "deleteOnSuccess" && key != "successTimeout" && key != "successCondition" {
					v.Hooks[key], _ = resolveOnePath(val.(string), dir, downloadDest)
				}
			}
		}

		for i := range v.ValuesFiles {
			v.ValuesFiles[i], _ = resolveOnePath(v.ValuesFiles[i], dir, downloadDest)
		}
		for i := range v.SecretsFiles {
			v.SecretsFiles[i], _ = resolveOnePath(v.SecretsFiles[i], dir, downloadDest)
		}

		if v.Chart != "" {
			var repoOrDir = filepath.Dir(v.Chart)
			_, isRepo := s.HelmRepos[repoOrDir]
			isRepo = isRepo || stringInSlice(repoOrDir, s.PreconfiguredHelmRepos)
			if !isRepo {
				// if there is no repo for the chart, we assume it's intended to be a local path

				// support env vars in path
				v.Chart = os.ExpandEnv(v.Chart)
				// respect absolute paths to charts but resolve relative paths
				if !filepath.IsAbs(v.Chart) {
					v.Chart, _ = filepath.Abs(filepath.Join(dir, v.Chart))
				}
			}
		}
		s.Apps[k] = v
	}
	// resolving paths for Bearer Token path in settings
	if s.Settings.BearerTokenPath != "" {
		s.Settings.BearerTokenPath, _ = resolveOnePath(s.Settings.BearerTokenPath, dir, downloadDest)
	}
	// resolve paths for global hooks
	if len(s.Settings.GlobalHooks) != 0 {
		for key, val := range s.Settings.GlobalHooks {
			if key != "deleteOnSuccess" && key != "successTimeout" && key != "successCondition" {
				s.Settings.GlobalHooks[key], _ = resolveOnePath(val.(string), dir, downloadDest)
			}
		}
	}
	// resolving paths for k8s certificate files
	for k := range s.Certificates {
		s.Certificates[k], _ = resolveOnePath(s.Certificates[k], "", downloadDest)
	}

}

// resolveOnePath takes the input file (URL, cloud bucket, or local file relative path),
// the directory containing the DSF and the temp directory where files will be fetched to
// and downloads/fetches the file locally into helmsman temp directory and returns
// its absolute path
func resolveOnePath(file string, dir string, downloadDest string) (string, error) {
	destFileName := filepath.Join(downloadDest, path.Base(file))
	return filepath.Abs(downloadFile(file, dir, destFileName))
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
func validateEnvVars(s string, filename string) (bool, string) {
	if !flags.skipValidation {
		log.Info("validating environment variables in " + filename)
		var key string
		r, _ := regexp.Compile("\\${([a-zA-Z_][a-zA-Z0-9_-]*)}|\\$([a-zA-Z_][a-zA-Z0-9_-]*)")
		scanner := bufio.NewScanner(strings.NewReader(s))
		for scanner.Scan() {
			text := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(text, "#") {
				for _, v := range r.FindAllStringSubmatch(strings.ReplaceAll(text, "$$", "!?"), -1) {
					if v[1] != "" {
						key = v[1]
					} else {
						key = v[2]
					}
					if _, ok := os.LookupEnv(key); !ok {
						return false, v[0] + " is used as an env variable but is currently unset. Either set it or escape it like so: $" + v[0]
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Critical(err.Error())
		}
	}
	return true, ""
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

		log.Info("" + file + " will be used from local file system.")
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

// isLocalChart checks if a chart specified in the DSF is a local directory or not
func isLocalChart(chart string) bool {
	_, err := os.Stat(chart)
	return err == nil
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

	command := command{
		Cmd:         cmd,
		Args:        args,
		Description: "Decrypting " + name,
	}

	result := command.exec()
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

// isValidFile checks if the file exists in the given path or accessible via http and is of allowed file extension (e.g. yaml, json ...)
func isValidFile(filePath string, allowedFileTypes []string) error {
	if strings.HasPrefix(filePath, "http") {
		if _, err := url.ParseRequestURI(filePath); err != nil {
			return errors.New(filePath + " must be valid URL path to a raw file.")
		}
	} else if _, pathErr := os.Stat(filePath); pathErr != nil {
		return errors.New(filePath + " must be valid relative (from dsf file) file path.")
	} else if !isOfType(filePath, allowedFileTypes) {
		return errors.New(filePath + " must be of one the following file formats: " + strings.Join(allowedFileTypes, ", "))
	}
	return nil
}
