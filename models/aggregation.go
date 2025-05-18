package models

// TotalSales
type TotalSales struct {
	Value float64 `json:"total_sales"`
}

// Popular Items
type PopularItems struct {
	Items []PopularItem `json:"popular_items"`
}

type PopularItem struct {
	ItemName     string `json:"item_name"`
	OrderedTimes int    `json:"ordered_times"`
}

// Search
type Search struct {
	MenuItems  []SearchMenu  `json:"menu_items"`
	OrderItems []SearchOrder `json:"orders"`
}

type SearchOrder struct {
	Id           string `json:"id"`
	CustomerName string `json:"customer_name"`
	ItemName     string `json:"item_name"`
	ItemPrice    string `json:"item_price"`
}

type SearchMenu struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description:"`
	Price       float64 `json:"price"`
}

// OrderedItemsByPeriod

type ListOrderedItemByPeriods struct {
	Period string                `json:"period"`
	Month  string                `json:"month"`
	Year   string                `json:"year"`
	Items  []OrderedItemByPeriod `json:"orderedItems"`
}

type OrderedItemByPeriod struct {
	Date  string
	Count int
}
