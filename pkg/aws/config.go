package aws

import "fmt"

// GenerateConfig generates an AWS config file
func GenerateConfig(awsAccessKeyID, awsSecretAccessKey string) string {
	return fmt.Sprintf(
		`[default]
aws_access_key_id=%v
aws_secret_access_key=%v`,
		awsAccessKeyID,
		awsSecretAccessKey)
}
