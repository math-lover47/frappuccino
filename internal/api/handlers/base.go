package handlers

import (
	"frappuccino/internal/services"
	"frappuccino/utils"
	"log/slog"
	"net/http"
)

type BaseHandler struct {
	logger *slog.Logger
}

func NewBaseHandler(logger *slog.Logger) *BaseHandler {
	return &BaseHandler{logger: logger}
}

type Handler struct {
	CustomerHandler    *CustomerHandler
	InventoryHandler   *InventoryHandler
	MenuHandler        *MenuHandler
	OrderHandler       *OrderHandler
	AggregationHandler *AggregationHandler
}

func New(service *services.Base, base *BaseHandler) *Handler {
	return &Handler{
		CustomerHandler:    NewCustomerHandler(service.CustomerService, base),
		InventoryHandler:   NewInventoryHandler(service.InventoryService, base),
		MenuHandler:        NewMenuHandler(service.MenuService, base),
		OrderHandler:       NewOrderHandler(service.OrderService, base),
		AggregationHandler: NewAggregationHandler(service.AggregationService, base),
	}
}

func (b *BaseHandler) handleError(w http.ResponseWriter, r *http.Request, code utils.INT, message utils.TEXT, err error) {
	if err != nil {
		b.logger.Error(string(message), "error", err, "code", code, "url", r.URL.Path)
	} else {
		b.logger.Error(string(message), "code", code, "url", r.URL.Path)
	}

	jsonErr := utils.APIError{
		Code:     code,
		Message:  message,
		Resource: utils.TEXT(r.URL.Path),
	}
	jsonErr.Send(w)
}
