package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"repback/internal/app/config"
	"repback/internal/app/dsn"
	"repback/internal/app/handler"
	"repback/internal/app/repository"
	"repback/internal/pkg"
)

// @title BITOP
// @version 1.0
// @description Bmstu Open IT Platform

// @contact.name API Support
// @contact.url https://vk.com/bmstu_schedule
// @contact.email bitop@spatecon.ru

// @license.name AS IS (NO WARRANTY)

// @host localhost:8080
// @schemes https http
// @BasePath /

func main() {
	router := gin.Default()
	// Добавьте CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)
	logrus.Infof("MinIO Config: endpoint=%s, access_key=%s, bucket=%s",
		conf.MinIOEndpoint, conf.MinIOAccessKey, conf.MinIOBucket)

	minioClient, err := conf.InitMinIO()
	if err != nil {
		logrus.Fatalf("error initializing MinIO: %v", err)
	}
	logrus.Info("MinIO client initialized successfully")

	rep, errRep := repository.New(postgresString, minioClient, conf.MinIOBucket)
	if errRep != nil {
		logrus.Fatalf("error creating repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)
	application := pkg.NewApp(conf, router, hand)

	application.RunApp()
}
