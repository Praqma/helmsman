package gcs

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	// Imports the Google Cloud Storage client package.
	"cloud.google.com/go/storage"
	"github.com/logrusorgru/aurora"
	"golang.org/x/net/context"
)

// Auth checks for GCLOUD_CREDENTIALS in the environment
// returns true if they exist and creates a json credentials file and sets the GOOGLE_APPLICATION_CREDENTIALS env var
// returns false if credentials are not found
func Auth() bool {
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
		log.Println("INFO: GOOGLE_APPLICATION_CREDENTIALS is already set in the environment.")
		return true
	}

	if os.Getenv("GCLOUD_CREDENTIALS") != "" {
		credFile := "/tmp/gcloud_credentials.json"
		// write the credentials content into a json file
		d := []byte(os.Getenv("GCLOUD_CREDENTIALS"))
		err := ioutil.WriteFile(credFile, d, 0644)

		if err != nil {
			log.Fatal(aurora.Bold(aurora.Red("ERROR: Cannot create credentials file: " + err.Error())))
		}

		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credFile)
		return true
	}
	return false
}

// ReadFile reads a file from storage bucket and saves it in a desired location.
func ReadFile(bucketName string, filename string, outFile string) {
	if !Auth() {
		log.Fatal(aurora.Bold(aurora.Red("ERROR: Failed to find the GCLOUD_CREDENTIALS env var. Please make sure it is set in the environment.")))
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(aurora.Bold(aurora.Red("ERROR: Failed to configure Storage bucket: " + err.Error())))
	}
	storageBucket := client.Bucket(bucketName)

	// Creates an Object handler for our file
	obj := storageBucket.Object(filename)

	// Read the object.
	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatal(aurora.Bold(aurora.Red("ERROR: Failed to create object reader: " + err.Error())))
	}
	defer r.Close()

	// create output file and write to it
	var writers []io.Writer
	file, err := os.Create(outFile)
	if err != nil {
		log.Fatal(aurora.Bold(aurora.Red("ERROR: Failed to create an output file: " + err.Error())))
	}
	writers = append(writers, file)
	defer file.Close()

	dest := io.MultiWriter(writers...)
	if _, err := io.Copy(dest, r); err != nil {
		log.Fatal(aurora.Bold(aurora.Red("ERROR: Failed to read object content: " + err.Error())))
	}
	log.Println("INFO: Successfully downloaded " + filename + " from GCS as " + outFile)
}
