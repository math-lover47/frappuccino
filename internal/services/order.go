package services

import "frappuccino/internal/repo"

type OrderServiceIfc interface{}

type OrderService struct {
	OrderRepo repo.OrderRepoIfc
}

func NewOrderService(OrderRepo repo.OrderRepoIfc) *OrderService {
	return &OrderService{OrderRepo: OrderRepo}
}
