package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"repback/internal/app/ds"
	"strconv"
)

func (h *Handler) GetFuels(ctx *gin.Context) {
	var fuels []ds.Fuel
	var err error
	searchString := ctx.Query("searchQuery")

	if searchString == "" {
		fuels, err = h.Repository.GetFuels()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		fuels, err = h.Repository.GetFuelsByTitle(searchString)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"fuels":       fuels,
		"searchQuery": searchString,
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

func (h *Handler) GetReqFuels(ctx *gin.Context) {
	var fuels []ds.Fuel
	var err error
	fuels, err = h.Repository.GetReqFuels()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "req.html", gin.H{
		"fuels": fuels,
	})
}
