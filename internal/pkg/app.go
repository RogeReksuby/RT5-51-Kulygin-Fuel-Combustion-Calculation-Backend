package pkg

import (
	"fmt"
	"net/http"
	"os"

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

	// Временное решение - используем только HTTP
	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	logrus.Infof("Starting HTTP server on %s", serverAddress)

	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatalf("Failed to start server: %v", err)
	}
}
func (a *Application) startHTTPRedirect() {
	httpRouter := gin.Default()

	// Простой редирект с HTTP на HTTPS
	httpRouter.Any("/*path", func(c *gin.Context) {
		target := "https://" + c.Request.Host + c.Request.URL.Path
		if len(c.Request.URL.RawQuery) > 0 {
			target += "?" + c.Request.URL.RawQuery
		}
		c.Redirect(http.StatusPermanentRedirect, target)
	})

	httpAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	logrus.Infof("Starting HTTP redirect server on %s", httpAddress)

	if err := httpRouter.Run(httpAddress); err != nil {
		logrus.Warnf("HTTP redirect server error: %v", err)
	}
}

func (a *Application) startHTTPS() {
	// Проверяем существование сертификатов
	certFile := a.Config.HTTPSCertFile
	keyFile := a.Config.HTTPSKeyFile

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		logrus.Warnf("SSL certificate not found at %s, generating self-signed certificates...", certFile)
		if err := a.generateSelfSignedCert(); err != nil {
			logrus.Fatalf("Failed to generate SSL certificates: %v", err)
		}
	}

	// Создаем HTTPS сервер
	server := &http.Server{
		Addr:    a.Config.HTTPSAddress,
		Handler: a.Router,
	}

	httpsAddress := a.Config.HTTPSAddress
	logrus.Infof("Starting HTTPS server on %s", httpsAddress)
	logrus.Infof("Cert file: %s, Key file: %s", certFile, keyFile)

	// Запускаем HTTPS сервер
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		logrus.Fatalf("Failed to start HTTPS server: %v", err)
	}
}

func (a *Application) generateSelfSignedCert() error {
	return a.Config.GenerateSelfSignedCert()
}
