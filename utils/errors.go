package utils

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrIdNotFound     = errors.New("ID not found")
	ErrConflictFields = errors.New("Conflict duplicate fields")
	ErrMenuItem       = errors.New("Menu Item does not exist")

	ErrInvalidQuantity       = errors.New("quantity cannot be negative")
	ErrInvalidReorderLevel   = errors.New("reorder level cannot be negative")
	ErrInvalidIngredientId   = errors.New("Id be positive")
	ErrInvalidIngredientName = errors.New("ingredient name cannot be empty")
)

type APIError struct {
	Code     INT  `json:"code"`
	Message  TEXT `json:"message"`
	Resource TEXT `json:"resource"`
}

func (response *APIError) Send(w http.ResponseWriter) {
	j, err := json.MarshalIndent(response, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.Code))
	w.Write(j)
}
