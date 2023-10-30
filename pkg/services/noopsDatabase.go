package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/marcosArruda/purchases-multi-country/pkg/models"
)

type (
	noOpsDatabase struct {
		sm ServiceManager
	}
)

func NewNoOpsDatabase() Database {
	return &noOpsDatabase{}
}

func (n *noOpsDatabase) Start(ctx context.Context) error {
	return nil
}

func (n *noOpsDatabase) Close(ctx context.Context) error {
	return nil
}

func (n *noOpsDatabase) Healthy(ctx context.Context) error {
	return nil
}

func (n *noOpsDatabase) WithServiceManager(sm ServiceManager) Database {
	n.sm = sm
	return n
}

func (n *noOpsDatabase) ServiceManager() ServiceManager {
	return n.sm
}

func (n *noOpsDatabase) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	if ctx != nil {
		if ctx.Value("error") != nil {
			return nil, errors.New("some error")
		}
		return &sql.Tx{}, nil
	}
	return nil, nil
}
func (n *noOpsDatabase) CommitTransaction(tx *sql.Tx) error {
	if tx != nil {
		return nil
	}
	return errors.New("some error")
}

func (n *noOpsDatabase) RollbackTransaction(tx *sql.Tx) error {
	if tx != nil {
		return nil
	}
	return errors.New("some error")
}

func (n *noOpsDatabase) InsertPurchase(ctx context.Context, tx *sql.Tx, p *models.Purchase) error {
	return nil
}

func (n *noOpsDatabase) BatchInsertExchanges(ctx context.Context, tx *sql.Tx, exchanges []*models.ExchangeForDate) error {
	return nil
}

func (n *noOpsDatabase) ExistsBySignature(ctx context.Context, signature string) (bool, error) {
	return false, nil
}

func (n *noOpsDatabase) GetPurchaseById(ctx context.Context, id string) (*models.Purchase, error) {
	if id != "1" {
		return nil, errors.New("some error")
	}
	return &models.Purchase{
		Id:          "1",
		Description: "Some transaction",
		Amount:      "20.13",
		Date:        "2023-09-30",
	}, nil
}

func (n *noOpsDatabase) ListAllPurchases(ctx context.Context) ([]*models.Purchase, error) {
	if ctx == nil {
		return nil, errors.New("some error")
	}
	return make([]*models.Purchase, 0), nil
}

func (n *noOpsDatabase) InsertExchange(ctx context.Context, tx *sql.Tx, ex *models.ExchangeForDate) error {
	return nil
}

func (n *noOpsDatabase) GetExchangeRateForCountryCurrencyAndDate(ctx context.Context, countrycurrency string, date string) (*models.ExchangeForDate, error) {
	return nil, nil
}
