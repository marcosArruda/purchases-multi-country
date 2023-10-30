package exchangeservice

import (
	"context"
	"reflect"
	"testing"

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

	basicConvertedAmount = &models.ConvertedAmount{
		Id:              basicPurchase.Id,
		Description:     basicPurchase.Description,
		PurchaseDate:    basicExchange.Date,
		OriginalAmount:  basicPurchase.Amount,
		ExchangeRate:    basicExchange.ExchangeRate,
		ConvertedAmount: "100.65",
	}
	basicConvertedAmountHigherNumber = &models.ConvertedAmount{
		Id:              basicPurchase.Id,
		Description:     basicPurchase.Description,
		PurchaseDate:    basicExchange.Date,
		OriginalAmount:  basicPurchase.Amount,
		ExchangeRate:    "11.43",
		ConvertedAmount: "230.09",
	}
	basicConvertedAmounts = []*models.ConvertedAmount{{
		Id:              basicPurchase.Id,
		Description:     basicPurchase.Description,
		PurchaseDate:    basicExchange.Date,
		OriginalAmount:  basicPurchase.Amount,
		ExchangeRate:    basicExchange.ExchangeRate,
		ConvertedAmount: "100.65",
	}}
)

func NewManagerForTests() (services.ServiceManager, context.Context) {
	asyncWorkChannel := make(chan func() error)
	stop := make(chan struct{})
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.AppEnvKey, "TESTS")
	ctx = context.WithValue(ctx, logs.AppNameKey, logs.AppName)
	ctx = context.WithValue(ctx, logs.AppVersionKey, logs.AppVersion)
	return services.NewManager(asyncWorkChannel, stop), ctx
}

func Test_exchangeServiceFinal_Start(t *testing.T) {

	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTests()
	tests := []struct {
		name    string
		n       *exchangeServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success", //here is just the success case is needed because ExchangeService.Start(ctx) does nothing with the ctx yet
			args:    args{ctx: ctx},
			n:       sm.WithExchangeService(NewExchangeService()).ExchangeService().(*exchangeServiceFinal),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Start(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("exchangeServiceFinal.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_exchangeServiceFinal_Close(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	_, ctx := NewManagerForTests()
	tests := []struct {
		name    string
		n       *exchangeServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success", //here is just the success case is needed because ExchangeService.Close(ctx) does nothing with the ctx yet
			args:    args{ctx: ctx},
			n:       NewExchangeService().(*exchangeServiceFinal),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Close(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("exchangeServiceFinal.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_exchangeServiceFinal_Healthy(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	_, ctx := NewManagerForTests()
	tests := []struct {
		name    string
		n       *exchangeServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success", //here is just the success case is needed because ExchangeService.Healthy(ctx) does nothing with the ctx yet
			args:    args{ctx: ctx},
			n:       NewExchangeService().(*exchangeServiceFinal),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Healthy(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("exchangeServiceFinal.Healthy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_exchangeServiceFinal_WithServiceManager(t *testing.T) {
	type args struct {
		sm services.ServiceManager
	}
	sm, _ := NewManagerForTests()
	pf := NewExchangeService()
	tests := []struct {
		name string
		n    *exchangeServiceFinal
		args args
		want services.ExchangeService
	}{
		{
			name: "success",
			args: args{sm: sm},
			n:    pf.(*exchangeServiceFinal),
			want: pf,
		},
		{
			name: "successNil",
			args: args{sm: nil},
			n:    pf.(*exchangeServiceFinal),
			want: pf,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.WithServiceManager(tt.args.sm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("exchangeServiceFinal.WithServiceManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_exchangeServiceFinal_ServiceManager(t *testing.T) {
	sm, _ := NewManagerForTests()
	tests := []struct {
		name string
		n    *exchangeServiceFinal
		want services.ServiceManager
	}{
		{
			name: "success",
			n:    NewExchangeService().WithServiceManager(sm).(*exchangeServiceFinal),
			want: sm,
		},
		{
			name: "successNil",
			n:    NewExchangeService().WithServiceManager(nil).(*exchangeServiceFinal),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.ServiceManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("exchangeServiceFinal.ServiceManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_exchangeServiceFinal_HandleNewPurchase(t *testing.T) {
	type args struct {
		ctx context.Context
		p   *models.Purchase
	}
	sm, ctx := NewManagerForTests()
	pf := NewExchangeService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *exchangeServiceFinal
		args    args
		want    *models.Purchase
		wantErr bool
	}{
		{
			name:    "success",
			args:    args{ctx: ctx, p: basicPurchase},
			n:       pf.(*exchangeServiceFinal),
			wantErr: false,
		},
		{
			name:    "returnNilCallingTreasuryAccessService",
			args:    args{ctx: ctx, p: nil},
			n:       pf.(*exchangeServiceFinal),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.n.HandleNewPurchase(tt.args.ctx, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: exchangeServiceFinal.HandleNewPurchase() error = %v", tt.name, err)
				return
			}
		})
	}
}

func Test_exchangeServiceFinal_SearchPurchasesById(t *testing.T) {
	type args struct {
		ctx             context.Context
		Id              string
		countrycurrency string
	}
	sm, ctx := NewManagerForTests()
	pf := NewExchangeService().WithServiceManager(sm)
	var want *models.ConvertedAmount = basicConvertedAmount
	tests := []struct {
		name    string
		n       *exchangeServiceFinal
		args    args
		want    *models.ConvertedAmount
		wantErr bool
	}{
		{
			name:    "success",
			args:    args{ctx: ctx, Id: basicPurchase.Id, countrycurrency: basicExchange.CountryCurrencyDesc},
			n:       pf.(*exchangeServiceFinal),
			want:    want,
			wantErr: false,
		},
		{
			name:    "anyError",
			args:    args{ctx: ctx, Id: basicPurchase.Id, countrycurrency: "error"},
			n:       pf.(*exchangeServiceFinal),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.SearchPurchasesById(tt.args.ctx, tt.args.Id, tt.args.countrycurrency)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: exchangeServiceFinal.SearchPurchasesById() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: exchangeServiceFinal.SearchPurchasesById() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_exchangeServiceFinal_convertPurchaseByExchangeRate(t *testing.T) {
	type args struct {
		ctx          context.Context
		p            *models.Purchase
		exchangeRate string
	}
	sm, ctx := NewManagerForTests()
	pf := NewExchangeService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *exchangeServiceFinal
		args    args
		want    *models.ConvertedAmount
		wantErr bool
	}{
		{
			name:    "success",
			args:    args{ctx: ctx, exchangeRate: "5.00", p: basicPurchase},
			n:       pf.(*exchangeServiceFinal),
			want:    basicConvertedAmount,
			wantErr: false,
		},
		{
			name:    "successHigherNumber",
			args:    args{ctx: ctx, exchangeRate: "11.43", p: basicPurchase},
			n:       pf.(*exchangeServiceFinal),
			want:    basicConvertedAmountHigherNumber,
			wantErr: false,
		},
		{
			name:    "anyError",
			args:    args{ctx: nil, exchangeRate: "error"},
			n:       pf.(*exchangeServiceFinal),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.convertPurchaseByExchangeRate(tt.args.ctx, tt.args.p, tt.args.exchangeRate)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: exchangeServiceFinal.GetAllPurchases() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: exchangeServiceFinal.GetAllPurchases() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_exchangeServiceFinal_GetAllPurchases(t *testing.T) {
	type args struct {
		ctx             context.Context
		countrycurrency string
	}
	sm, ctx := NewManagerForTests()
	pf := NewExchangeService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *exchangeServiceFinal
		args    args
		want    []*models.ConvertedAmount
		wantErr bool
	}{
		{
			name:    "success",
			args:    args{ctx: ctx, countrycurrency: "Brazil-Real"},
			n:       pf.(*exchangeServiceFinal),
			want:    basicConvertedAmounts,
			wantErr: false,
		},
		{
			name:    "anyError",
			args:    args{ctx: nil, countrycurrency: "error"},
			n:       pf.(*exchangeServiceFinal),
			want:    services.EmptyConvertedPurchasesSlice,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.GetAllPurchases(tt.args.ctx, tt.args.countrycurrency)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: exchangeServiceFinal.GetAllPurchases() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("%s: exchangeServiceFinal.GetAllPurchases() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
