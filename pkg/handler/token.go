package handler

import (
	"encoding/json"
	"fmt"
	"github.com/hortonworks/aws-iam-authenticator-service/pkg/kube"
	"net/http"
	"os"

	"github.com/hortonworks/aws-iam-authenticator-service/pkg/aws"
	"github.com/hortonworks/aws-iam-authenticator-service/pkg/util"
	log "github.com/sirupsen/logrus"
)

// Token converts input parameters to an EKS token
func Token(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var tokenRequest kube.TokenRequest
	if err := decoder.Decode(&tokenRequest); err != nil {
		log.Errorf("Invalid request, err: %s", err.Error())
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid token request: %s", err.Error()))
		return
	}
	clusterName := tokenRequest.ClusterName
	if len(clusterName) == 0 {
		writeError(w, http.StatusBadRequest, "missing clusterName")
		return
	}
	awsAccessKeyID := tokenRequest.AwsAccessKeyId
	if len(awsAccessKeyID) == 0 {
		writeError(w, http.StatusBadRequest, "missing awsAccessKeyID")
		return
	}
	awsSecretAccessKey := tokenRequest.AwsSecretAccessKey
	if len(awsSecretAccessKey) == 0 {
		writeError(w, http.StatusBadRequest, "missing awsSecretAccessKey")
		return
	}
	log.Debugf("Generating config file for: %s", clusterName)
	configContent := aws.GenerateConfig(awsAccessKeyID, awsSecretAccessKey)
	var err error
	dir, configFile, err := util.CreateTemporaryFile(clusterName, configContent)
	if err != nil {
		writeError(w, http.StatusUnauthorized, fmt.Sprintf("Failed to create config file: %s", err.Error()))
		return
	}
	defer func() {
		err = os.RemoveAll(dir)
		if err != nil {
			log.Error(err.Error())
		}
	}()
	log.Debugf("Requesting token for: %s", clusterName)
	token, err := aws.GetEKSToken(clusterName, configFile)
	if err != nil {
		writeError(w, http.StatusUnauthorized, fmt.Sprintf("Failed to get token: %s", err.Error()))
		return
	}
	if token == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	response, err := json.Marshal(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, fmt.Sprintf("Failed to marshall token: %s", err.Error()))
		return
	}
	_, err = fmt.Fprint(w, string(response))
	if err != nil {
		log.Error(err.Error())
	}
}

func writeError(w http.ResponseWriter, statusCode int, statusReason string) {
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(statusReason)); err != nil {
		log.Errorf("Couldn't write response: %s", err.Error())
	}
}
