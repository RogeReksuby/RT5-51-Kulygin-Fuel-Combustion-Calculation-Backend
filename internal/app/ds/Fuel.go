package ds

type Fuel struct {
	ID        int     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string  `gorm:"type:varchar(100);not null" json:"title"`
	Heat      float64 `gorm:"type:decimal(10,2);not null" json:"heat"`
	MolarMass float64 `gorm:"type:decimal(10,2)" json:"molar_mass"`
	Density   float64 `gorm:"type:decimal(10,2)" json:"density"`
	CardImage string  `gorm:"type:varchar(255)" json:"card_image"`
	ShortDesc string  `gorm:"type:varchar(200)" json:"short_desc"`
	FullDesc  string  `gorm:"type:text" json:"full_desc"`
	IsGas     bool    `gorm:"type:boolean;default:false" json:"is_gas"`
	IsDelete  bool    `gorm:"type:boolean;default:false;not null" json:"is_delete"`
}
