package ds

// AsyncResult для приема результатов от Django
type AsyncCalculationResult struct {
	CombustionID uint    `json:"combustion_id" binding:"required"`
	FuelID       uint    `json:"fuel_id" binding:"required"`
	Result       float64 `json:"result" binding:"required"` // intermediate_energy
	Token        string  `json:"token" binding:"required"`  // для проверки
}

// CombustionWithCount для отображения прогресса
type CombustionWithCount struct {
	CombustionCalculation
	CalculatedCount int `json:"calculated_count"` // сколько промежуточных энергий рассчитано
	TotalCount      int `json:"total_count"`      // всего связей
}
