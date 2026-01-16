package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"repback/internal/app/ds"
)

// UpdateAsyncResultAPI - прием результатов от Django сервиса
func (h *Handler) UpdateAsyncResultAPI(ctx *gin.Context) {
	var input ds.AsyncCalculationResult
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат данных: %v", err))
		return
	}

	// ✅ Проверка токена из конфига
	expectedToken := h.Config.AsyncServiceToken
	if input.Token != expectedToken {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("неверный токен"))
		return
	}

	if err := h.Repository.UpdateIntermediateEnergy(input.CombustionID, input.FuelID, input.Result); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	logrus.Infof("✅ Сохранена промежуточная энергия для combustion_id=%d, fuel_id=%d: %.4f",
		input.CombustionID, input.FuelID, input.Result)

	combustionWithCount, err := h.Repository.GetCombustionWithCount(input.CombustionID)
	if err != nil {
		logrus.Errorf("Ошибка получения прогресса: %v", err)
		ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "Результат сохранен"})
		return
	}

	if combustionWithCount.CalculatedCount >= combustionWithCount.TotalCount && combustionWithCount.TotalCount > 0 {
		logrus.Infof("🎉 Все расчеты завершены для combustion_id=%d (%d/%d). Завершаем заявку...",
			input.CombustionID, combustionWithCount.CalculatedCount, combustionWithCount.TotalCount)

		totalEnergy, err := h.Repository.CalculateFinalResult(input.CombustionID)
		if err != nil {
			logrus.Errorf("Ошибка расчета финального результата: %v", err)
			ctx.JSON(http.StatusOK, gin.H{"status": "success"})
			return
		}

		moderatorID := combustionWithCount.ModeratorID
		if moderatorID == 0 {
			moderatorID, err = h.Repository.GetModeratorID(input.CombustionID)
			if err != nil {
				logrus.Errorf("Не удалось получить moderator_id: %v", err)
				ctx.JSON(http.StatusOK, gin.H{"status": "success"})
				return
			}
		}

		if moderatorID > 0 {
			if err := h.Repository.CompleteOrRejectCombustion(input.CombustionID, moderatorID, true, totalEnergy); err != nil {
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

// StartAsyncCalculationAPI
func (h *Handler) StartAsyncCalculationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	combustionID := uint(id)

	combustionData, err := h.Repository.GetCombustionForAsync(combustionID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("заявка не найдена: %v", err))
		return
	}

	if combustionData.Status != "сформирован" {
		h.errorHandler(ctx, http.StatusBadRequest,
			fmt.Errorf("заявка должна быть в статусе 'сформирован', текущий статус: '%s'", combustionData.Status))
		return
	}

	fuels, err := h.Repository.GetCombustionFuelsForAsync(combustionID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка получения данных о топливе: %v", err))
		return
	}
	if len(fuels) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("в заявке нет топлива"))
		return
	}

	// ✅ Генерация токена — сейчас фиксированная, но можно заменить на h.Config.AsyncServiceToken
	token := h.generateToken()

	if err := h.Repository.SetAsyncToken(combustionID, token); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка сохранения токена: %v", err))
		return
	}

	logrus.Infof("🚀 Запуск асинхронного расчета для combustion_id=%d", combustionID)
	logrus.Infof("   Топлив для расчета: %d", len(fuels))
	logrus.Infof("   Молярный объем: %.4f", combustionData.MolarVolume)

	for _, fuel := range fuels {
		go func(f ds.AsyncFuelData) {
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

			if err := h.callDjangoService(data, token); err != nil {
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

// ✅ Метод генерации токена (экспортируемый через *Handler)
func (h *Handler) generateToken() string {
	return h.Config.AsyncServiceToken // ← теперь через конфиг
}

// ✅ Метод вызова Django — ПОЛНОСТЬЮ ПЕРЕПИСАН с заголовком
func (h *Handler) callDjangoService(data map[string]interface{}, token string) error {
	djangoURL := "http://localhost:8001/calculate/"
	data["callback_url"] = "http://localhost:8080/api/async/update-result"

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %v", err)
	}

	req, err := http.NewRequest("POST", djangoURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP-запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", token) // ← КЛЮЧЕВОЙ ЗАГОЛОВОК

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка HTTP запроса к Django: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Django вернул ошибку %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("ошибка чтения ответа от Django: %v", err)
	}

	logrus.Infof("✅ Запрос в Django: combustion_id=%v, fuel_id=%v, status=%v",
		data["combustion_id"], data["fuel_id"], result["status"])
	return nil
}
