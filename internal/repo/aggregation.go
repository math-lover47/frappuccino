package repo

import (
	"context"
	"database/sql"
	"fmt"
	"frappuccino/models"
)

type AggregationRepoIfc interface {
	GetTotalSales(ctx context.Context) (float64, error)
	GetPopularItems(ctx context.Context) (models.PopularItems, error)
	GetSearchItems(ctx context.Context, q string, filter []string, maxPrice float64, minPrice float64) (models.Search, error)
	GetListOfOrderedItems(ctx context.Context, period string, month string, year string) (models.ListOrderedItemByPeriods, error)
}

type AggregationRepo struct {
	db *sql.DB
}

func NewAggregationRepo(db *sql.DB) *AggregationRepo {
	return &AggregationRepo{db: db}
}

func (ar *AggregationRepo) TotalSales(ctx context.Context) (float64, error) {
	var totalSales float64

	err := ar.db.QueryRowContext(ctx, `SELECT sum(total_price) FROM orders WHERE order_status = 'COMPLETED';`).Scan(
		&totalSales,
	)
	if err != nil {
		return 0, err
	}

	return totalSales, nil
}

func (ar *AggregationRepo) GetPopularItems(ctx context.Context) (models.PopularItems, error) {
	var popularItems models.PopularItems

	rows, err := ar.db.QueryContext(ctx,
		`SELECT item_name, count(item_name) 
		FROM order_items 
		GROUP BY item_name
		ORDER BY count(item_name) DESC
		LIMIT 10;`,
	)
	if err != nil {
		return models.PopularItems{}, err
	}

	for rows.Next() {
		var popularItem models.PopularItem

		err = rows.Scan(&popularItem.ItemName, &popularItem.OrderedTimes)
		if err != nil {
			return models.PopularItems{}, err
		}
		popularItems.Items = append(popularItems.Items, popularItem)
	}

	return popularItems, nil
}

func (ar *AggregationRepo) GetSearchItems(ctx context.Context, q string, filter []string, maxPrice float64, minPrice float64) (models.Search, error) {
	var Search models.Search
	var err error
	for _, val := range filter {
		if val == "MENU" {
			Search.MenuItems, err = ar.getMenu(ctx, q, maxPrice, minPrice)
		} else if val == "ORDER" {
			Search.OrderItems, err = ar.getOrder(ctx, q, maxPrice, minPrice)
		} else if len(val) == 0 {
			Search.OrderItems, err = ar.getOrder(ctx, q, maxPrice, minPrice)
			if err != nil {
				return models.Search{}, err
			}
			Search.MenuItems, err = ar.getMenu(ctx, q, maxPrice, minPrice)
			if err != nil {
				return models.Search{}, err
			}
			break
		}
		if err != nil {
			return models.Search{}, err
		}
	}
	return Search, nil
}

func (ar *AggregationRepo) getOrder(ctx context.Context, q string, maxPrice float64, minPrice float64) ([]models.SearchOrder, error) {
	query := `SELECT oi.order_id, c.full_name, oi.item_name, oi.unit_price 
				FROM order_items as oi
				JOIN orders as o on o.order_id = oi.order_id
				JOIN customer as c on c.customer_id = o.customer_id
				WHERE  (c.full_name ILIKE '%' || $1 || '%' OR  oi.item_name ILIKE '%' || $1 || '%')`
	var rows *sql.Rows
	var err error

	if minPrice != -1 && maxPrice != -1 {
		fmt.Println("1")
		query += ` AND oi.price BETWEEN $2 AND $3`
		rows, err = ar.db.QueryContext(ctx, query, q, minPrice, maxPrice)
	} else if minPrice != -1 {
		fmt.Println("2")
		query += ` AND oi.price >= $2`
		rows, err = ar.db.QueryContext(ctx, query, q, minPrice)
	} else if maxPrice != -1 {
		fmt.Println("3")
		query += ` AND oi.price <= $2`
		rows, err = ar.db.QueryContext(ctx, query, q, maxPrice)
	} else {
		rows, err = ar.db.QueryContext(ctx, query, q)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allOrders []models.SearchOrder
	for rows.Next() {
		var rowResult models.SearchOrder

		err := rows.Scan(&rowResult.Id, &rowResult.CustomerName, &rowResult.ItemName, &rowResult.ItemPrice)
		if err != nil {
			return nil, err
		}
		allOrders = append(allOrders, rowResult)
	}
	return allOrders, nil
}

func (ar *AggregationRepo) getMenu(ctx context.Context, q string, maxPrice float64, minPrice float64) ([]models.SearchMenu, error) {
	query := `SELECT menu_item_id, item_name, item_description, price 
				FROM menu_items
				WHERE  (item_name ILIKE  '%' || $1 || '%' OR item_description ILIKE '%' || $1 || '%')`
	var rows *sql.Rows
	var err error
	fmt.Println("here 1")
	if minPrice != -1 && maxPrice != -1 {
		query += ` AND price BETWEEN $2 AND $3`
		rows, err = ar.db.QueryContext(ctx, query, q, minPrice, maxPrice)
	} else if minPrice != -1 {
		query += ` AND price >= $2`
		fmt.Println("here 2")
		rows, err = ar.db.QueryContext(ctx, query, q, minPrice)
	} else if maxPrice != -1 {
		query += ` AND price <= $2`
		rows, err = ar.db.QueryContext(ctx, query, q, maxPrice)
	} else {
		rows, err = ar.db.QueryContext(ctx, query, q)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allOrders []models.SearchMenu

	for rows.Next() {
		var rowResult models.SearchMenu
		err := rows.Scan(&rowResult.Id, &rowResult.Name, &rowResult.Description, &rowResult.Price)
		if err != nil {
			return nil, err
		}
		allOrders = append(allOrders, rowResult)
	}
	return allOrders, nil
}

func (ar *AggregationRepo) GetListOfOrderedItems(ctx context.Context, period string, month string, year string) (models.ListOrderedItemByPeriods, error) {
	var list models.ListOrderedItemByPeriods
	if period == "day" {
		rows, err := ar.db.QueryContext(ctx,
			`WITH base_date AS (
				SELECT TO_DATE($1 || ' ' || $2, 'Month YYYY') AS first_day
			),
			days AS (
				SELECT 
					generate_series(
						1, 
						date_part(
							'day', 
							(DATE_TRUNC('month', (SELECT first_day FROM base_date)) 
							+ INTERVAL '1 month - 1 day')
						)::int
					) AS n
			)
			SELECT 
				d.n, 
				COUNT(o.order_id)
			FROM days d
			LEFT JOIN orders o 
				ON DATE(o.created_at) = (SELECT first_day FROM base_date) + (d.n - 1) * INTERVAL '1 day'
			GROUP BY d.n
			ORDER BY d.n;
			`,
			month,
			year,
		)
		if err != nil {
			return models.ListOrderedItemByPeriods{}, err
		}
		for rows.Next() {
			var item models.OrderedItemByPeriod
			err = rows.Scan(&item.Date, &item.Count)
			if err != nil {
				return models.ListOrderedItemByPeriods{}, err
			}
			list.Items = append(list.Items, item)
		}
	} else if period == "month" {
		rows, err := ar.db.Query(
			`WITH all_months AS (
				SELECT generate_series(1, 12) AS month_num
			),
			monthly_orders AS (
				SELECT 
					EXTRACT(MONTH FROM o.created_at)::int AS month_num,
					COUNT(o.order_id) AS order_count
				FROM orders o
				WHERE EXTRACT(YEAR FROM o.created_at) = $1
				GROUP BY EXTRACT(MONTH FROM o.created_at)
			)
			SELECT 
				TO_CHAR(TO_DATE(m.month_num::text, 'MM'), 'Month') AS month_name,
				COALESCE(mo.order_count, 0) AS total_orders
			FROM all_months m
			LEFT JOIN monthly_orders mo 
				ON m.month_num = mo.month_num
			ORDER BY m.month_num;`,
			year,
		)
		if err != nil {
			return models.ListOrderedItemByPeriods{}, err
		}
		for rows.Next() {
			var item models.OrderedItemByPeriod
			err = rows.Scan(&item.Date, &item.Count)
			if err != nil {
				return models.ListOrderedItemByPeriods{}, err
			}
			list.Items = append(list.Items, item)
		}
	}
	return list, nil
}
