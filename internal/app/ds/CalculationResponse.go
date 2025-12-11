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

// Структуры для асинхронного расчета
type AsyncFuelData struct {
	FuelID     uint    `json:"fuel_id"`
	FuelVolume float64 `json:"fuel_volume"`
	Heat       float64 `json:"heat"`
	MolarMass  float64 `json:"molar_mass"`
	Density    float64 `json:"density"`
	IsGas      bool    `json:"is_gas"`
}

type AsyncCombustionData struct {
	MolarVolume float64 `json:"molar_volume"`
	Status      string  `json:"status"`
}
