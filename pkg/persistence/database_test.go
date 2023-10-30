package persistence

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/marcosArruda/purchases-multi-country/pkg/logs"
	"github.com/marcosArruda/purchases-multi-country/pkg/models"
	"github.com/marcosArruda/purchases-multi-country/pkg/services"
)

var (
	basicPurchase = &models.Purchase{
		Id:          "abcd-fghi",
		Description: "Some transaction",
		Amount:      "20.13",
		Date:        "2023-09-30",
	}

	basicExchange = &models.ExchangeForDate{
		CountryCurrencyDesc: "Brazil-Real",
		ExchangeRate:        "5.00",
		Date:                "2023-09-30",
	}
)

func NewManagerForTestsDatabase() (services.ServiceManager, context.Context) {
	asyncWorkChannel := make(chan func() error)
	stop := make(chan struct{})

	os.Setenv("DB_NAME", "dummyName")
	os.Setenv("DB_USER", "dummyUser")
	os.Setenv("DB_PASSWORD", "dummyPassword")
	os.Setenv("DB_HOSTPORT", "dummyHostPort")

	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.AppEnvKey, "TESTS")
	ctx = context.WithValue(ctx, logs.AppNameKey, logs.AppName)
	ctx = context.WithValue(ctx, logs.AppVersionKey, logs.AppVersion)
	return services.NewManager(asyncWorkChannel, stop), ctx
}

func buildMock(t *testing.T, errorIn int) *sql.DB {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	if errorIn == -2 {
		mock.ExpectClose()
	}

	expect := []*sqlmock.ExpectedExec{}
	expect = append(expect, mock.ExpectExec("CREATE TABLE IF NOT EXISTS purchase").WillReturnResult(sqlmock.NewResult(1, 1)))
	expect = append(expect, mock.ExpectExec("CREATE TABLE IF NOT EXISTS exchange").WillReturnResult(sqlmock.NewResult(1, 1)))

	mock.ExpectQuery("SELECT VERSION").WillReturnRows(mock.NewRows([]string{"version"}).AddRow("1.0"))

	if errorIn >= 0 {
		expect[errorIn].WillReturnError(errors.New("some error"))
	}
	return db
}

func buildTransactionsMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

func Test_mysqlDatabaseFinal_buildConnection(t *testing.T) {
	type args struct {
		ctx context.Context
		db  *sql.DB
	}
	sm, ctx := NewManagerForTestsDatabase()
	dbService := sm.WithDatabase(NewDatabase()).Database()
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	tests := []struct {
		name    string
		n       *mysqlDatabaseFinal
		args    args
		wantErr bool
	}{
		{
			name:    "successMocked",
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{ctx: ctx, db: db},
			wantErr: false,
		},
		{
			name:    "successPROD",
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{ctx: ctx, db: nil},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.buildConnection(tt.args.ctx, tt.args.db); (err != nil) != tt.wantErr {
				t.Errorf("mysqlDatabaseFinal.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mysqlDatabaseFinal_Start(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTestsDatabase()
	dbService := sm.WithDatabase(NewDatabase()).Database()

	tests := []struct {
		name    string
		n       *mysqlDatabaseFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{ctx: context.WithValue(ctx, MockDbKey, buildMock(t, -1))},
			wantErr: false,
		},
		{
			name:    "errorPurchase",
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{ctx: context.WithValue(ctx, MockDbKey, buildMock(t, 0))},
			wantErr: true,
		},
		{
			name:    "errorExchange",
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{ctx: context.WithValue(ctx, MockDbKey, buildMock(t, 1))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Start(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("mysqlDatabaseFinal.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
			defer tt.args.ctx.Value(MockDbKey).(*sql.DB).Close()
		})
	}
}

func Test_mysqlDatabaseFinal_Close(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTestsDatabase()
	dbService := sm.WithDatabase(NewDatabase()).Database()
	dbService.Start(context.WithValue(ctx, MockDbKey, buildMock(t, -2)))
	tests := []struct {
		name    string
		n       *mysqlDatabaseFinal
		args    args
		wantErr bool
	}{
		{
			name: "success", //just success because Database.Close() just closes the connection
			// and is intermitent if the connection was not yet created.
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{ctx: ctx},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Close(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("mysqlDatabaseFinal.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mysqlDatabaseFinal_Healthy(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTestsDatabase()
	dbService := sm.WithDatabase(NewDatabase()).Database()
	dbService.Start(context.WithValue(ctx, MockDbKey, buildMock(t, -1)))
	tests := []struct {
		name    string
		n       *mysqlDatabaseFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{ctx},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Healthy(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("mysqlDatabaseFinal.Healthy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mysqlDatabaseFinal_WithServiceManager(t *testing.T) {
	type args struct {
		sm services.ServiceManager
	}
	sm, _ := NewManagerForTestsDatabase()
	dbService := sm.WithDatabase(NewDatabase()).Database()
	tests := []struct {
		name string
		n    *mysqlDatabaseFinal
		args args
		want services.Database
	}{
		{
			name: "success",
			n:    dbService.(*mysqlDatabaseFinal),
			args: args{sm: sm},
			want: dbService,
		},
		{
			name: "successNil",
			n:    dbService.(*mysqlDatabaseFinal),
			args: args{sm: nil},
			want: dbService,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.WithServiceManager(tt.args.sm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mysqlDatabaseFinal.WithServiceManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mysqlDatabaseFinal_ServiceManager(t *testing.T) {
	sm, _ := NewManagerForTestsDatabase()
	tests := []struct {
		name string
		n    *mysqlDatabaseFinal
		want services.ServiceManager
	}{
		{
			name: "success",
			n:    sm.WithDatabase(NewDatabase()).Database().(*mysqlDatabaseFinal),
			want: sm,
		},
		{
			name: "successNil",
			n:    NewDatabase().(*mysqlDatabaseFinal),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.ServiceManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mysqlDatabaseFinal.ServiceManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mysqlDatabaseFinal_BeginTransaction(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTestsDatabase()
	dbService := sm.WithDatabase(NewDatabase()).Database()
	db, mock := buildTransactionsMock(t)
	mock.ExpectBegin()
	ctx = context.WithValue(ctx, MockDbKey, db)
	dbService.Start(ctx)
	tests := []struct {
		name    string
		n       *mysqlDatabaseFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{ctx: ctx},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.n.BeginTransaction(tt.args.ctx)
			if (err == nil) != tt.wantErr {
				t.Errorf("mysqlDatabaseFinal.BeginTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
				return
			}
		})
	}
}

func Test_mysqlDatabaseFinal_CommitTransaction(t *testing.T) {
	type args struct {
		tx *sql.Tx
	}
	sm, _ := NewManagerForTestsDatabase()
	dbService := sm.WithDatabase(NewDatabase()).Database()
	db, mock := buildTransactionsMock(t)
	mock.ExpectBegin()
	mock.ExpectCommit()
	tx, _ := db.Begin()

	tests := []struct {
		name    string
		n       *mysqlDatabaseFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{tx: tx},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.CommitTransaction(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("mysqlDatabaseFinal.CommitTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
				return
			}
		})
	}
}

func Test_mysqlDatabaseFinal_RollbackTransaction(t *testing.T) {
	type args struct {
		tx *sql.Tx
	}
	sm, _ := NewManagerForTestsDatabase()
	dbService := sm.WithDatabase(NewDatabase()).Database()
	db, mock := buildTransactionsMock(t)
	mock.ExpectBegin()
	mock.ExpectRollback()
	tx, _ := db.Begin()
	tests := []struct {
		name    string
		n       *mysqlDatabaseFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			n:       dbService.(*mysqlDatabaseFinal),
			args:    args{tx: tx},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.RollbackTransaction(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("mysqlDatabaseFinal.RollbackTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
				return
			}
		})
	}
}

func Test_mysqlDatabaseFinal_GetPurchaseById(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    *models.Purchase
		dbFunc  func() *sql.DB
		wantErr bool
	}{
		{
			name: "success",
			args: args{id: "1"},
			dbFunc: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS purchase").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS exchange").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("FROM purchase").WithArgs("1").WillReturnRows(sqlmock.NewRows([]string{"id", "description", "amount", "date"}).
					FromCSVString("abcd-fghi,Some transaction,20.13,2023-09-30"))
				mock.ExpectQuery("FROM exchange").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"date", "countrycurrency", "exchangerate"}).
					FromCSVString("2023-09-30,Brazil-Real,20.13"))
				return db
			},
			want:    basicPurchase,
			wantErr: false,
		},
		{
			name: "purchaseErrNoRows",
			args: args{id: "2"},
			dbFunc: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS purchase").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS exchange").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("FROM purchase").WithArgs("2").WillReturnError(sql.ErrNoRows)
				return db
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "purchaseErrAny",
			args: args{id: "3"},
			dbFunc: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS purchase").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS exchange").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("FROM purchase").WithArgs("3").WillReturnError(errors.New("some error"))

				return db
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "purchaseErrNoRows",
			args: args{id: "4"},
			dbFunc: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS purchase").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS exchange").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("FROM purchase").WithArgs("4").WillReturnRows(sqlmock.NewRows([]string{"id", "description", "amount", "date", "signature"}).
					FromCSVString("abcd-fghi,Some transaction,20.13,2023-09-30,20.13_2023-09-30_Sometransaction"))
				mock.ExpectQuery("FROM exchange").WithArgs("4").
					WillReturnError(sql.ErrNoRows)
				return db
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "exchangeErrAny",
			args: args{id: "5"},
			dbFunc: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS purchase").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS exchange").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("FROM purchase").WithArgs("5").WillReturnRows(sqlmock.NewRows([]string{"id", "description", "amount", "date", "signature"}).
					FromCSVString("abcd-fghi,Some transaction,20.13,2023-09-30,20.13_2023-09-30_Sometransaction"))
				mock.ExpectQuery("FROM exchange").WithArgs("5").
					WillReturnError(errors.New("some error"))
				return db
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, ctx := NewManagerForTestsDatabase()
			ctxTmp := context.WithValue(ctx, MockDbKey, tt.dbFunc())
			dbService := sm.WithDatabase(NewDatabase()).Database().(*mysqlDatabaseFinal)
			sm.Start(ctxTmp)
			got, err := dbService.GetPurchaseById(ctxTmp, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: mysqlDatabaseFinal.GetPurchaseById() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !tt.wantErr && !purchaseSuperficialDeepEqual(got, tt.want) {
				t.Errorf("%s: mysqlDatabaseFinal.GetPurchaseById() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_mysqlDatabaseFinal_InsertPurchase(t *testing.T) {
	type args struct {
		newPurchase *models.Purchase
	}
	tests := []struct {
		name    string
		args    args
		dbFunc  func() *sql.DB
		wantErr bool
	}{
		{
			name: "success", //just manual work to make other cases. For now, I will pass, but I KNOW that in production apps we need to cover ALL..
			args: args{newPurchase: basicPurchase},
			dbFunc: func() *sql.DB {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS purchase").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE IF NOT EXISTS exchange").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectBegin()
				mock.ExpectPrepare("INSERT INTO purchase").
					ExpectExec().WithArgs(basicPurchase.Id, basicPurchase.Description, basicPurchase.Amount, basicPurchase.Date, basicPurchase.Signature()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
				return db
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, ctx := NewManagerForTestsDatabase()
			ctxTmp := context.WithValue(ctx, MockDbKey, tt.dbFunc())
			dbService := sm.WithDatabase(NewDatabase()).Database().(*mysqlDatabaseFinal)
			sm.Start(ctxTmp)
			tx, _ := sm.Database().BeginTransaction(ctxTmp)
			if err := dbService.InsertPurchase(ctxTmp, tx, tt.args.newPurchase); (err != nil) != tt.wantErr {
				t.Errorf("mysqlDatabaseFinal.InsertPurchase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_mysqlDatabaseFinal_ListAllPurchases(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		n       *mysqlDatabaseFinal
		args    args
		want    []*models.Purchase
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.ListAllPurchases(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("mysqlDatabaseFinal.ListAllPurchases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mysqlDatabaseFinal.ListAllPurchases() = %v, want %v", got, tt.want)
			}
		})
	}
}

func purchaseSuperficialDeepEqual(p1 *models.Purchase, p2 *models.Purchase) bool {
	return p1.Id == p2.Id && p1.Description == p2.Description && p1.Date == p2.Date && p1.Amount == p2.Amount
}
