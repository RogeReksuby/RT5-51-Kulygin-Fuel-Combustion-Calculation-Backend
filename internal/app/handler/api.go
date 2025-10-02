package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"repback/internal/app/ds"
	"strconv"
)

func (h *Handler) GetCombustionCalculationsAPI(ctx *gin.Context) {
	var filter struct {
		Status    string `form:"status"`
		StartDate string `form:"start_date"`
		EndDate   string `form:"end_date"`
	}

	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	calculations, err := h.Repository.GetCombustionCalculations(filter.Status, filter.StartDate, filter.EndDate)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Преобразуем в response структуру с логинами
	response := make([]ds.CombustionResponse, len(calculations))
	for i, calc := range calculations {
		response[i] = ds.CombustionResponse{
			ID:           calc.ID,
			Status:       calc.Status,
			DateCreate:   calc.DateCreate.Format("02.01.2006"),
			DateUpdate:   calc.DateUpdate.Format("02.01.2006"),
			CreatorLogin: calc.Creator.Login,
			MolarVolume:  calc.MolarVolume,
			FinalResult:  calc.FinalResult,
		}

		// Дата завершения (если есть)
		if calc.DateFinish.Valid {
			response[i].DateFinish = calc.DateFinish.Time.Format("02.01.2006")
		}

		// Логин модератора (если есть)
		if calc.Moderator.ID != 0 {
			response[i].ModeratorLogin = calc.Moderator.Login
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
		"count":  len(response),
	})
}

// GetCombustionCalculationAPI - GET одной заявки с услугами
func (h *Handler) GetCombustionCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Получаем заявку и услуги отдельно
	calculation, fuels, err := h.Repository.GetCombustionCalculationByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	response := struct {
		ID             uint
		Status         string
		DateCreate     string
		DateUpdate     string
		DateFinish     string
		CreatorLogin   string
		ModeratorLogin string
		MolarVolume    float64
		FinalResult    float64
		Fuels          []ds.Fuel
	}{
		ID:           calculation.ID,
		Status:       calculation.Status,
		DateCreate:   calculation.DateCreate.Format("02.01.2006"),
		DateUpdate:   calculation.DateUpdate.Format("02.01.2006"),
		CreatorLogin: calculation.Creator.Login,
		MolarVolume:  calculation.MolarVolume,
		FinalResult:  calculation.FinalResult,
		Fuels:        make([]ds.Fuel, len(fuels)), // используем отдельно загруженные fuels
	}

	// Дата завершения (если есть)
	if calculation.DateFinish.Valid {
		response.DateFinish = calculation.DateFinish.Time.Format("02.01.2006")
	}

	// Логин модератора (если есть)
	if calculation.Moderator.ID != 0 {
		response.ModeratorLogin = calculation.Moderator.Login
	}

	// Заполняем услуги
	for i, fuel := range fuels {
		response.Fuels[i] = ds.Fuel{
			ID:        fuel.ID,
			Title:     fuel.Title,
			Heat:      fuel.Heat,
			MolarMass: fuel.MolarMass,
			CardImage: fuel.CardImage,
			ShortDesc: fuel.ShortDesc,
			FullDesc:  fuel.FullDesc,
			IsGas:     fuel.IsGas,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// UpdateCombustionMolarVolumeAPI - PUT изменение MolarVolume
func (h *Handler) UpdateCombustionMolarVolumeAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var input struct {
		MolarVolume float64 `json:"molar_volume" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Обновляем MolarVolume
	err = h.Repository.UpdateCombustionMolarVolume(uint(id), input.MolarVolume)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err) // 400 т.к. бизнес-логика не прошла
		return
	}

	// Получаем обновленную заявку для ответа
	updatedCalculation, _, err := h.Repository.GetCombustionCalculationByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    updatedCalculation,
		"message": "MolarVolume успешно обновлен",
	})
}

// FormCombustionCalculationAPI - PUT формирование заявки
func (h *Handler) FormCombustionCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Формируем заявку
	err = h.Repository.FormCombustionCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Получаем обновленную заявку для ответа
	updatedCalculation, fuels, err := h.Repository.GetCombustionCalculationByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    updatedCalculation,
		"fuels":   fuels,
		"message": "Заявка успешно сформирована",
	})
}

// CompleteOrRejectCombustionAPI - PUT завершение/отклонение заявки
func (h *Handler) CompleteOrRejectCombustionAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var input struct {
		IsComplete bool `json:"is_complete" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	moderatorID := uint(2)

	err = h.Repository.CompleteOrRejectCombustion(uint(id), moderatorID, input.IsComplete)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	updatedCalculation, fuels, err := h.Repository.GetCombustionCalculationByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	message := "Заявка отклонена"
	if input.IsComplete {
		message = "Заявка завершена"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    updatedCalculation,
		"fuels":   fuels,
		"message": message,
	})
}
