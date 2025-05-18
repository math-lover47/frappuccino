package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"frappuccino/models"
	"frappuccino/utils"
)

type InventoryRepoIfc interface {
	Create(ctx context.Context, Ingredient models.Inventory) (models.Inventory, error)
	GetAll(ctx context.Context) ([]models.Inventory, error)
	GetByID(ctx context.Context, IngredientId string) (models.Inventory, error)
	UpdateByID(ctx context.Context, Ingredient models.Inventory) error
	DeleteByID(ctx context.Context, IngerdientID string) error
}

type InventoryRepo struct {
	db *sql.DB
}

func NewInventoryRepo(db *sql.DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

func (ir *InventoryRepo) Create(ctx context.Context, Ingredient models.Inventory) (models.Inventory, error) {
	tx, err := ir.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Inventory{}, err
	}
	defer tx.Rollback()

	err = ir.db.QueryRowContext(ctx,
		`INSERT INTO inventory (ingredient_name, unit, quantity, reorder_level)
        VALUES ($1, $2, $3, $4)
        RETURNING ingredient_id, created_at, updated_at`,
		Ingredient.IngredientName,
		Ingredient.Unit,
		Ingredient.Quantity,
		Ingredient.ReorderLevel,
	).Scan(
		&Ingredient.IngredientId,
		&Ingredient.CreatedAt,
		&Ingredient.UpdatedAt,
	)
	if err != nil {
		return models.Inventory{}, err
	}

	return Ingredient, tx.Commit()
}

func (ir *InventoryRepo) GetAll(ctx context.Context) ([]models.Inventory, error) {
	rows, err := ir.db.QueryContext(ctx, `SELECT * FROM inventory`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventory []models.Inventory
	for rows.Next() {
		var ingredient models.Inventory
		err := rows.Scan(&ingredient.IngredientId, &ingredient.IngredientName, &ingredient.Unit, &ingredient.Quantity, &ingredient.ReorderLevel, &ingredient.CreatedAt, &ingredient.UpdatedAt)
		if err != nil {
			return nil, err
		}
		inventory = append(inventory, ingredient)
	}
	return inventory, nil
}

func (ir *InventoryRepo) GetByID(ctx context.Context, IngredientId string) (models.Inventory, error) {
	var ingredient models.Inventory
	err := ir.db.QueryRowContext(ctx, `SELECT * FROM inventory WHERE ingredient_id=$1`, IngredientId).Scan(&ingredient.IngredientId, &ingredient.IngredientName, &ingredient.Unit, &ingredient.Quantity, &ingredient.ReorderLevel, &ingredient.CreatedAt, &ingredient.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Inventory{}, err
		}
		return models.Inventory{}, err
	}

	return ingredient, nil
}

func (ir *InventoryRepo) UpdateByID(ctx context.Context, ingredient models.Inventory) error {
	tx, err := ir.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE inventory
	SET ingredient_name = $1,
		unit = $2,
		quantity= $3,
		reorder_level =$4,
		updated_at = NOW()
	WHERE ingredient_id =$5
	`,
		ingredient.IngredientName,
		ingredient.Unit,
		ingredient.Quantity,
		ingredient.ReorderLevel,
		ingredient.IngredientId,
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

func (ir *InventoryRepo) DeleteByID(ctx context.Context, IngerdientID string) error {
	tx, err := ir.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`DELETE FROM inventory WHERE ingredient_id= $1`,
		IngerdientID,
	)
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
