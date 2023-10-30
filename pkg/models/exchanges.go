package models

import "fmt"

/*
FILTERED REQUEST: https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v1/accounting/od/rates_of_exchange?fields=country_currency_desc,exchange_rate,record_date&filter=country_currency_desc:in:(Canada-Dollar,Mexico-Peso),record_date:gte:2020-01-01
	params:
	/v1/accounting/od/rates_of_exchange?
	fields:
		- country_currency_desc
		- exchange_rate
		- record_date
	filter:
		- country_currency_desc?
			:in:(Canada-Dollar,Mexico-Peso)
			:record_date:gte:2020-01-01

	single date, single country: https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v1/accounting/od/rates_of_exchange?fields=country_currency_desc,exchange_rate,record_date&filter=country_currency_desc:in:(Mexico-Peso),record_date:eq:2023-09-30
	single date, all countries: https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v1/accounting/od/rates_of_exchange?fields=country_currency_desc,exchange_rate,record_date&filter=record_date:eq:2023-09-30
	https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v1/accounting/od/rates_of_exchange?fields=country_currency_desc,exchange_rate,record_date&filter=record_date:eq:2023-09-29

	lower of equal date, all countries: https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v1/accounting/od/rates_of_exchange?fields=record_date,country_currency_desc,exchange_rate,effective_date&filter=effective_date:lte:2023-09-30&sort=-effective_date&page[number]=1&page[size]=300
*/

type (
	Purchase struct {
		/*
			Description: must not exceed 50 characters
			Transaction date: must be a valid date format
			Purchase amount: must be a valid positive amount rounded to the nearest cent
			Unique identifier: must uniquely identify the purchase
		*/

		Id          string `json:"id"`
		Description string `json:"description"`
		Amount      string `json:"amount"`
		Date        string `json:"date"`
		signature   string `json:"-"`
	}

	ConvertedAmount struct {
		/*
			https://fiscaldata.treasury.gov/datasets/treasury-reporting-rates-exchange/treasury-reporting-rates-of-exchange
			The retrieved purchase should include the identifier, the description, the transaction date, the original US dollar purchase
			amount, the exchange rate used, and the converted amount based upon the specified currency’s exchange rate for the
			date of the purchase.

			Endpoints:
				BASE URL:
					https://api.fiscaldata.treasury.gov/services/api/fiscal_service/
				ENDPOINT:
					v1/accounting/od/rates_of_exchange
				FULL URL:
					https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v1/accounting/od/rates_of_exchange

			Currency conversion requirements:
				● When converting between currencies, you do not need an exact date match, but must use a currency conversion
					rate less than or equal to the purchase date from within the last 6 months.
				● If no currency conversion rate is available within 6 months equal to or before the purchase date, an error should
					be returned stating the purchase cannot be converted to the target currency.
				● The converted purchase amount to the target currency should be rounded to two decimal places (i.e., cent).
		*/
		Id              string `json:"id"`
		Description     string `json:"description"`
		PurchaseDate    string `json:"purchase_date"`
		OriginalAmount  string `json:"original_amount"`
		ExchangeRate    string `json:"exchange_rate"`
		ConvertedAmount string `json:"converted_amount"`
	}

	DataVal struct {
		CountryCurrencyDesc string `json:"country_currency_desc"`
		ExchangeRate        string `json:"exchange_rate"`
		RecordDate          string `json:"record_date"`
		EffectiveDate       string `json:"effective_date"`
	}

	LabelsVal struct {
		CountryCurrencyDesc string `json:"country_currency_desc"`
		ExchangeRate        string `json:"exchange_rate"`
		RecordDate          string `json:"record_date"`
		EffectiveDate       string `json:"effective_date"`
	}

	DataTypesVal struct {
		CountryCurrencyDesc string `json:"country_currency_desc"`
		ExchangeRate        string `json:"exchange_rate"`
		RecordDate          string `json:"record_date"`
		EffectiveDate       string `json:"effective_date"`
	}

	DataFormatsVal struct {
		CountryCurrencyDesc string `json:"country_currency_desc"`
		ExchangeRate        string `json:"exchange_rate"`
		RecordDate          string `json:"record_date"`
		EffectiveDate       string `json:"effective_date"`
	}

	MetaVal struct {
		Count       int             `json:"count"`
		Labels      *LabelsVal      `json:"labels"`
		Datatypes   *DataTypesVal   `json:"dataTypes"`
		DataFormats *DataFormatsVal `json:"dataFormats"`
		TotalCount  int             `json:"total-count"`
		TotalPages  int             `json:"total-pages"`
	}

	LinksVal struct {
		Self  string `json:"self"`
		First string `json:"first"`
		Prev  string `json:"prev"`
		Next  string `json:"next"`
		Last  string `json:"last"`
	}

	ExchangesReturn struct {
		Data  []*DataVal `json:"data"`
		Meta  *MetaVal   `json:"meta"`
		Links *LinksVal  `json:"links"`
	}

	ExchangeForDate struct {
		//ID                  string `json:"id"`
		Date                string `json:"date"`
		CountryCurrencyDesc string `json:"country_currency_desc"`
		ExchangeRate        string `json:"exchange_rate"`
	}
)

func (p *Purchase) Signature() string {
	if p.signature == "" {
		p.signature = fmt.Sprintf("%s_%s_%s", p.Amount, p.Date, firstN(p.Description, 20))
	}

	return p.signature
}

func firstN(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
