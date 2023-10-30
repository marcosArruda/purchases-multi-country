package services

import (
	"context"
	"errors"

	"github.com/marcosArruda/purchases-multi-country/pkg/models"
)

type (
	noOpsPersistenceService struct {
		sm ServiceManager
	}
)

func NewNoOpsPersistenceService() PersistenceService {
	return &noOpsPersistenceService{}
}

func (n *noOpsPersistenceService) Start(ctx context.Context) error {
	return nil
}

func (n *noOpsPersistenceService) Close(ctx context.Context) error {
	return nil
}

func (n *noOpsPersistenceService) Healthy(ctx context.Context) error {
	return nil
}

func (n *noOpsPersistenceService) WithServiceManager(sm ServiceManager) PersistenceService {
	n.sm = sm
	return n
}

func (n *noOpsPersistenceService) ServiceManager() ServiceManager {
	return n.sm
}

func (n *noOpsPersistenceService) InsertPurchase(ctx context.Context, p *models.Purchase) error {
	return nil
}

func (n *noOpsPersistenceService) BatchInsertExchanges(ctx context.Context, p *models.Purchase, exchanges []*models.ExchangeForDate) error {
	return nil
}

func (n *noOpsPersistenceService) GetPurchaseById(ctx context.Context, id string) (*models.Purchase, error) {
	return &models.Purchase{
		Id:          id,
		Description: "Some transaction",
		Amount:      "20.13",
		Date:        "2023-09-30",
	}, nil
}

func (n *noOpsPersistenceService) ExistsBySignature(ctx context.Context, signature string) (bool, error) {
	return false, nil
}

func (n *noOpsPersistenceService) ListAllPurchases(ctx context.Context) ([]*models.Purchase, error) {
	return []*models.Purchase{{
		Id:          "abcd-fghi",
		Description: "Some transaction",
		Amount:      "20.13",
		Date:        "2023-09-30",
	}}, nil
}

func (n *noOpsPersistenceService) GetExchangeRateForCountryCurrency(ctx context.Context, countrycurrency string) (*models.ExchangeForDate, error) {
	if countrycurrency == "error" {
		return nil, errors.New("some error")
	}
	return &models.ExchangeForDate{
		Date:                "2023-09-30",
		CountryCurrencyDesc: "Brazil-Real",
		ExchangeRate:        "5.00",
	}, nil
}

func (n *noOpsPersistenceService) InsertExchange(ctx context.Context, p *models.Purchase, exchange *models.ExchangeForDate) error {
	return nil
}

func (n *noOpsPersistenceService) GetExchangeRateForCountryCurrencyAndDate(ctx context.Context, countrycurrency string, date string) (*models.ExchangeForDate, error) {
	if countrycurrency == "error" {
		return nil, errors.New("some error")
	}
	return &models.ExchangeForDate{
		Date:                "2023-09-30",
		CountryCurrencyDesc: "Brazil-Real",
		ExchangeRate:        "5.00",
	}, nil
}
