package repo

import "database/sql"

type AggregationRepoIfc interface{}

type AggregationRepo struct {
	db *sql.DB
}

func NewAggregationRepo(db *sql.DB) *AggregationRepo {
	return &AggregationRepo{db: db}
}
