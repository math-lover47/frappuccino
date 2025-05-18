package models

import "frappuccino/utils"

// type OrderStatus string

// type PaymentMethod string

// const (
// 	OrderStatusPending   OrderStatus = "PENDING"
// 	OrderStatusCompleted OrderStatus = "COMPLETED"
// 	OrderStatusCancelled OrderStatus = "CANCELLED"
// )

// const (
// 	PaymentMethodCash PaymentMethod = "CASH"
// 	PaymentMethodCard PaymentMethod = "CARD"
// )

type Orders struct {
	OrderId             utils.TEXT `json:"order_id"`
	CustomerId          utils.TEXT `json:"customer_id"`
	SpecialInstructions utils.TEXT `json:"special_instructions"`
	TotalPrice          utils.DEC  `json:"total_price"`
	OrderStatus         utils.TEXT `json:"order_status"`
	PaymentMethod       utils.TEXT `json:"payment_method"`
	CreatedAt           utils.TIME `json:"created_at"`
	UpdatedAt           utils.TIME `json:"updated_at"`
}

type OrderItems struct {
	OrderItemId    utils.TEXT  `json:"order_item_id"`
	MenuItemId     utils.TEXT  `json:"menu_item_id"`
	OrderId        utils.TEXT  `json:"order_id"`
	Customizations utils.JSONB `json:"customizations"`
	ItemName       utils.TEXT  `json:"item_name"`
	Quantity       utils.DEC   `json:"quantity"`
	UnitPrice      utils.DEC   `json:"unit_price"`
	TotalPrice     utils.DEC   `json:"total_price"`
}

type OrderStatusHistory struct {
	OrderStatusHistoryId utils.TEXT `json:"order_status_history"`
	OrderId              utils.TEXT `json:"order_id"`
	Notes                utils.TEXT `json:"notes"`
	OrderStatus          utils.TEXT `json:"order_status"`
	UpdatedAt            utils.TIME `json:"updated_at"`
}
