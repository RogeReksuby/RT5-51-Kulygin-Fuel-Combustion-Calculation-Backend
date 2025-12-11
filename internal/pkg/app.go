package pkg

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"repback/internal/app/config"
	"repback/internal/app/handler"
)

type Application struct {
	Config  *config.Config
	Router  *gin.Engine
	Handler *handler.Handler
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler) *Application {
	return &Application{
		Config:  c,
		Router:  r,
		Handler: h,
	}
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")
	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)

	// Проверяем наличие сертификатов mkcert
	certFile := a.Config.HTTPSCertFile
	keyFile := a.Config.HTTPSKeyFile

	hasCertificates := false
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			hasCertificates = true
		}
	}

	if hasCertificates {
		logrus.Infof("Found mkcert certificates: %s, %s", certFile, keyFile)

		// Запускаем оба сервера одновременно
		go a.startHTTPRedirect()

		// Основной HTTPS сервер (блокирующий вызов)
		a.startHTTPS()
	} else {
		// Если сертификатов нет, используем обычный HTTP
		logrus.Warn("SSL certificates not found, starting HTTP server only")

		serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
		logrus.Infof("Starting HTTP server on %s", serverAddress)

		if err := a.Router.Run(serverAddress); err != nil {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}
}

func (a *Application) startHTTPRedirect() {
	// Создаем отдельный роутер ТОЛЬКО для редиректа
	redirectRouter := gin.New()
	redirectRouter.Use(gin.Recovery())

	// Middleware для логирования редиректов
	redirectRouter.Use(func(c *gin.Context) {
		logrus.Infof("HTTP Redirect: %s %s -> https://%s%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.Host,
			c.Request.URL.Path)
		c.Next()
	})

	// РЕШАЕМ ПРОБЛЕМУ: Явно проверяем, что это HTTP, а не HTTPS
	redirectRouter.Use(func(c *gin.Context) {
		// Если это HTTPS запрос (не должен сюда попадать) - отдаем ошибку
		if c.Request.TLS != nil {
			logrus.Warnf("HTTPS request to HTTP redirect server: %s", c.Request.Host)
			c.JSON(400, gin.H{
				"error":     "This is HTTP redirect server. Use HTTPS directly.",
				"https_url": "https://" + strings.Replace(c.Request.Host, ":8080", ":8443", 1) + c.Request.URL.Path,
			})
			c.Abort()
			return
		}
		c.Next()
	})

	// Редирект ВСЕХ запросов с HTTP на HTTPS
	redirectRouter.Any("/*path", func(c *gin.Context) {
		// Формируем HTTPS URL
		httpsHost := strings.Replace(c.Request.Host, ":8080", ":8443", 1)
		if !strings.Contains(httpsHost, ":") {
			// Если порт не указан, добавляем стандартный HTTPS порт
			httpsHost = httpsHost + ":8443"
		}

		target := "https://" + httpsHost + c.Request.URL.Path
		if len(c.Request.URL.RawQuery) > 0 {
			target += "?" + c.Request.URL.RawQuery
		}

		logrus.Infof("Redirecting HTTP -> HTTPS: %s", target)
		c.Redirect(http.StatusPermanentRedirect, target)
	})

	// Запускаем HTTP сервер редиректа на стандартном HTTP порту
	httpAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	logrus.Infof("Starting HTTP redirect server on %s", httpAddress)
	logrus.Infof("All HTTP traffic will be redirected to HTTPS on port 8443")

	// Важно: используем обычный HTTP сервер без TLS
	server := &http.Server{
		Addr:    httpAddress,
		Handler: redirectRouter,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("HTTP redirect server failed: %v", err)
	}
}

func (a *Application) startHTTPS() {
	certFile := a.Config.HTTPSCertFile
	keyFile := a.Config.HTTPSKeyFile

	// Создаем HTTPS сервер
	server := &http.Server{
		Addr:    a.Config.HTTPSAddress,
		Handler: a.Router,
	}

	httpsAddress := a.Config.HTTPSAddress
	logrus.Infof("Starting HTTPS server on %s", httpsAddress)
	logrus.Infof("Using certificates: %s, %s", certFile, keyFile)

	// Middleware для логирования HTTPS запросов
	a.Router.Use(func(c *gin.Context) {
		logrus.Infof("HTTPS Request: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	})

	// Запускаем HTTPS сервер
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("Failed to start HTTPS server: %v", err)
	}
}
