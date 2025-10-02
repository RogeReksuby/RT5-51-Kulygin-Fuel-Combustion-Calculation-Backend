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

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
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
