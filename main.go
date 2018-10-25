package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	heptio "github.com/heptio/authenticator/pkg/token"
	log "github.com/sirupsen/logrus"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Debug(r)
			os.Exit(1)
		}
	}()

	http.HandleFunc("/", handler)
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

const awsCredentialsDirBase = "./management-State"
const awsCredentialsDirectory = awsCredentialsDirBase + "/aws"
const awsCredentialsPath = awsCredentialsDirectory + "/credentials"
const awsSharedCredentialsFile = "AWS_SHARED_CREDENTIALS_FILE"

func getEKSToken(clusterName, awsAccessKeyID, awsSecretAccessKey string) (string, error) {
	generator, err := heptio.NewGenerator()
	if err != nil {
		return "", fmt.Errorf("error creating generator: %v", err)
	}

	os.Setenv(awsSharedCredentialsFile, awsCredentialsPath)

	defer func() {
		os.Remove(awsCredentialsPath)
		os.Remove(awsCredentialsDirectory)
		os.Remove(awsCredentialsDirBase)
		os.Unsetenv(awsSharedCredentialsFile)
	}()
	err = os.MkdirAll(awsCredentialsDirectory, 0744)
	if err != nil {
		return "", fmt.Errorf("error creating credentials directory: %v", err)
	}

	err = ioutil.WriteFile(awsCredentialsPath, []byte(fmt.Sprintf(
		`[default]
aws_access_key_id=%v
aws_secret_access_key=%v`,
		awsAccessKeyID,
		awsSecretAccessKey)), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing credentials file: %v", err)
	}

	return generator.Get(clusterName)
}
