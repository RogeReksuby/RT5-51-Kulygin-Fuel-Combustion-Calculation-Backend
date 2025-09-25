package ds

type CombustionsFuels struct {
	ID uint `gorm:"primaryKey"`
	// здесь создаем Unique key, указывая общий uniqueIndex
	RequestID uint `gorm:"not null;uniqueIndex:idx_request_fuel"`
	FuelID    uint `gorm:"not null;uniqueIndex:idx_request_fuel"`

	FuelVolume           float64 `gorm:"type:decimal(10,4)"`
	IntermediateEnergies float64 `gorm:"type:decimal(15,4)"`

	Request CombustionCalculation `gorm:"foreignKey:RequestID"`
	Fuel    Fuel                  `gorm:"foreignKey:FuelID"`
}
