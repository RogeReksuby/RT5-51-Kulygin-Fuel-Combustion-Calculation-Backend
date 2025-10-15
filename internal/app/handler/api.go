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
		"data":    updatedCalculation,
		"fuels":   fuels,
		"message": message,
	})
}

// RemoveFuelFromCombustionAPI - DELETE удаление услуги из заявки
func (h *Handler) RemoveFuelFromCombustionAPI(ctx *gin.Context) {

	calculationID := h.Repository.GetRequestID(1)

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

// UpdateFuelInCombustionAPI - PUT изменение данных в связи м-м
func (h *Handler) UpdateFuelInCombustionAPI(ctx *gin.Context) {

	calculationID := h.Repository.GetRequestID(1)

	var input struct {
		FuelID     uint    `json:"fuel_id" binding:"required"`
		FuelVolume float64 `json:"fuel_volume" binding:"required"`
	}
	var err error
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

// DeleteCombustionCalculationAPI - DELETE удаление заявки
func (h *Handler) DeleteCombustionCalculationAPI(ctx *gin.Context) {
	id := h.Repository.GetRequestID(1)

	err := h.Repository.DeleteCombustionCalculation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Заявка успешно удалена",
	})
}

// UpdateUserAPI - обновление только переданных полей
func (h *Handler) UpdateUserAPI(ctx *gin.Context) {
	userID := uint(1)

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
// @Tags Auth
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
// @Tags Auth
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
		"data":    updatedFuel,
		"message": "Изображение успешно загружено",
	})
}

func (h *Handler) GetCombCartIconAPI(ctx *gin.Context) {

	requestID := h.Repository.GetRequestID(1)
	cartCount := h.Repository.GetCartCount()

	ctx.JSON(http.StatusOK, gin.H{
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

func (h *Handler) GetUserProfileAPI(ctx *gin.Context) {
	// В реальном приложении ID пользователя берется из JWT токена или сессии
	// Для лабораторной работы используем фиксированного пользователя
	userID := uint(1) // Фиксированный ID пользователя

	user, err := h.Repository.GetUserProfile(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": user,
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
		"message": "Топливо успешно удалено",
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
		"data":    updatedFuel,
		"message": "Топливо успешно обновлено",
	})
}

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

func (h *Handler) AddFuelToCartAPI(ctx *gin.Context) {
	idFuelStr := ctx.Param("id")
	id, err := strconv.Atoi(idFuelStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	err = h.Repository.AddFuelToCart(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Услуга добавлена в заявку",
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
		"data": fuel,
	})
}

// GetFuelsAPI returns a list of fuels
// @Summary Получить список топлива
// @Description Возвращает все виды топлива. Поддерживается фильтрация по названию через query-параметр ?title=...
// @Tags fuels
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
