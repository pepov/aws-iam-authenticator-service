package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"syscall"

	"bou.ke/monkey"
	heptio "github.com/heptio/authenticator/pkg/token"
	"github.com/jtolds/gls"
	log "github.com/sirupsen/logrus"
)

const (
	awsAccessKeyIDEnv           = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnv       = "AWS_SECRET_ACCESS_KEY"
	awsSharedCredentialsFileEnv = "AWS_SHARED_CREDENTIALS_FILE"
	awsProfileEnv               = "AWS_PROFILE"
)

var (
	Version   string
	BuildTime string
	Debug     bool
)

var threadLocal = gls.NewContextManager()

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	if true {
		log.SetLevel(log.DebugLevel)
	}

	for _, k := range []string{awsAccessKeyIDEnv, awsSecretAccessKeyEnv, awsSharedCredentialsFileEnv, awsProfileEnv} {
		log.Debugf("Unset env variable %s", k)
		os.Unsetenv(k)
	}

	log.Debug("Replacing os.Getenv")
	monkey.Patch(os.Getenv, getEnv)

	http.HandleFunc("/", handler)
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("/status called")
		fmt.Fprint(w, "ok")
	})
	log.Debug("start listening on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Debug("/ called")
	var err error
	query := r.URL.Query()
	clusterName := query.Get("clusterName")
	if len(clusterName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		if _, err = w.Write([]byte("400 Bad Request: missing clusterName")); err != nil {
			log.Errorf("Couldn't write response %s", err.Error())
		}
		return
	}
	awsAccessKeyID := query.Get("awsAccessKeyID")
	if len(awsAccessKeyID) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		if _, err = w.Write([]byte("400 Bad Request: missing awsAccessKeyID")); err != nil {
			log.Errorf("Couldn't write response %s", err.Error())
		}
		return
	}
	awsSecretAccessKey := query.Get("awsSecretAccessKey")
	if len(awsSecretAccessKey) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		if _, err = w.Write([]byte("400 Bad Request: missing awsSecretAccessKey")); err != nil {
			log.Errorf("Couldn't write response %s", err.Error())
		}
		return
	}
	log.Debugf("Requesting token for %s", clusterName)
	token, err := getEKSToken(clusterName, awsAccessKeyID, awsSecretAccessKey)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get token %s", err.Error())
		log.Error(errorMsg)
		w.WriteHeader(http.StatusUnauthorized)
		if _, err = w.Write([]byte("401 Unauthorized: " + errorMsg)); err != nil {
			log.Errorf("Couldn't write response %s", err.Error())
		}
		return
	}
	if len(token) == 0 {
		w.WriteHeader(http.StatusNoContent)
	}
	fmt.Fprint(w, token)
}

func getEnv(key string) string {
	if key == awsSharedCredentialsFileEnv {
		value, _ := threadLocal.GetValue(key)
		return value.(string)
	}
	v, _ := syscall.Getenv(key)
	return v
}

func getEKSToken(clusterName, awsAccessKeyID, awsSecretAccessKey string) (string, error) {
	dir, err := ioutil.TempDir("", clusterName)
	if err != nil {
		log.Error(err.Error())
		return "", fmt.Errorf("Error during temp dir creation: %v", err)
	}
	defer os.RemoveAll(dir)

	configFile := dir + "/config"

	err = ioutil.WriteFile(configFile, []byte(fmt.Sprintf(
		`[default]
aws_access_key_id=%v
aws_secret_access_key=%v`,
		awsAccessKeyID,
		awsSecretAccessKey)), 0600)
	if err != nil {
		log.Error(err.Error())
		return "", fmt.Errorf("Error during writing credentials file: %v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	tokenChan := make(chan string)
	errChan := make(chan error)

	go threadLocal.SetValues(gls.Values{awsSharedCredentialsFileEnv: configFile}, func() {
		defer wg.Done()

		log.Debugf("Constructing new generator for %s", clusterName)
		generator, err := heptio.NewGenerator()
		if err != nil {
			log.Error(err.Error())
			errChan <- fmt.Errorf("Error during generator initialization: %v", err)
			return
		}
		log.Debugf("Retrieving token for %s", clusterName)
		token, err := generator.Get(clusterName)
		if err != nil {
			log.Error(err.Error())
			errChan <- fmt.Errorf("Error during token request: %v", err)
			return
		}
		log.Debugf("Token is retrieved for %s", clusterName)
		tokenChan <- token
	})

	go func() {
		log.Debugf("Waiting for WaitGroup on %s", clusterName)
		wg.Wait()
		log.Debugf("WaitGroup is done for %s", clusterName)
		close(tokenChan)
		close(errChan)
	}()

	log.Debugf("Waiting for token for %s", clusterName)
	select {
	case token, _ := <-tokenChan:
		return token, nil
	case err, _ := <-errChan:
		return "", err
	}
}
