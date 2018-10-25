package kube

// TokenResponse token response
type TokenResponse struct {
	APIVersion string      `json:"apiVersion"`
	Kind       string      `json:"kind"`
	Status     TokenStatus `json:"status"`
}

// TokenStatus token status
type TokenStatus struct {
	Token string `json:"token"`
}
