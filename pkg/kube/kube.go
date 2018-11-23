package kube

import "time"

// TokenRequest request to authenticate with AWS credentials
type TokenRequest struct {
	ClusterName        string `json:"clusterName"`
	AwsAccessKeyId     string `json:"awsAccessKeyId"`
	AwsSecretAccessKey string `json:"awsSecretAccessKey"`
}

// TokenResponse token response
type TokenResponse struct {
	APIVersion string      `json:"apiVersion"`
	Kind       string      `json:"kind"`
	Status     TokenStatus `json:"status"`
}

// TokenStatus token status
type TokenStatus struct {
	Token               string    `json:"token"`
	ExpirationTimestamp time.Time `json:"expirationTimestamp"`
}
