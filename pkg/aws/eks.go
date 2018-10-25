package aws

import (
	"fmt"
	"os"
	"sync"
	"syscall"

	"bou.ke/monkey"
	heptio "github.com/heptio/authenticator/pkg/token"
	"github.com/jtolds/gls"
	"github.com/mhmxs/aws-iam-authenticator-service/pkg/kube"
	log "github.com/sirupsen/logrus"
)

var threadLocal = gls.NewContextManager()

func init() {
	log.Debug("Replacing os.Getenv")
	monkey.Patch(os.Getenv, getEnv)
}

// GetEKSToken retrieves EKS token
func GetEKSToken(clusterName, configFile string) (*kube.TokenResponse, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	tokenChan := make(chan string)
	errChan := make(chan error)

	go threadLocal.SetValues(gls.Values{awsSharedCredentialsFileEnv: configFile}, func() {
		defer wg.Done()

		log.Debugf("Constructing new generator for: %s", clusterName)
		generator, err := heptio.NewGenerator()
		if err != nil {
			errChan <- fmt.Errorf("Error during generator initialization: %v", err)
			return
		}
		log.Debugf("Retrieving token for: %s", clusterName)
		token, err := generator.Get(clusterName)
		if err != nil {
			errChan <- fmt.Errorf("Error during token request: %v", err)
			return
		}
		log.Debugf("Token is retrieved for: %s", clusterName)
		tokenChan <- token
	})

	go func() {
		log.Debugf("Waiting for WaitGroup on: %s", clusterName)
		wg.Wait()
		log.Debugf("WaitGroup is done for: %s", clusterName)
		close(tokenChan)
		close(errChan)
	}()

	log.Debugf("Waiting for token for: %s", clusterName)
	select {
	case token, _ := <-tokenChan:
		return &kube.TokenResponse{
			APIVersion: "client.authentication.k8s.io/v1alpha1",
			Kind:       "ExecCredential",
			Status: kube.TokenStatus{
				Token: token,
			},
		}, nil
	case err, _ := <-errChan:
		return nil, err
	}
}

func getEnv(key string) string {
	if key == awsSharedCredentialsFileEnv {
		value, _ := threadLocal.GetValue(key)
		return value.(string)
	}
	v, _ := syscall.Getenv(key)
	return v
}
