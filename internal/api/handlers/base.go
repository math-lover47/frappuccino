package handlers

import (
	"frappuccino/internal/services"
	"frappuccino/utils"
	"log/slog"
	"net/http"
)

type BaseHandler struct {
	logger slog.Logger
}

func NewBaseHandler(logger slog.Logger) *BaseHandler {
	return &BaseHandler{logger: logger}
}

func (b *BaseHandler) handleError(w http.ResponseWriter, r *http.Request, code utils.INT, message utils.TEXT, err error) {
	if err != nil {
		b.logger.Error(string(message), "error", err, "code", code, "url", r.URL.Path)
	} else {
		b.logger.Error(string(message), "code", code, r.URL.Path)
	}

	jsonErr := utils.APIError{
		Code:     code,
		Message:  message,
		Resource: utils.TEXT(r.URL.Path),
	}

	jsonErr.Send(w)
}

type Handler struct {
	CustomerHandler    *CustomerHandler
	InventoryHandler   *InventoryHandler
	MenuHandler        *MenuHandler
	OrderHandler       *OrderHandler
	AggregationHandler *AggregationHandler
}

func New(service *services.Service, base *BaseHandler) *Handler {
	return &Handler{
		CustomerHandler:    NewCustomerHandler(services.CustomerService, base),
		InventoryHandler:   NewInventoryHandler(services.InventoryService, base),
		MenuHandler:        NewMenuHandler(services.MenuService, base),
		OrderHandler:       NewOrderHandler(services.OrderService, base),
		AggregationHandler: NewAggregationHandler(services.AggregationService, base),
	}
}
