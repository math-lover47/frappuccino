package models

import "frappuccino/utils"

type Inventory struct {
	IngredientId   utils.TEXT `json:"ingredient_id"`
	IngredientName utils.TEXT `json:"ingredient_name"`
	Unit           utils.TEXT `json:"unit"`
	Quantity       utils.DEC  `json:"quantity"`
	ReorderLevel   utils.DEC  `json:"reorder_level"`
	CreatedAt      utils.TIME `json:"created_at"`
	UpdatedAt      utils.TIME `json:"updated_at"`
}

type InventoryTransactions struct {
	InventoryTransactionId     utils.TEXT `json:"inventory_transaction_id"`
	IngredientId               utils.TEXT `json:"ingredient_id"`
	InventoryTransactionAction utils.TEXT `json:"inventory_transaction_action"`
	Quantity                   utils.DEC  `json:"quantity"`
	CreatedAt                  utils.TIME `json:"created_at"`
}

type Page struct {
	CurrentPage int  `json:"current_page"`
	HasNextPage bool `json:"has_next_page"`
	PageSize    int  `json:"page_size"`
	TotalPages  int  `json:"total_pages"`
	Data        []Data
}

type Data struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
}
