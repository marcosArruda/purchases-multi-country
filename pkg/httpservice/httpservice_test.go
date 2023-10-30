package httpservice

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/marcosArruda/purchases-multi-country/pkg/logs"
	"github.com/marcosArruda/purchases-multi-country/pkg/services"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func NopCloser(r io.Reader) io.ReadCloser {
	return nopCloser{r}
}

func SetUpRouter() *gin.Engine {
	router := gin.Default()
	return router
}

func NewManagerForTests() (services.ServiceManager, context.Context) {
	asyncWorkChannel := make(chan func() error)
	stop := make(chan struct{})
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.AppEnvKey, "TESTS")
	ctx = context.WithValue(ctx, logs.AppNameKey, logs.AppName)
	ctx = context.WithValue(ctx, logs.AppVersionKey, logs.AppVersion)
	return services.NewManager(asyncWorkChannel, stop), ctx
}

func NewGinContextForTests(reqPath string, withError bool) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	u := &url.URL{
		Path: reqPath,
	}

	header := make(http.Header)
	if withError {
		header.Add("Countrycurrency", "error")
	} else {
		header.Add("Countrycurrency", "Brazil-Real")
	}
	v := url.Values{}
	ctx.Request = &http.Request{
		Header: header,
		URL:    u,
	}

	ctx.Request.URL.RawQuery = v.Encode()
	return ctx
}

func NewGinContextForTestsPOST(reqPath string, withError bool) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	u := &url.URL{
		Path: reqPath,
	}

	header := make(http.Header)
	if withError {
		header.Add("Countrycurrency", "error")
	} else {
		header.Add("Countrycurrency", "Brazil-Real")
	}

	v := url.Values{}
	ctx.Request = &http.Request{
		Header: header,
		URL:    u,
		Body:   io.NopCloser(strings.NewReader("{'id': 'abcd-fghi', 'description': 'Some transaction', 'amount': '20.13', date: '2023-09-30'}")),
	}

	ctx.Request.URL.RawQuery = v.Encode()
	return ctx
}

func Test_httpServiceFinal_Start(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTests()
	httpService := sm.WithHttpService(NewHttpService()).HttpService()
	tests := []struct {
		name    string
		n       *httpServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			n:       httpService.(*httpServiceFinal),
			args:    args{ctx},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Start(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("httpServiceFinal.Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_httpServiceFinal_Close(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	sm, ctx := NewManagerForTests()
	httpService := sm.WithHttpService(NewHttpService()).HttpService()
	httpService.Start(ctx)
	tests := []struct {
		name    string
		n       *httpServiceFinal
		args    args
		wantErr bool
	}{
		{
			name:    "success",
			n:       httpService.(*httpServiceFinal),
			args:    args{ctx},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.n.Close(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("httpServiceFinal.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_httpServiceFinal_GetPurchaseById(t *testing.T) {
	type args struct {
		c *gin.Context
	}

	sm, _ := NewManagerForTests()
	httpService := sm.WithHttpService(NewHttpService()).HttpService()
	ginCtx := NewGinContextForTests("/some-request-path/1/", false)
	tests := []struct {
		name string
		n    *httpServiceFinal
		args args
	}{
		{
			name: "success",
			n:    httpService.(*httpServiceFinal),
			args: args{ginCtx},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.GetPurchaseById(tt.args.c)
		})
	}
}

func Test_httpServiceFinal_GetAllPurchases(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	sm, _ := NewManagerForTests()
	httpService := sm.WithHttpService(NewHttpService()).HttpService()
	ginCtx := NewGinContextForTests("/some-request-path/1/", false)
	tests := []struct {
		name string
		n    *httpServiceFinal
		args args
	}{
		{
			name: "success",
			n:    httpService.(*httpServiceFinal),
			args: args{ginCtx},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.GetAllPurchases(tt.args.c)
		})
	}
}

func Test_httpServiceFinal_PostPurchase(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	sm, _ := NewManagerForTests()
	httpService := sm.WithHttpService(NewHttpService()).HttpService()
	ginCtx := NewGinContextForTestsPOST("/some-request-path/1/", false)
	tests := []struct {
		name string
		n    *httpServiceFinal
		args args
	}{
		{
			name: "success",
			n:    httpService.(*httpServiceFinal),
			args: args{ginCtx},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.n.PostPurchase(tt.args.c)
		})
	}
}
