package handlers

import (
	"encoding/json"
	"net/http"
)

type RequestError struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Value    string `json:"value"`
	Message  string `json:"msg"`
}

func jsonError(w http.ResponseWriter, code int, errs []RequestError) {
	resp, _ := json.Marshal(map[string]any{
		"errors": errs,
	})

	http.Error(w, string(resp), code)
}
