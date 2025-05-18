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
)

type OrderHandler struct {
	service services.OrderServiceIfc
	*BaseHandler
}

func NewOrderHandler(service services.OrderServiceIfc, baseHandler *BaseHandler) *OrderHandler {
	return &OrderHandler{
		service:     service,
		BaseHandler: baseHandler,
	}
}

func (o *OrderHandler) validateOrder(order models.Orders) error {
	if order.CustomerId == "" {
		return errors.New("customer id is required")
	}
	if len(order.OrderItems) == 0 {
		return errors.New("order must contain at least one item")
	}
	if order.PaymentMethod == "" {
		return errors.New("payment method is required")
	}
	return nil
}

func (o *OrderHandler) Post(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var newOrder models.OrderItems
	data, err := io.ReadAll(r.Body)
	if err != nil {
		o.handleError(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err = json.Unmarshal(data, &newOrder); err != nil {
		o.handleError(w, r, http.StatusBadRequest, "Failed to parse JSON", err)
		return
	}
	if err := o.validateOrder(newOrder); err != nil {
		o.handleError(w, r, http.StatusBadRequest, "invalid order data", err)
		return
	}
	if err = o.service.AddOrder(newOrder); err != nil {
		if errors.Is(err, utils.ErrMenuItem) {
			o.handleError(w, r, http.StatusBadRequest, "Menu Item does not exist", err)
			return
		}
		o.handleError(w, r, http.StatusInternalServerError, "Failed to add order", err)
		return
	}
	o.logger.Info("New order is added successfully!", slog.String("users phone number", newOrder.PhoneNumber))

	successResponse := utils.APIResponse{
		Code:    http.StatusCreated,
		Message: "Order created successfully",
	}
	successResponse.Send(w)
}

func (o *OrderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orders, err := o.service.GetAllOrders()
	if err != nil {
		o.handleError(w, r, http.StatusInternalServerError, "Failed to get orders", err)
		return
	}

	o.logger.Info("All orders are successfully retrieved", slog.Int("count", len(orders)))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (o *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	order, err := o.service.UpdateByID(ctx, id)
	if err != nil {
		o.handleError(w, r, http.StatusNotFound, "Order not found", err)
		return
	}
	o.logger.Info("Order has been successfully taken", slog.String("Phone number", order.PhoneNumber))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (o *OrderHandler) Put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var orderChanges models.Orders
	id := r.PathValue("id")

	orderChanges.OrderId = utils.TEXT(id)

	data, err := io.ReadAll(r.Body)
	if err != nil {
		o.handleError(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err := json.Unmarshal(data, &orderChanges); err != nil {
		o.handleError(w, r, http.StatusBadRequest, "Failed to parse JSON", err)
		return
	}
	if err = o.service.UpdateByID(ctx, &orderChanges); err != nil {
		o.handleError(w, r, http.StatusInternalServerError, "Failed to update order", err)
		return
	}

	successResponse := utils.APIResponse{
		Code:    http.StatusOK,
		Message: "Order's changes applied successfully",
	}
	successResponse.Send(w)
}

func (o *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	if err := o.service.DeleteByID(ctx, id); err != nil {
		o.handleError(w, r, http.StatusInternalServerError, "Failed to delete order", err)
		return
	}

	successResponse := utils.APIResponse{
		Code:    http.StatusOK,
		Message: "Order deleted successfully",
	}
	successResponse.Send(w)
}

func (o *OrderHandler) PostCloseOrderById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := r.PathValue("id")
	if err := o.service.CloseOrderById(id); err != nil {
		o.handleError(w, r, http.StatusInternalServerError, "Failed to close order", err)
		return
	}

	successResponse := utils.APIResponse{
		Code:    http.StatusOK,
		Message: "Order issued",
	}
	successResponse.Send(w)
}

func (o *OrderHandler) BatchProcess(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var ordersArray []models.Orders

	data, err := io.ReadAll(r.Body)
	if err != nil {
		o.handleError(w, r, http.StatusBadRequest, "Invalid request body", err)
		return
	}
	if err := json.Unmarshal(data, &ordersArray); err != nil {
		o.handleError(w, r, http.StatusBadRequest, "Failed to parse JSON", err)
		return
	}
	if err := o.service.BatchProcess(ordersArray); err != nil {
		o.handleError(w, r, http.StatusInternalServerError, "Failed to process batch orders", err)
		return
	}

	o.logger.Info("New orders are added successfully!")
	successResponse := utils.APIResponse{
		Code:    http.StatusCreated,
		Message: "Orders created successfully",
	}
	successResponse.Send(w)
}

func (orderHandler *OrderHandler) NumberOfOrderedItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	result, err := orderHandler.service.NumberOfOrderedItems(startDate, endDate)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	w.Header().Set("Content-Type", "apoplication/json")
	json.NewEncoder(w).Encode(result)
}
