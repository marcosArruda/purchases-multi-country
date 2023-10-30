package services

import (
	"context"

	"github.com/marcosArruda/purchases-multi-country/pkg/models"
)

type (
	noOpsExchangeService struct {
		sm ServiceManager
	}
)

func NewNoOpsExchangeService() ExchangeService {
	return &noOpsExchangeService{}
}

func (n *noOpsExchangeService) Start(ctx context.Context) error {
	return nil
}

func (n *noOpsExchangeService) Close(ctx context.Context) error {
	return nil
}

func (n *noOpsExchangeService) Healthy(ctx context.Context) error {
	return nil
}

func (n *noOpsExchangeService) WithServiceManager(sm ServiceManager) ExchangeService {
	n.sm = sm
	return n
}

func (n *noOpsExchangeService) ServiceManager() ServiceManager {
	return n.sm
}

func (n *noOpsExchangeService) HandleNewPurchase(ctx context.Context, p *models.Purchase) error {
	return nil
}

func (n *noOpsExchangeService) SearchPurchasesById(ctx context.Context, id string, countrycurrency string) (*models.ConvertedAmount, error) {
	return nil, nil
}

func (n *noOpsExchangeService) GetAllPurchases(ctx context.Context, countrycurrency string) ([]*models.ConvertedAmount, error) {
	return nil, nil
}

func (n *noOpsExchangeService) CollectExchangeRatesForPurchase(ctx context.Context, p *models.Purchase) ([]*models.ExchangeForDate, error) {
	return nil, nil
}

func (n *noOpsExchangeService) SearchPurchasesByDate(ctx context.Context, id int) error {
	return nil
}

func (n *noOpsExchangeService) SearchPurchasesByDescription(ctx context.Context, desc string) error {
	return nil
}
