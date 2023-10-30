package services

import (
	"context"

	"github.com/marcosArruda/purchases-multi-country/pkg/models"
)

type (
	noOpsTreasuryAccessService struct {
		sm ServiceManager
	}
)

func NewNoOpsTreasuryAccessService() TreasuryAccessService {
	return &noOpsTreasuryAccessService{}
}

func (n *noOpsTreasuryAccessService) Start(ctx context.Context) error {
	return nil
}

func (n *noOpsTreasuryAccessService) Close(ctx context.Context) error {
	return nil
}

func (n *noOpsTreasuryAccessService) Healthy(ctx context.Context) error {
	return nil
}

func (n *noOpsTreasuryAccessService) WithServiceManager(sm ServiceManager) TreasuryAccessService {
	n.sm = sm
	return n
}

func (n *noOpsTreasuryAccessService) ServiceManager() ServiceManager {
	return n.sm
}

func (n *noOpsTreasuryAccessService) GetExchangesForDate(ctx context.Context, id string) ([]*models.ExchangeForDate, error) {
	return nil, nil
}

func (n *noOpsTreasuryAccessService) GetSpecificExchangeForDateAndCurrency(ctx context.Context, date string, countrycurrency string) (*models.ExchangeForDate, error) {
	return nil, nil
}
