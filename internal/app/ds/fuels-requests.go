package ds

type RequestFuel struct {
	ID uint `gorm:"primaryKey"`
	// здесь создаем Unique key, указывая общий uniqueIndex
	RequestID uint `gorm:"not null;uniqueIndex:idx_request_fuel"`
	FuelID    uint `gorm:"not null;uniqueIndex:idx_request_fuel"`

	Request Request `gorm:"foreignKey:RequestID"`
	Fuel    Fuel    `gorm:"foreignKey:FuelID"`
}
