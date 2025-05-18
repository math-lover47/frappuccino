package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"frappuccino/internal/services"
	"frappuccino/models"
	"frappuccino/utils"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

type InventoryHandler struct {
	service services.InventoryServiceIfc
	*BaseHandler
}

func NewInventoryHandler(service services.InventoryServiceIfc, baseHandler *BaseHandler) *InventoryHandler {
	return &InventoryHandler{service: service, BaseHandler: baseHandler}
}

func (ih *InventoryHandler) validateIngredient(ingredient models.Inventory) bool {
	return ingredient.IngredientName != "" && ingredient.Unit != "" && ingredient.Quantity > 0 && ingredient.ReorderLevel > 0
}

func (ih *InventoryHandler) Post(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var newInventoryItem models.Inventory

	status := "POST"
	data, err := io.ReadAll(r.Body)
	if err != nil {
		ih.handleError(w, r, http.StatusInternalServerError, "Failed to read request body", err)
		return
	}

	err = json.Unmarshal(data, &newInventoryItem)
	if err != nil {
		ih.handleError(w, r, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	if !ih.validateIngredient(newInventoryItem) {
		message := "Invalid inventory item data"
		if newInventoryItem.IngredientName == "" {
			message = "Name is empty"
		} else if newInventoryItem.Unit == "" {
			message = "Unit is empty"
		} else if newInventoryItem.Quantity <= 0 {
			message = "Quantity must be greater than zero"
		} else if newInventoryItem.ReorderLevel <= 0 {
			message = "Reorder level must be greater than zero"
		}

		ih.handleError(w, r, http.StatusBadRequest, utils.TEXT(message), nil)
		return
	}
	_, err = ih.service.Create(ctx, &newInventoryItem)
	if err != nil {
		ih.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	err = ih.service.CreateTransaction(ctx, &newInventoryItem, status)
	ih.logger.Info("New inventory item added successfully",
		slog.String("name", string(newInventoryItem.IngredientName)),
		slog.String("unit", string(newInventoryItem.Unit)),
		slog.Float64("quantity", float64(newInventoryItem.Quantity)),
		slog.String("url", r.URL.Path),
	)

	successResponse := utils.APIResponse{
		Code:    http.StatusCreated,
		Message: "Inventory item created successfully",
	}
	successResponse.Send(w)
}

func (ih *InventoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	inventoryItems, err := ih.service.GetAll(ctx)
	if err != nil {
		ih.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	ih.logger.Info("Fetched all inventory items",
		slog.Int("count", len(inventoryItems)),
		slog.String("url", r.URL.Path),
	)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventoryItems)
}

func (ih *InventoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")

	var newInventory models.Inventory

	newInventory.IngredientId = utils.TEXT(id)

	status := "GET"
	inventoryItem, err := ih.service.Create(ctx, &newInventory)
	if err != nil {
		if errors.Is(err, utils.ErrIdNotFound) {
			ih.handleError(w, r, http.StatusNotFound, "ID not found", err)
			return
		}
		ih.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	fmt.Println(inventoryItem)
	err = ih.service.CreateTransaction(ctx, inventoryItem, status)
	ih.logger.Info("Fetched inventory item by ID",
		slog.String("id", id),
		slog.String("name", string(inventoryItem.IngredientName)),
		slog.String("unit", string(inventoryItem.Unit)),
		slog.Float64("quantity", float64(inventoryItem.Quantity)),
		slog.String("url", r.URL.Path),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventoryItem)
}

func (ih *InventoryHandler) Put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	status := "PUT"
	var newInventoryItem models.Inventory

	data, err := io.ReadAll(r.Body)
	if err != nil {
		ih.handleError(w, r, http.StatusInternalServerError, "Failed to read request body", err)
		return
	}

	err = json.Unmarshal(data, &newInventoryItem)
	if err != nil {
		ih.handleError(w, r, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}
	if !ih.validateIngredient(newInventoryItem) {
		ih.handleError(w, r, http.StatusBadRequest, "Invalid inventory item data", nil)
		return
	}
	err = ih.service.UpdateByID(ctx, &newInventoryItem)
	if err != nil {
		ih.handleError(w, r, http.StatusNotFound, "ID not found", err)
		return
	}

	err = ih.service.CreateTransaction(ctx, &newInventoryItem, status)
	if err != nil {
		ih.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	ih.logger.Info("Inventory item updated successfully",
		slog.String("id", id),
		slog.String("name", string(newInventoryItem.IngredientName)),
		slog.String("unit", string(newInventoryItem.Unit)),
		slog.Float64("quantity", float64(newInventoryItem.Quantity)),
		slog.String("url", r.URL.Path),
	)

	successResponse := utils.APIResponse{
		Code:    http.StatusOK,
		Message: "Inventory item updated successfully",
	}
	successResponse.Send(w)
}

func (ih *InventoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	err := ih.service.DeleteByID(ctx, id)
	if err != nil {
		ih.handleError(w, r, http.StatusNotFound, "ID not found", err)
		return
	}

	ih.logger.Info("Inventory item deleted successfully",
		slog.String("id", id),
		slog.String("url", r.URL.Path),
	)

	successResponse := utils.APIResponse{
		Code:    http.StatusNoContent,
		Message: "Inventory item deleted successfully",
	}
	successResponse.Send(w)
}

func (ih *InventoryHandler) GETLeftOvers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pageStr := r.PathValue("page")
	pageSizeStr := r.PathValue("pageSize")
	page, err := strconv.Atoi(pageStr)

	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	leftOvers, err := ih.service.GetLeftOvers(ctx, page, pageSize)
	if err != nil {
		ih.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	// ih.logger.Info("Dodelat")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leftOvers)
}
