package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"net/http"
	"repback/internal/app/ds"
	"repback/internal/app/role"
	"strconv"
	"strings"
	"time"
)

// GetCombustionCalculationsAPI godoc
// @Summary Получить список расчетов горения
// @Description Возвращает список расчетов горения с возможностью фильтрации. Обычные пользователи видят только свои расчеты, модераторы видят все расчеты в системе.
// @Tags Combustions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param status query string false "Фильтр по статусу расчета" Enums(черновик, сформирован, завершён, отклонён, удалён)
// @Param start_date query string false "Фильтр по дате создания (начало периода в формате DD.MM.YYYY)"
// @Param end_date query string false "Фильтр по дате создания (конец периода в формате DD.MM.YYYY)"
// @Success 200 {object} map[string][]ds.CombustionResponse "Успешный ответ со списком расчетов"
// @Failure 400 {object} map[string]string "Неверные параметры запроса"
// @Failure 403 {object} map[string]string "Доступ запрещен или пользователь не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/combustions [get]
func (h *Handler) GetCombustionCalculationsAPI(ctx *gin.Context) {
	userId, err := h.GetUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}
	userIsModerator, err := h.GetUserIsModeratorFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}

	var filter struct {
		Status    string `form:"status"`
		StartDate string `form:"start_date"`
		EndDate   string `form:"end_date"`
	}

	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	calculations, err := h.Repository.GetCombustionCalculations(userId, userIsModerator, filter.Status, filter.StartDate, filter.EndDate)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

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

		// дата завершения если есть
		if calc.DateFinish.Valid {
			response[i].DateFinish = calc.DateFinish.Time.Format("02.01.2006")
		}

		// логин модератора если есть
		if calc.Moderator.ID != 0 {
			response[i].ModeratorLogin = calc.Moderator.Login
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// GetCombustionCalculationAPI godoc
// @Summary Получить детали расчета горения по ID
// @Description Возвращает детальную информацию о конкретном расчете горения включая список используемого топлива
// @Tags Combustions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID расчета горения"
// @Success 200 {object} map[string]interface{} "Успешный ответ с данными расчета"
// @Failure 400 {object} map[string]string "Неверный ID расчета"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Расчет не найден"
// @Router /api/combustions/{id} [get]
func (h *Handler) GetCombustionCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

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
		Fuels:        make([]ds.Fuel, len(fuels)),
	}

	if calculation.DateFinish.Valid {
		response.DateFinish = calculation.DateFinish.Time.Format("02.01.2006")
	}

	if calculation.Moderator.ID != 0 {
		response.ModeratorLogin = calculation.Moderator.Login
	}

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
		"data": response,
	})
}

// UpdateCombustionMolarVolumeAPI godoc
// @Summary Обновить молярный объем расчета горения
// @Description Обновляет значение молярного объема для расчета горения. Доступно только для расчетов в статусе "черновик".
// @Tags Combustions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID расчета горения" example(1)
// @Param input body map[string]float64 true "Данные для обновления" example({"molar_volume": 22.414})
// @Success 200 {object} map[string]interface{} "Молярный объем успешно обновлен" example({"data":{"ID":1,"Status":"черновик","MolarVolume":22.414},"message":"MolarVolume успешно обновлен"})
// @Failure 400 {object} map[string]string "Неверные данные запроса" example({"error":"molar_volume: cannot be empty"})
// @Failure 403 {object} map[string]string "Доступ запрещен" example({"error":"Требуется авторизация"})
// @Failure 404 {object} map[string]string "Расчет не найден" example({"error":"Расчет с ID 1 не найден"})
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера" example({"error":"Ошибка при обновлении"})
// @Router /api/combustions/{id} [put]
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

	err = h.Repository.UpdateCombustionMolarVolume(uint(id), input.MolarVolume)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err) // 400 т.к. бизнес-логика не прошла
		return
	}

	updatedCalculation, _, err := h.Repository.GetCombustionCalculationByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    updatedCalculation,
		"message": "MolarVolume успешно обновлен",
	})
}

// FormCombustionCalculationAPI godoc
// @Summary Сформировать расчет горения
// @Description Переводит расчет горения из статуса "черновик" в статус "сформирован". Расчет должен содержать хотя бы одно топливо и заполненный молярный объем.
// @Tags Combustions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID расчета горения" example(1)
// @Success 200 {object} map[string]interface{} "Расчет успешно сформирован" example({"data":{"ID":1,"Status":"сформирован"},"fuels":[{"id":1,"title":"Природный газ"}],"message":"Заявка успешно сформирована"})
// @Failure 400 {object} map[string]string "Ошибка формирования" example({"error":"Добавьте хотя бы одну услугу для формирования заявки"})
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Расчет не найден"
// @Router /api/combustions/{id}/form [put]
func (h *Handler) FormCombustionCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.FormCombustionCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	updatedCalculation, fuels, err := h.Repository.GetCombustionCalculationByID(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    updatedCalculation,
		"fuels":   fuels,
		"message": "Заявка успешно сформирована",
	})
}

// CompleteOrRejectCombustionAPI godoc
// @Summary Завершить или отклонить расчёт горения
// @Description Модератор может завершить (одобрить) или отклонить расчёт горения. Требуется роль модератора.
// @Tags Combustions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID расчёта горения" example(1)
// @Param request body object{is_complete=bool} true "Решение модератора" example({"is_complete":true})
// @Success 200 {object} map[string]interface{} "Операция выполнена успешно" example({"data":{"ID":1,"Status":"завершён"},"fuels":[{"id":1,"title":"Природный газ"}],"message":"Заявка завершена"})
// @Failure 400 {object} map[string]string "Ошибка валидации или бизнес-логики" example({"error":"Невозможно изменить статус расчёта"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Требуется роль модератора"})
// @Failure 404 {object} map[string]string "Расчёт не найден" example({"error":"Расчёт горения с указанным ID не найден"})
// @Router /api/combustions/{id}/moderate [put]
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

	moderatorID, err := h.GetUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

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
		"data":    updatedCalculation,
		"fuels":   fuels,
		"message": message,
	})
}

// RemoveFuelFromCombustionAPI godoc
// @Summary Удалить топливо из заявки на расчёт горения
// @Description Удаляет указанное топливо из текущей активной заявки пользователя. Топливо задаётся через query-параметр fuel_id.
// @Tags Fuel-combustions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param fuel_id query int true "ID топлива для удаления" example(5)
// @Success 200 {object} map[string]interface{} "Топливо успешно удалено" example({"message":"Услуга удалена из заявки"})
// @Failure 400 {object} map[string]string "Ошибка запроса" example({"error":"Некорректный ID топлива"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Недостаточно прав для выполнения операции"})
// @Failure 404 {object} map[string]string "Топливо или заявка не найдены" example({"error":"Топливо не найдено в вашей заявке"})
// @Router /api/fuel-combustions [delete]
func (h *Handler) RemoveFuelFromCombustionAPI(ctx *gin.Context) {

	userId, err := h.GetUserIDFromContext(ctx)

	calculationID := h.Repository.GetRequestID(userId)

	fuelIDStr := ctx.Query("fuel_id")
	fuelID, err := strconv.Atoi(fuelIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.RemoveFuelFromCombustion(uint(calculationID), uint(fuelID))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Услуга удалена из заявки",
	})
}

// UpdateFuelInCombustionAPI godoc
// @Summary Обновить объём топлива в заявке
// @Description Обновляет объём указанного топлива в текущей активной заявке пользователя.
// @Tags Fuel-combustions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{fuel_id=int,fuel_volume=number} true "Данные для обновления топлива" example({"fuel_id":3,"fuel_volume":150.5})
// @Success 200 {object} map[string]interface{} "Топливо успешно обновлено" example({"message":"Данные услуги в заявке обновлены"})
// @Failure 400 {object} map[string]string "Ошибка валидации или бизнес-логики" example({"error":"Некорректный объём топлива или топливо не найдено в заявке"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 404 {object} map[string]string "Топливо не найдено" example({"error":"Указанное топливо отсутствует в вашей заявке"})
// @Router /api/fuel-combustions [put]
func (h *Handler) UpdateFuelInCombustionAPI(ctx *gin.Context) {
	userID, err := h.GetUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
	}
	calculationID := h.Repository.GetRequestID(userID)

	var input struct {
		FuelID     uint    `json:"fuel_id" binding:"required"`
		FuelVolume float64 `json:"fuel_volume" binding:"required"`
	}

	if err = ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.UpdateFuelInCombustion(uint(calculationID), input.FuelID, input.FuelVolume)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Данные услуги в заявке обновлены",
	})
}

// DeleteCombustionCalculationAPI godoc
// @Summary Удалить текущую заявку на расчёт горения
// @Description Удаляет активную (черновик) заявку на расчёт горения текущего пользователя. Заявка определяется автоматически по ID пользователя.
// @Tags Combustions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Заявка успешно удалена" example({"message":"Заявка успешно удалена"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Нет прав на удаление заявки"})
// @Failure 404 {object} map[string]string "Заявка не найдена" example({"error":"Активная заявка не найдена"})
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера" example({"error":"Не удалось удалить заявку"})
// @Router /api/combustions [delete]
func (h *Handler) DeleteCombustionCalculationAPI(ctx *gin.Context) {
	userID, err := h.GetUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}

	id := h.Repository.GetRequestID(userID)

	err = h.Repository.DeleteCombustionCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Заявка успешно удалена",
	})
}

// UpdateUserAPI godoc
// @Summary Частично обновить профиль пользователя
// @Description Обновляет указанные поля профиля текущего пользователя. Все поля опциональны — можно отправить только те, что нужно изменить.
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{login=string,name=string,is_moderator=bool} false "Поля для обновления (любое подмножество)" example({"name":"Иванов Иван","login":"ivanov"})
// @Success 200 {object} map[string]interface{} "Профиль успешно обновлён" example({"data":{"ID":5,"login":"ivanov","name":"Иванов Иван","is_moderator":false},"message":"Данные обновлены"})
// @Failure 400 {object} map[string]string "Ошибка валидации или бизнес-логики" example({"error":"нет полей для обновления"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Нет прав на изменение профиля"})
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера" example({"error":"Не удалось обновить данные"})
// @Router /api/users/profile [put]
func (h *Handler) UpdateUserAPI(ctx *gin.Context) {
	userID, err := h.GetUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}
	var input struct {
		Login       *string `json:"login,omitempty"`
		Name        *string `json:"name,omitempty"`
		IsModerator *bool   `json:"is_moderator,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	updates := make(map[string]interface{})

	if input.Login != nil {
		updates["login"] = *input.Login
	}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.IsModerator != nil {
		updates["is_moderator"] = *input.IsModerator
	}

	if len(updates) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("нет полей для обновления"))
		return
	}

	user, err := h.Repository.UpdateUser(userID, updates)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    user,
		"message": "Данные обновлены",
	})
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required" example:"user123"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	User        *ds.Users `json:"user"`
}

// LoginUserAPI godoc
// @Summary Аутентификация пользователя
// @Description Выполняет вход пользователя в систему и возвращает JWT токен для доступа к защищенным endpoint'ам
// @Tags Users
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 {object} LoginResponse "Успешная аутентификация"
// @Failure 400 {object} object{status=string,description=string} "Bad Request"
// @Failure 403 {object} object{status=string,description=string} "Forbidden"
// @Failure 500 {object} object{status=string,description=string} "Внутренняя ошибка сервера"
// @Router /api/users/login [post]
func (h *Handler) LoginUserAPI(ctx *gin.Context) {
	var input struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Аутентифицируем пользователя
	user, err := h.Repository.AuthenticateUser(input.Login, input.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}
	userRole := role.FromUser(user.ID, user.IsModerator)

	jwtConfig := h.Config.GetJWTConfig()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		UserID:      user.ID,
		Login:       user.Login,
		IsModerator: user.IsModerator,
		Name:        user.Name,
		Role:        userRole,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(jwtConfig.ExpiresIn).Unix(), // используем из конфига
			IssuedAt:  time.Now().Unix(),
			Issuer:    jwtConfig.Issuer, // используем из конфига
		},
	})
	tokenString, err := token.SignedString([]byte(jwtConfig.SecretKey))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка создания токена: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, ds.LoginResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int64(jwtConfig.ExpiresIn.Seconds()), // используем из конфига
		User:        user,
	})
}

// LogoutUserAPI godoc
// @Summary Выход пользователя
// @Description Добавляет JWT токен в черный список
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string "Успешный выход"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /api/users/logout [post]
func (h *Handler) LogoutUserAPI(ctx *gin.Context) {
	tokenString := ctx.GetHeader("Authorization")
	if tokenString == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Выход выполнен",
		})
		return
	}

	if !strings.HasPrefix(tokenString, jwtPrefix) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Выход выполнен",
		})
		return
	}

	// Отрезаем префикс
	tokenString = tokenString[len(jwtPrefix):]

	// Парсим токен чтобы получить expiration time
	claims := &ds.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWTSecretKey), nil
	})

	if err != nil || !token.Valid {
		// Если токен невалиден, все равно считаем выход успешным
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Выход выполнен",
		})
		return
	}

	// Добавляем токен в черный список
	if h.Repository.RedisClient != nil {
		// Время жизни в черном списке = оставшееся время жизни токена
		remainingTTL := time.Unix(claims.ExpiresAt, 0).Sub(time.Now())
		if remainingTTL > 0 {
			err = h.Repository.RedisClient.WriteJWTToBlacklist(ctx.Request.Context(), tokenString, remainingTTL)
			if err != nil {
				// Логируем ошибку, но все равно возвращаем успех
				logrus.Errorf("Ошибка добавления токена в черный список: %v", err)
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Выход выполнен успешно",
	})
}

// UploadFuelImageAPI godoc
// @Summary Загрузить изображение для топлива
// @Description Загружает изображение для указанного топлива. Требуется роль модератора. Файл должен быть передан в поле 'image' формы.
// @Tags Fuels
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID топлива" example(3)
// @Param image formData file true "Файл изображения (JPEG, PNG)"
// @Success 200 {object} map[string]interface{} "Изображение успешно загружено" example({"data":{"ID":3,"title":"Природный газ","image_url":"/uploads/fuel_3.jpg"},"message":"Изображение успешно загружено"})
// @Failure 400 {object} map[string]string "Ошибка запроса" example({"error":"файл изображения обязателен"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Требуется роль модератора"})
// @Failure 404 {object} map[string]string "Топливо не найдено" example({"error":"Топливо с указанным ID не существует"})
// @Failure 500 {object} map[string]string "Ошибка сервера" example({"error":"Не удалось сохранить изображение"})
// @Router /api/fuels/{id}/image [post]
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
		"data":    updatedFuel,
		"message": "Изображение успешно загружено",
	})
}

// GetCombCartIconAPI godoc
// @Summary Получить данные для иконки корзины расчёта горения
// @Description Возвращает ID текущей активной заявки и количество добавленных топлив (услуг) для авторизованного пользователя.
// @Tags Combustions
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Данные корзины" example({"id_combustion":12,"items_count":3})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Нет прав на просмотр данных"})
// @Router /api/combustions/cart-icon [get]
func (h *Handler) GetCombCartIconAPI(ctx *gin.Context) {

	userID, err := h.GetUserIDFromContext(ctx)
	if err != nil || userID == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"id_combustion": 0,
			"items_count":   0,
		})
		return
	}
	requestID := h.Repository.GetRequestID(userID)
	cartCount := h.Repository.GetCartCount(userID)

	ctx.JSON(http.StatusOK, gin.H{
		"id_combustion": requestID,
		"items_count":   cartCount,
	})
}

// RegisterUserAPI godoc
// @Summary Регистрация нового пользователя
// @Description Создаёт нового пользователя и возвращает JWT-токен для авторизации. Поле is_moderator игнорируется — только администратор может назначать модераторов.
// @Tags Users
// @Accept json
// @Produce json
// @Param request body object{login=string,password=string,name=string,is_moderator=bool} true "Данные для регистрации" example({"login":"ivanov","password":"secret123","name":"Иванов Иван","is_moderator":false})
// @Success 201 {object} ds.LoginResponse "Пользователь успешно зарегистрирован"
// @Failure 400 {object} map[string]string "Ошибка валидации или логики" example({"error":"Пользователь с таким логином уже существует"})
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера" example({"error":"Не удалось создать токен"})
// @Router /api/users/register [post]
func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	var input struct {
		Login       string `json:"login" binding:"required"`
		Password    string `json:"password" binding:"required"`
		IsModerator bool   `json:"is_moderator,omitempty"`
		Name        string `json:"name,omitempty"`
	}

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

	// Генерируем JWT токен (как в LoginUserAPI)
	userRole := role.FromUser(newUser.ID, newUser.IsModerator)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		UserID:      newUser.ID,
		Login:       newUser.Login,
		IsModerator: newUser.IsModerator,
		Name:        newUser.Name,
		Role:        userRole,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(h.Config.JWTExpiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    h.Config.JWTIssuer,
		},
	})

	tokenString, err := token.SignedString([]byte(h.Config.JWTSecretKey))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка создания токена: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, ds.LoginResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.Config.JWTExpiresIn.Seconds()),
		User:        newUser,
	})
}

// GetUserProfileAPI godoc
// @Summary Получить профиль текущего пользователя
// @Description Возвращает данные профиля авторизованного пользователя (без пароля и других служебных полей).
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Профиль успешно получен" example({"data":{"ID":5,"login":"ivanov","name":"Иванов Иван","is_moderator":false}})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Неверный или просроченный токен"})
// @Failure 404 {object} map[string]string "Пользователь не найден" example({"error":"Пользователь не существует"})
// @Router /api/users/profile [get]
func (h *Handler) GetUserProfileAPI(ctx *gin.Context) {
	// В реальном приложении ID пользователя берется из JWT токена или сессии
	// Для лабораторной работы используем фиксированного пользователя
	userID, err := h.GetUserIDFromContext(ctx) // Фиксированный ID пользователя
	if err != nil {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}

	user, err := h.Repository.GetUserProfile(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// DeleteFuelAPI godoc
// @Summary Удалить топливо
// @Description Удаляет запись топлива по указанному ID. Требуется роль модератора. Выполняется мягкое удаление (запись помечается как удалённая).
// @Tags Fuels
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID топлива" example(5)
// @Success 200 {object} map[string]interface{} "Топливо успешно удалено" example({"message":"Топливо успешно удалено"})
// @Failure 400 {object} map[string]string "Некорректный ID" example({"error":"invalid syntax"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Требуется роль модератора"})
// @Failure 404 {object} map[string]string "Топливо не найдено" example({"error":"Топливо с указанным ID не существует"})
// @Failure 500 {object} map[string]string "Ошибка сервера" example({"error":"Не удалось выполнить удаление"})
// @Router /api/fuels/{id} [delete]
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
		"message": "Топливо успешно удалено",
	})
}

// UpdateFuelAPI godoc
// @Summary Обновить данные топлива
// @Description Частично обновляет информацию о топливе по указанному ID. Все поля опциональны — можно отправить только те, что нужно изменить. Требуется роль модератора.
// @Tags Fuels
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID топлива" example(3)
// @Param request body object{title=string,heat=number,molar_mass=number,card_image=string,short_desc=string,full_desc=string,is_gas=bool} false "Поля для обновления (любое подмножество)" example({"title":"Природный газ (обновлённый)","heat":35.8,"is_gas":true})
// @Success 200 {object} map[string]interface{} "Топливо успешно обновлено" example({"data":{"ID":3,"title":"Природный газ (обновлённый)","heat":35.8,"molar_mass":16.04,"card_image":"/uploads/gas.jpg","short_desc":"Чистое топливо","full_desc":"Используется в промышленности и быту","is_gas":true},"message":"Топливо успешно обновлено"})
// @Failure 400 {object} map[string]string "Ошибка валидации" example({"error":"неверный формат JSON"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Требуется роль модератора"})
// @Failure 404 {object} map[string]string "Топливо не найдено" example({"error":"Топливо с указанным ID не существует"})
// @Failure 500 {object} map[string]string "Ошибка сервера" example({"error":"Не удалось обновить запись"})
// @Router /api/fuels/{id} [put]
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
		"data":    updatedFuel,
		"message": "Топливо успешно обновлено",
	})
}

// CreateFuelAPI godoc
// @Summary Создать новое топливо
// @Description Создаёт новую запись топлива. Обязательные поля: title, heat. Остальные — опциональны. Требуется роль модератора.
// @Tags Fuels
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body object{title=string,heat=number,molar_mass=number,density=number,short_desc=string,full_desc=string,is_gas=bool} true "Данные нового топлива" example({"title":"Природный газ","heat":35.8,"molar_mass":16.04,"density":0.8,"short_desc":"Чистое топливо","full_desc":"Широко используется в быту и промышленности","is_gas":true})
// @Success 201 {object} map[string]interface{} "Топливо успешно создано" example({"data":{"ID":10,"title":"Природный газ","heat":35.8,"molar_mass":16.04,"density":0.8,"short_desc":"Чистое топливо","full_desc":"Широко используется в быту и промышленности","is_gas":true},"message":"Топливо успешно создано"})
// @Failure 400 {object} map[string]string "Ошибка валидации" example({"error":"Key: 'title' Error:Field validation for 'title' failed on the 'required' tag"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Требуется роль модератора"})
// @Failure 500 {object} map[string]string "Ошибка сервера" example({"error":"Не удалось сохранить топливо в базу"})
// @Router /api/fuels [post]
func (h *Handler) CreateFuelAPI(ctx *gin.Context) {
	var fuelInput struct {
		Title     string  `json:"title" binding:"required"`
		Heat      float64 `json:"heat" binding:"required"`
		MolarMass float64 `json:"molar_mass,omitempty"`
		Density   float64 `json:"density,omitempty"`
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
		Density:   fuelInput.Density,
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
		"data":    newFuel,
		"message": "Топливо успешно создано",
	})
}

// AddFuelToCartAPI godoc
// @Summary Добавить топливо в заявку на расчёт горения
// @Description Добавляет указанное топливо в текущую активную заявку (корзину) авторизованного пользователя.
// @Tags Fuels
// @Security BearerAuth
// @Produce json
// @Param id path int true "ID топлива" example(7)
// @Success 200 {object} map[string]interface{} "Топливо успешно добавлено" example({"message":"Услуга добавлена в заявку"})
// @Failure 400 {object} map[string]string "Ошибка логики" example({"error":"Топливо уже добавлено в заявку"})
// @Failure 401 {object} map[string]string "Неавторизован" example({"error":"Требуется авторизация"})
// @Failure 403 {object} map[string]string "Доступ запрещён" example({"error":"Нет прав на изменение заявки"})
// @Failure 404 {object} map[string]string "Топливо не найдено" example({"error":"Топливо с указанным ID не существует"})
// @Router /api/fuels/{id}/add-to-comb [post]
func (h *Handler) AddFuelToCartAPI(ctx *gin.Context) {
	userID, err := h.GetUserIDFromContext(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusForbidden, err)
		return
	}

	idFuelStr := ctx.Param("id")
	id, err := strconv.Atoi(idFuelStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	err = h.Repository.AddFuelToCart(uint(id), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Услуга добавлена в заявку",
	})
}

// GetFuelAPI godoc
// @Summary Получить данные топлива по ID
// @Description Возвращает полную информацию о топливе по указанному идентификатору. Доступен без авторизации.
// @Tags Fuels
// @Produce json
// @Param id path int true "ID топлива" example(3)
// @Success 200 {object} map[string]interface{} "Данные топлива" example({"data":{"ID":3,"title":"Природный газ","heat":35.8,"molar_mass":16.04,"density":0.8,"short_desc":"Чистое топливо","full_desc":"Широко используется в быту и промышленности","is_gas":true}})
// @Failure 400 {object} map[string]string "Некорректный ID" example({"error":"invalid syntax"})
// @Failure 404 {object} map[string]string "Топливо не найдено" example({"error":"Топливо с указанным ID не существует"})
// @Router /api/fuels/{id} [get]
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
		"data": fuel,
	})
}

// GetFuelsAPI returns a list of fuels
// @Summary Получить список топлива
// @Description Возвращает все виды топлива. Поддерживается фильтрация по названию через query-параметр ?title=...
// @Tags Fuels
// @Produce json
// @Param title query string false "Фильтр по названию топлива (частичное совпадение)"
// @Router /api/fuels [get]
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
		"data":  fuels,
		"count": len(fuels),
	})
}

func (h *Handler) GetUserIDFromContext(ctx *gin.Context) (uint, error) {
	userID, exists := ctx.Get("userID")
	if !exists {
		return 0, fmt.Errorf("требуется авторизация")
	}

	return userID.(uint), nil
}

func (h *Handler) GetUserIsModeratorFromContext(ctx *gin.Context) (bool, error) {
	user, exists := ctx.Get("isModerator")
	if !exists {
		return false, fmt.Errorf("требуется авторизация")
	}

	return user.(bool), nil
}
