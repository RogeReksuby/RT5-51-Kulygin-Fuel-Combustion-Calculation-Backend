package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"repback/internal/app/repository"
	"strconv"
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

func (h *Handler) GetFuels(ctx *gin.Context) {
	var fuels []repository.Fuel
	var err error
	searchFuelString := ctx.Query("searchFuelQuery")

	if searchFuelString == "" {
		fuels, err = h.Repository.GetFuels()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		fuels, err = h.Repository.GetFuelByTitle(searchFuelString)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"fuels":           fuels,
		"searchFuelQuery": searchFuelString,
		"comb_id":         1,
		"count_of_fuels":  len(h.Repository.GetReqArrayOfID()),
	})
}

func (h *Handler) GetReqFuels(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusSeeOther, "/fuels")
	}
	if id != 1 {
		ctx.Redirect(http.StatusSeeOther, "/fuels")
	}
	var fuels []repository.Fuel

	fuels, err = h.Repository.GetReqFuels()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "combustion.html", gin.H{
		"fuels": fuels,
	})
}

func (h *Handler) GetFuel(ctx *gin.Context) {
	idFuelStr := ctx.Param("id")
	idFuel, err := strconv.Atoi(idFuelStr)
	if err != nil {
		logrus.Error(err)
	}

	fuel, err := h.Repository.GetFuel(idFuel)
	if err != nil {
		logrus.Error(err)
	}
	ctx.HTML(http.StatusOK, "fuel.html", gin.H{
		"fuel": fuel,
	})
}
