package handler

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Status responses the status of the service
func Status(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprint(w, "ok")
	if err != nil {
		log.Error(err.Error())
	}
}
