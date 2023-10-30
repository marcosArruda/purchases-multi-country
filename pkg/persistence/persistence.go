package persistence

import (
	"context"
	"fmt"

	"github.com/marcosArruda/purchases-multi-country/pkg/models"
	"github.com/marcosArruda/purchases-multi-country/pkg/services"
)

type (
	persistenceServiceFinal struct {
		sm services.ServiceManager
	}
)

func NewPersistenceService() services.PersistenceService {
	return &persistenceServiceFinal{}
}

func (n *persistenceServiceFinal) Start(ctx context.Context) error {
	n.sm.LogsService().Info(ctx, "Persistence Started!")
	return nil
}

func (n *persistenceServiceFinal) Close(ctx context.Context) error {
	return nil
}

func (n *persistenceServiceFinal) Healthy(ctx context.Context) error {
	return nil
}

func (n *persistenceServiceFinal) WithServiceManager(sm services.ServiceManager) services.PersistenceService {
	n.sm = sm
	return n
}

func (n *persistenceServiceFinal) ServiceManager() services.ServiceManager {
	return n.sm
}

func (n *persistenceServiceFinal) GetPurchaseById(ctx context.Context, id string) (*models.Purchase, error) {
	return n.ServiceManager().Database().GetPurchaseById(ctx, id)
}

func (n *persistenceServiceFinal) ExistsBySignature(ctx context.Context, signature string) (bool, error) {
	return n.sm.Database().ExistsBySignature(ctx, signature)
}

func (n *persistenceServiceFinal) InsertPurchase(ctx context.Context, p *models.Purchase) error {
	db := n.ServiceManager().Database()
	n.sm.LogsService().Info(ctx, fmt.Sprintf("Inserting new purchase {id: %s, signature: %s}", p.Id, p.Signature()))
	tx, err := db.BeginTransaction(ctx)
	if err != nil {
		db.RollbackTransaction(tx)
		return err
	}
	err = db.InsertPurchase(ctx, tx, p)
	if err != nil {
		db.RollbackTransaction(tx)
		return err
	}
	return nil
}

func (n *persistenceServiceFinal) BatchInsertExchanges(ctx context.Context, p *models.Purchase, exchanges []*models.ExchangeForDate) error {
	db := n.ServiceManager().Database()
	n.sm.LogsService().Info(ctx, fmt.Sprintf("Batch Inserting new exchanges for signature: '%s', exchanges num: %d", p.Signature(), len(exchanges)))
	tx, err := db.BeginTransaction(ctx)
	if err != nil {
		db.RollbackTransaction(tx)
		return err
	}
	err = db.BatchInsertExchanges(ctx, tx, exchanges)
	if err != nil {
		db.RollbackTransaction(tx)
		return err
	}
	return nil
}

func (n *persistenceServiceFinal) InsertExchange(ctx context.Context, p *models.Purchase, exchange *models.ExchangeForDate) error {
	db := n.ServiceManager().Database()
	n.sm.LogsService().Info(ctx, fmt.Sprintf("Inserting new exchange for '%s' and purchase signature: '%s'", exchange.CountryCurrencyDesc, p.Signature()))
	tx, err := db.BeginTransaction(ctx)
	if err != nil {
		db.RollbackTransaction(tx)
		return err
	}
	err = db.InsertExchange(ctx, tx, exchange)
	if err != nil {
		db.RollbackTransaction(tx)
		return err
	}
	return nil
}

func (n *persistenceServiceFinal) GetExchangeRateForCountryCurrencyAndDate(ctx context.Context, countrycurrency string, date string) (*models.ExchangeForDate, error) {
	return n.sm.Database().GetExchangeRateForCountryCurrencyAndDate(ctx, countrycurrency, date)
}

func (n *persistenceServiceFinal) ListAllPurchases(ctx context.Context) ([]*models.Purchase, error) {
	return n.sm.Database().ListAllPurchases(ctx)
}
