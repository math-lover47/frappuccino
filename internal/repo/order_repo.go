package repo

import (
	"context"
	"database/sql"

	"frappuccino/models"
)

type OrderRepo interface {
	Create(ctx context.Context, customer models.Customer) (models.Customer, error)
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetItemByID(ctx context.Context, CustomerId string) (models.Customer, error)
	UpdateItemByID(ctx context.Context, customer models.Customer) error
	DeleteItemByID(ctx context.Context, CustomerId string) error
}

type orderRepo struct {
	*Repository
}

func NewOrderRepository(db *sql.DB) OrderRepo {
	return &orderRepo{NewRepository(db)}
}
