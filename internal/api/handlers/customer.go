package handlers

import (
	"encoding/json"
	"errors"
	"frappuccino/internal/services"
	"frappuccino/models"
	"frappuccino/utils"
	"io"
	"log/slog"
	"net/http"
)

type CustomerHandler struct {
	service services.CustomerServiceIfc
	*BaseHandler
}

func NewCustomerHandler(service services.CustomerServiceIfc, baseHandler *BaseHandler) *CustomerHandler {
	return &CustomerHandler{service: service, BaseHandler: baseHandler}
}

func (ch *CustomerHandler) validateCustomer(customer models.Customer) bool {
	return customer.FullName != "" && customer.PhoneNumber != "" && customer.Email != ""
}

func (ch *CustomerHandler) Post(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var newCustomer models.Customer
	data, err := io.ReadAll(r.Body)
	if err != nil {
		ch.handleError(w, r, http.StatusInternalServerError, "Failed to read request body", err)
		return
	}

	if err := json.Unmarshal(data, &newCustomer); err != nil {
		ch.handleError(w, r, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	if !ch.validateCustomer(newCustomer) {
		ch.handleError(w, r, http.StatusBadRequest, "Name or PhoneNumber is empty", nil)
		return
	}

	if _, err := ch.service.Create(ctx, &newCustomer); err != nil {
		ch.handleError(w, r, http.StatusInternalServerError, "Unexpected error", err)
		return
	}

	ch.logger.Info("New customer added successfully", slog.String("customer_id", string(newCustomer.CustomerId)))

	successResponse := utils.APIResponse{
		Code:    http.StatusCreated,
		Message: "Customer created successfully",
	}
	successResponse.Send(w)
}

func (ch *CustomerHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	allCustomer, err := ch.service.GetAll(ctx)
	if err != nil {
		ch.handleError(w, r, http.StatusNotFound, "Unexpected Error", err)
		return
	}
	ch.logger.Info("Fetched all")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allCustomer)
}

func (ch *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	customer, err := ch.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, utils.ErrIdNotFound) {
			ch.handleError(w, r, http.StatusNotFound, "ID not found", err)
			return
		}
		ch.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	ch.logger.Info("Fetched Customer by ID",
		slog.String("id", id),
		slog.String("name", string(customer.FullName)),
		slog.String("url", r.URL.Path),
	)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func (ch *CustomerHandler) Put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var newCustomer models.Customer
	data, err := io.ReadAll(r.Body)
	if err != nil {
		ch.handleError(w, r, http.StatusInternalServerError, "Failed to read request body", err)
		return
	}

	id := r.PathValue("id")
	if err := json.Unmarshal(data, &newCustomer); err != nil {
		ch.handleError(w, r, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}
	if !ch.validateCustomer(newCustomer) {
		ch.handleError(w, r, http.StatusBadRequest, "Name or PhoneNumber is empty", nil)
		return
	}

	newCustomer.CustomerId = utils.TEXT(id)
	if err := ch.service.UpdateById(ctx, &newCustomer); err != nil {
		if errors.Is(err, utils.ErrIdNotFound) {
			ch.handleError(w, r, http.StatusNotFound, "ID not found", err)
			return
		} else if errors.Is(err, utils.ErrConflictFields) {
			ch.handleError(w, r, http.StatusConflict, "Conflict Fields", err)
			return
		}
		ch.handleError(w, r, http.StatusInternalServerError, "Unexpected error", err)
	}

	successResponse := utils.APIResponse{
		Code:    http.StatusOK,
		Message: "Customer updated successfully",
	}
	successResponse.Send(w)
}

func (ch *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")

	if err := ch.service.DeleteCustomerById(ctx, id); err != nil {
		if errors.Is(err, utils.ErrIdNotFound) {
			ch.handleError(w, r, http.StatusNotFound, "ID not found", err)
			return
		}
		ch.handleError(w, r, http.StatusInternalServerError, "Unexpected Error", err)
		return
	}
	successResponse := utils.APIResponse{
		Code:    http.StatusOK,
		Message: "Customer deleted successfully",
	}
	successResponse.Send(w)
}
