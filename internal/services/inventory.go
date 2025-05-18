package services

import (
	"context"
	"frappuccino/internal/repo"
	"frappuccino/models"
	"frappuccino/utils"
)

type InventoryServiceIfc interface {
	Create(ctx context.Context, ingredient models.Inventory) (models.Inventory, error)
	GetAll(ctx context.Context) ([]models.Inventory, error)
	GetByID(ctx context.Context, ingredientId string) (models.Inventory, error)
	UpdateByID(ctx context.Context, ingredient models.Inventory) error
	DeleteByID(ctx context.Context, ingerdientID string) error
	CreateTransaction(ctx context.Context, inventoryItem *models.Inventory, istatus string) error
	GetLeftOvers(ctx context.Context, pagenum int, pagesize int) (models.Page, error)
}

type InventoryService struct {
	inventoryRepo repo.InventoryRepoIfc
}

func NewInventoryService(inventoryRepo repo.InventoryRepoIfc) *InventoryService {
	return &InventoryService{inventoryRepo: inventoryRepo}
}

func (is *InventoryService) Create(ctx context.Context, ingredient models.Inventory) (models.Inventory, error) {
	if ingredient.Quantity < 0 {
		return models.Inventory{}, utils.ErrInvalidQuantity
	}
	if ingredient.ReorderLevel < 0 {
		return models.Inventory{}, utils.ErrInvalidReorderLevel
	}

	return is.inventoryRepo.Create(ctx, ingredient)
}

func (is *InventoryService) GetAll(ctx context.Context) ([]models.Inventory, error) {
	return is.inventoryRepo.GetAll(ctx)
}

func (is *InventoryService) GetByID(ctx context.Context, IngredientId string) (models.Inventory, error) {
	if len(IngredientId) <= 0 {
		return models.Inventory{}, utils.ErrInvalidIngredientId
	}
	return is.inventoryRepo.GetByID(ctx, IngredientId)
}

func (is *InventoryService) UpdateByID(ctx context.Context, ingredient models.Inventory) error {
	if ingredient.IngredientId == "" {
		return utils.ErrInvalidIngredientId
	}
	if ingredient.IngredientName == "" {
		return utils.ErrInvalidIngredientName
	}
	if ingredient.Quantity < 0 {
		return utils.ErrInvalidQuantity
	}
	if ingredient.ReorderLevel < 0 {
		return utils.ErrInvalidReorderLevel
	}
	return is.inventoryRepo.UpdateByID(ctx, ingredient)
}

func (is *InventoryService) DeleteByID(ctx context.Context, IngredientId string) error {
	if IngredientId == "" {
		return utils.ErrInvalidIngredientId
	}
	return is.inventoryRepo.DeleteByID(ctx, IngredientId)
}

func (is *InventoryService) CreateTransaction(ctx context.Context, inventoryItem *models.Inventory, status string) error {
	return is.inventoryRepo.CreateTransaction(ctx, inventoryItem, status)
}

func (is *InventoryService) GetLeftOvers(ctx context.Context, page int, pageSize int) (models.Page, error) {
	return is.inventoryRepo.GetLeftOvers(ctx, page, pageSize)
}
