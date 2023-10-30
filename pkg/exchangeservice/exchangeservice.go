package exchangeservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/marcosArruda/purchases-multi-country/pkg/models"
	"github.com/marcosArruda/purchases-multi-country/pkg/services"
	"github.com/shopspring/decimal"
)

type (
	exchangeServiceFinal struct {
		sm services.ServiceManager
	}
)

func NewExchangeService() services.ExchangeService {
	return &exchangeServiceFinal{}
}

func (n *exchangeServiceFinal) Start(ctx context.Context) error {
	n.sm.LogsService().Info(ctx, "Exchange Service Started!")
	return nil
}

func (n *exchangeServiceFinal) Close(ctx context.Context) error {
	return nil
}

func (n *exchangeServiceFinal) Healthy(ctx context.Context) error {
	return nil
}

func (n *exchangeServiceFinal) WithServiceManager(sm services.ServiceManager) services.ExchangeService {
	n.sm = sm
	return n
}

func (n *exchangeServiceFinal) ServiceManager() services.ServiceManager {
	return n.sm
}

func (n *exchangeServiceFinal) HandleNewPurchase(ctx context.Context, p *models.Purchase) error {
	if p == nil {
		return errors.New("cannot insert nil Purchase")
	}
	if exists, _ := n.sm.PersistenceService().ExistsBySignature(ctx, p.Signature()); exists {
		return nil
	}
	if p.Id == "" {
		p.Id = uuid.NewString()
	}

	err := n.sm.PersistenceService().InsertPurchase(ctx, p)
	if err != nil {
		return err
	}

	asynContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() {
		n.sm.AsyncWorkChannel() <- func() error { //async collect and persist  ...
			defer cancel()
			n.sm.LogsService().Info(ctx, "starting async collect of exchanges..")
			exchanges, err := n.CollectExchangeRatesForPurchase(ctx, p)
			if err != nil {
				n.sm.LogsService().Error(ctx, err.Error())
				return err
			}

			return n.sm.PersistenceService().BatchInsertExchanges(asynContext, p, exchanges)
		}
	}()
	return nil
}

func (n *exchangeServiceFinal) GetAllPurchases(ctx context.Context, countrycurrency string) ([]*models.ConvertedAmount, error) {
	purchases, err := n.sm.PersistenceService().ListAllPurchases(ctx)
	if err != nil {
		n.sm.LogsService().Error(ctx, err.Error())
		return services.EmptyConvertedPurchasesSlice, err
	}

	var converteds []*models.ConvertedAmount
	for _, v := range purchases {
		exchange, err := n.sm.PersistenceService().GetExchangeRateForCountryCurrencyAndDate(ctx, countrycurrency, v.Date)
		if err != nil {
			n.sm.LogsService().Error(ctx, err.Error())
			return nil, err
		}

		c, err := n.convertPurchaseByExchangeRate(ctx, v, exchange.ExchangeRate)
		if err != nil {
			n.sm.LogsService().Error(ctx, err.Error())
			return services.EmptyConvertedPurchasesSlice, err
		}
		converteds = append(converteds, c)
	}
	return converteds, nil
}

func (n *exchangeServiceFinal) SearchPurchasesById(ctx context.Context, id string, countrycurrency string) (*models.ConvertedAmount, error) {
	purchase, err := n.sm.PersistenceService().GetPurchaseById(ctx, id)
	if err != nil {
		n.sm.LogsService().Error(ctx, err.Error())
		return nil, err
	}

	exchange, err := n.sm.PersistenceService().GetExchangeRateForCountryCurrencyAndDate(ctx, countrycurrency, purchase.Date)
	if err != nil {
		n.sm.LogsService().Error(ctx, err.Error())
		return nil, err
	}
	if exchange == nil {
		exchange, err = n.CollectSpecificExchangeRateForPurchase(ctx, purchase, countrycurrency)
		if err != nil {
			n.sm.LogsService().Error(ctx, err.Error())
			return nil, err
		}
	}

	return n.convertPurchaseByExchangeRate(ctx, purchase, exchange.ExchangeRate)
}

func (n *exchangeServiceFinal) CollectExchangeRatesForPurchase(ctx context.Context, p *models.Purchase) ([]*models.ExchangeForDate, error) {

	exchanges, err := n.sm.TreasuryAccessService().GetExchangesForDate(ctx, p.Date)
	if err != nil {
		msg := fmt.Sprintf("Error Calling the TreasuryAccess API: %s", err.Error())
		n.sm.LogsService().Error(ctx, msg)
		return nil, err
	}

	return exchanges, nil
}

func (n *exchangeServiceFinal) CollectSpecificExchangeRateForPurchase(ctx context.Context, p *models.Purchase, countrycurrency string) (*models.ExchangeForDate, error) {

	exchange, err := n.sm.TreasuryAccessService().GetSpecificExchangeForDateAndCurrency(ctx, p.Date, countrycurrency)
	if err != nil {
		msg := fmt.Sprintf("Error Calling the TreasuryAccess API: %s", err.Error())
		n.sm.LogsService().Error(ctx, msg)
		return nil, err
	}

	err = n.sm.PersistenceService().InsertExchange(ctx, p, exchange)
	if err != nil {
		msg := fmt.Sprintf("Error Inserting specific exchange for signature '%s': %s", p.Signature(), err.Error())
		n.sm.LogsService().Error(ctx, msg)
		return nil, err
	}

	return exchange, nil
}

func (n *exchangeServiceFinal) convertPurchaseByExchangeRate(ctx context.Context, p *models.Purchase, exchangeRateVal string) (*models.ConvertedAmount, error) {
	exchangeRate, err := decimal.NewFromString(exchangeRateVal)
	if err != nil {
		return nil, fmt.Errorf("error converting the exchangeRate '%s' to decimal: %s ", exchangeRateVal, err.Error())
	}

	originalAmount, err := decimal.NewFromString(p.Amount)
	if err != nil {
		return nil, fmt.Errorf("error converting the original Amount '%s' to decimal: %s ", p.Amount, err.Error())
	}

	convertedAmount := originalAmount.Mul(exchangeRate).Round(2)

	return &models.ConvertedAmount{
		Id:              p.Id,
		Description:     p.Description,
		OriginalAmount:  p.Amount,
		PurchaseDate:    p.Date,
		ExchangeRate:    exchangeRateVal,
		ConvertedAmount: convertedAmount.String(),
	}, nil
}
