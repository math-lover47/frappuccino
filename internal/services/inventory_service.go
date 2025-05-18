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
	GetIngredientByID(ctx context.Context, IngredientId string) (models.Inventory, error)
	UpdateIngredientByID(ctx context.Context, ingredient models.Inventory) error
	DeleteIngredientByID(ctx context.Context, IngerdientID string) error
}

type InventoryService struct {
	inventoryRepo repo.InventoryRepo
}

func NewInventoryService(inventoryRepo repo.InventoryRepo) *InventoryService {
	return &InventoryService{inventoryRepo: inventoryRepo}
}

func (s *InventoryService) Create(ctx context.Context, ingredient models.Inventory) (models.Inventory, error) {
	if ingredient.Quantity < 0 {
		return models.Inventory{}, utils.ErrInvalidQuantity
	}
	if ingredient.ReorderLevel < 0 {
		return models.Inventory{}, utils.ErrInvalidReorderLevel
	}

	return s.inventoryRepo.Create(ctx, ingredient)
}

func (s *InventoryService) GetAll(ctx context.Context) ([]models.Inventory, error) {
	return s.inventoryRepo.GetAll(ctx)
}

func (s *InventoryService) GetIngredientByID(ctx context.Context, IngredientId string) (models.Inventory, error) {
	if len(IngredientId) <= 0 {
		return models.Inventory{}, utils.ErrInvalidIngredientId
	}
	return s.inventoryRepo.GetIngredientByID(ctx, IngredientId)
}

func (s *InventoryService) UpdateIngredientByID(ctx context.Context, ingredient models.Inventory) error {
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
	return s.inventoryRepo.UpdateIngredientByID(ctx, ingredient)
}

func (s *InventoryService) DeleteIngredientByID(ctx context.Context, IngredientId string) error {
	if IngredientId == "" {
		return utils.ErrInvalidIngredientId
	}
	return s.inventoryRepo.DeleteIngredientByID(ctx, IngredientId)
}
