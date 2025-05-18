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
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Вставка элемента меню
	err = mr.db.QueryRowContext(ctx,
		`INSERT INTO menu_items (item_name, item_description, price, categories)
	     VALUES ($1, $2, $3, $4)
		 RETURNING menu_item_id, created_at, updated_at`,
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

	for _, ingredient := range menuItem.Ingredients {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO ingredients (menu_item_id, ingredient_name, quantity)
			 VALUES ($1, $2, $3)`,
			menuItem.MenuItemId, ingredient.IngredientName, ingredient.Quantity,
		)
		if err != nil {
			return models.MenuItems{}, err
		}
	}

	// Завершаем транзакцию
	return menuItem, err
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

		ingredientRows, err := mr.db.QueryContext(ctx,
			`SELECT ingredient_name, quantity FROM ingredients WHERE menu_item_id = $1`, menuItem.MenuItemId)
		if err != nil {
			return nil, err
		}
		defer ingredientRows.Close()

		var ingredients []models.Ingredients
		for ingredientRows.Next() {
			var ingredient models.Ingredients
			err := ingredientRows.Scan(&ingredient.IngredientName, &ingredient.Quantity)
			if err != nil {
				return nil, err
			}
			ingredients = append(ingredients, ingredient)
		}
		menuItem.Ingredients = ingredients

		menu = append(menu, menuItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
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

	// Получаем ингредиенты для данного элемента меню
	ingredientRows, err := mr.db.QueryContext(ctx,
		`SELECT ingredient_name, quantity FROM ingredients WHERE menu_item_id = $1`, menuItemId)
	if err != nil {
		return models.MenuItems{}, err
	}
	defer ingredientRows.Close()

	var ingredients []models.Ingredients
	for ingredientRows.Next() {
		var ingredient models.Ingredients
		err := ingredientRows.Scan(&ingredient.IngredientName, &ingredient.Quantity)
		if err != nil {
			return models.MenuItems{}, err
		}
		ingredients = append(ingredients, ingredient)
	}

	menuItem.Ingredients = ingredients

	return menuItem, nil
}

func (mr *MenuRepo) UpdateByID(ctx context.Context, menuItem models.MenuItems) error {
	tx, err := mr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Обновление данных меню
	res, err := tx.ExecContext(ctx,
		`UPDATE menu_items
		SET 
			item_name = $1,
			item_description = $2,
			price = $3,
			categories = $4,
			updated_at = NOW()
		WHERE menu_item_id = $5`,
		menuItem.ItemName,
		menuItem.ItemDescription,
		menuItem.Price,
		pq.Array(menuItem.Categories), // Если используется pq.Array для массивов
		menuItem.MenuItemId,
	)
	if err != nil {
		return err // Непосредственно возвращаем ошибку
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return utils.ErrIdNotFound // Элемент не найден
	}

	// Обновление или добавление ингредиентов в инвентаре
	// Для этого предполагаем, что у тебя есть список ингредиентов, который нужно обновить
	for _, ingredient := range menuItem.Ingredients {
		// Проверяем, есть ли уже этот ингредиент в инвентаре
		var currentQuantity int
		err := tx.QueryRowContext(ctx,
			`SELECT quantity FROM inventory WHERE menu_item_id = $1 AND ingredient_name = $2`,
			menuItem.MenuItemId, ingredient.IngredientName,
		).Scan(&currentQuantity)

		// Если ингредиент существует, обновляем его количество
		if err == nil {
			_, err := tx.ExecContext(ctx,
				`UPDATE inventory
				SET quantity = $1
				WHERE menu_item_id = $2 AND ingredient_name = $3`,
				ingredient.Quantity, menuItem.MenuItemId, ingredient.IngredientName,
			)
			if err != nil {
				return err
			}
		} else if err == sql.ErrNoRows {
			// Если ингредиента нет в инвентаре, добавляем его
			_, err := tx.ExecContext(ctx,
				`INSERT INTO inventory (menu_item_id, ingredient_name, quantity)
				VALUES ($1, $2, $3)`,
				menuItem.MenuItemId, ingredient.IngredientName, ingredient.Quantity,
			)
			if err != nil {
				return err
			}
		} else {
			return err // Ошибка, отличная от "не найдено"
		}
	}

	// Завершаем транзакцию
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
