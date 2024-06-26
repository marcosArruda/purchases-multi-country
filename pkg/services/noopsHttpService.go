package services

import (
	"context"

	"github.com/gin-gonic/gin"
)

type (
	noOpsHttpService struct {
		sm ServiceManager
	}
)

func NewNoOpsHttpService() HttpService {
	return &noOpsHttpService{}
}

func (n *noOpsHttpService) Start(ctx context.Context) error {
	gin.SetMode(gin.DebugMode)
	return nil
}

func (n *noOpsHttpService) Close(ctx context.Context) error {
	return nil
}

func (n *noOpsHttpService) Healthy(ctx context.Context) error {
	return nil
}

func (n *noOpsHttpService) WithServiceManager(sm ServiceManager) HttpService {
	n.sm = sm
	return n
}

func (n *noOpsHttpService) ServiceManager() ServiceManager {
	return n.sm
}

func (n *noOpsHttpService) PostPurchase(c *gin.Context) {}

func (n *noOpsHttpService) GetPurchaseById(c *gin.Context) {}

func (n *noOpsHttpService) GetAllPurchases(c *gin.Context) {}
