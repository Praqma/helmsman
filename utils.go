package main

import (
  "bytes"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "os"
  "path/filepath"
  "strconv"
  "strings"

  "github.com/BurntSushi/toml"
  "github.com/Praqma/helmsman/aws"
  "github.com/Praqma/helmsman/gcs"
)

// printMap prints to the console any map of string keys and values.
func printMap(m map[string]string) {
  for key, value := range m {
    fmt.Println(key, " : ", value)
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

  if _, err := toml.DecodeFile(file, s); err != nil {
    return false, err.Error()
  }
  return true, "INFO: Parsed [[ " + file + " ]] successfully and found [ " + strconv.Itoa(len(s.Apps)) + " ] apps."

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
    log.Fatal("ERROR: failed to read [ " + filepath + " ] file content: " + err.Error())
  }
  return string(data)
}

// printHelp prints Helmsman commands
func printHelp() {
  fmt.Println("Helmsman version: " + version)
  fmt.Println("Helmsman is a Helm Charts as Code tool which allows you to automate the deployment/management of your Helm charts.")
  fmt.Println("Usage: helmsman [options]")
  fmt.Println()
  fmt.Println("Options:")
  fmt.Println("--f                specifies the desired state TOML file.")
  fmt.Println("--debug            prints basic logs during execution.")
  fmt.Println("--apply            generates and applies an action plan.")
  fmt.Println("--verbose          prints more verbose logs during execution.")
  fmt.Println("--ns-override      override defined namespaces with a provided one.")
  fmt.Println("--skip-validation  generates and applies an action plan.")
  fmt.Println("--help             prints Helmsman help.")
  fmt.Println("--v                prints Helmsman version.")
}

// logVersions prints the versions of kubectl and helm to the logs
func logVersions() {
  cmd := command{
    Cmd:         "bash",
    Args:        []string{"-c", "kubectl version"},
    Description: "Kubectl version: ",
  }

  exitCode, result := cmd.exec(debug, false)
  if exitCode != 0 {
    log.Fatal("ERROR: while checking kubectl version: " + result)
  }

  log.Println("VERBOSE: kubectl version: \n " + result + "\n")

  cmd = command{
    Cmd:         "bash",
    Args:        []string{"-c", "helm version"},
    Description: "Helm version: ",
  }

  exitCode, result = cmd.exec(debug, false)
  if exitCode != 0 {
    log.Fatal("ERROR: while checking helm version: " + result)
  }
  log.Println("VERBOSE: helm version: \n" + result + "\n")
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

// validateServiceAccount checks if k8s service account exists in a given namespace
func validateServiceAccount(sa string, namespace string) (bool, string) {
  if namespace == "" {
    namespace = "default"
  }
  ns := " -n " + namespace

  cmd := command{
    Cmd:         "bash",
    Args:        []string{"-c", "kubectl get serviceaccount " + sa + ns},
    Description: "validating that serviceaccount [ " + sa + " ] exists in namespace [ " + namespace + " ].",
  }

  if exitCode, err := cmd.exec(debug, verbose); exitCode != 0 {
    return false, err
  }
  return true, ""
}

// downloadFile downloads a file from GCS or AWS buckets and name it with a given outfile
// if downloaded, returns the outfile name. If the file path is local file system path, it is returned as is.
func downloadFile(path string, outfile string) string {
  if strings.HasPrefix(path, "s3") {

    tmp := getBucketElements(path)
    aws.ReadFile(tmp["bucketName"], tmp["filePath"], outfile)

  } else if strings.HasPrefix(path, "gs") {

    tmp := getBucketElements(path)
    gcs.ReadFile(tmp["bucketName"], tmp["filePath"], outfile)

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
    log.Fatal("ERROR: while copying " + source + " to " + destination + " : " + err.Error())
  }
  defer from.Close()

  to, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, 0666)
  if err != nil {
    log.Fatal("ERROR: while copying " + source + " to " + destination + " : " + err.Error())
  }
  defer to.Close()

  _, err = io.Copy(to, from)
  if err != nil {
    log.Fatal("ERROR: while copying " + source + " to " + destination + " : " + err.Error())
  }
}

// deleteFile deletes a file
func deleteFile(path string) {
  log.Println("INFO: cleaning up ... deleting " + path)
  if err := os.Remove(path); err != nil {
    log.Fatal("ERROR: could not delete file: " + path)
  }
}

