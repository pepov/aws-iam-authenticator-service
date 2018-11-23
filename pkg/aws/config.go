package aws

import "fmt"

// GenerateConfig generates an AWS config file
func GenerateConfig(awsAccessKeyID, awsSecretAccessKey string) string {
	return fmt.Sprintf(`
[default]
aws_access_key_id=%s
aws_secret_access_key=%s
`,
		awsAccessKeyID,
		awsSecretAccessKey)
}
