package models

import "frappuccino/utils"

type MenuItems struct {
	MenuItemId      utils.TEXT    `json:"menu_item_id"`
	ItemName        utils.TEXT    `json:"item_name"`
	ItemDescription utils.TEXT    `json:"item_description"`
	Price           utils.DEC     `json:"price"`
	Categories      utils.TEXTARR `json:"categories"`
	Ingredients     []Ingredients
	CreatedAt       utils.TIME `json:"created_at"`
	UpdatedAt       utils.TIME `json:"updated_at"`
}

type MenuItemsIngredients struct {
	MenuItemIngredientId utils.TEXT `json:"menu_item_ingredient_id"`
	MenuItemId           utils.TEXT `json:"menu_item_id"`
	IngredientId         utils.TEXT `json:"ingredient_id"`
	IngredientName       utils.TEXT `json:"ingredient_name"`
	Quantity             utils.DEC  `json:"quantity"`
}

type Ingredients struct {
	IngredientName string  `json:"ingredient_name"`
	Quantity       float64 `json:"quantity"`
}

func (m *MenuItems) Marshal(menu *MenuItems) {
	menu.MenuItemId = m.MenuItemId
	menu.ItemName = m.ItemName
	menu.ItemDescription = m.ItemDescription
	menu.Price = m.Price
	menu.Categories = m.Categories
	menu.CreatedAt = m.CreatedAt
	menu.UpdatedAt = m.UpdatedAt
}

func (m *MenuItems) Unmarshal(menu *MenuItems) {
	m.MenuItemId = menu.MenuItemId
	m.ItemName = menu.ItemName
	m.ItemDescription = menu.ItemDescription
	m.Price = menu.Price
	m.Categories = menu.Categories
	m.CreatedAt = menu.CreatedAt
	m.UpdatedAt = menu.UpdatedAt
}
