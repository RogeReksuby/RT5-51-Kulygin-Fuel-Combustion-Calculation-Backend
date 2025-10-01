package repository

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"repback/internal/app/ds"
	"time"
)

func (r *Repository) GetFuels() ([]ds.Fuel, error) {
	var fuels []ds.Fuel
	// тут запрос SELECT *
	err := r.db.Where("is_delete = false").Find(&fuels).Error
	if err != nil {
		return nil, err
	}
	if len(fuels) == 0 {
		return nil, fmt.Errorf("пустой массив")
	}
	return fuels, nil

}

func (r *Repository) RequestStatusById(id int) (string, error) {
	var reqStatus string
	err := r.db.Model(&ds.CombustionCalculation{}).Where("id = ?", id).Select("status").First(&reqStatus).Error
	if err != nil {
		return "", err
	}
	return reqStatus, nil
}

func (r *Repository) GetFuel(id int) (ds.Fuel, error) {
	fuel := ds.Fuel{}
	err := r.db.Where("is_delete = false AND id = ?", id).First(&fuel).Error
	if err != nil {
		return ds.Fuel{}, err
	}
	return fuel, nil
}

func (r *Repository) GetFuelsByTitle(title string) ([]ds.Fuel, error) {
	var fuels []ds.Fuel
	err := r.db.Where("is_delete = false AND title ILIKE ?", "%"+title+"%").Find(&fuels).Error
	if err != nil {
		return nil, err
	}
	return fuels, nil
}

func (r *Repository) GetRequestID(userID uint) int {
	var requestID int
	err := r.db.Model(&ds.CombustionCalculation{}).Where("creator_id = ? AND status = ?", userID, "черновик").Select("id").First(&requestID).Error
	if err != nil {
		return 0
	}
	return requestID
}

func (r *Repository) GetReqFuels(requestID uint) ([]ds.Fuel, error) {
	var fuels []ds.Fuel

	var fuelIDs []int64
	err := r.db.Model(&ds.CombustionsFuels{}).Where("request_id = ?", requestID).Pluck("fuel_id", &fuelIDs).Error
	if err != nil {
		return nil, err
	}

	for _, fuelID := range fuelIDs {
		fuel, err := r.GetFuel(int(fuelID))
		if err != nil {
			return nil, err
		}
		fuels = append(fuels, fuel)
	}
	return fuels, nil
}

// поправь давай емае
func (r *Repository) GetReqFuelsOld() ([]ds.Fuel, error) {
	// имитация получения списка id топлива в заявке
	reqs := []int{2, 4}
	var reqFuels []ds.Fuel
	fuels, err := r.GetFuels()
	if err != nil {
		return nil, err
	}
	for _, id := range reqs {
		for _, fuel := range fuels {
			if fuel.ID == id {
				reqFuels = append(reqFuels, fuel)
			}
		}
	}
	return reqFuels, nil

}

func (r *Repository) GetCartCount() int64 {
	var requestID uint
	var count int64
	creatorID := 1

	err := r.db.Model(&ds.CombustionCalculation{}).Where("creator_id = ? AND status = ?", creatorID, "черновик").Select("id").First(&requestID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.CombustionsFuels{}).Where("request_id = ?", requestID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in list_chats:", err)
	}

	return count
}

func (r *Repository) DeleteFuel(fuelId uint) error {
	err := r.db.Model(&ds.Fuel{}).Where("id = ?", fuelId).UpdateColumn("is_delete", true).Error
	fmt.Println(fuelId)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении услуги с id %d: %w", fuelId, err)
	}

	return nil
}

func (r *Repository) AddToCart(fuelID uint) error {
	userID := 1
	moderatorID := 2
	var requestID uint
	var count int64
	//err := r.db.Model(&ds.CombustionCalculation{}).Where("id = ?", userID).Select("id").First(&requestID).Error
	err := r.db.Model(&ds.CombustionCalculation{}).Where("creator_id = ? AND status = ?", userID, "черновик").Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		newReq := ds.CombustionCalculation{
			Status:      "черновик",
			DateCreate:  time.Now(),
			DateUpdate:  time.Now(),
			CreatorID:   uint(userID),
			ModeratorID: uint(moderatorID),
		}
		err := r.db.Create(&newReq).Error
		if err != nil {
			return err
		}
	}

	err = r.db.Model(&ds.CombustionCalculation{}).Where("creator_id = ? AND status = ?", userID, "черновик").Select("id").First(&requestID).Error
	if err != nil {
		return err
	}

	newFuelReq := ds.CombustionsFuels{
		RequestID: requestID,
		FuelID:    fuelID,
	}

	err = r.db.Create(&newFuelReq).Error
	if err != nil {
		return err
	}

	return nil

}

func (r *Repository) RemoveRequest(requestID uint) error {

	deleteQuery := "UPDATE combustion_calculations SET status = $1, date_finish = $2, date_update = $3 WHERE id = $4"
	r.db.Exec(deleteQuery, "удалён", time.Now(), time.Now(), requestID)
	return nil

}

func (r *Repository) CreateFuel(fuel *ds.Fuel) error {
	if fuel.Title == "" {
		return fmt.Errorf("название топлива обязательно")
	}

	err := r.db.Select(
		"Title", "Heat", "MolarMass", "CardImage",
		"ShortDesc", "FullDesc", "IsGas", "IsDelete",
	).Create(fuel).Error
	if err != nil {
		return fmt.Errorf("ошибка при создании топлива: %w", err)
	}

	return nil
}

func (r *Repository) UpdateFuel(id uint, fuelData *ds.Fuel) error {

	var existingFuel ds.Fuel
	err := r.db.Where("id = ? AND is_delete = false", id).First(&existingFuel).Error
	if err != nil {
		return fmt.Errorf("топливо с ID %d не найдено", id)
	}

	updates := map[string]interface{}{
		"title":      fuelData.Title,
		"heat":       fuelData.Heat,
		"molar_mass": fuelData.MolarMass,
		"card_image": fuelData.CardImage,
		"short_desc": fuelData.ShortDesc,
		"full_desc":  fuelData.FullDesc,
		"is_gas":     fuelData.IsGas,
	}

	for key, value := range updates {
		if value == "" || value == nil {
			delete(updates, key)
		}
	}

	err = r.db.Model(&ds.Fuel{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("ошибка при обновлении топлива: %w", err)
	}

	return nil
}
