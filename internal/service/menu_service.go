package service

import (
	"context"
	"fmt"
	"log"

	"frappuccino/internal/repo"
	"frappuccino/models"
)

type MenuService interface {
	Create(ctx context.Context, item models.MenuItems) (models.MenuItems, error)
	GetAll(ctx context.Context) ([]models.MenuItems, error)
	GetItemByID(ctx context.Context, MenuItemId string) (models.MenuItems, error)
	UpdateItemByID(ctx context.Context, item models.MenuItems) error
	DeleteItemByID(ctx context.Context, MenuItemId string) error
}

type menuService struct {
	menuRepo repo.MenuRepo
}

func NewMenuService(menuRepo repo.MenuRepo) MenuService {
	return &menuService{menuRepo: menuRepo}
}

func (s *menuService) Create(ctx context.Context, item models.MenuItems) (models.MenuItems, error) {
	log.Println("Creating new menu item:", item.ItemName)
	created, err := s.menuRepo.Create(ctx, item)
	if err != nil {
		log.Printf("Failed to create menu item '%s': %v", item.ItemName, err)
		return models.MenuItems{}, fmt.Errorf("could not create menu item: %w", err)
	}
	log.Println("Menu item created successfully:", created.MenuItemId)
	return created, nil
}

func (s *menuService) GetAll(ctx context.Context) ([]models.MenuItems, error) {
	log.Println("Fetching all menu items")
	menu, err := s.menuRepo.GetAll(ctx)
	if err != nil {
		log.Printf("Failed to fetch menu items: %v", err)
		return nil, fmt.Errorf("could not retrieve menu: %w", err)
	}
	log.Printf("Retrieved %d menu items", len(menu))
	return menu, nil
}

func (s *menuService) GetItemByID(ctx context.Context, MenuItemId string) (models.MenuItems, error) {
	log.Printf("Fetching menu item by ID: %s", MenuItemId)
	item, err := s.menuRepo.GetItemByID(ctx, MenuItemId)
	if err != nil {
		log.Printf("Failed to fetch menu item [%s]: %v", MenuItemId, err)
		return models.MenuItems{}, fmt.Errorf("could not get menu item: %w", err)
	}
	log.Printf("Retrieved menu item [%s]: %s", item.MenuItemId, item.ItemName)
	return item, nil
}

func (s *menuService) UpdateItemByID(ctx context.Context, item models.MenuItems) error {
	log.Printf("Updating menu item [%s]", item.MenuItemId)
	err := s.menuRepo.UpdateItemByID(ctx, item)
	if err != nil {
		log.Printf("Failed to update menu item [%s]: %v", item.MenuItemId, err)
		return fmt.Errorf("could not update menu item: %w", err)
	}
	log.Printf("Menu item [%s] updated successfully", item.MenuItemId)
	return nil
}

func (s *menuService) DeleteItemByID(ctx context.Context, MenuItemId string) error {
	log.Printf("Deleting menu item [%s]", MenuItemId)
	err := s.menuRepo.DeleteItemByID(ctx, MenuItemId)
	if err != nil {
		log.Printf("Failed to delete menu item [%s]: %v", MenuItemId, err)
		return fmt.Errorf("could not delete menu item: %w", err)
	}
	log.Printf("Menu item [%s] deleted successfully", MenuItemId)
	return nil
}
