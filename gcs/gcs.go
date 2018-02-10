package gcs

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"

	// Imports the Google Cloud Storage client package.
	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
)

func checkCredentialsEnvVar() bool {
	if os.Getenv("GCLOUD_CREDENTIALS") != "" {
		// write the credentials content into a json file
		file, err := os.Create("/tmp/credentials.json")
		if err != nil {
			log.Fatal("ERROR: Cannot create credentials file: ", err)
		} else {
			// making sure special characters are not escaped
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetEscapeHTML(false)
			enc.Encode(os.Getenv("GCLOUD_CREDENTIALS"))

			// writing the credentials content to a file
			err := ioutil.WriteFile(file.Name(), buf.Bytes(), 0644)
			if err != nil {
				log.Fatal("ERROR: Cannot write the credentials file: ", err)
			}
		}
		defer file.Close()

		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/credentials.json")
		return true
	}
	return false
}

// ReadFile reads a file from storage bucket and saves it in a desired location.
func ReadFile(bucketName string, filename string, outFile string) {
	if !checkCredentialsEnvVar() {
		log.Fatal("Failed to find the GCLOUD_CREDENTIALS env var. Please make sure it is set in the environment.")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to configure Storage bucket: %v", err)
	}
	storageBucket := client.Bucket(bucketName)

	// Creates an Object handler for our file
	obj := storageBucket.Object(filename)

	// Read the object.
	r, err := obj.NewReader(ctx)
	if err != nil {
		log.Fatalf("Failed to create object reader: %v", err)
	}
	defer r.Close()

	// create output file and write to it
	var writers []io.Writer
	file, err := os.Create(outFile)
	if err != nil {
		log.Fatalf("Failed to create an output file: %v", err)
	}
	writers = append(writers, file)
	defer file.Close()

	dest := io.MultiWriter(writers...)
	if _, err := io.Copy(dest, r); err != nil {
		log.Fatalf("Failed to read object content: %v", err)
	}
	log.Println("INFO: Successfully downloaded " + filename + " from GCS as " + outFile)
}
