package repository

import "repback/internal/app/ds"

// Обновить промежуточную энергию (вызывается Django)
func (r *Repository) UpdateIntermediateEnergy(combustionID, fuelID uint, result float64) error {
	return r.db.Model(&ds.CombustionsFuels{}).
		Where("request_id = ? AND fuel_id = ?", combustionID, fuelID).
		Updates(map[string]interface{}{
			"intermediate_energies": result,
			"is_calculated":         true,
		}).Error
}

// Получить заявку с количеством расчетов
func (r *Repository) GetCombustionWithCount(id uint) (*ds.CombustionWithCount, error) {
	var combustion ds.CombustionCalculation
	if err := r.db.Preload("Creator").Preload("Moderator").
		First(&combustion, id).Error; err != nil {
		return nil, err
	}

	// Считаем рассчитанные энергии
	var calculatedCount, totalCount int64
	r.db.Model(&ds.CombustionsFuels{}).
		Where("request_id = ? AND is_calculated = ?", id, true).
		Count(&calculatedCount)

	r.db.Model(&ds.CombustionsFuels{}).
		Where("request_id = ?", id).
		Count(&totalCount)

	return &ds.CombustionWithCount{
		CombustionCalculation: combustion,
		CalculatedCount:       int(calculatedCount),
		TotalCount:            int(totalCount),
	}, nil
}

// Установить токен для асинхронного расчета
func (r *Repository) SetAsyncToken(combustionID uint, token string) error {
	return r.db.Model(&ds.CombustionCalculation{}).
		Where("id = ?", combustionID).
		Updates(map[string]interface{}{
			"async_token":        token,
			"calculation_status": "processing",
		}).Error
}

// Проверить токен
func (r *Repository) CheckAsyncToken(combustionID uint, token string) bool {
	var combustion ds.CombustionCalculation
	if err := r.db.Select("async_token").First(&combustion, combustionID).Error; err != nil {
		return false
	}
	return combustion.AsyncToken == token
}

// GetCombustionFuelsForAsync с типизированной структурой
func (r *Repository) GetCombustionFuelsForAsync(combustionID uint) ([]ds.AsyncFuelData, error) {
	var results []ds.AsyncFuelData

	err := r.db.Table("combustions_fuels as cf").
		Select(`
			cf.fuel_id,
			cf.fuel_volume,
			f.heat,
			f.molar_mass,
			f.density,
			f.is_gas
		`).
		Joins("LEFT JOIN fuels f ON f.id = cf.fuel_id").
		Where("cf.request_id = ?", combustionID).
		Scan(&results).Error

	return results, err
}

// GetMolarVolume - получить молярный объем заявки
func (r *Repository) GetMolarVolume(combustionID uint) (float64, error) {
	var combustion struct {
		MolarVolume float64
	}

	err := r.db.Model(&ds.CombustionCalculation{}).
		Select("molar_volume").
		Where("id = ?", combustionID).
		First(&combustion).Error

	if err != nil {
		return 0, err
	}

	return combustion.MolarVolume, nil
}

// GetCombustionForAsync с типизированной структурой
func (r *Repository) GetCombustionForAsync(combustionID uint) (*ds.AsyncCombustionData, error) {
	var result ds.AsyncCombustionData

	err := r.db.Model(&ds.CombustionCalculation{}).
		Select("molar_volume, status").
		Where("id = ?", combustionID).
		First(&result).Error

	return &result, err
}

// SaveModeratorForAsync - сохранить ID модератора для асинхронного завершения
func (r *Repository) SaveModeratorForAsync(combustionID, moderatorID uint) error {
	return r.db.Model(&ds.CombustionCalculation{}).
		Where("id = ?", combustionID).
		Update("moderator_id", moderatorID).Error
}

// GetCombustionFuelsWithEnergies - получить все связи заявки с промежуточными энергиями
func (r *Repository) GetCombustionFuelsWithEnergies(combustionID uint) ([]ds.CombustionsFuels, error) {
	var fuelComb []ds.CombustionsFuels
	err := r.db.Where("request_id = ?", combustionID).Find(&fuelComb).Error
	return fuelComb, err
}

// CalculateFinalResult - рассчитать финальный результат из промежуточных энергий
func (r *Repository) CalculateFinalResult(combustionID uint) (float64, error) {
	var fuelComb []ds.CombustionsFuels
	err := r.db.Where("request_id = ?", combustionID).Find(&fuelComb).Error
	if err != nil {
		return 0, err
	}

	var totalEnergy float64
	for _, fuelFromComb := range fuelComb {
		totalEnergy += fuelFromComb.IntermediateEnergies
	}

	return totalEnergy, nil
}

// GetModeratorID - получить ID модератора заявки
func (r *Repository) GetModeratorID(combustionID uint) (uint, error) {
	var combustion struct {
		ModeratorID uint
	}

	err := r.db.Model(&ds.CombustionCalculation{}).
		Select("moderator_id").
		Where("id = ?", combustionID).
		First(&combustion).Error

	return combustion.ModeratorID, err
}
