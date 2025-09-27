package ds

type Fuel struct {
	ID        int     `gorm:"primaryKey;autoIncrement"`
	Title     string  `gorm:"type:varchar(100);not null"`
	Heat      float64 `gorm:"type:decimal(10,2);not null"`
	MolarMass float64 `gorm:"type:decimal(10,2);not null"`
	CardImage string  `gorm:"type:varchar(255)"`
	ShortDesc string  `gorm:"type:varchar(200)"`
	FullDesc  string  `gorm:"type:text"`
	IsGas     bool    `gorm:"type:boolean;default:false"`
	IsDelete  bool    `gorm:"type:boolean;default:false;not null"`
}
