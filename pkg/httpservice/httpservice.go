package httpservice

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/marcosArruda/purchases-multi-country/pkg/logs"
	"github.com/marcosArruda/purchases-multi-country/pkg/models"
	"github.com/marcosArruda/purchases-multi-country/pkg/services"

	"github.com/gin-gonic/gin"
)

const (
	countrycurrencyKey = "Countrycurrency"
)

type (
	httpServiceFinal struct {
		sm         services.ServiceManager
		router     *gin.Engine
		srv        *http.Server
		regexpRule *regexp.Regexp
		quit       chan os.Signal
	}
)

func NewHttpService() services.HttpService {
	return &httpServiceFinal{regexpRule: regexp.MustCompile(`[^a-zA-Z0-9 ]+`)}
}

func (n *httpServiceFinal) Start(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	n.router = gin.Default()

	n.router.POST("/purchases", n.PostPurchase)
	n.router.GET("/purchases/:id", n.GetPurchaseById)
	n.router.GET("/purchases", n.GetAllPurchases)

	n.srv = &http.Server{
		Addr:    ":8080",
		Handler: n.router,
	}

	n.quit = make(chan os.Signal, 1)
	signal.Notify(n.quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// http interface connection
		if err := n.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			n.sm.LogsService().Error(ctx, fmt.Sprintf("listen error: %s\n", err.Error()))
			n.quit <- syscall.SIGINT
		}
		n.sm.LogsService().Info(ctx, "Http Server Listening!")
	}()

	go func() {
		if ctx.Value(logs.AppEnvKey) == "TESTS" {
			time.Sleep(5 * time.Second)
			n.Close(ctx)
		}
	}()

	<-n.quit
	n.sm.LogsService().Info(ctx, "shuting down server ...")

	ctxT, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := n.srv.Shutdown(ctxT); err != nil {
		n.sm.LogsService().Error(ctx, fmt.Sprintf("Something went wrong executing server shutdown: %s\n", err.Error()))
	}
	<-ctxT.Done()
	n.sm.LogsService().Warn(ctx, "timed out after 2 seconds.")

	n.sm.LogsService().Warn(ctx, "server exiting")

	return nil
}

func (n *httpServiceFinal) Close(ctx context.Context) error {
	n.quit <- syscall.SIGINT
	return nil
}

func (n *httpServiceFinal) Healthy(ctx context.Context) error {
	return nil
}

func (n *httpServiceFinal) WithServiceManager(sm services.ServiceManager) services.HttpService {
	n.sm = sm
	return n
}

func (n *httpServiceFinal) ServiceManager() services.ServiceManager {
	return n.sm
}

func (n *httpServiceFinal) PostPurchase(c *gin.Context) {
	n.sm.LogsService().Info(c.Request.Context(), c.FullPath()+" Call received")
	var body models.Purchase
	err := c.BindJSON(&body)
	if body.Id == "" {
		n.sm.LogsService().Error(c.Request.Context(), fmt.Sprintf("Error scanning the body received: %s", err.Error()))
	}
	if err != nil {
		n.sm.LogsService().Error(c.Request.Context(), fmt.Sprintf("Error scanning the body received: %s", err.Error()))
	}

	n.sm.LogsService().Info(c.Request.Context(), "Delegating to ExchangeService to handle the new transaction")
	err = n.sm.ExchangeService().HandleNewPurchase(c.Request.Context(), &body)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Something went wrong: %s", err.Error())})
		return
	}
	n.sm.LogsService().Info(c.Request.Context(), "persisted the new purchase")
	c.IndentedJSON(http.StatusOK, nil)
}

func (n *httpServiceFinal) GetPurchaseById(c *gin.Context) {
	n.sm.LogsService().Info(c.Request.Context(), c.FullPath()+" Call received")
	id := c.Param("id")
	countrycurrency := c.Request.Header[countrycurrencyKey][0]
	fmt.Println("countrycurrency=" + countrycurrency)
	n.sm.LogsService().Info(c.Request.Context(), "Delegating to ExchangeService to find the purchase")

	p, err := n.sm.ExchangeService().SearchPurchasesById(c.Request.Context(), id, countrycurrency)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Something went wrong: %s", err.Error())})
		return
	}
	n.sm.LogsService().Info(c.Request.Context(), fmt.Sprintf("Got the correct purchase, returning it: %v", p))
	c.IndentedJSON(http.StatusOK, p)
}

func (n *httpServiceFinal) GetAllPurchases(c *gin.Context) {
	n.sm.LogsService().Info(c.Request.Context(), c.FullPath()+" Call received")
	countrycurrency := c.Request.Header[countrycurrencyKey][0]
	fmt.Println("countrycurrency=" + countrycurrency)

	ps, err := n.sm.ExchangeService().GetAllPurchases(c.Request.Context(), countrycurrency)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Error getting all Purchases: %s", err.Error())})
		return
	}

	c.IndentedJSON(http.StatusOK, ps)

}
