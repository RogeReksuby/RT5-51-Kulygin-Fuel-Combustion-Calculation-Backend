package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"repback/internal/app/ds"
	"strconv"
	"strings"
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
		"fuels":                fuels,
		"cart_count":           h.Repository.GetCartCount(),
		"searchFuelTitleQuery": searchString,
		"reqID":                h.Repository.GetRequestID(uint(1)),
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
	requestIDStr := ctx.Param("id")
	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		logrus.Error(err)
	}

	reqStatus, err := h.Repository.RequestStatusById(requestID)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/fuels")
	}

	// если заявка по которой переходим удалена, то перенаправляем на главную
	if reqStatus == "удалён" {
		ctx.Redirect(http.StatusFound, "/fuels")
	}

	fuels, err = h.Repository.GetReqFuels(uint(requestID))
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "combustion.html", gin.H{
		"fuels": fuels,
		"idReq": requestID,
	})
}

func (h *Handler) DeleteChat(ctx *gin.Context) {
	strId := ctx.PostForm("fuel_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	err = h.Repository.DeleteFuel(uint(id))
	if err != nil && !strings.Contains(err.Error(), "duplicate key value vioalates unique constraint") {
		return
	}
	ctx.Redirect(http.StatusFound, "/fuels")
}

func (h *Handler) AddToCart(ctx *gin.Context) {
	strId := ctx.PostForm("fuel_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	err = h.Repository.AddToCart(uint(id))
	ctx.Redirect(http.StatusFound, "/fuels")
}

func (h *Handler) RemoveRequest(ctx *gin.Context) {
	idReqStr := ctx.Param("id")
	idReq, err := strconv.Atoi(idReqStr)
	if err != nil {
		logrus.Error(err)
	}

	err = h.Repository.RemoveRequest(uint(idReq))
	ctx.Redirect(http.StatusFound, "/fuels")

}
