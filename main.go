package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	heptio "github.com/kubernetes-sigs/aws-iam-authenticator"
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

func getEKSToken(clusterName, awsAccessKeyID, awsSecretAccessKey string) (string, error) {
	generator, err := heptio.NewGenerator()
	if err != nil {
		return "", fmt.Errorf("error creating generator: %v", err)
	}

	defer awsCredentialsLocker.Unlock()
	awsCredentialsLocker.Lock()
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
		awsAccessKeyId,
		awsSecretAccessKey)), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing credentials file: %v", err)
	}

	return generator.Get(clusterName)
}
