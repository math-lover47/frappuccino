package repo

import (
	"context"
	"database/sql"
	"errors"
	"frappuccino/models"
	"frappuccino/utils"

	"github.com/lib/pq"
)

type MenuRepoIfc interface {
	Create(ctx context.Context, menuItem models.MenuItems) (models.MenuItems, error)
	GetAll(ctx context.Context) ([]models.MenuItems, error)
	GetByID(ctx context.Context, menuItemId string) (models.MenuItems, error)
	UpdateByID(ctx context.Context, menuItem models.MenuItems) error
	DeleteByID(ctx context.Context, menuItemId string) error

	CreatePriceHistory(ctx context.Context, menuItemId string, Price float64) error
	CreateIngredient(ctx context.Context, Ingredient *models.MenuItemsIngredients, menuItemName string) error
	GetMenuItemPriceByName(ctx context.Context, menuItemName string) (float64, error)
}

type MenuRepo struct {
	db *sql.DB
}

func NewMenuRepo(db *sql.DB) *MenuRepo {
	return &MenuRepo{db: db}
}

func (mr *MenuRepo) Create(ctx context.Context, menuItem models.MenuItems) (models.MenuItems, error) {
	tx, err := mr.db.BeginTx(ctx, nil)
	if err != nil {
		return models.MenuItems{}, err
	}
	defer tx.Rollback()

	err = mr.db.QueryRowContext(ctx,
		`INSERT INTO menu_items (item_name,item_description,price,categories)
	     VALUES ($1,$2,$3,$4)
		 RETURNING menu_item_id,created_at,updated_at`,
		menuItem.ItemName,
		menuItem.ItemDescription,
		menuItem.Price,
		pq.Array(menuItem.Categories),
	).Scan(
		&menuItem.MenuItemId,
		&menuItem.CreatedAt,
		&menuItem.UpdatedAt,
	)

	if err != nil {
		return models.MenuItems{}, err
	}

	return menuItem, tx.Commit()
}

func (mr *MenuRepo) GetAll(ctx context.Context) ([]models.MenuItems, error) {
	rows, err := mr.db.QueryContext(ctx, `SELECT * FROM menu_items`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menu []models.MenuItems
	for rows.Next() {
		var menuItem models.MenuItems
		err := rows.Scan(
			&menuItem.MenuItemId,
			&menuItem.ItemName,
			&menuItem.ItemDescription,
			&menuItem.Price,
			pq.Array(&menuItem.Categories),
			&menuItem.CreatedAt,
			&menuItem.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		menu = append(menu, menuItem)
	}

	return menu, nil
}

func (mr *MenuRepo) GetByID(ctx context.Context, menuItemId string) (models.MenuItems, error) {
	var menuItem models.MenuItems
	err := mr.db.QueryRowContext(ctx,
		`SELECT * FROM menu_items WHERE menu_item_id = $1`,
		menuItemId,
	).Scan(
		&menuItem.MenuItemId,
		&menuItem.ItemName,
		&menuItem.ItemDescription,
		&menuItem.Price,
		pq.Array(&menuItem.Categories),
		&menuItem.CreatedAt,
		&menuItem.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.MenuItems{}, utils.ErrIdNotFound
		}
		return models.MenuItems{}, err
	}
	return menuItem, nil
}

func (mr *MenuRepo) UpdateByID(ctx context.Context, menuItem models.MenuItems) error {
	tx, err := mr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`UPDATE menu_items
		SET 
			item_name = $1,
			item_description =$2,
			price =$3,
			categories =$4,
			updated_at = NOW()
		WHERE menu_item_id = $5
	`,
		menuItem.ItemName,
		menuItem.ItemDescription,
		menuItem.Price,
		menuItem.Categories,
		menuItem.MenuItemId,
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

func (mr *MenuRepo) DeleteByID(ctx context.Context, menuItemId string) error {
	tx, err := mr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `DELETE FROM menu_item_id WHERE id= $1`, menuItemId)
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

func (mr *MenuRepo) CreatePriceHistory(ctx context.Context, menuItemId string, price float64) error {
	tx, err := mr.db.Begin()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO price_history (menu_item_id, price)
			VALUES ($1, $2)`,
		menuItemId,
		price,
	)

	if err != nil {
		return err
	}
	return tx.Commit()
}

func (mr *MenuRepo) CreateIngredient(ctx context.Context, ingredient *models.MenuItemsIngredients, menuItemName string) error {
	tx, err := mr.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO menu_item_ingredients(menu_item_id, ingredient_id, ingredient_name,quantity)
        VALUES(
			(SELECT menu_item_id FROM menu_items WHERE item_name = $1), 
			(SELECT ingredient_id FROM inventory WHERE ingredient_name = $2), 
			$3, 
			$4
		);`,
		menuItemName,
		ingredient.IngredientName,
		ingredient.IngredientName,
		ingredient.Quantity,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (mr *MenuRepo) GetMenuItemPriceByName(ctx context.Context, menuItemName string) (float64, error) {
	var menuItemPrice float64

	err := mr.db.QueryRowContext(ctx, `SELECT price FROM menu_items WHERE item_name=$1`,
		menuItemName,
	).Scan(
		&menuItemPrice,
	)
	if err != nil {
		return 0, err
	}

	return menuItemPrice, nil
}
