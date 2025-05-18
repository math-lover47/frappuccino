package api

import (
	"frappuccino/internal/api/handlers"
	"net/http"
)

func Router(handlers *handlers.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /customer", handlers.CustomerHandler.POST)
	mux.HandleFunc("GET /customer", handlers.CustomerHandler.GET)
	mux.HandleFunc("GET /customer/{id}", handlers.CustomerHandler.GETBYID)
	mux.HandleFunc("PUT /customer/{id}", handlers.CustomerHandler.PUT)
	mux.HandleFunc("DELETE /customer/{id}", handlers.CustomerHandler.DELETE)

	mux.HandleFunc("POST /inventory", handlers.InventoryHandler.POST)
	mux.HandleFunc("GET /inventory", handlers.InventoryHandler.GET)
	mux.HandleFunc("GET /inventory/{id}", handlers.InventoryHandler.GETBYID)
	mux.HandleFunc("PUT /inventory/{id}", handlers.InventoryHandler.PUT)
	mux.HandleFunc("DELETE /inventory/{id}", handlers.InventoryHandler.DELETE)

	mux.HandleFunc("POST /menu", handlers.MenuHandler.POST)
	mux.HandleFunc("GET /menu", handlers.MenuHandler.GET)
	mux.HandleFunc("GET /menu/{id}", handlers.MenuHandler.GETBYID)
	mux.HandleFunc("PUT /menu/{id}", handlers.MenuHandler.PUT)
	mux.HandleFunc("DELETE /menu/{id}", handlers.MenuHandler.DELETE)

	mux.HandleFunc("POST /order", handlers.OrderHandler.POST)
	mux.HandleFunc("GET /order", handlers.OrderHandler.GET)
	mux.HandleFunc("GET /order/{id}", handlers.OrderHandler.GETBYID)
	mux.HandleFunc("PUT /order/{id}", handlers.OrderHandler.PUT)
	mux.HandleFunc("DELETE /order/{id}", handlers.OrderHandler.DELETE)

	mux.HandleFunc("GET /reports/total-sales", handlers.AggregationHandler.GETTotalItems)
	mux.HandleFunc("GET /reports/popular-items", handlers.AggregationHandler.GETPopularItems)
	mux.HandleFunc("GET /reports/search", handlers.AggregationHandler.GETBYSearch)
	mux.HandleFunc("GET /reports/orderedItemsNyPeriod", handlers.AggregationHandler.GETListOfOrderedItems)

	return mux
}
