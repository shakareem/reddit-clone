package handlers

import (
	"encoding/json"
	"net/http"
)

func Error(w http.ResponseWriter, err string, code int) {
	resp, _ := json.Marshal(map[string]any{
		"status": code,
		"error":  err,
	})

	http.Error(w, string(resp), code)
}
