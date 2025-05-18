package service

import (
	"context"

	"frappuccino/internal/repo"
	"frappuccino/models"
)

type InventoryService interface {
	Create(ctx context.Context, ingredient models.Inventory) (models.Inventory, error)
	GetAll(ctx context.Context) ([]models.Inventory, error)
	GetIngredientByID(ctx context.Context, IngredientId string) (models.Inventory, error)
	UpdateIngredientByID(ctx context.Context, ingredient models.Inventory) error
	DeleteIngredientByID(ctx context.Context, IngerdientID string) error
}

type inventoryService struct {
	inventoryRepo repo.InventoryRepo
}

func NewInventoryService(inventoryRepo repo.InventoryRepo) InventoryService {
	return &inventoryService{inventoryRepo: inventoryRepo}
}

func (s *inventoryService) Create(ctx context.Context, ingredient models.Inventory) (models.Inventory, error) {
	if ingredient.Quantity < 0 {
		return models.Inventory{}, models.ErrInvalidQuantity
	}
	if ingredient.ReorderLevel < 0 {
		return models.Inventory{}, models.ErrInvalidReorderLevel
	}

	return s.inventoryRepo.Create(ctx, ingredient)
}

func (s *inventoryService) GetAll(ctx context.Context) ([]models.Inventory, error) {
	return s.inventoryRepo.GetAll(ctx)
}

func (s *inventoryService) GetIngredientByID(ctx context.Context, IngredientId string) (models.Inventory, error) {
	if len(IngredientId) <= 0 {
		return models.Inventory{}, models.ErrInvalidIngredientId
	}
	return s.inventoryRepo.GetIngredientByID(ctx, IngredientId)
}

func (s *inventoryService) UpdateIngredientByID(ctx context.Context, ingredient models.Inventory) error {
	if ingredient.IngredientId == "" {
		return models.ErrInvalidIngredientId
	}
	if ingredient.IngredientName == "" {
		return models.ErrInvalidIngredientName
	}
	if ingredient.Quantity < 0 {
		return models.ErrInvalidQuantity
	}
	if ingredient.ReorderLevel < 0 {
		return models.ErrInvalidReorderLevel
	}
	return s.inventoryRepo.UpdateIngredientByID(ctx, ingredient)
}

func (s *inventoryService) DeleteIngredientByID(ctx context.Context, IngredientId string) error {
	if IngredientId == "" {
		return models.ErrInvalidIngredientId
	}
	return s.inventoryRepo.DeleteIngredientByID(ctx, IngredientId)
}
