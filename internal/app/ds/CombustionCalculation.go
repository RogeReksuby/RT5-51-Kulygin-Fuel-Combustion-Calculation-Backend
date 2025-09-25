package ds

import (
	"database/sql"
	"time"
)

type CombustionCalculation struct {
	ID          uint      `gorm:"primaryKey"`
	Status      string    `gorm:"type:varchar(15);not null"`
	DateCreate  time.Time `gorm:"not null"`
	DateUpdate  time.Time
	DateFinish  sql.NullTime `gorm:"default:null"`
	CreatorID   uint         `gorm:"not null"`
	ModeratorID uint         `gorm:"default:null"`

	// Новые поля для данных расчета
	MolarVolume float64 `gorm:"type:decimal(10,4);default:22.414"` // молярный объем (л/моль)
	FinalResult float64 `gorm:"type:decimal(15,4)"`                // итоговый результат (кДж)

	// Связи
	Creator   Users `gorm:"foreignKey:CreatorID"`
	Moderator Users `gorm:"foreignKey:ModeratorID"`
}
