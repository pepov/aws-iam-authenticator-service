package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/mhmxs/aws-iam-authenticator-service/config"
	"github.com/mhmxs/aws-iam-authenticator-service/pkg/handler"
	log "github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	debug, err := strconv.ParseBool(config.Debug)
	if err != nil && debug {
		log.SetLevel(log.DebugLevel)
	}

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		log.Errorf("Invalid port number: %s fallback to 8080", config.Port)
		port = 8080
	}

	http.HandleFunc("/", handler.Token)
	http.HandleFunc("/status", handler.Status)
	log.Debugf("start listening on %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
