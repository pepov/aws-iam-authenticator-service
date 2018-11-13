package aws

import (
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	awsAccessKeyIDEnv           = "AWS_ACCESS_KEY_ID"
	awsSecretAccessKeyEnv       = "AWS_SECRET_ACCESS_KEY"
	awsProfileEnv               = "AWS_PROFILE"
	awsSharedCredentialsFileEnv = "AWS_SHARED_CREDENTIALS_FILE"
)

func init() {
	for _, k := range []string{awsAccessKeyIDEnv, awsSecretAccessKeyEnv, awsProfileEnv, awsSharedCredentialsFileEnv} {
		log.Debugf("Unset env variable: %s", k)
		err := os.Unsetenv(k)
		if err != nil {
			panic("Unable to unset variable: " + k)
		}
	}
}
