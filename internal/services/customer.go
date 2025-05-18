package services

import (
	"context"
	"frappuccino/internal/repo"
	"frappuccino/models"
	"log"
)

type CustomerServiceIfc interface {
	Create(ctx context.Context, customer *models.Customer) (*models.Customer, error)
	GetAll(ctx context.Context) ([]models.Customer, error)
	GetByID(ctx context.Context, customerId string) (models.Customer, error)
	UpdateById(ctx context.Context, customer *models.Customer) error
	DeleteCustomerById(ctx context.Context, customerId string) error
	GetByFullNameAndPhone(ctx context.Context, fullname string, phone string) (string, error)
}

type CustomerService struct {
	customerRepo repo.CustomerRepoIfc
}

func NewCustomerService(customerRepo repo.CustomerRepoIfc) *CustomerService {
	return &CustomerService{customerRepo: customerRepo}
}

func (cs *CustomerService) Create(ctx context.Context, customer *models.Customer) (*models.Customer, error) {
	log.Println("Creating new customer:", customer.FullName)
	created, err := cs.customerRepo.Create(ctx, customer)
	if err != nil {
		return nil, err
	}
	log.Println("Customer created successfully:", created.CustomerId)
	return created, nil
}

func (cs *CustomerService) GetAll(ctx context.Context) ([]models.Customer, error) {
	log.Println("Fetching all customers")
	customers, err := cs.customerRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	log.Printf("Retrieved %d customers", len(customers))
	return customers, nil
}

func (cs *CustomerService) GetByID(ctx context.Context, customerId string) (models.Customer, error) {
	log.Printf("Fetching customer by ID: %s", customerId)
	customer, err := cs.customerRepo.GetByID(ctx, customerId)
	if err != nil {
		return models.Customer{}, err
	}
	log.Printf("Retrieved customer [%s]: %s", customer.CustomerId, customer.FullName)
	return customer, nil
}

func (cs *CustomerService) UpdateById(ctx context.Context, customer *models.Customer) error {
	log.Printf("Updating customer [%s]", customer.CustomerId)
	err := cs.customerRepo.UpdateById(ctx, customer)
	if err != nil {
		return err
	}
	log.Printf("Customer [%s] updated successfully", customer.CustomerId)
	return nil
}

func (cs *CustomerService) DeleteCustomerById(ctx context.Context, customerId string) error {
	log.Printf("Deleting customer [%s]", customerId)
	err := cs.customerRepo.DeleteById(ctx, customerId)
	if err != nil {
		return err
	}
	log.Printf("Customer [%s] deleted successfully", customerId)
	return nil
}

func (cs *CustomerService) GetByFullNameAndPhone(ctx context.Context, fullname string, phonenumber string) (string, error) {
	log.Printf("Fetching customer ID by fullname and phone number: %s, %s", fullname, phonenumber)
	customerId, err := cs.customerRepo.GetByFullNameAndPhone(ctx, fullname, phonenumber)
	if err != nil {
		return "", err
	}
	log.Printf("Retrieved customer ID [%s]", customerId)
	return customerId, nil
}
