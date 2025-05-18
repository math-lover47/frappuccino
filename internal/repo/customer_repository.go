package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"frappuccino/models"
)

type CustomerRepo interface {
	Create(ctx context.Context, customer models.Customer) (models.Customer, error)
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetItemByID(ctx context.Context, CustomerId string) (models.Customer, error)
	UpdateItemByID(ctx context.Context, customer models.Customer) error
	DeleteItemByID(ctx context.Context, CustomerId string) error
}

type customerRepo struct {
	*Repository
}

func NewCustomerRepository(db *sql.DB) CustomerRepo {
	return &customerRepo{NewRepository(db)}
}

func (r *customerRepo) Create(ctx context.Context, customer models.Customer) (models.Customer, error) {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO customers (full_name,phone_number,email,preferences)
	     VALUES ($1,$2,$3,$4)
		 RETURNING customer_id,created_at,updated_at`, customer.FullName, customer.PhoneNumber, customer.Email, customer.Preferences).Scan(&customer.CustomerId, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		return models.Customer{}, fmt.Errorf("failed to create Customer: %w", err)
	}
	return customer, nil
}

func (r *customerRepo) GetAll(ctx context.Context) ([]models.Customer, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT * FROM customers`)
	if err != nil {
		return nil, fmt.Errorf("failer to query Customer: %w", err)
	}
	defer rows.Close()
	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(&customer.CustomerId, &customer.FullName, &customer.PhoneNumber, &customer.Email, &customer.Preferences, &customer.CreatedAt, &customer.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Customer: %w", err)
		}
		customers = append(customers, customer)
	}
	return customers, nil
}

func (r *customerRepo) GetItemByID(ctx context.Context, CustomerId string) (models.Customer, error) {
	var customer models.Customer
	err := r.db.QueryRowContext(ctx, `
		SELECT * FROM customers WHERE customer_id = $1`, CustomerId).Scan(&customer.CustomerId, &customer.FullName, &customer.PhoneNumber, &customer.Email, &customer.Preferences, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Customer{}, fmt.Errorf("Customer not found: %w", err)
		}
		return models.Customer{}, fmt.Errorf("failed to get Customer: %w", err)
	}
	return customer, nil
}

func (r *customerRepo) UpdateItemByID(ctx context.Context, customer models.Customer) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `
	UPDATE customers 
	SET 
		full_name = $1,
		phone_number =$2,
		email =$3,
		preferences =$4,
		updated_at = NOW()
	WHERE customer_id = $5
	`, customer.FullName, customer.PhoneNumber, customer.Email, customer.Preferences, customer.CustomerId)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *customerRepo) DeleteItemByID(ctx context.Context, CustomerId string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `DELETE FROM customer_id WHERE id= $1`, CustomerId)
	if err != nil {
		return fmt.Errorf("failed to delete Customer: %w", err)
	}

	// Verify exactly one row was deleted
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	// Commit transaction if everything succeeded
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
