package handlers

import (
	"encoding/json"
	"frappuccino/internal/services"
	"frappuccino/models"
	"log"
	"net/http"
)

type InventoryHandler struct {
	inventoryService services.InventoryService // было invenrotyService
}

func NewInventoryHandler(service services.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: service}
}

func (h *InventoryHandler) CreateInventoryIngredient(w http.ResponseWriter, r *http.Request) {
	var input models.Inventory
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	log.Printf("input %v", input)

	item, err := h.inventoryService.Create(r.Context(), input)
	if err != nil {
		log.Printf("failed to create ingredient: %v", err) // <- вот здесь логируем ошибку
		http.Error(w, "failed to create ingredient", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	inventory, err := h.inventoryService.GetAll(r.Context())
	if err != nil {
		http.Error(w, "failed to get inventory", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(inventory)
}

func (h *InventoryHandler) GetIngredientByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	item, err := h.inventoryService.GetIngredientByID(r.Context(), idStr)
	if err != nil {
		http.Error(w, "ingredient not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (h *InventoryHandler) UpdateIngredient(w http.ResponseWriter, r *http.Request) {
	var input models.Inventory
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err := h.inventoryService.UpdateIngredientByID(r.Context(), input)
	if err != nil {
		http.Error(w, "failed to update ingredient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"ingredient updated successfully"}`))
}

func (h *InventoryHandler) DeleteIngredient(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	err := h.inventoryService.DeleteIngredientByID(r.Context(), idStr)
	if err != nil {
		http.Error(w, "failed to delete ingredient: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"ingredient deleted successfully"}`))
}
