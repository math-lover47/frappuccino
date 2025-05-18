package repo

import (
	"context"
	"database/sql"
	"fmt"
	"frappuccino/models"
	"frappuccino/utils"
)

type OrderRepoIfc interface {
	Create(ctx context.Context, order *models.Orders) (*models.Orders, error)
	GetAll(ctx context.Context) ([]models.Orders, error)
	GetOrderByID(ctx context.Context, orderId string) (models.Orders, error)
	UpdateItemByID(ctx context.Context, order *models.Orders) error
	DeleteItemByID(ctx context.Context, orderId string) error
	checkAndUpdateInventory(ctx context.Context, tx *sql.Tx, orderItems []models.OrderItems) error
	getOrderItemsByOrderID(ctx context.Context, orderId string) ([]models.OrderItems, error)
}

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (or *OrderRepo) Create(ctx context.Context, order *models.Orders) (*models.Orders, error) {
	// Начало транзакции
	tx, err := or.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback() // Откатить транзакцию в случае ошибки
		} else {
			err = tx.Commit() // Подтвердить транзакцию в случае успеха
		}
	}()

	var totalPrice utils.DEC

	// Суммируем стоимость всех элементов заказа
	for _, item := range order.OrderItems {
		totalPrice += item.Quantity * item.UnitPrice
	}

	// Теперь totalPrice содержит итоговую сумму заказа
	order.TotalPrice = totalPrice

	// Вставка данных заказа в таблицу orders
	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders (customer_id, special_instructions, total_price, order_status, payment_method)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING order_id, created_at, updated_at`,
		order.CustomerId,
		order.SpecialInstructions,
		order.TotalPrice,
		order.OrderStatus,
		order.PaymentMethod,
	).Scan(&order.OrderId, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Проверка и обновление остатков в инвентаре для каждого элемента заказа
	err = or.checkAndUpdateInventory(ctx, tx, order.OrderItems)
	if err != nil {
		return nil, err
	}

	// Вставка данных элементов заказа (OrderItems)
	for _, item := range order.OrderItems {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, menu_item_id, customizations, item_name, quantity, unit_price)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			order.OrderId, // Привязка к заказу
			item.MenuItemId,
			item.Customizations,
			item.ItemName,
			item.Quantity,
			item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
	}

	return order, nil
}

func (or *OrderRepo) GetAll(ctx context.Context) ([]models.Orders, error) {
	rows, err := or.db.QueryContext(ctx, `
		SELECT o.order_id, 
		       o.customer_id, 
		       o.total_price, 
		       o.order_status, 
		       o.order_payment_method, 
		       o.created_at, 
		       o.updated_at
		FROM orders o
		ORDER BY o.created_at DESC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Orders
	for rows.Next() {
		var order models.Orders
		if err := rows.Scan(&order.OrderId, &order.CustomerId, &order.TotalPrice, &order.OrderStatus, &order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (or *OrderRepo) UpdateItemByID(ctx context.Context, order *models.Orders) error {
	// Начинаем транзакцию
	tx, err := or.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Обновление позиций заказа (если необходимо, можно сделать по отдельности для каждой позиции)
	for _, item := range order.OrderItems {
		_, err = tx.ExecContext(ctx, `
			UPDATE order_items
			SET quantity = $1, unit_price = $2
			WHERE order_item_id = $3;
		`, item.Quantity, item.UnitPrice, item.OrderItemId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	var totalPrice utils.DEC

	// Суммируем стоимость всех элементов заказа
	for _, item := range order.OrderItems {
		totalPrice += item.Quantity * item.UnitPrice
	}

	// Теперь totalPrice содержит итоговую сумму заказа
	order.TotalPrice = totalPrice
	// Обновление заказа
	_, err = tx.ExecContext(ctx, `
UPDATE orders
SET special_instructions = $1, 
	total_price = $2, 
	order_status = $3, 
	order_payment_method = $4,
	updated_at = NOW()
WHERE order_id = $5;
`, order.SpecialInstructions, order.TotalPrice, order.OrderStatus, order.PaymentMethod, order.OrderId)
	if err != nil { // Функция для проверки остатков и обновления инвентаря

		tx.Rollback()
		return err
	}
	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (or *OrderRepo) DeleteItemByID(ctx context.Context, orderId string) error {
	// Начинаем транзакцию
	tx, err := or.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Удаляем все позиции заказа
	_, err = tx.ExecContext(ctx, `DELETE FROM order_items WHERE order_id = $1`, orderId)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Удаляем заказ
	_, err = tx.ExecContext(ctx, `DELETE FROM orders WHERE order_id = $1`, orderId)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (or *OrderRepo) getOrderItemsByOrderID(ctx context.Context, orderId string) ([]models.OrderItems, error) {
	// Выполняем запрос на получение всех позиций заказа
	rows, err := or.db.QueryContext(ctx,
		`SELECT order_item_id, menu_item_id, order_id, customizations, item_name, quantity, unit_price 
		FROM order_items WHERE order_id = $1`, orderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Создаём слайс для хранения позиций заказа
	var orderItems []models.OrderItems
	for rows.Next() {
		var item models.OrderItems
		// Сканируем данные в структуру
		err := rows.Scan(
			&item.OrderItemId,
			&item.MenuItemId,
			&item.OrderId,
			&item.Customizations,
			&item.ItemName,
			&item.Quantity,
			&item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
		// Добавляем позицию в список
		orderItems = append(orderItems, item)
	}

	// Проверка на ошибки после обработки строк
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orderItems, nil
}

// Функция для проверки остатков и обновления инвентаря
func (or *OrderRepo) checkAndUpdateInventory(ctx context.Context, tx *sql.Tx, orderItems []models.OrderItems) error {
	for _, item := range orderItems {
		// Получаем ингредиенты для данного блюда
		var ingredients []models.Inventory
		rows, err := tx.QueryContext(ctx,
			`SELECT i.ingredient_id, i.ingredient_name, i.unit, i.quantity
			FROM menu_item_ingredients mi
			JOIN ingredients i ON mi.ingredient_id = i.ingredient_id
			WHERE mi.menu_item_id = $1`,
			item.MenuItemId,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		// Составляем список ингредиентов для текущего блюда
		for rows.Next() {
			var ingredient models.Inventory
			err := rows.Scan(&ingredient.IngredientId, &ingredient.IngredientName, &ingredient.Unit, &ingredient.Quantity)
			if err != nil {
				return err
			}
			ingredients = append(ingredients, ingredient)
		}

		// Проверка, достаточно ли остатков в инвентаре для каждого ингредиента
		for _, ingredient := range ingredients {
			requiredQuantity := ingredient.Quantity * item.Quantity // Количество, необходимое для всех позиций этого блюда
			if requiredQuantity > ingredient.Quantity {
				return fmt.Errorf("недостаточно ингредиента %s для выполнения заказа", ingredient.IngredientName)
			}
		}

		// Обновляем инвентарь, уменьшая количество ингредиентов
		for _, ingredient := range ingredients {
			requiredQuantity := ingredient.Quantity * item.Quantity // Количество, необходимое для всех позиций этого блюда

			_, err := tx.ExecContext(ctx,
				`UPDATE ingredients
				SET quantity = quantity - $1, updated_at = NOW()
				WHERE ingredient_id = $2`,
				requiredQuantity, ingredient.IngredientId,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (or *OrderRepo) GetOrderByID(ctx context.Context, orderId string) (models.Orders, error) {
	// Запрос для получения заказа по его ID
	var order models.Orders
	err := or.db.QueryRowContext(ctx, `
		SELECT order_id, customer_id, total_price, order_status, order_payment_method, created_at, updated_at
		FROM orders
		WHERE order_id = $1
	`, orderId).Scan(&order.OrderId, &order.CustomerId, &order.TotalPrice, &order.OrderStatus, &order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt)
	// Обработка ошибок
	if err != nil {
		if err == sql.ErrNoRows {
			// Если заказ не найден
			return models.Orders{}, fmt.Errorf("order with id %s not found", orderId)
		}
		return models.Orders{}, err // другие ошибки (например, проблемы с подключением к базе)
	}

	// Получаем все позиции заказа
	orderItems, err := or.getOrderItemsByOrderID(ctx, orderId)
	if err != nil {
		return models.Orders{}, err
	}

	order.OrderItems = orderItems

	return order, nil
}
