package main

import (
	"net/http"

	"github.com/mhmxs/aws-iam-authenticator-service/pkg/handler"
	log "github.com/sirupsen/logrus"
)

var (
	Version   string
	BuildTime string
	Debug     bool
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	if true {
		log.SetLevel(log.DebugLevel)
	}

	http.HandleFunc("/", handler.Token)
	http.HandleFunc("/status", handler.Status)
	log.Debug("start listening on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
