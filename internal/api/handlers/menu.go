package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"frappuccino/internal/services"
	"frappuccino/models"
	"frappuccino/utils"
	"io"
	"net/http"
)

type MenuHandler struct {
	service services.MenuServiceIfc
	*BaseHandler
}

func NewMenuHandler(service services.MenuServiceIfc, baseHansler *BaseHandler) *MenuHandler {
	return &MenuHandler{service: service, BaseHandler: baseHansler}
}

func (mh *MenuHandler) Post(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		mh.handleError(w, r, http.StatusInternalServerError, "Failed to read request body", err)
		return
	}
	var newMenuItem models.MenuItems
	err = json.Unmarshal(data, &newMenuItem)
	if err != nil {
		mh.handleError(w, r, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}
	if newMenuItem.MenuItemId != "" {
		newMenuItem.MenuItemId = ""
	}
	err = mh.service.Create(ctx, &newMenuItem)
	if err != nil {
		mh.handleError(w, r, http.StatusInternalServerError, "Failed to add menu item", err)
		return
	}
	mh.logger.Info(
		"Successfully added a new Menu Item",
		"name", newMenuItem.ItemName,
		"URL", r.URL.Path)

	successResponse := utils.APIResponse{
		Code:    http.StatusCreated,
		Message: "Menu Item added successfully",
	}
	successResponse.Send(w)
}

func (mh *MenuHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	menuitems, err := mh.service.GetAll(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuitems)
}

func (mh *MenuHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	menuItem, err := mh.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, utils.ErrIDNotFound) {
			mh.handleError(w, r, http.StatusNotFound, "ID not found", err)
			return
		}
		mh.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItem)
}

func (mh *MenuHandler) Put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	data, err := io.ReadAll(r.Body)
	if err != nil {
		mh.handleError(w, r, http.StatusInternalServerError, "Failed to read request body", err)
		return
	}
	var newMenuItem models.MenuItems
	err = json.Unmarshal(data, &newMenuItem)
	if err != nil {
		mh.handleError(w, r, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}
	err = mh.service.UpdateByID(ctx, &newMenuItem)
	if err != nil {
		if errors.Is(err, utils.ErrIdNotFound) {
			mh.handleError(w, r, http.StatusNotFound, "ID not found", err)
			return
		}
		mh.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	mh.logger.Info("Menu Item updated successfully",
		"id", id,
		"name", newMenuItem.ItemName,
		"url", r.URL.Path)

	successResponse := utils.APIResponse{
		Code:    http.StatusOK,
		Message: "Menu Item updated successfully",
	}
	successResponse.Send(w)
}

func (mh *MenuHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	err := mh.service.DeleteByID(ctx, id)
	if err != nil {
		if errors.Is(err, utils.ErrIdNotFound) {
			mh.handleError(w, r, http.StatusNotFound, "ID not found", err)
			return
		}
		mh.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	mh.logger.Info("Menu Item deleted successfully",
		"id", id,
		"url", r.URL.Path)
	successResponse := utils.APIResponse{
		Code:    http.StatusNoContent,
		Message: "Menu Item deleted successfully",
	}
	successResponse.Send(w)
}
