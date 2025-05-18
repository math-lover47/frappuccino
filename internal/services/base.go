package services

import "frappuccino/internal/repo"

type Base struct {
	CustomerService    CustomerServiceIfc
	InventoryService   InventoryServiceIfc
	MenuService        MenuServiceIfc
	OrderService       OrderServiceIfc
	AggregationService AggregationServiceIfc
}

func New(repo *repo.Repo) *Base {
	return &Base{
		CustomerService:    NewCustomerService(repo.CustomerRepo),
		InventoryService:   NewInventoryService(repo.InventoryRepo),
		MenuService:        NewMenuService(repo.MenuRepo),
		OrderService:       NewOrderService(repo.OrderRepo),
		AggregationService: NewAggregationService(repo.AggregationRepo),
	}
}
