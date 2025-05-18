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
	var service Base
	service.CustomerService = NewCustomerService(repo.CustomerRepo)
	service.AggregationService = NewAggregationService(repo.AggregationRepo)
	service.InventoryService = NewInventoryService(repo.InventoryRepo)
	service.MenuService = NewMenuService(repo.MenuRepo)
	service.OrderService = NewOrderService(repo.OrderRepo)
	return &service
}
