package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"frappuccino/internal/service"
	"frappuccino/models"
)

type MenuHandler struct {
	menuService service.MenuService
}

func NewMenuHandler(service service.MenuService) *MenuHandler {
	return &MenuHandler{menuService: service}
}

func (h *MenuHandler) CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	var input models.MenuItems
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	log.Printf("input %v", input)

	item, err := h.menuService.Create(r.Context(), input)
	if err != nil {
		log.Printf("failed to create ingredient: %v", err) // <- вот здесь логируем ошибку
		http.Error(w, "failed to create ingredient", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *MenuHandler) GetAllMenu(w http.ResponseWriter, r *http.Request) {
	items, err := h.menuService.GetAll(r.Context())
	if err != nil {
		http.Error(w, "failed to get menu", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *MenuHandler) GetIngredientByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id") // FIX
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	item, err := h.menuService.GetItemByID(r.Context(), idStr)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (h *MenuHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	var input models.MenuItems
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err := h.menuService.UpdateItemByID(r.Context(), input)
	if err != nil {
		http.Error(w, "failed to update Item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Item updated successfully"}`))
}

func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id") // FIX
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	err := h.menuService.DeleteItemByID(r.Context(), idStr)
	if err != nil {
		http.Error(w, "failed to delete item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"item deleted successfully"}`))
}
