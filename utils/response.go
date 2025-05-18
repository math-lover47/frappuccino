package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Code    INT  `json:"code"`
	Message TEXT `json:"message"`
}

func (response *APIResponse) Send(w http.ResponseWriter) {
	j, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.Code))
	w.Write(j)
}
