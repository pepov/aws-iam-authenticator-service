package handler

import (
	"fmt"
	"net/http"
)

// Status responses the status of the service
func Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}
