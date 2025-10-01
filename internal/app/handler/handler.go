package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"repback/internal/app/repository"
)

type Handler struct {
	// то есть первое - имя поля структуры, второе - указатель на Repository из пакета repository
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/fuels", h.GetFuels)
	router.GET("/fuel/:id", h.GetFuel)
	router.GET("/combustion/:id", h.GetReqFuels)
	router.POST("/delete-fuel", h.DeleteChat)
	router.POST("/add-to-comb", h.AddFuelToCart)
	router.POST("/remove-comb/:id", h.RemoveRequest)
	api := router.Group("/api")
	{
		api.GET("/fuels", h.GetFuelsAPI)
		api.GET("/fuels/:id", h.GetFuelAPI)
		api.POST("/fuels", h.CreateFuelAPI)
		api.PUT("/fuels/:id", h.UpdateFuelAPI)
		api.DELETE("/fuels/:id", h.DeleteFuelAPI)
		api.POST("/fuels/:id/image", h.UploadFuelImageAPI)
		api.POST("/fuels/:id/add-to-comb", h.AddFuelToCartAPI)

		api.GET("combustions/cart-icon", h.GetCombCartIconAPI)

		api.POST("users/register", h.RegisterUserAPI)
	}
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

//
