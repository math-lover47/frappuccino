package handlers

import (
	"encoding/json"
	"fmt"
	"frappuccino/internal/services"
	"net/http"
	"strconv"
)

type AggregationHandler struct {
	service services.AggregationServiceIfc
	*BaseHandler
}

func NewAggregationHandler(service services.AggregationServiceIfc, baseHandler *BaseHandler) *AggregationHandler {
	return &AggregationHandler{
		service:     service,
		BaseHandler: baseHandler,
	}
}

func (ah *AggregationHandler) GetTotalSales(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	totalSales, err := ah.service.GetTotalSales(ctx)
	if err != nil {
		ah.handleError(w, r, http.StatusInternalServerError, "Failed to fetch total sales", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(totalSales)
}

func (ah *AggregationHandler) GetPopularItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	popularItems, err := ah.service.GetPopularItems(ctx)
	if err != nil {
		ah.handleError(w, r, http.StatusInternalServerError, "Failed to fetch popular items", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(popularItems)
}

func (ah *AggregationHandler) GetBySearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	q := r.URL.Query().Get("q")
	filter := r.URL.Query().Get("filter")
	minPriceString := r.URL.Query().Get("minPrice")
	maxPriceString := r.URL.Query().Get("maxPrice")
	var maxPrice, minPrice float64
	var err error

	if len(minPriceString) == 0 {
		minPrice = -1
	} else {
		minPrice, err = strconv.ParseFloat(minPriceString, 64)
		if err != nil {
			ah.handleError(w, r, http.StatusBadRequest, "Invalid minPrice value", err)
			return
		}
	}

	if len(maxPriceString) == 0 {
		maxPrice = -1
	} else {
		maxPrice, err = strconv.ParseFloat(maxPriceString, 64)
		if err != nil {
			ah.handleError(w, r, http.StatusBadRequest, "Invalid maxPrice value", err)
			return
		}
	}

	if len(q) == 0 {
		ah.handleError(w, r, http.StatusBadRequest, "Query parameter 'q' is required", nil)
		return
	}

	result, err := ah.service.GetSearchItems(ctx, q, filter, maxPrice, minPrice)
	if err != nil {
		ah.handleError(w, r, http.StatusInternalServerError, "Search failed", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (aggregationHandler *AggregationHandler) GetListOfOrderedItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	period := r.URL.Query().Get("period")
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")

	if len(year) == 0 {
		year = "2022"
	}
	result, err := aggregationHandler.service.GetListOfOrderedItems(ctx, period, month, year)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
