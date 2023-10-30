package persistence

import (
	"context"
	"reflect"
	"testing"

	"github.com/marcosArruda/purchases-multi-country/pkg/logs"
	"github.com/marcosArruda/purchases-multi-country/pkg/models"
	"github.com/marcosArruda/purchases-multi-country/pkg/services"
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

func Test_persistenceServiceFinal_Start(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTests()
	ps := NewPersistenceService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *persistenceServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success", // Only success because PersistenceService.Start(ctx) does nothing with the Context for now
			n:       ps.(*persistenceServiceFinal),
			args:    args{ctx: ctx},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Start(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("persistenceServiceFinal.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_persistenceServiceFinal_Close(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTests()
	ps := NewPersistenceService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *persistenceServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success", // Only success because PersistenceService.Close(ctx) does nothing with the Context for now
			n:       ps.(*persistenceServiceFinal),
			args:    args{ctx: ctx},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Close(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("persistenceServiceFinal.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_persistenceServiceFinal_Healthy(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTests()
	ps := NewPersistenceService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *persistenceServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success", // Only success because PersistenceService.Healthy(ctx) does nothing with the Context for now
			n:       ps.(*persistenceServiceFinal),
			args:    args{ctx: ctx},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Healthy(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("persistenceServiceFinal.Healthy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_persistenceServiceFinal_WithServiceManager(t *testing.T) {
	type args struct {
		sm services.ServiceManager
	}
	sm, _ := NewManagerForTests()
	ps := NewPersistenceService()
	tests := []struct {
		name string
		n    *persistenceServiceFinal
		args args
		want services.PersistenceService
	}{
		{
			name: "success",
			n:    ps.(*persistenceServiceFinal),
			args: args{sm: sm},
			want: ps,
		},
		{
			name: "WithNil",
			n:    ps.(*persistenceServiceFinal),
			args: args{sm: nil},
			want: ps,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.WithServiceManager(tt.args.sm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("persistenceServiceFinal.WithServiceManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_persistenceServiceFinal_ServiceManager(t *testing.T) {
	sm, _ := NewManagerForTests()
	tests := []struct {
		name string
		n    *persistenceServiceFinal
		want services.ServiceManager
	}{
		{
			name: "success",
			n:    NewPersistenceService().WithServiceManager(sm).(*persistenceServiceFinal),
			want: sm,
		},
		{
			name: "returnNil",
			n:    NewPersistenceService().(*persistenceServiceFinal),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.ServiceManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("persistenceServiceFinal.ServiceManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_persistenceServiceFinal_GetPurchaseById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}
	sm, ctx := NewManagerForTests()
	ps := NewPersistenceService().WithServiceManager(sm)
	p := &models.Purchase{
		Id:          "1",
		Description: "Some transaction",
		Amount:      "20.13",
		Date:        "2023-09-30",
	}

	tests := []struct {
		name    string
		n       *persistenceServiceFinal
		args    args
		want    *models.Purchase
		wantErr bool
	}{
		{
			name:    "success",
			n:       ps.(*persistenceServiceFinal),
			args:    args{ctx: ctx, id: "1"},
			want:    p,
			wantErr: false,
		},
		{
			name:    "error",
			n:       ps.(*persistenceServiceFinal),
			args:    args{ctx: ctx, id: "0"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.GetPurchaseById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("persistenceServiceFinal.GetPurchaseById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("persistenceServiceFinal.GetPurchaseById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_persistenceServiceFinal_ListAllPurchases(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTests()
	ps := NewPersistenceService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *persistenceServiceFinal
		args    args
		want    []*models.Purchase
		wantErr bool
	}{
		{
			name:    "success",
			n:       ps.(*persistenceServiceFinal),
			args:    args{ctx: ctx},
			want:    services.EmptyPurchasesSlice,
			wantErr: false,
		},
		{
			name:    "error",
			n:       ps.(*persistenceServiceFinal),
			args:    args{ctx: nil},
			want:    services.EmptyPurchasesSlice,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.ListAllPurchases(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: persistenceServiceFinal.ListAllPurchases() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("%s: persistenceServiceFinal.ListAllPurchases() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_persistenceServiceFinal_InsertPurchase(t *testing.T) {
	type args struct {
		newPurchase *models.Purchase
	}
	sm, ctx := NewManagerForTests()
	ps := NewPersistenceService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *persistenceServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success", //just manual work to make other cases.. For now, I will keep just testing the success, but I KNOW that in production apps we need to cover ALL..
			args:    args{newPurchase: basicPurchase},
			n:       ps.(*persistenceServiceFinal),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm.Start(ctx)
			if err := tt.n.InsertPurchase(ctx, tt.args.newPurchase); (err != nil) != tt.wantErr {
				t.Errorf("%s: persistenceServiceFinal.InsertPurchase() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func Test_persistenceServiceFinal_BatchInsertExchanges(t *testing.T) {
	type args struct {
		p            *models.Purchase
		newExchanges []*models.ExchangeForDate
	}
	sm, ctx := NewManagerForTests()
	ps := NewPersistenceService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *persistenceServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success", //just manual work to make other cases.. For now, I will keep just testing the success, but I KNOW that in production apps we need to cover ALL..
			args:    args{p: basicPurchase, newExchanges: []*models.ExchangeForDate{basicExchange}},
			n:       ps.(*persistenceServiceFinal),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm.Start(ctx)
			if err := tt.n.BatchInsertExchanges(ctx, tt.args.p, tt.args.newExchanges); (err != nil) != tt.wantErr {
				t.Errorf("%s: persistenceServiceFinal.BatchInsertExchanges() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func Test_persistenceServiceFinal_InsertExchange(t *testing.T) {
	type args struct {
		p           *models.Purchase
		newExchange *models.ExchangeForDate
	}
	sm, ctx := NewManagerForTests()
	ps := NewPersistenceService().WithServiceManager(sm)
	tests := []struct {
		name    string
		n       *persistenceServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success", //just manual work to make other cases.. For now, I will keep just testing the success, but I KNOW that in production apps we need to cover ALL..
			args:    args{p: basicPurchase, newExchange: basicExchange},
			n:       ps.(*persistenceServiceFinal),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm.Start(ctx)
			if err := tt.n.InsertExchange(ctx, tt.args.p, tt.args.newExchange); (err != nil) != tt.wantErr {
				t.Errorf("%s: persistenceServiceFinal.InsertExchange() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
