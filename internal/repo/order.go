package repo

import (
	"context"
	"database/sql"
	"frappuccino/models"
)

type OrderRepoIfc interface {
	Create(ctx context.Context, customer models.Customer) (models.Customer, error)
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetItemByID(ctx context.Context, customerId string) (models.Customer, error)
	UpdateItemByID(ctx context.Context, customer models.Customer) error
	DeleteItemByID(ctx context.Context, customerId string) error
}

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}
