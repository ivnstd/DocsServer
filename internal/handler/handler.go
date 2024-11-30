package handler

import (
	"docs_server/internal/service"

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

	api := router.Group("/api")

	api.POST("/register", h.Register)
	api.POST("/auth", h.Auth)
	api.DELETE("/auth/:token", h.EndSession)

	documents := api.Group("/docs")
	{
		documents.POST("", h.CreateDocument)
		documents.GET("", h.GetDocumentsList)
		documents.HEAD("", h.GetDocumentsList)
		documents.GET("/:id", h.GetDocument)
		documents.HEAD("/:id", h.GetDocument)
		documents.DELETE("/:id", h.DeleteDocument)
	}

	return router
}
