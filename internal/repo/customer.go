package repo

import (
	"context"
	"database/sql"
	"errors"
	"frappuccino/models"
	"frappuccino/utils"
)

type CustomerRepoIfc interface {
	Create(ctx context.Context, customer models.Customer) (models.Customer, error)
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetByID(ctx context.Context, customerId string) (models.Customer, error)
	UpdateById(ctx context.Context, customer models.Customer) error
	DeleteById(ctx context.Context, customerId string) error
	GetByFullNameAndPhone(ctx context.Context, fullname string, phonenumber string) (string, error)
}

type CustomerRepo struct {
	db *sql.DB
}

func NewCustomerRepo(db *sql.DB) *CustomerRepo {
	return &CustomerRepo{db: db}
}

func (cr *CustomerRepo) Create(ctx context.Context, customer models.Customer) (models.Customer, error) {
	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Customer{}, err
	}
	defer tx.Rollback()

	err = cr.db.QueryRowContext(ctx,
		`INSERT INTO customers (full_name,phone_number,email,preferences)
	     VALUES ($1,$2,$3,$4)
		 RETURNING customer_id,created_at,updateByIdd_at`,
		customer.FullName,
		customer.PhoneNumber,
		customer.Email,
		customer.Preferences,
	).Scan(
		&customer.CustomerId,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		return models.Customer{}, err
	}

	return customer, tx.Commit()
}

func (cr *CustomerRepo) GetAll(ctx context.Context) ([]models.Customer, error) {
	rows, err := cr.db.QueryContext(ctx, `SELECT * FROM customers`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allcustomers []models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.CustomerId,
			&customer.FullName,
			&customer.PhoneNumber,
			&customer.Email,
			&customer.Preferences,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		allcustomers = append(allcustomers, customer)
	}

	return allcustomers, nil
}

func (cr *CustomerRepo) GetByID(ctx context.Context, customerId string) (models.Customer, error) {
	var customer models.Customer
	err := cr.db.QueryRowContext(ctx,
		`SELECT * FROM customers WHERE customer_id = $1`,
		customerId,
	).Scan(
		&customer.CustomerId,
		&customer.FullName,
		&customer.PhoneNumber,
		&customer.Email,
		&customer.Preferences,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Customer{}, utils.ErrIdNotFound
		}
		return models.Customer{}, err
	}
	return customer, nil
}

func (cr *CustomerRepo) UpdateById(ctx context.Context, customer models.Customer) error {
	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE customers 
			SET 
				full_name = $1,
				phone_number =$2,
				email =$3,
				preferences =$4,
				updateByIdd_at = NOW()
			WHERE customer_id = $5
		`,
		customer.FullName,
		customer.PhoneNumber,
		customer.Email,
		customer.Preferences,
		customer.CustomerId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.ErrIdNotFound
		} else {
			return utils.ErrConflictFields
		}
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return utils.ErrIdNotFound
	}

	return tx.Commit()
}

func (cr *CustomerRepo) DeleteById(ctx context.Context, customerId string) error {
	tx, err := cr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM customer_id WHERE id= $1`, customerId)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrIdNotFound
	}

	return tx.Commit()
}

func (cr *CustomerRepo) GetByFullNameAndPhone(ctx context.Context, fullname string, phonenumber string) (string, error) {
	var CustomerId string
	err := cr.db.QueryRowContext(ctx,
		`WITH existing AS (
		SELECT customer_id 
		FROM customer 
		WHERE full_name = $1 AND phone_number = $2
		),
		inserted AS (
		INSERT INTO customer (full_name, phone_number)
		SELECT $1, $2
		WHERE NOT EXISTS (SELECT 1 FROM existing)
		RETURNING customer_id
		)
		SELECT customer_id FROM inserted
		UNION ALL
		SELECT customer_id FROM existing;
	`,
		fullname,
		phonenumber).Scan(
		&CustomerId,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", utils.ErrIdNotFound
		}
		return "", err
	}

	return CustomerId, nil
}
