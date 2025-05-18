package services

import (
	"context"
	"frappuccino/internal/repo"
	"frappuccino/models"
	"log"
)

type CustomerServiceIfc interface {
	Create(ctx context.Context, customer models.Customer) (models.Customer, error)
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetByID(ctx context.Context, CustomerId string) (models.Customer, error)
	UpdateById(ctx context.Context, customer models.Customer) error
	DeleteById(ctx context.Context, CustomerId string) error
	GetByFullNameAndPhone(name string, phone string) (string, error)
}

type CustomerService struct {
	customerRepo repo.CustomerRepoIfc
}

func NewCustomerService(customerRepo repo.CustomerRepoIfc) *CustomerService {
	return &CustomerService{customerRepo: customerRepo}
}

func (cs *CustomerService) Create(ctx context.Context, customer models.Customer) (models.Customer, error) {
	log.Println("Creating new Customer item:", customer.FullName)
	created, err := cs.customerRepo.Create(ctx, customer)
	if err != nil {
		return models.Customer{}, err
	}
	log.Println("Customer item created successfully:", created.CustomerId)
	return created, nil
}

func (cs *CustomerService) GetAll(ctx context.Context) ([]models.Customer, error) {
	log.Println("Fetching all menu items")
	menu, err := cs.customerRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("Retrieved %d menu items", len(menu))
	return menu, nil
}

func (cs *CustomerService) GetByID(ctx context.Context, CustomerId string) (models.Customer, error) {
	log.Printf("Fetching menu item by ID: %s", CustomerId)
	customer, err := cs.customerRepo.GetByID(ctx, CustomerId)
	if err != nil {
		return models.Customer{}, err
	}
	log.Printf("Retrieved menu item [%s]: %s", customer.CustomerId, customer.FullName)
	return customer, nil
}

func (cs *CustomerService) UpdateById(ctx context.Context, customer models.Customer) error {
	log.Printf("Updating menu item [%s]", customer.CustomerId)
	err := cs.customerRepo.UpdateById(ctx, customer)
	if err != nil {
		return err
	}
	log.Printf("Menu item [%s] updated successfully", customer.CustomerId)
	return nil
}

func (cs *CustomerService) DeleteByID(ctx context.Context, CustomerId string) error {
	log.Printf("Deleting menu item [%s]", CustomerId)
	err := cs.customerRepo.DeleteById(ctx, CustomerId)
	if err != nil {
		return err
	}
	log.Printf("Menu item [%s] deleted successfully", CustomerId)
	return nil
}

func (cs *CustomerService) GetCustomerByNameAndPhone(ctx context.Context, fullname string, phonenumber string) (string, error) {
	log.Printf("Fetching menu item id by fullname and phone number: %s, %s", fullname, phonenumber)
	customerId, err := cs.customerRepo.GetByFullNameAndPhone(ctx, fullname, phonenumber)
	if err != nil {
		return "", err
	}
	log.Printf("Retrieved menu item id [%s]: %s", customerId)
	return customerId, nil
}
