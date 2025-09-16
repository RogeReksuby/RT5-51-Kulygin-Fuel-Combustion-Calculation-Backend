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

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error creating repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()

}
