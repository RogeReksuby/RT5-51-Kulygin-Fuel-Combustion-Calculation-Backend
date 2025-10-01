package handler

import (
	"fmt"
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

func (h *Handler) GetFuelsAPI(ctx *gin.Context) {
	var fuels []ds.Fuel
	var err error

	searchString := ctx.Query("title")

	if searchString == "" {
		fuels, err = h.Repository.GetFuels()
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	} else {
		fuels, err = h.Repository.GetFuelsByTitle(searchString)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   fuels,
		"count":  len(fuels),
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

func (h *Handler) GetFuelAPI(ctx *gin.Context) {
	idFuelStr := ctx.Param("id")
	idFuel, err := strconv.Atoi(idFuelStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	fuel, err := h.Repository.GetFuel(idFuel)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   fuel,
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

func (h *Handler) AddFuelToCart(ctx *gin.Context) {
	strId := ctx.PostForm("fuel_id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	err = h.Repository.AddFuelToCart(uint(id))
	ctx.Redirect(http.StatusFound, "/fuels")
}

func (h *Handler) AddFuelToCartAPI(ctx *gin.Context) {
	idFuelStr := ctx.Param("id")
	id, err := strconv.Atoi(idFuelStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	err = h.Repository.AddFuelToCart(uint(id))
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Услуга добавлена в заявку",
	})
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

func (h *Handler) CreateFuelAPI(ctx *gin.Context) {
	var fuelInput struct {
		Title     string  `json:"title" binding:"required"`
		Heat      float64 `json:"heat" binding:"required"`
		MolarMass float64 `json:"molar_mass" binding:"required"`
		ShortDesc string  `json:"short_desc,omitempty"`
		FullDesc  string  `json:"full_desc,omitempty"`
		IsGas     bool    `json:"is_gas,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&fuelInput); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	newFuel := ds.Fuel{
		Title:     fuelInput.Title,
		Heat:      fuelInput.Heat,
		MolarMass: fuelInput.MolarMass,
		ShortDesc: fuelInput.ShortDesc,
		FullDesc:  fuelInput.FullDesc,
		IsGas:     fuelInput.IsGas,
	}

	err := h.Repository.CreateFuel(&newFuel)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"data":    newFuel,
		"message": "Топливо успешно создано",
	})
}

func (h *Handler) UpdateFuelAPI(ctx *gin.Context) {

	idFuelStr := ctx.Param("id")
	id, err := strconv.Atoi(idFuelStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var fuelInput struct {
		Title     string  `json:"title,omitempty"`
		Heat      float64 `json:"heat,omitempty"`
		MolarMass float64 `json:"molar_mass,omitempty"`
		CardImage string  `json:"card_image,omitempty"`
		ShortDesc string  `json:"short_desc,omitempty"`
		FullDesc  string  `json:"full_desc,omitempty"`
		IsGas     bool    `json:"is_gas,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&fuelInput); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	updateData := ds.Fuel{
		Title:     fuelInput.Title,
		Heat:      fuelInput.Heat,
		MolarMass: fuelInput.MolarMass,
		CardImage: fuelInput.CardImage,
		ShortDesc: fuelInput.ShortDesc,
		FullDesc:  fuelInput.FullDesc,
		IsGas:     fuelInput.IsGas,
	}

	err = h.Repository.UpdateFuel(uint(id), &updateData)
	if err != nil {
		fmt.Println("grg1")
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	updatedFuel, err := h.Repository.GetFuel(int(id))
	if err != nil {
		fmt.Println("grg2")
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	fmt.Println("grg3")
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    updatedFuel,
		"message": "Топливо успешно обновлено",
	})
}

func (h *Handler) DeleteFuelAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Проверяем существование записи перед удалением
	_, err = h.Repository.GetFuel(int(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Используем ваш существующий метод DeleteFuel (мягкое удаление)
	err = h.Repository.DeleteFuel(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Топливо успешно удалено",
	})
}

// UploadFuelImageAPI - REST API метод для загрузки изображения услуги
func (h *Handler) UploadFuelImageAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Получаем файл из формы
	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("файл изображения обязателен"))
		return
	}

	// Загружаем изображение
	err = h.Repository.UploadFuelImage(uint(id), file)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Получаем обновленные данные услуги
	updatedFuel, err := h.Repository.GetFuel(int(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    updatedFuel,
		"message": "Изображение успешно загружено",
	})
}

func (h *Handler) GetCombCartIconAPI(ctx *gin.Context) {

	requestID := h.Repository.GetRequestID(1)
	cartCount := h.Repository.GetCartCount()

	ctx.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"id_combustion": requestID,
		"items_count":   cartCount,
	})
}

func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	var input struct {
		Login       string `json:"login" binding:"required"`
		Password    string `json:"password" binding:"required"`
		IsModerator bool   `json:"is_moderator,omitempty"`
		Name        string `json:"name,omitempty"`
	}

	// Парсим JSON из тела запроса
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Регистрируем пользователя
	newUser, err := h.Repository.RegisterUser(input.Login, input.Password, input.Name, input.IsModerator)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"data":    newUser,
		"message": "Пользователь успешно зарегистрирован",
	})
}
