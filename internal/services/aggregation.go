package services

import (
	"context"
	"frappuccino/internal/repo"
	"frappuccino/models"
	"strings"
)

type AggregationServiceIfc interface {
	GetTotalSales(ctx context.Context) (models.TotalSales, error)
	GetPopularItems(ctx context.Context) (models.PopularItems, error)
	GetSearchItems(ctx context.Context, q string, filter string, maxPrice float64, minPrice float64) (models.Search, error)
	GetListOfOrderedItems(ctx context.Context, period string, month string, year string) (models.ListOrderedItemByPeriods, error)
}

type AggregationService struct {
	AggregationRepo repo.AggregationRepoIfc
}

func NewAggregationService(AggregationRepo repo.AggregationRepoIfc) *AggregationService {
	return &AggregationService{AggregationRepo: AggregationRepo}
}

func (as *AggregationService) GetTotalSales(ctx context.Context) (models.TotalSales, error) {
	value, err := as.AggregationRepo.GetTotalSales(ctx)
	if err != nil {
		return models.TotalSales{}, err
	}
	var totalSales models.TotalSales
	totalSales.Value = value
	return totalSales, nil
}

func (as *AggregationService) GetPopularItems(ctx context.Context) (models.PopularItems, error) {
	popularItems, err := as.AggregationRepo.GetPopularItems(ctx)
	if err != nil {
		return models.PopularItems{}, err
	}
	return popularItems, nil
}

func (as *AggregationService) GetSearchItems(ctx context.Context, q string, filter string, maxPrice float64, minPrice float64) (models.Search, error) {
	filters := strings.Split(filter, ",")
	result, err := as.AggregationRepo.GetSearchItems(ctx, q, filters, maxPrice, minPrice)
	if err != nil {
		return models.Search{}, err
	}
	return result, nil
}

func (as *AggregationService) GetListOfOrderedItems(ctx context.Context, period string, month string, year string) (models.ListOrderedItemByPeriods, error) {
	list, err := as.AggregationRepo.GetListOfOrderedItems(ctx, period, month, year)
	if err != nil {
		return models.ListOrderedItemByPeriods{}, err
	}
	list.Period = period
	list.Month = month
	list.Year = year
	return list, nil
}
