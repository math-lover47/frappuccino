package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"frappuccino/models"

	"github.com/lib/pq"
)

type MenuRepo interface {
	Create(ctx context.Context, item models.MenuItems) (models.MenuItems, error)
	GetAll(ctx context.Context) ([]models.MenuItems, error)
	GetItemByID(ctx context.Context, MenuItemId string) (models.MenuItems, error)
	UpdateItemByID(ctx context.Context, item models.MenuItems) error
	DeleteItemByID(ctx context.Context, MenuItemId string) error
}

type menuRepo struct {
	*Repository
}

func NewMenuRepository(db *sql.DB) MenuRepo {
	return &menuRepo{NewRepository(db)}
}

func (r *menuRepo) Create(ctx context.Context, item models.MenuItems) (models.MenuItems, error) {
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO menu_item (item_name,item_description,price,categories)
	     VALUES ($1,$2,$3,$4)
		 RETURNING menu_item_id,created_at,updated_at`, item.ItemName, item.ItemDescription, item.Price, pq.Array(item.Categories)).Scan(&item.MenuItemId, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return models.MenuItems{}, fmt.Errorf("failed to create menu item: %w", err)
	}
	return item, nil
}

func (r *menuRepo) GetAll(ctx context.Context) ([]models.MenuItems, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT * FROM menu_item`)
	if err != nil {
		return nil, fmt.Errorf("failer to query Menu: %w", err)
	}
	defer rows.Close()
	var menu []models.MenuItems
	for rows.Next() {
		var item models.MenuItems
		err := rows.Scan(&item.MenuItemId, &item.ItemName, &item.ItemDescription, &item.Price, pq.Array(&item.Categories), &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Menu: %w", err)
		}
		menu = append(menu, item)
	}
	return menu, nil
}

func (r *menuRepo) GetItemByID(ctx context.Context, MenuItemId string) (models.MenuItems, error) {
	var item models.MenuItems
	err := r.db.QueryRowContext(ctx, `
		SELECT * FROM menu_item WHERE menu_item_id = $1`, MenuItemId).Scan(&item.MenuItemId, &item.ItemName, &item.ItemDescription, &item.Price, pq.Array(&item.Categories), &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.MenuItems{}, fmt.Errorf("Item not found: %w", err)
		}
		return models.MenuItems{}, fmt.Errorf("failed to get Item: %w", err)
	}
	return item, nil
}

func (r *menuRepo) UpdateItemByID(ctx context.Context, item models.MenuItems) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `
	UPDATE menu_item 
	SET 
		item_name = $1,
		item_description =$2,
		price =$3,
		categories =$4,
		updated_at = NOW()
	WHERE menu_item_id = $5
	`, item.ItemName, item.ItemDescription, item.Price, item.Categories, item.MenuItemId)
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

func (r *menuRepo) DeleteItemByID(ctx context.Context, MenuItemId string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `DELETE FROM menu_item_id WHERE id= $1`, MenuItemId)
	if err != nil {
		return fmt.Errorf("failed to delete MenuItems: %w", err)
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
