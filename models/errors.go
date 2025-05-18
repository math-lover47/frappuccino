package models

import "errors"

var (
	ErrInvalidQuantity       = errors.New("quantity cannot be negative")
	ErrInvalidReorderLevel   = errors.New("reorder level cannot be negative")
	ErrInvalidIngredientId   = errors.New("Id be positive")
	ErrInvalidIngredientName = errors.New("ingredient name cannot be empty")
)

type APIError struct{}
