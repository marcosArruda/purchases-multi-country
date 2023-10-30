package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/marcosArruda/purchases-multi-country/pkg/messages"
	"github.com/marcosArruda/purchases-multi-country/pkg/models"
	"github.com/marcosArruda/purchases-multi-country/pkg/services"
)

type (
	mysqlDatabaseFinal struct {
		sm services.ServiceManager
		db *sql.DB
		//mockDb bool
	}
	Key string
)

var (
	version             string
	MockDbKey           Key = "mockDb"
	purchaseCreateTable     = `CREATE TABLE IF NOT EXISTS purchase (
		id VARCHAR(255) PRIMARY KEY,
		description VARCHAR(255) NOT NULL,
		amount VARCHAR(50) NOT NULL,
		date VARCHAR(40),
		signature VARCHAR(255),
		INDEX (date)
	)`

	exchangeCreateTable = `CREATE TABLE IF NOT EXISTS exchange (
		date VARCHAR(40),
		country_currency_desc VARCHAR(255) NOT NULL,
		exchange_rate VARCHAR(50) NOT NULL,
		PRIMARY KEY (country_currency_desc,date)
	)`
)

func NewDatabase() services.Database {
	return &mysqlDatabaseFinal{}
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	//https://gorm.io/docs/connecting_to_the_database.html
}

func (n *mysqlDatabaseFinal) buildConnection(ctx context.Context, mockDb *sql.DB) error {
	if mockDb == nil {
		dbName := os.Getenv("DB_NAME")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbHostPort := os.Getenv("DB_HOSTPORT")
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPass, dbHostPort, dbName))
		if err != nil {
			n.sm.LogsService().Error(ctx, err.Error())
			return err
		}
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Minute * 5)
		n.db = db
	} else {
		n.db = mockDb
	}
	return nil
}

func (n *mysqlDatabaseFinal) Start(ctx context.Context) error {
	sm := n.ServiceManager()
	mdb := ctx.Value(MockDbKey)
	var err error
	if mdb == nil {
		err = n.buildConnection(ctx, nil)
	} else {
		err = n.buildConnection(ctx, ctx.Value(MockDbKey).(*sql.DB))
	}
	if err != nil {
		return err
	}

	sm.LogsService().Info(ctx, "Database Started!")
	sm.LogsService().Info(ctx, "Creating Tables ...")
	if err = n.createTablesIfNotExists(ctx); err != nil {
		sm.LogsService().Error(ctx, "Error creating tables: "+err.Error())
		return err
	}
	sm.LogsService().Info(ctx, "Basic Tables Created!")
	return nil
}

func (n *mysqlDatabaseFinal) createTablesIfNotExists(ctx context.Context) error {
	_, err := n.db.ExecContext(ctx, purchaseCreateTable)
	if err != nil {
		return err
	}

	_, err = n.db.ExecContext(ctx, exchangeCreateTable)
	if err != nil {
		return err
	}
	return nil
}

func (n *mysqlDatabaseFinal) Close(ctx context.Context) error {
	return n.db.Close()
}

func (n *mysqlDatabaseFinal) Healthy(ctx context.Context) error {
	return n.db.QueryRow("SELECT VERSION()").Scan(&version)
}

func (n *mysqlDatabaseFinal) WithServiceManager(sm services.ServiceManager) services.Database {
	n.sm = sm
	return n
}

func (n *mysqlDatabaseFinal) ServiceManager() services.ServiceManager {
	return n.sm
}

func (n *mysqlDatabaseFinal) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	//txCtx, _ := context.WithTimeout(ctx, 10*time.Second)
	//defer cancel()

	tx, err := n.db.Begin()
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("Error creating transaction: %s", err.Error()))
		return nil, err
	}
	return tx, nil
}

func (n *mysqlDatabaseFinal) CommitTransaction(tx *sql.Tx) error {
	return tx.Commit()
}

func (n *mysqlDatabaseFinal) RollbackTransaction(tx *sql.Tx) error {
	return tx.Rollback()
}

func (n *mysqlDatabaseFinal) InsertPurchase(ctx context.Context, tx *sql.Tx, p *models.Purchase) error {
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO purchase(id, description, amount, date, signature) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("Error when preparing SQL statement: %s", err.Error()))
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, p.Id, p.Description, p.Amount, p.Date, p.Signature())
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("Error when inserting row into purchase table: %s", err.Error()))
		return err
	}

	if err := tx.Commit(); err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("Error commiting purchase insert transaction: %s", err.Error()))
		return err
	}
	n.sm.LogsService().Info(ctx, fmt.Sprintf("Purchase Inserted! Purchase Signature: '%s'", p.Signature()))
	return nil
}

func (n *mysqlDatabaseFinal) InsertExchange(ctx context.Context, tx *sql.Tx, ex *models.ExchangeForDate) error {
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO exchange(date, country_currency_desc, exchange_rate) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE exchange_rate = VALUES(exchange_rate)")
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("Error when preparing SQL statement: %s", err.Error()))
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, ex.Date, ex.CountryCurrencyDesc, ex.ExchangeRate)
	if err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("Error when inserting row into exchange table: %s", err.Error()))
		return err
	}

	if err := tx.Commit(); err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("Error commiting exchange insert transaction: %s", err.Error()))
		return err
	}
	n.sm.LogsService().Info(ctx, fmt.Sprintf("Exchange Inserted! For countrycurrency: '%s' and date: '%s'", ex.CountryCurrencyDesc, ex.Date))
	return nil
}

func (n *mysqlDatabaseFinal) BatchInsertExchanges(ctx context.Context, tx *sql.Tx, exchanges []*models.ExchangeForDate) error {
	valueStrings := []string{}
	valueArgs := []interface{}{}
	for _, ex := range exchanges {
		valueStrings = append(valueStrings, "(?, ?, ?)")
		valueArgs = append(valueArgs, ex.Date)
		valueArgs = append(valueArgs, ex.CountryCurrencyDesc)
		valueArgs = append(valueArgs, ex.ExchangeRate)
	}
	smt := `INSERT INTO exchange(date, country_currency_desc, exchange_rate) VALUES %s ON DUPLICATE KEY UPDATE exchange_rate = VALUES(exchange_rate)`
	smt = fmt.Sprintf(smt, strings.Join(valueStrings, ","))
	fmt.Println("smttt:", smt)
	_, err := tx.Exec(smt, valueArgs...)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (n *mysqlDatabaseFinal) ExistsBySignature(ctx context.Context, signature string) (bool, error) {
	count := 0
	n.db.QueryRow("SELECT count(1) FROM purchase WHERE signature = ?", signature).Scan(count)
	return count > 0, nil
}

func (n *mysqlDatabaseFinal) GetPurchaseById(ctx context.Context, id string) (*models.Purchase, error) {
	p := &models.Purchase{}
	err := n.db.QueryRow("SELECT id, description, amount, date FROM purchase WHERE id = ?", id).Scan(&p.Id, &p.Description, &p.Amount, &p.Date)
	if err == sql.ErrNoRows {
		return nil, messages.ErrNoPurchaseFound
	}
	if err != nil {
		msg := fmt.Sprintf("Something went wrong searching by the Purchase with ID %s: %s", id, err.Error())
		return nil, &messages.PurchaseError{Msg: msg, PurchaseId: id}
	}
	return p, nil
}

func (n *mysqlDatabaseFinal) ListAllPurchases(ctx context.Context) ([]*models.Purchase, error) {
	pRows, err := n.db.Query("SELECT id, description, amount, date FROM purchase")
	if err != nil {
		if err == sql.ErrNoRows {
			return services.EmptyPurchasesSlice, messages.ErrNoPurchaseFound
		}
		return n.emptyAndGenericError(err)
	}
	defer pRows.Close()
	var purchases []*models.Purchase
	for pRows.Next() {
		var p models.Purchase
		if err := pRows.Scan(&p.Id, &p.Description, &p.Amount, &p.Date); err != nil {
			return n.emptyAndGenericError(err)
		}
		purchases = append(purchases, &p)
	}
	if err := pRows.Err(); err != nil {
		return n.emptyAndGenericError(err)
	}
	return purchases, nil
}

func (n *mysqlDatabaseFinal) GetExchangeRateForCountryCurrencyAndDate(ctx context.Context, countrycurrency string, date string) (*models.ExchangeForDate, error) {
	p := &models.ExchangeForDate{}
	err := n.db.QueryRow("SELECT date, country_currency_desc, exchange_rate from exchange WHERE DATE(date) <= DATE(?) AND country_currency_desc = ? ORDER BY DATE(date) DESC", date, countrycurrency).Scan(&p.Date, &p.CountryCurrencyDesc, &p.ExchangeRate)
	if err == sql.ErrNoRows {
		return nil, messages.ErrNoExchangeFound
	}
	if err != nil {
		msg := fmt.Sprintf("Something went wrong searching by the Exchange {contrycurrency: %s, date: %s}: %s", countrycurrency, date, err.Error())
		return nil, &messages.ExchangeError{Msg: msg, ExchangeDate: date, ExchangeCurrency: countrycurrency}
	}
	return p, nil
}

func (n *mysqlDatabaseFinal) emptyAndGenericError(err error) ([]*models.Purchase, error) {
	baseMsg := "Something went wrong searching by All purchases: "
	msg := fmt.Sprintf("%s%s", baseMsg, err.Error())
	return services.EmptyPurchasesSlice, &messages.PurchaseError{Msg: msg}
}
