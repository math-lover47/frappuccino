package models

import "frappuccino/utils"

type MenuItems struct {
	MenuItemId      utils.TEXT    `json:"menu_item_id"`
	ItemName        utils.TEXT    `json:"item_name"`
	ItemDescription utils.TEXT    `json:"item_description"`
	Price           utils.DEC     `json:"price"`
	Categories      utils.TEXTARR `json:"categories"`
	CreatedAt       utils.TIME    `json:"created_at"`
	UpdatedAt       utils.TIME    `json:"updated_at"`
}

type MenuItemsIngredients struct {
	MenuItemIngredientId utils.TEXT `json:"menu_item_ingredient_id"`
	MenuItemId           utils.TEXT `json:"menu_item_id"`
	IngredientId         utils.TEXT `json:"ingredient_id"`
	IngredientName       utils.TEXT `json:"ingredient_name"`
	Quantity             utils.DEC  `json:"quantity"`
}

// func (m *Menu) Marshal(dtoMenu *dto.Menu) {
// }

// func (m *Menu) Unmarshal(dtoMenu *dto.Menu) {
// }
