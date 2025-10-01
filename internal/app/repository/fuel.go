package repository

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
	"mime/multipart"
	"repback/internal/app/ds"
	"strings"
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
	// Сначала получаем информацию о топливе чтобы узнать путь к изображению
	var fuel ds.Fuel
	err := r.db.Where("id = ?", fuelId).First(&fuel).Error
	if err != nil {
		return fmt.Errorf("топливо с ID %d не найдено: %w", fuelId, err)
	}

	// Удаляем изображение из MinIO если оно есть
	if fuel.CardImage != "" {
		if err := r.DeleteFuelImage(fuel.CardImage); err != nil {
			// Логируем ошибку, но продолжаем удаление записи
			logrus.Errorf("Не удалось удалить изображение для топлива %d: %v", fuelId, err)
		}
	}

	err = r.db.Model(&ds.Fuel{}).Where("id = ?", fuelId).UpdateColumn("is_delete", true).Error
	fmt.Println(fuelId)
	if err != nil {
		return fmt.Errorf("Ошибка при удалении услуги с id %d: %w", fuelId, err)
	}

	return nil
}

func (r *Repository) AddFuelToCart(fuelID uint) error {
	userID := 1
	var requestID uint
	var count int64
	err := r.db.Model(&ds.CombustionCalculation{}).Where("creator_id = ? AND status = ?", userID, "черновик").Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		newReq := ds.CombustionCalculation{
			Status:     "черновик",
			DateCreate: time.Now(),
			CreatorID:  uint(userID),
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

func (r *Repository) DeleteFuelImage(imagePath string) error {
	if imagePath == "" {
		return nil // нет изображения - ничего не делаем
	}

	// Извлекаем имя файла из пути
	objectName := r.extractObjectName(imagePath)
	if objectName == "" {
		return nil
	}

	// Удаляем объект из MinIO
	err := r.minioClient.RemoveObject(context.Background(), r.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка при удалении изображения из MinIO: %w", err)
	}

	return nil
}

// extractObjectName - извлекает имя объекта из полного пути
func (r *Repository) extractObjectName(imagePath string) string {
	// Предполагаем, что путь хранится как "bucket-name/folder/image.jpg" или просто "image.jpg"
	parts := strings.Split(imagePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1] // возвращаем последнюю часть пути
	}
	return imagePath
}

// UploadFuelImage - загрузка/обновление изображения для услуги
func (r *Repository) UploadFuelImage(fuelID uint, fileHeader *multipart.FileHeader) error {
	// Проверяем существование услуги
	var fuel ds.Fuel
	err := r.db.Where("id = ? AND is_delete = false", fuelID).First(&fuel).Error
	if err != nil {
		return fmt.Errorf("услуга с ID %d не найдена", fuelID)
	}

	// Удаляем старое изображение если есть
	if fuel.CardImage != "" {
		if err := r.DeleteFuelImage(fuel.CardImage); err != nil {
			logrus.Errorf("Не удалось удалить старое изображение: %v", err)
		}
	}

	// Оставляем оригинальное название файла
	fileName := fmt.Sprintf("fuel_%d_%s", fuelID, fileHeader.Filename)

	// Открываем файл
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	// Загружаем в MinIO
	_, err = r.minioClient.PutObject(
		context.Background(),
		r.bucketName,
		fileName,
		file,
		fileHeader.Size,
		minio.PutObjectOptions{
			ContentType: fileHeader.Header.Get("Content-Type"),
		},
	)
	if err != nil {
		return fmt.Errorf("ошибка загрузки в MinIO: %w", err)
	}

	// Обновляем путь к изображению в базе
	fuel.CardImage = "http://127.0.0.1:9000/ripimages/" + fileName
	err = r.db.Save(&fuel).Error
	if err != nil {
		// Если не удалось сохранить в БД, удаляем из MinIO
		r.minioClient.RemoveObject(context.Background(), r.bucketName, fileName, minio.RemoveObjectOptions{})
		return fmt.Errorf("ошибка сохранения пути к изображению: %w", err)
	}

	return nil
}

func (r *Repository) RegisterUser(login, password, name string, isModerator bool) (*ds.Users, error) {
	// Проверяем что логин не пустой
	if login == "" {
		return nil, fmt.Errorf("логин не может быть пустым")
	}

	// Проверяем что пароль не пустой
	if password == "" {
		return nil, fmt.Errorf("пароль не может быть пустым")
	}

	// Проверяем уникальность логина
	var existingUser ds.Users
	err := r.db.Where("login = ?", login).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("пользователь с логином '%s' уже существует", login)
	}

	// Создаем пользователя (пароль сохраняется как есть)
	newUser := ds.Users{
		Login:       login,
		Password:    password,
		IsModerator: isModerator,
		Name:        name,
	}

	err = r.db.Model(&ds.Users{}).Create(map[string]interface{}{
		"login":        login,
		"password":     password,
		"name":         name,
		"is_moderator": isModerator,
	}).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	newUser.Password = ""

	return &newUser, nil
}

func (r *Repository) GetUserProfile(userID uint) (*ds.Users, error) {
	var user ds.Users

	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден")
	}

	// Не возвращаем пароль
	user.Password = ""

	return &user, nil
}

func (r *Repository) AuthenticateUser(login, password string) (*ds.Users, error) {
	var user ds.Users

	// Ищем пользователя по логину
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("неверный логин или пароль")
	}

	// Проверяем пароль (без хеширования)
	if user.Password != password {
		return nil, fmt.Errorf("неверный логин или пароль")
	}

	// Не возвращаем пароль
	user.Password = ""

	return &user, nil
}

func (r *Repository) UpdateUser(userID uint, updates map[string]interface{}) (*ds.Users, error) {
	// Проверяем уникальность логина если он передается
	if login, exists := updates["login"]; exists && login != "" {
		var existingUser ds.Users
		err := r.db.Where("login = ? AND id != ?", login, userID).First(&existingUser).Error
		if err == nil {
			return nil, fmt.Errorf("логин '%s' уже занят", login)
		}
	}

	// Обновляем только переданные поля
	if len(updates) > 0 {
		err := r.db.Model(&ds.Users{}).Where("id = ?", userID).Updates(updates).Error
		if err != nil {
			return nil, fmt.Errorf("ошибка обновления: %w", err)
		}
	}

	// Получаем обновленного пользователя
	var user ds.Users
	r.db.Where("id = ?", userID).First(&user)
	user.Password = ""

	return &user, nil
}

// DeleteCombustionCalculation - удаление заявки (мягкое удаление)
func (r *Repository) DeleteCombustionCalculation(calculationID uint) error {
	// Проверяем существование заявки
	var calculation ds.CombustionCalculation
	err := r.db.Where("id = ?", calculationID).First(&calculation).Error
	if err != nil {
		return fmt.Errorf("заявка с ID %d не найдена", calculationID)
	}

	// Мягкое удаление - меняем статус на "удалён" и обновляем дату
	err = r.db.Model(&ds.CombustionCalculation{}).Where("id = ?", calculationID).Updates(map[string]interface{}{
		"status":      "удалён",
		"date_update": time.Now(),
	}).Error

	if err != nil {
		return fmt.Errorf("ошибка при удалении заявки: %w", err)
	}

	return nil
}
