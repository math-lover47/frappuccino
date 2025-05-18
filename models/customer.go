package models

import (
	"frappuccino/utils"
)

type Customer struct {
	CustomerId  utils.TEXT  `json:"customer_id"`
	FullName    utils.TEXT  `json:"full_name"`
	PhoneNumber utils.TEXT  `json:"phone_number"`
	Email       utils.TEXT  `json:"email"`
	Preferences utils.JSONB `json:"preferences"`
	CreatedAt   utils.TIME  `json:"created_at"`
	UpdatedAt   utils.TIME  `json:"updated_at"`
}
