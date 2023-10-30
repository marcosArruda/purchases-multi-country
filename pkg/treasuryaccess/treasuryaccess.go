package treasuryaccess

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/marcosArruda/purchases-multi-country/pkg/models"
	"github.com/marcosArruda/purchases-multi-country/pkg/services"
)

type (
	BasicHttpClient interface {
		Do(req *http.Request) (*http.Response, error)
	}
	treasuryAccessClientFinal struct {
		sm                   services.ServiceManager
		searchableHttpClient BasicHttpClient
	}
)

var (
	baseURL = "https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v1/accounting/od/rates_of_exchange?fields=record_date,country_currency_desc,exchange_rate,effective_date&filter=%s&sort=-effective_date&page[number]=1&page[size]=200"
)

func NewTreasuryAccessService() services.TreasuryAccessService {
	return &treasuryAccessClientFinal{searchableHttpClient: http.DefaultClient}
}

func (n *treasuryAccessClientFinal) Start(ctx context.Context) error {
	n.sm.LogsService().Info(ctx, "TreasuryAccess Service Started Started!")
	return nil
}

func (n *treasuryAccessClientFinal) Close(ctx context.Context) error {
	return nil
}

func (n *treasuryAccessClientFinal) Healthy(ctx context.Context) error {
	return nil
}

func (n *treasuryAccessClientFinal) WithServiceManager(sm services.ServiceManager) services.TreasuryAccessService {
	n.sm = sm
	return n
}

func (n *treasuryAccessClientFinal) ServiceManager() services.ServiceManager {
	return n.sm
}

func (n *treasuryAccessClientFinal) GetExchangesForDate(ctx context.Context, date string) ([]*models.ExchangeForDate, error) {
	filterDates, err := n.getDateRangeFilter(ctx, date)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: error creating filters: %s", err.Error()))
		return nil, err
	}
	url := fmt.Sprintf(baseURL, filterDates)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: could not create request: %s", err.Error()))
		return nil, err
	}
	res, err := n.searchableHttpClient.Do(req)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: error making http request: %s", err.Error()))
		return nil, err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: could not read response body: %s", err.Error()))
		return nil, err
	}
	var body models.ExchangesReturn
	err = json.Unmarshal(resBody, &body)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: could not unmarshall the body: %s", err.Error()))
		return nil, err
	}
	return n.convertTreasuryResponse(ctx, &body), nil
}

func (n *treasuryAccessClientFinal) GetSpecificExchangeForDateAndCurrency(ctx context.Context, date string, countrycurrency string) (*models.ExchangeForDate, error) {
	filterCurrency := fmt.Sprintf("country_currency_desc:in:(%s)", countrycurrency)
	filterDates, err := n.getDateRangeFilter(ctx, date)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: error creating filters: %s", err.Error()))
		return nil, err
	}
	url := fmt.Sprintf(baseURL, filterDates+","+filterCurrency)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: could not create request: %s", err.Error()))
		return nil, err
	}
	res, err := n.searchableHttpClient.Do(req)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: error making http request: %s", err.Error()))
		return nil, err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: could not read response body: %s", err.Error()))
		return nil, err
	}
	var body models.ExchangesReturn
	err = json.Unmarshal(resBody, &body)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("client: could not unmarshall the body: %s", err.Error()))
		return nil, err
	}
	return n.convertTreasuryResponse(ctx, &body)[0], nil
}

func (n *treasuryAccessClientFinal) convertTreasuryResponse(ctx context.Context, body *models.ExchangesReturn) []*models.ExchangeForDate {

	var exForDate []*models.ExchangeForDate
	for _, d := range body.Data {
		exForDate = append(exForDate, &models.ExchangeForDate{
			Date:                d.EffectiveDate,
			CountryCurrencyDesc: d.CountryCurrencyDesc,
			ExchangeRate:        d.ExchangeRate,
		})
	}
	return exForDate
}

var YYYYMMDD = "2006-01-02"

func (n *treasuryAccessClientFinal) getDateRangeFilter(ctx context.Context, date string) (string, error) {
	d, err := time.Parse(YYYYMMDD, date)
	if err != nil {
		return "", err
	}
	sixMbefore := d.AddDate(0, -6, 0).Format(YYYYMMDD)
	theFilter := fmt.Sprintf("effective_date:gte:%s,effective_date:lte:%s", sixMbefore, date)

	return theFilter, nil
}

/*
func (n *treasuryAccessClientFinal) replaceCountryCurrency(s string) string {

	r := strings.ReplaceAll(s, " ", "")
	r = strings.ReplaceAll(r, "-", "")
	r = strings.ReplaceAll(r, "&", "")
	r = strings.ReplaceAll(r, "'", "")
	r = strings.ReplaceAll(r, "(", "")
	r = strings.ReplaceAll(r, ")", "")

	return r
}
*/
