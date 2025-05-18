package services

import "frappuccino/internal/repo"

type AggregationServiceIfc interface{}

type AggregationService struct {
	AggregationRepo repo.AggregationRepoIfc
}

func NewAggregationService(AggregationRepo repo.AggregationRepoIfc) *AggregationService {
	return &AggregationService{AggregationRepo: AggregationRepo}
}
