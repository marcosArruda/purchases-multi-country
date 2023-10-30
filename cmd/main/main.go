package main

import (
	"context"
	"fmt"

	"github.com/marcosArruda/purchases-multi-country/pkg/exchangeservice"
	"github.com/marcosArruda/purchases-multi-country/pkg/httpservice"
	"github.com/marcosArruda/purchases-multi-country/pkg/logs"
	"github.com/marcosArruda/purchases-multi-country/pkg/persistence"
	"github.com/marcosArruda/purchases-multi-country/pkg/services"
	"github.com/marcosArruda/purchases-multi-country/pkg/treasuryaccess"
)

func main() {
	ctx := context.Background()
	//time.Sleep(5 * time.Second)
	asyncWorkChannel := make(chan func() error)
	stop := make(chan struct{})

	sm := services.NewManager(asyncWorkChannel, stop).
		WithLogsService(logs.NewLogsService()).
		WithDatabase(persistence.NewDatabase()).
		WithPersistenceService(persistence.NewPersistenceService()).
		WithExchangeService(exchangeservice.NewExchangeService()).
		WithTreasuryAccessService(treasuryaccess.NewTreasuryAccessService()).
		WithHttpService(httpservice.NewHttpService())

	// This is the goroutine that will execute any async work
	go func() {
	basicLoop:
		for {
			select {
			case w := <-asyncWorkChannel:
				if err := w(); err != nil {
					sm.LogsService().Warn(ctx, fmt.Sprintf("an async work failed: %s", err.Error()))
				}
			case <-stop: // triggered when the stop channel is closed
				break basicLoop // stop listening
			}
		}
	}()

	sm.Start(ctx)
}
