package messages

import (
	"errors"
)

var (
	ErrSwApiUnavailableError = errors.New("something went wrong accessing treasury data")
	ErrNoPurchaseFound       = errors.New("no Purchase found")
	ErrNoExchangeFound       = errors.New("no Exchange found")
)

type (
	PurchaseError struct {
		Msg        string
		PurchaseId string
	}

	ExchangeError struct {
		Msg              string
		ExchangeDate     string
		ExchangeCurrency string
	}
)

func (p *PurchaseError) Error() string {
	return p.Msg
}

func (f *ExchangeError) Error() string {
	return f.Msg
}
