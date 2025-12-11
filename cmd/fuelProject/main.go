package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"repback/internal/app/config"
	"repback/internal/app/dsn"
	"repback/internal/app/handler"
	"repback/internal/app/redis"
	"repback/internal/app/repository"
	"repback/internal/pkg"
)

// @title Расчёт горения топлива
// @version 1.0
// @description Система для формирования и модерации заявок на расчёт параметров горения топлива. Позволяет пользователям добавлять топливо в заявку, заполнять молярный объём, а модераторам — проверять и утверждать расчёты.

// @contact.name Репозиторий
// @contact.url https://github.com/RogeReksuby/web_rip

// @license.name AS IS (NO WARRANTY)

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите JWT токен в формате: Bearer {your_token}

// @host localhost:8080
// @schemes http
// @BasePath /

func main() {
	router := gin.Default()

	// === CORS MIDDLEWARE ===
	router.Use(func(c *gin.Context) {
		// Список разрешённых источников (origins)
		allowedOrigins := []string{
			"https://rogereksuby.github.io",
			"https://localhost:8080",
			"https://localhost:3000",
			"https://192.168.1.173:8080",
			"https://192.168.56.1:8080",
			"https://localhost:8443",
			"https://192.168.1.173:3000",
			"https://192.168.56.1:3000",
		}

		// Получаем источник текущего запроса
		requestOrigin := c.Request.Header.Get("Origin")

		// Проверяем, разрешён ли источник
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if requestOrigin == allowedOrigin {
				c.Header("Access-Control-Allow-Origin", requestOrigin)
				originAllowed = true
				break
			}
		}

		// Если источник не из разрешённого списка, НЕ устанавливаем заголовок.
		// Это значит, что браузер заблокирует ответ.

		// Важно: Для запросов с куками/авторизацией
		if originAllowed {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-CSRF-Token, localtonet-skip-warning")
		c.Header("Access-Control-Max-Age", "86400")

		// Обрабатываем preflight OPTIONS запросы
		if c.Request.Method == "OPTIONS" {
			c.Header("Content-Length", "0")
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Fuel Calculation API",
			"version": "1.0",
			"docs":    "/swagger/index.html",
			"endpoints": gin.H{
				"http":  "http://localhost:8080",
				"https": "https://localhost:8443",
			},
			"note": "HTTP automatically redirects to HTTPS",
		})
	})

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	redisClient, err := redis.New(context.Background(), *conf)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)
	logrus.Infof("MinIO Config: endpoint=%s, access_key=%s, bucket=%s",
		conf.MinIOEndpoint, conf.MinIOAccessKey, conf.MinIOBucket)

	minioClient, err := conf.InitMinIO()
	if err != nil {
		logrus.Fatalf("error initializing MinIO: %v", err)
	}
	logrus.Info("MinIO client initialized successfully")

	rep, errRep := repository.New(postgresString, minioClient, conf.MinIOBucket, redisClient)
	if errRep != nil {
		logrus.Fatalf("error creating repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)
	application := pkg.NewApp(conf, router, hand)

	// === ДОБАВЛЕНО: ПРОКСИ ДЛЯ MINIO ИЗОБРАЖЕНИЙ ===
	router.GET("/minio/*path", func(c *gin.Context) {
		path := c.Param("path")

		// Формируем URL до MinIO
		minioURL := fmt.Sprintf("http://localhost:9000%s", path)

		// Создаем HTTP клиент
		client := &http.Client{}

		// Создаем запрос к MinIO
		req, err := http.NewRequest("GET", minioURL, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request to MinIO"})
			return
		}

		// Выполняем запрос к MinIO
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to MinIO: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		// Копируем заголовки из MinIO
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Копируем статус код
		c.Status(resp.StatusCode)

		// Копируем тело ответа
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			logrus.Errorf("Failed to copy MinIO response: %v", err)
		}
	})

	// Запускаем приложение
	application.RunApp()
}
