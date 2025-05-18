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
	Create(ctx context.Context, ingredient models.Inventory) (models.Inventory, error)
	GetAll(ctx context.Context) ([]models.Inventory, error)
	GetByID(ctx context.Context, ingredientId string) (models.Inventory, error)
	UpdateByID(ctx context.Context, ingredient models.Inventory) error
	DeleteByID(ctx context.Context, ingerdientID string) error
	CreateTransaction(ctx context.Context, inventoryItem *models.Inventory, status string) error
	GetLeftOvers(ctx context.Context, pagenum int, pagesize int) (models.Page, error)
}

type InventoryRepo struct {
	db *sql.DB
}

func NewInventoryRepo(db *sql.DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

func (ir *InventoryRepo) Create(ctx context.Context, ingredient models.Inventory) (models.Inventory, error) {
	tx, err := ir.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Inventory{}, err
	}
	defer tx.Rollback()

	err = ir.db.QueryRowContext(ctx,
		`INSERT INTO inventory (ingredient_name, unit, quantity, reorder_level)
        VALUES ($1, $2, $3, $4)
        RETURNING ingredient_id, created_at, updated_at`,
		ingredient.IngredientName,
		ingredient.Unit,
		ingredient.Quantity,
		ingredient.ReorderLevel,
	).Scan(
		&ingredient.IngredientId,
		&ingredient.CreatedAt,
		&ingredient.UpdatedAt,
	)
	if err != nil {
		return models.Inventory{}, err
	}

	return ingredient, tx.Commit()
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

func (ir *InventoryRepo) GetByID(ctx context.Context, ingredientId string) (models.Inventory, error) {
	var ingredient models.Inventory
	err := ir.db.QueryRowContext(ctx, `SELECT * FROM inventory WHERE ingredient_id=$1`, ingredientId).Scan(&ingredient.IngredientId, &ingredient.IngredientName, &ingredient.Unit, &ingredient.Quantity, &ingredient.ReorderLevel, &ingredient.CreatedAt, &ingredient.UpdatedAt)
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

func (ir *InventoryRepo) DeleteByID(ctx context.Context, ingerdientID string) error {
	tx, err := ir.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`DELETE FROM inventory WHERE ingredient_id= $1`,
		ingerdientID,
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

func (ir *InventoryRepo) CreateTransaction(ctx context.Context, inventoryItem *models.Inventory, status string) error {
	tx, err := ir.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO inventory_transactions (ingredient_id, quantity, inventory_transaction_action)
        VALUES($1, $2, $3)`,
		inventoryItem.IngredientId,
		inventoryItem.Quantity,
		status,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ir *InventoryRepo) GetLeftOvers(ctx context.Context, pagenum int, pagesize int) (models.Page, error) {
	rows, err := ir.db.QueryContext(ctx,
		`WITH total AS (
		SELECT COUNT(*) AS total_count FROM inventory
		)
		SELECT 
			i.ingredient_name, i.quantity, total.total_count
		FROM inventory i, total
		ORDER BY i.quantity DESC
		LIMIT $1 OFFSET $2;
	`,
		pagesize,
		(pagenum-1)*pagesize,
	)
	if err != nil {
		return models.Page{}, err
	}
	defer rows.Close()

	var leftovers []models.Data
	var totalCount int

	for rows.Next() {
		var item models.Data
		if err := rows.Scan(&item.Name, &item.Quantity, &totalCount); err != nil {
			return models.Page{}, err
		}
		leftovers = append(leftovers, item)
	}

	totalpages := (totalCount + pagesize - 1) / pagesize

	response := models.Page{
		CurrentPage: pagenum,
		HasNextPage: pagenum < totalpages,
		PageSize:    pagesize,
		TotalPages:  totalpages,
		Data:        leftovers,
	}

	return response, nil
}
