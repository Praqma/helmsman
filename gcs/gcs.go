package gcs

import (
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
		d := []byte(os.Getenv("GCLOUD_CREDENTIALS"))
		err := ioutil.WriteFile("/tmp/credentials.json", d, 0644)

		if err != nil {
			log.Fatal("ERROR: Cannot create credentials file: ", err)
		}

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
