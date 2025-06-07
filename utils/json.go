package utils

import (
	"encoding/json"
	"net/http"

	"github.com/charmbracelet/log"
)

func RespondError(w http.ResponseWriter, logger *log.Logger, code int, msg string, err error) {
	if err != nil {
		logger.Error(err)
		logger.Printf("Responding with %d error: %v", code, msg)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}
	RespondJSON(w, logger, code, errorResponse{
		Error: msg,
	})
}

func RespondJSON(w http.ResponseWriter, logger *log.Logger, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func Respond(w http.ResponseWriter, logger *log.Logger, code int, payload string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// convert payload to byte slice
	if payload == "" {
		payload = "{}"
	}
	w.Write([]byte(payload))
}
