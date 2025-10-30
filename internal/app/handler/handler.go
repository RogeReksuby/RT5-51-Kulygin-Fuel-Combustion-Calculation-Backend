package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"repback/internal/app/config"
	"repback/internal/app/repository"
	"repback/internal/app/role"
)

import _ "repback/cmd/fuelProject/docs"

type Handler struct {
	// то есть первое - имя поля структуры, второе - указатель на Repository из пакета repository
	Repository *repository.Repository
	Config     *config.Config
}

func NewHandler(r *repository.Repository) *Handler {
	var conf, _ = config.NewConfig()
	return &Handler{
		Repository: r,
		Config:     conf,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/fuels", h.GetFuels)
	router.GET("/fuel/:id", h.GetFuel)
	router.GET("/combustion/:id", h.GetReqFuels)
	router.POST("/delete-fuel", h.DeleteChat)
	router.POST("/add-to-comb", h.AddFuelToCart)
	router.POST("/remove-comb/:id", h.RemoveRequest)
	api := router.Group("/api")
	{
		// === ПУБЛИЧНЫЕ API МАРШРУТЫ (без авторизации) ===
		api.GET("/fuels", h.GetFuelsAPI)
		api.GET("/fuels/:id", h.GetFuelAPI)
		api.POST("/users/register", h.RegisterUserAPI)
		api.POST("/users/login", h.LoginUserAPI)

		api.Use(h.WithAuthCheckCart(role.Buyer, role.Moderator)).GET("/combustions/cart-icon", h.GetCombCartIconAPI)
		// === ЗАЩИЩЕННЫЕ МАРШРУТЫ (требуют авторизации) ===
		auth := api.Group("")
		auth.Use(h.WithAuthCheck(role.Buyer, role.Moderator))
		{
			// Пользовательские операции
			auth.GET("/users/profile", h.GetUserProfileAPI)
			auth.POST("/users/logout", h.LogoutUserAPI)
			auth.PUT("/users/profile", h.UpdateUserAPI)

			// Работа с корзиной и заявками
			//auth.GET("/combustions/cart-icon", h.GetCombCartIconAPI)
			auth.POST("/fuels/:id/add-to-comb", h.AddFuelToCartAPI)
			auth.GET("/combustions/:id", h.GetCombustionCalculationAPI)
			auth.PUT("/combustions/:id", h.UpdateCombustionMolarVolumeAPI)
			auth.DELETE("/combustions", h.DeleteCombustionCalculationAPI)
			auth.DELETE("/fuel-combustions", h.RemoveFuelFromCombustionAPI)
			auth.PUT("/fuel-combustions", h.UpdateFuelInCombustionAPI)
			auth.GET("/combustions", h.GetCombustionCalculationsAPI)
			auth.PUT("/combustions/:id/form", h.FormCombustionCalculationAPI)
		}

		// === МАРШРУТЫ ДЛЯ МОДЕРАТОРОВ ===
		moderator := api.Group("")
		moderator.Use(h.WithAuthCheck(role.Moderator))
		{
			// Управление топливом
			moderator.POST("/fuels", h.CreateFuelAPI)
			moderator.PUT("/fuels/:id", h.UpdateFuelAPI)
			moderator.DELETE("/fuels/:id", h.DeleteFuelAPI)
			moderator.POST("/fuels/:id/image", h.UploadFuelImageAPI)
			moderator.PUT("/combustions/:id/moderate", h.CompleteOrRejectCombustionAPI)
		}
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
