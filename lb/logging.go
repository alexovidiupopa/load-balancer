package lb

import (
	"log"
	"net/http"
)

func logRequest(r *http.Request) {
	log.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
}
