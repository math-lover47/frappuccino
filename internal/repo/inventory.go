package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"frappuccino/models"
)

type InventoryRepoIfc interface {
	Create(ctx context.Context, ingredient models.Inventory) (models.Inventory, error)
	GetAll(ctx context.Context) ([]models.Inventory, error)
	GetIngredientByID(ctx context.Context, IngredientId string) (models.Inventory, error)
	UpdateIngredientByID(ctx context.Context, ingredient models.Inventory) error
	DeleteIngredientByID(ctx context.Context, IngerdientID string) error
}

type InventoryRepo struct {
	db *sql.DB
}

func NewInventoryRepo(db *sql.DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

func (r *InventoryRepo) Create(ctx context.Context, ingredient models.Inventory) (models.Inventory, error) {
	var item models.Inventory
	if ingredient.IngredientName == "" {
		return models.Inventory{}, errors.New("ingredient_name cannot be empty")
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO inventory (ingredient_name, unit, quantity, reorder_level)
        VALUES ($1, $2, $3, $4)
        RETURNING ingredient_id, created_at, updated_at`,
		ingredient.IngredientName, ingredient.Unit, ingredient.Quantity, ingredient.ReorderLevel).
		Scan(&item.IngredientId, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return models.Inventory{}, err
	}

	// Заполнение объекта item
	item.IngredientName = ingredient.IngredientName
	item.Unit = ingredient.Unit
	item.Quantity = ingredient.Quantity
	item.ReorderLevel = ingredient.ReorderLevel

	return item, nil
}

func (r *InventoryRepo) GetAll(ctx context.Context) ([]models.Inventory, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT * FROM inventory`)
	if err != nil {
		return nil, fmt.Errorf("failer to query inventory: %w", err)
	}
	defer rows.Close()
	var inventory []models.Inventory
	for rows.Next() {
		var ingredient models.Inventory
		err := rows.Scan(&ingredient.IngredientId, &ingredient.IngredientName, &ingredient.Unit, &ingredient.Quantity, &ingredient.ReorderLevel, &ingredient.CreatedAt, &ingredient.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ingredient: %w", err)
		}
		inventory = append(inventory, ingredient)
	}
	return inventory, nil
}

func (r *InventoryRepo) GetIngredientByID(ctx context.Context, IngredientId string) (models.Inventory, error) {
	var ingredient models.Inventory
	err := r.db.QueryRowContext(ctx, `SELECT * FROM inventory WHERE ingredient_id=$1`, IngredientId).Scan(&ingredient.IngredientId, &ingredient.IngredientName, &ingredient.Unit, &ingredient.Quantity, &ingredient.ReorderLevel, &ingredient.CreatedAt, &ingredient.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Inventory{}, fmt.Errorf("ingredient not found: %w", err)
		}
		return models.Inventory{}, fmt.Errorf("failed to get ingredient: %w", err)
	}

	return ingredient, nil
}

func (r *InventoryRepo) UpdateIngredientByID(ctx context.Context, ingredient models.Inventory) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
	UPDATE inventory
	SET ingredient_name = $1,
	unit = $2,
	quantity= $3,
	reorder_level =$4,
	updated_at = NOW()
	WHERE ingredient_id =$5
	`, ingredient.IngredientName, ingredient.Unit, ingredient.Quantity, ingredient.ReorderLevel, ingredient.IngredientId)
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

func (r *InventoryRepo) DeleteIngredientByID(ctx context.Context, IngerdientID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `DELETE FROM inventory WHERE ingredient_id= $1`, IngerdientID)
	if err != nil {
		return fmt.Errorf("failed to delete ingredient: %w", err)
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
