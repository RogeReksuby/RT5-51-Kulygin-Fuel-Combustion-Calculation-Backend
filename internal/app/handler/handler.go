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
	router.GET("/req", h.GetReqFuels)
	router.POST("/delete-fuel", h.DeleteChat)
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
