package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mhmxs/aws-iam-authenticator-service/pkg/aws"
	"github.com/mhmxs/aws-iam-authenticator-service/pkg/util"
	log "github.com/sirupsen/logrus"
)

// Token converts input parameters to an EKS token
func Token(w http.ResponseWriter, r *http.Request) {
	var err error
	query := r.URL.Query()
	clusterName := query.Get("clusterName")
	if len(clusterName) == 0 {
		writeError(w, http.StatusBadRequest, "missing clusterName")
		return
	}
	awsAccessKeyID := query.Get("awsAccessKeyID")
	if len(awsAccessKeyID) == 0 {
		writeError(w, http.StatusBadRequest, "missing awsAccessKeyID")
		return
	}
	awsSecretAccessKey := query.Get("awsSecretAccessKey")
	if len(awsSecretAccessKey) == 0 {
		writeError(w, http.StatusBadRequest, "missing awsSecretAccessKey")
		return
	}
	log.Debugf("Generating config file for: %s", clusterName)
	configContent := aws.GenerateConfig(awsAccessKeyID, awsSecretAccessKey)
	dir, configFile, err := util.CreateTemporaryFile(clusterName, configContent)
	if err != nil {
		writeError(w, http.StatusUnauthorized, fmt.Sprintf("Failed to create config file: %s", err.Error()))
		return
	}
	defer os.RemoveAll(dir)
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
	fmt.Fprint(w, string(response))
}

func writeError(w http.ResponseWriter, statusCode int, statusReason string) {
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(statusReason)); err != nil {
		log.Errorf("Couldn't write response: %s", err.Error())
	}
}
