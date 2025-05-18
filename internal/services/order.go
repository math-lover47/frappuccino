package services

import (
	"context"
	"frappuccino/internal/repo"
	"frappuccino/models"
	"log"
)

type OrderServiceIfc interface {
	Create(ctx context.Context, order *models.Orders) (*models.Orders, error)
	GetAll(ctx context.Context) ([]models.Orders, error)
	GetByID(ctx context.Context, orderId string) (models.Orders, error)
	UpdateByID(ctx context.Context, order *models.Orders) error
	DeleteByID(ctx context.Context, orderId string) error
}

type OrderService struct {
	OrderRepo repo.OrderRepoIfc
}

func NewOrderService(OrderRepo repo.OrderRepoIfc) *OrderService {
	return &OrderService{OrderRepo: OrderRepo}
}

// Create создает новый заказ
func (os *OrderService) Create(ctx context.Context, order *models.Orders) (*models.Orders, error) {
	log.Println("Creating new order for customer:", order.CustomerId)
	createdOrder, err := os.OrderRepo.Create(ctx, order)
	if err != nil {
		log.Println("Error creating order:", err)
		return nil, err
	}
	log.Println("Order created successfully:", createdOrder.OrderId)
	return createdOrder, nil
}

// GetAll возвращает все заказы
func (os *OrderService) GetAll(ctx context.Context) ([]models.Orders, error) {
	log.Println("Fetching all orders")
	orders, err := os.OrderRepo.GetAll(ctx)
	if err != nil {
		log.Println("Error fetching orders:", err)
		return nil, err
	}
	log.Printf("Retrieved %d orders", len(orders))
	return orders, nil
}

// GetByID возвращает заказ по ID
func (os *OrderService) GetByID(ctx context.Context, orderId string) (models.Orders, error) {
	log.Printf("Fetching order by ID: %s", orderId)
	order, err := os.OrderRepo.GetOrderByID(ctx, orderId)
	if err != nil {
		log.Println("Error fetching order:", err)
		return models.Orders{}, err
	}
	log.Printf("Retrieved order [%s] for customer [%s]", order.OrderId, order.CustomerId)
	return order, nil
}

// UpdateByID обновляет заказ по ID
func (os *OrderService) UpdateByID(ctx context.Context, order *models.Orders) error {
	log.Printf("Updating order [%s]", order.OrderId)
	err := os.OrderRepo.UpdateItemByID(ctx, order)
	if err != nil {
		log.Println("Error updating order:", err)
		return err
	}
	log.Printf("Order [%s] updated successfully", order.OrderId)
	return nil
}

// DeleteByID удаляет заказ по ID
func (os *OrderService) DeleteByID(ctx context.Context, orderId string) error {
	log.Printf("Deleting order [%s]", orderId)
	err := os.OrderRepo.DeleteItemByID(ctx, orderId)
	if err != nil {
		log.Println("Error deleting order:", err)
		return err
	}
	log.Printf("Order [%s] deleted successfully", orderId)
	return nil
}

func (orderService *OrderService) CloseOrderById(id string) error {
	return orderService.OrderRepo.CloseOrderById(id)
}

func (orderService *OrderService) BatchProcess(NewOrders []models.Order) error {
	for _, val := range NewOrders {
		err := orderService.OrderRepo(val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (orderService *OrderService) NumberOfOrderedItems(startDate string, endDate string) ([]models.OrderedItem, error) {
	result, err := orderService.OrderRepo.NumberOfOrderedItems(startDate, endDate)
	if err != nil {
		return nil, err
	}
	return result, nil
}
