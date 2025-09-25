package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"repback/internal/app/ds"
	"repback/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(&ds.Fuel{}, &ds.CombustionCalculation{}, &ds.Users{}, &ds.CombustionsFuels{})
	if err != nil {
		panic("failed to migrate database")
	}

}
