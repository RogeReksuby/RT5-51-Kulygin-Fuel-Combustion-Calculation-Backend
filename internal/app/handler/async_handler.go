package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"repback/internal/app/ds"
)

// UpdateAsyncResultAPI - прием результатов от Django сервиса
// @Summary Прием результатов асинхронного расчета
// @Description Принимает результаты расчета промежуточной энергии от Django сервиса
// @Tags async
// @Accept json
// @Produce json
// @Param data body ds.AsyncCalculationResult true "Результат расчета"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /api/async/update-result [post]
func (h *Handler) UpdateAsyncResultAPI(ctx *gin.Context) {
	var input ds.AsyncCalculationResult

	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат данных: %v", err))
		return
	}

	// Проверка токена
	expectedToken := "abc123def456"
	if input.Token != expectedToken {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("неверный токен"))
		return
	}

	// Обновляем промежуточную энергию через репозиторий
	if err := h.Repository.UpdateIntermediateEnergy(input.CombustionID, input.FuelID, input.Result); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	logrus.Infof("✅ Сохранена промежуточная энергия для combustion_id=%d, fuel_id=%d: %.4f",
		input.CombustionID, input.FuelID, input.Result)

	// Проверяем, все ли расчеты завершены через репозиторий
	combustionWithCount, err := h.Repository.GetCombustionWithCount(input.CombustionID)
	if err != nil {
		logrus.Errorf("Ошибка получения прогресса: %v", err)
		// Все равно отвечаем 200, т.к. основной результат сохранен
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Результат сохранен",
		})
		return
	}

	// Если ВСЕ расчеты завершены - завершаем заявку
	if combustionWithCount.CalculatedCount >= combustionWithCount.TotalCount && combustionWithCount.TotalCount > 0 {
		logrus.Infof("🎉 Все расчеты завершены для combustion_id=%d (%d/%d). Завершаем заявку...",
			input.CombustionID, combustionWithCount.CalculatedCount, combustionWithCount.TotalCount)

		// 1. Рассчитываем final_result через репозиторий
		totalEnergy, err := h.Repository.CalculateFinalResult(input.CombustionID)
		if err != nil {
			logrus.Errorf("Ошибка расчета финального результата: %v", err)
			ctx.JSON(http.StatusOK, gin.H{"status": "success"})
			return
		}

		// 2. Получаем moderator_id через репозиторий
		moderatorID := combustionWithCount.ModeratorID
		if moderatorID == 0 {
			// Если в CombustionWithCount нет, пробуем получить отдельно
			moderatorID, err = h.Repository.GetModeratorID(input.CombustionID)
			if err != nil {
				logrus.Errorf("Не удалось получить moderator_id: %v", err)
				ctx.JSON(http.StatusOK, gin.H{"status": "success"})
				return
			}
		}

		// 3. Завершаем заявку через репозиторий
		if moderatorID > 0 {
			if err := h.Repository.CompleteOrRejectCombustion(
				input.CombustionID,
				moderatorID,
				true,
				totalEnergy,
			); err != nil {
				logrus.Errorf("Ошибка завершения заявки: %v", err)
			} else {
				logrus.Infof("✅ Заявка %d автоматически завершена. Final result: %.2f кДж",
					input.CombustionID, totalEnergy)
			}
		} else {
			logrus.Warnf("Не найден moderator_id для заявки %d", input.CombustionID)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Результат сохранен",
	})
}

// StartAsyncCalculationAPI - запуск асинхронного расчета
// @Summary Запуск асинхронного расчета
// @Description Запускает асинхронный расчет промежуточных энергий для всех топлив в заявке
// @Tags moderators
// @Security ApiKeyAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/combustions/{id}/start-async [post]
func (h *Handler) StartAsyncCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	combustionID := uint(id)

	// 1. Получаем данные заявки через репозиторий
	combustionData, err := h.Repository.GetCombustionForAsync(combustionID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("заявка не найдена: %v", err))
		return
	}

	// Проверяем статус заявки
	if combustionData.Status != "сформирован" {
		h.errorHandler(ctx, http.StatusBadRequest,
			fmt.Errorf("заявка должна быть в статусе 'сформирован', текущий статус: '%s'", combustionData.Status))
		return
	}

	// 2. Получаем все связи заявки через репозиторий
	fuels, err := h.Repository.GetCombustionFuelsForAsync(combustionID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка получения данных о топливе: %v", err))
		return
	}

	if len(fuels) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("в заявке нет топлива"))
		return
	}

	// 3. Генерируем токен для этой сессии расчета
	token := generateToken()

	// 4. Сохраняем токен в БД через репозиторий
	if err := h.Repository.SetAsyncToken(combustionID, token); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка сохранения токена: %v", err))
		return
	}

	logrus.Infof("🚀 Запуск асинхронного расчета для combustion_id=%d", combustionID)
	logrus.Infof("   Топлив для расчета: %d", len(fuels))
	logrus.Infof("   Молярный объем: %.4f", combustionData.MolarVolume)

	// 5. Запускаем расчет для каждого топлива в отдельной горутине
	for _, fuel := range fuels {
		go func(f ds.AsyncFuelData) {
			// Формируем данные для Django
			data := map[string]interface{}{
				"combustion_id": combustionID,
				"fuel_id":       f.FuelID,
				"fuel_volume":   f.FuelVolume,
				"heat":          f.Heat,
				"molar_mass":    f.MolarMass,
				"density":       f.Density,
				"is_gas":        f.IsGas,
				"molar_volume":  combustionData.MolarVolume,
			}

			// Вызываем Django сервис
			if err := callDjangoService(data, token); err != nil {
				logrus.Errorf("Ошибка вызова Django для fuel_id=%d: %v", f.FuelID, err)
			}
		}(fuel)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Асинхронный расчет запущен",
		"data": gin.H{
			"combustion_id":  combustionID,
			"fuels_count":    len(fuels),
			"molar_volume":   combustionData.MolarVolume,
			"token":          token,
			"estimated_time": "5-10 секунд на каждый расчет",
			"callback_url":   "http://localhost:8080/api/async/update-result",
		},
	})
}

// GetCombustionWithCountAPI - получение заявки с количеством расчетов
// @Summary Получить заявку с прогрессом расчета
// @Description Возвращает заявку с информацией о количестве выполненных расчетов
// @Tags combustions
// @Security ApiKeyAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} ds.CombustionWithCount
// @Router /api/combustions/{id}/progress [get]
func (h *Handler) GetCombustionWithCountAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	combustion, err := h.Repository.GetCombustionWithCount(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, combustion)
}

// Вспомогательная функция для вызова Django сервиса
func callDjangoService(data map[string]interface{}, token string) error {
	// URL Django сервиса
	djangoURL := "http://localhost:8001/calculate/"

	// Добавляем callback_url
	data["callback_url"] = "http://localhost:8080/api/async/update-result"

	// Конвертируем в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}

	// Отправляем запрос
	resp, err := http.Post(djangoURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка HTTP запроса к Django: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Django сервис вернул ошибку: %d", resp.StatusCode)
	}

	// Читаем ответ
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("ошибка чтения ответа от Django: %v", err)
	}

	logrus.Infof("✅ Запрос отправлен в Django: combustion_id=%v, fuel_id=%v, status=%v",
		data["combustion_id"], data["fuel_id"], result["status"])

	return nil
}

// Генерация токена
func generateToken() string {
	return "abc123def456" // фиксированный токен как в задании (8 байт)
}
