package handlers

import (
	_ "github.com/BabyJhon/medods-test-task/docs" // путь к docs
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/BabyJhon/medods-test-task/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.GET("/auth", h.auth)
	router.GET("/user", h.user)
	router.POST("/revoke", h.revoke)
	router.POST("/refresh", h.refresh)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
