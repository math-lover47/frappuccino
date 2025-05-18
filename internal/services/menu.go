package services

import (
	"context"
	"frappuccino/internal/repo"
	"frappuccino/models"
	"log"
)

type MenuServiceIfc interface {
	Create(ctx context.Context, item *models.MenuItems) (*models.MenuItems, error)
	GetAll(ctx context.Context) ([]models.MenuItems, error)
	GetByID(ctx context.Context, MenuItemId string) (models.MenuItems, error)
	UpdateByID(ctx context.Context, item *models.MenuItems) error
	DeleteByID(ctx context.Context, MenuItemId string) error
	GetMenuItemPriceByName(ctx context.Context, name string) (float64, error)
}

type MenuService struct {
	menuRepo repo.MenuRepoIfc
}

func NewMenuService(menuRepo repo.MenuRepoIfc) *MenuService {
	return &MenuService{menuRepo: menuRepo}
}

func (ms *MenuService) Create(ctx context.Context, item *models.MenuItems) (*models.MenuItems, error) {
	log.Println("Creating new menu item:", item.ItemName)
	created, err := ms.menuRepo.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	log.Println("Menu item created successfully:", created.MenuItemId)
	return created, nil
}

func (ms *MenuService) GetAll(ctx context.Context) ([]models.MenuItems, error) {
	log.Println("Fetching all menu items")
	menu, err := ms.menuRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("Retrieved %d menu items", len(menu))
	return menu, nil
}

func (ms *MenuService) GetByID(ctx context.Context, MenuItemId string) (models.MenuItems, error) {
	log.Printf("Fetching menu item by ID: %s", MenuItemId)
	item, err := ms.menuRepo.GetByID(ctx, MenuItemId)
	if err != nil {
		return models.MenuItems{}, err
	}
	log.Printf("Retrieved menu item [%s]: %s", item.MenuItemId, item.ItemName)
	return item, nil
}

func (ms *MenuService) UpdateByID(ctx context.Context, item *models.MenuItems) error {
	log.Printf("Updating menu item [%s]", item.MenuItemId)
	err := ms.menuRepo.UpdateByID(ctx, item)
	if err != nil {
		return err
	}
	log.Printf("Menu item [%s] updated successfully", item.MenuItemId)
	return nil
}

func (ms *MenuService) DeleteByID(ctx context.Context, MenuItemId string) error {
	log.Printf("Deleting menu item [%s]", MenuItemId)
	err := ms.menuRepo.DeleteByID(ctx, MenuItemId)
	if err != nil {
		return err
	}
	log.Printf("Menu item [%s] deleted successfully", MenuItemId)
	return nil
}

func (ms *MenuService) GetMenuItemPriceByName(ctx context.Context, name string) (float64, error) {
	return ms.menuRepo.GetMenuItemPriceByName(ctx, name)
}
