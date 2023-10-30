package services

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/marcosArruda/purchases-multi-country/pkg/models"
)

var (
	EmptyPurchasesSlice          = make([]*models.Purchase, 0)
	EmptyConvertedPurchasesSlice = make([]*models.ConvertedAmount, 0)
)

type (
	GenericService interface {
		Start(ctx context.Context) error
		Close(ctx context.Context) error
		Healthy(ctx context.Context) error
	}

	LogsService interface {
		GenericService
		WithServiceManager(sm ServiceManager) LogsService
		ServiceManager() ServiceManager
		Info(ctx context.Context, s string)
		Warn(ctx context.Context, s string)
		Error(ctx context.Context, s string)
		Debug(ctx context.Context, s string)
	}

	Database interface {
		GenericService
		WithServiceManager(sm ServiceManager) Database
		ServiceManager() ServiceManager
		BeginTransaction(ctx context.Context) (*sql.Tx, error)
		CommitTransaction(tx *sql.Tx) error
		RollbackTransaction(tx *sql.Tx) error
		InsertPurchase(ctx context.Context, tx *sql.Tx, p *models.Purchase) error
		BatchInsertExchanges(ctx context.Context, tx *sql.Tx, exchanges []*models.ExchangeForDate) error
		GetPurchaseById(ctx context.Context, id string) (*models.Purchase, error)
		ExistsBySignature(ctx context.Context, signature string) (bool, error)
		ListAllPurchases(ctx context.Context) ([]*models.Purchase, error)
		InsertExchange(ctx context.Context, tx *sql.Tx, ex *models.ExchangeForDate) error
		GetExchangeRateForCountryCurrencyAndDate(ctx context.Context, countrycurrency string, date string) (*models.ExchangeForDate, error)
	}

	PersistenceService interface {
		GenericService
		WithServiceManager(sm ServiceManager) PersistenceService
		ServiceManager() ServiceManager
		InsertPurchase(ctx context.Context, p *models.Purchase) error
		BatchInsertExchanges(ctx context.Context, p *models.Purchase, exchanges []*models.ExchangeForDate) error
		GetPurchaseById(ctx context.Context, id string) (*models.Purchase, error)
		ExistsBySignature(ctx context.Context, signature string) (bool, error)
		ListAllPurchases(ctx context.Context) ([]*models.Purchase, error)
		GetExchangeRateForCountryCurrencyAndDate(ctx context.Context, countrycurrency string, date string) (*models.ExchangeForDate, error)
		InsertExchange(ctx context.Context, p *models.Purchase, exchange *models.ExchangeForDate) error
	}

	ExchangeService interface {
		GenericService
		WithServiceManager(sm ServiceManager) ExchangeService
		ServiceManager() ServiceManager
		HandleNewPurchase(ctx context.Context, p *models.Purchase) error
		GetAllPurchases(ctx context.Context, countrycurrency string) ([]*models.ConvertedAmount, error)
		SearchPurchasesById(ctx context.Context, id string, countrycurrency string) (*models.ConvertedAmount, error)
		CollectExchangeRatesForPurchase(ctx context.Context, p *models.Purchase) ([]*models.ExchangeForDate, error)
	}

	HttpService interface {
		GenericService
		WithServiceManager(sm ServiceManager) HttpService
		ServiceManager() ServiceManager
		PostPurchase(c *gin.Context)
		GetPurchaseById(c *gin.Context)
		GetAllPurchases(c *gin.Context)
	}

	TreasuryAccessService interface {
		GenericService
		WithServiceManager(sm ServiceManager) TreasuryAccessService
		ServiceManager() ServiceManager
		GetExchangesForDate(ctx context.Context, date string) ([]*models.ExchangeForDate, error)
		GetSpecificExchangeForDateAndCurrency(ctx context.Context, date string, countrycurrency string) (*models.ExchangeForDate, error)
	}

	ServiceManager interface {
		GenericService
		WithLogsService(ls LogsService) ServiceManager
		LogsService() LogsService
		WithDatabase(db Database) ServiceManager
		Database() Database
		WithPersistenceService(p PersistenceService) ServiceManager
		PersistenceService() PersistenceService
		WithExchangeService(p ExchangeService) ServiceManager
		ExchangeService() ExchangeService
		WithTreasuryAccessService(p TreasuryAccessService) ServiceManager
		TreasuryAccessService() TreasuryAccessService
		WithHttpService(h HttpService) ServiceManager
		HttpService() HttpService
		AsyncWorkChannel() chan func() error
	}

	serviceManagerFinal struct {
		logsService           LogsService
		asyncWorkChannel      chan func() error
		stop                  chan struct{}
		database              Database
		persistenceService    PersistenceService
		exchangeService       ExchangeService
		treasuryAccessService TreasuryAccessService
		httpService           HttpService
	}
)

func NewManager(asyncWorkChannel chan func() error, stop chan struct{}) ServiceManager {
	return &serviceManagerFinal{
		logsService:           NewNoOpsLogsService(),
		asyncWorkChannel:      asyncWorkChannel,
		stop:                  stop,
		database:              NewNoOpsDatabase(),
		persistenceService:    NewNoOpsPersistenceService(),
		exchangeService:       NewNoOpsExchangeService(),
		treasuryAccessService: NewNoOpsTreasuryAccessService(),
		httpService:           NewNoOpsHttpService(),
	}
}

func (m *serviceManagerFinal) Start(ctx context.Context) error {
	if err := m.logsService.Start(ctx); err != nil {
		return err
	}

	if err := m.database.Start(ctx); err != nil {
		m.logsService.Error(ctx, err.Error())
		return err
	}

	if err := m.persistenceService.Start(ctx); err != nil {
		m.logsService.Error(ctx, err.Error())
		return err
	}

	if err := m.exchangeService.Start(ctx); err != nil {
		m.logsService.Error(ctx, err.Error())
		return err
	}

	if err := m.treasuryAccessService.Start(ctx); err != nil {
		m.logsService.Error(ctx, err.Error())
		return err
	}

	if err := m.httpService.Start(ctx); err != nil {
		m.logsService.Error(ctx, err.Error())
		return err
	}

	return nil
}

func (m *serviceManagerFinal) Close(ctx context.Context) error {
	if err := m.logsService.Close(ctx); err != nil {
		return err
	}

	if err := m.persistenceService.Close(ctx); err != nil {
		m.logsService.Error(ctx, err.Error())
		return err
	}

	if err := m.httpService.Close(ctx); err != nil {
		m.logsService.Error(ctx, err.Error())
		return err
	}

	return nil
}

func (m *serviceManagerFinal) Healthy(ctx context.Context) error {
	if err := m.logsService.Healthy(ctx); err != nil {
		return err
	}

	if err := m.persistenceService.Healthy(ctx); err != nil {
		m.logsService.Error(ctx, err.Error())
		return err
	}

	if err := m.httpService.Healthy(ctx); err != nil {
		m.logsService.Error(ctx, err.Error())
		return err
	}

	return nil
}

func (m *serviceManagerFinal) WithLogsService(ls LogsService) ServiceManager {
	m.logsService = ls.WithServiceManager(m)
	return m
}

func (m *serviceManagerFinal) LogsService() LogsService {
	return m.logsService
}

func (m *serviceManagerFinal) WithHttpService(h HttpService) ServiceManager {
	m.httpService = h.WithServiceManager(m)
	return m
}

func (m *serviceManagerFinal) HttpService() HttpService {
	return m.httpService
}

func (m *serviceManagerFinal) WithPersistenceService(p PersistenceService) ServiceManager {
	m.persistenceService = p.WithServiceManager(m)
	return m
}

func (m *serviceManagerFinal) PersistenceService() PersistenceService {
	return m.persistenceService
}

func (m *serviceManagerFinal) WithDatabase(db Database) ServiceManager {
	m.database = db.WithServiceManager(m)
	return m
}

func (m *serviceManagerFinal) Database() Database {
	return m.database
}

func (m *serviceManagerFinal) WithExchangeService(p ExchangeService) ServiceManager {
	m.exchangeService = p.WithServiceManager(m)
	return m
}
func (m *serviceManagerFinal) ExchangeService() ExchangeService {
	return m.exchangeService
}

func (m *serviceManagerFinal) WithTreasuryAccessService(p TreasuryAccessService) ServiceManager {
	m.treasuryAccessService = p.WithServiceManager(m)
	return m
}
func (m *serviceManagerFinal) TreasuryAccessService() TreasuryAccessService {
	return m.treasuryAccessService
}

func (m *serviceManagerFinal) AsyncWorkChannel() chan func() error {
	return m.asyncWorkChannel
}
