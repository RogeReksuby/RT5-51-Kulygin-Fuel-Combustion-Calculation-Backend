package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"repback/internal/app/handler"
	"repback/internal/app/repository"
)

func StartServer() {
	log.Println("Server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	// создание дефолтного роутера с настройками по умолчанию
	// он умеет сам логгировать запросы и восстанавливать сервер
	r := gin.Default()

	// метод роутера для считывания файлов шаблонов, и путь по которому они лежат
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// регистрирует функцию - обработчик для GET запросов по пути
	// gin context содержит всю информацию о запросе, объект контекста
	// он предоставляет доступ к данным запроса и методы для формирования ответа (например JSON)
	r.GET("/fuels", handler.GetFuels)
	r.GET("/fuel/:id", handler.GetFuel)
	r.GET("/req", handler.GetReqFuels)
	r.Run()

	log.Println("Server down")
}
