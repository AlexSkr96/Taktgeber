package main

import (
	"net/http"
)

// handlerHealth is a simple liveness check — just confirms the server is up.
func (cfg *apiConfig) handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
