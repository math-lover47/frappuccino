package api

import (
	"frappuccino/internal/api/handlers"
	"net/http"
)

func Router(handlers *handlers.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /customer", handlers.CustomerHandler.Post)
	mux.HandleFunc("GET /customer", handlers.CustomerHandler.GetAll)
	mux.HandleFunc("GET /customer/{id}", handlers.CustomerHandler.Get)
	mux.HandleFunc("PUT /customer/{id}", handlers.CustomerHandler.Put)
	mux.HandleFunc("DELETE /customer/{id}", handlers.CustomerHandler.Delete)

	mux.HandleFunc("POST /inventory", handlers.InventoryHandler.Post)
	mux.HandleFunc("GET /inventory", handlers.InventoryHandler.GetAll)
	mux.HandleFunc("GET /inventory/{id}", handlers.InventoryHandler.Get)
	mux.HandleFunc("PUT /inventory/{id}", handlers.InventoryHandler.Put)
	mux.HandleFunc("DELETE /inventory/{id}", handlers.InventoryHandler.Delete)

	mux.HandleFunc("POST /menu", handlers.MenuHandler.Post)
	mux.HandleFunc("GET /menu", handlers.MenuHandler.GetAll)
	mux.HandleFunc("GET /menu/{id}", handlers.MenuHandler.Get)
	mux.HandleFunc("PUT /menu/{id}", handlers.MenuHandler.Put)
	mux.HandleFunc("DELETE /menu/{id}", handlers.MenuHandler.Delete)

	mux.HandleFunc("GET /inventory/getLeftOvers/{page}/{pageSize}", handlers.InventoryHandler.GETLeftOvers)

	mux.HandleFunc("POST /order", handlers.OrderHandler.Post)
	mux.HandleFunc("GET /order", handlers.OrderHandler.GetAll)
	mux.HandleFunc("GET /order/{id}", handlers.OrderHandler.Get)
	mux.HandleFunc("PUT /order/{id}", handlers.OrderHandler.Put)
	mux.HandleFunc("DELETE /order/{id}", handlers.OrderHandler.Delete)
	mux.HandleFunc("POST /order/{id}/close", handlers.OrderHandler.PostClose)
	mux.HandleFunc("GET /order/batch-process", handlers.OrderHandler.BatchProcess)
	mux.HandleFunc("GET /order/numberOfOrderedItems", handlers.OrderHandler.NumberOfOrderedItems)

	mux.HandleFunc("GET /reports/total-sales", handlers.AggregationHandler.GetTotalSales)
	mux.HandleFunc("GET /reports/popular-items", handlers.AggregationHandler.GetPopularItems)
	mux.HandleFunc("GET /reports/search", handlers.AggregationHandler.GetBySearch)
	mux.HandleFunc("GET /reports/orderedItemsNyPeriod", handlers.AggregationHandler.GetListOfOrderedItems)

	return mux
}
