package repo

import "database/sql"

type Repo struct {
	CustomerRepo    CustomerRepoIfc
	InventoryRepo   InventoryRepoIfc
	MenuRepo        MenuRepoIfc
	OrderRepo       OrderRepoIfc
	AggregationRepo AggregationRepoIfc
}

func NewRepository(db *sql.DB) *Repo {
	return &Repo{
		CustomerRepo:    NewCustomerRepo(db),
		InventoryRepo:   NewInventoryRepo(db),
		MenuRepo:        NewMenuRepo(db),
		OrderRepo:       NewOrderRepo(db),
		AggregationRepo: NewAggregationRepo(db),
	}
}
