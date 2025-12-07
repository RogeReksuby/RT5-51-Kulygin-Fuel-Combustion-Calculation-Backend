package config

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int

	// HTTPS конфигурация
	HTTPSAddress  string `mapstructure:"https_address"`
	HTTPSCertFile string `mapstructure:"https_cert_file"`
	HTTPSKeyFile  string `mapstructure:"https_key_file"`

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOUseSSL    bool
	MinIOBucket    string

	JWTSecretKey string        `mapstructure:"jwt_secret_key"`
	JWTExpiresIn time.Duration `mapstructure:"jwt_expires_in"`
	JWTIssuer    string        `mapstructure:"jwt_issuer"`

	RedisHost        string        `mapstructure:"redis_host"`
	RedisPort        int           `mapstructure:"redis_port"`
	RedisPassword    string        `mapstructure:"redis_password"`
	RedisUser        string        `mapstructure:"redis_user"`
	RedisDialTimeout time.Duration `mapstructure:"redis_dial_timeout"`
	RedisReadTimeout time.Duration `mapstructure:"redis_read_timeout"`
}

const (
	envRedisHost = "REDIS_HOST"
	envRedisPort = "REDIS_PORT"
	envRedisUser = "REDIS_USER"
	envRedisPass = "REDIS_PASSWORD"
)

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	// Значения по умолчанию
	viper.SetDefault("redis_host", "localhost")
	viper.SetDefault("redis_port", 6379)
	viper.SetDefault("redis_dial_timeout", "10s")
	viper.SetDefault("redis_read_timeout", "10s")
	viper.SetDefault("https_address", ":8443")
	viper.SetDefault("https_cert_file", "certs/server.crt")
	viper.SetDefault("https_key_file", "certs/server.key")
	viper.SetDefault("service_host", "localhost")
	viper.SetDefault("service_port", 8080)

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		ServiceHost:      viper.GetString("service_host"),
		ServicePort:      viper.GetInt("service_port"),
		HTTPSAddress:     viper.GetString("https_address"),
		HTTPSCertFile:    viper.GetString("https_cert_file"),
		HTTPSKeyFile:     viper.GetString("https_key_file"),
		MinIOEndpoint:    viper.GetString("endpoint"),
		MinIOAccessKey:   viper.GetString("access_key"),
		MinIOSecretKey:   viper.GetString("secret_key"),
		MinIOUseSSL:      viper.GetBool("use_ssl"),
		MinIOBucket:      viper.GetString("bucket"),
		JWTSecretKey:     viper.GetString("jwt_secret_key"),
		JWTExpiresIn:     viper.GetDuration("jwt_expires_in"),
		JWTIssuer:        viper.GetString("jwt_issuer"),
		RedisHost:        viper.GetString("redis_host"),
		RedisPort:        viper.GetInt("redis_port"),
		RedisPassword:    viper.GetString("redis_password"),
		RedisUser:        viper.GetString("redis_user"),
		RedisDialTimeout: viper.GetDuration("redis_dial_timeout"),
		RedisReadTimeout: viper.GetDuration("redis_read_timeout"),
	}

	if host := os.Getenv(envRedisHost); host != "" {
		cfg.RedisHost = host
	}
	if port := os.Getenv(envRedisPort); port != "" {
		cfg.RedisPort, err = strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("redis port must be int value: %w", err)
		}
	}
	if user := os.Getenv(envRedisUser); user != "" {
		cfg.RedisUser = user
	}
	if password := os.Getenv(envRedisPass); password != "" {
		cfg.RedisPassword = password
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	logrus.Infof("Config loaded: HTTP Host=%s, HTTP Port=%d", cfg.ServiceHost, cfg.ServicePort)
	logrus.Infof("HTTPS Config: Address=%s, Cert=%s, Key=%s", cfg.HTTPSAddress, cfg.HTTPSCertFile, cfg.HTTPSKeyFile)
	logrus.Infof("MinIO Config: endpoint=%s, access_key=%s, bucket=%s",
		cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOBucket)

	logrus.Info("config parsed successfully")

	return cfg, nil
}

// GenerateSelfSignedCert создает правильные самоподписанные сертификаты
func (c *Config) GenerateSelfSignedCert() error {
	// Создаем папку для сертификатов если её нет
	certDir := "certs"
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Генерируем приватный ключ
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096) // Увеличиваем до 4096 бит
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Создаем шаблон сертификата с правильными настройками
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"FuelCalc Development"},
			Country:       []string{"US"},
			Province:      []string{"California"},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
			CommonName:    "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 год
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost", "127.0.0.1", "192.168.1.173", c.ServiceHost},
		IPAddresses: []net.IP{
			net.IPv4(127, 0, 0, 1),
			net.IPv4(192, 168, 1, 173),
			net.IPv6loopback,
		},
	}

	// Создаем самоподписанный сертификат
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Сохраняем сертификат в PEM формате
	certOut, err := os.Create(c.HTTPSCertFile)
	if err != nil {
		return fmt.Errorf("failed to open cert file for writing: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Сохраняем приватный ключ в PEM формате
	keyOut, err := os.Create(c.HTTPSKeyFile)
	if err != nil {
		return fmt.Errorf("failed to open key file for writing: %w", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	logrus.Info("Self-signed SSL certificates generated successfully with proper configuration")
	return nil
}

// ... остальные методы без изменений
func (c *Config) InitMinIO() (*minio.Client, error) {
	minioClient, err := minio.New(c.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinIOAccessKey, c.MinIOSecretKey, ""),
		Secure: c.MinIOUseSSL,
	})
	if err != nil {
		return nil, err
	}

	// Проверяем существование бакета
	exists, err := minioClient.BucketExists(context.Background(), c.MinIOBucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("bucket %s does not exist", c.MinIOBucket)
	}

	return minioClient, nil
}

func (c *Config) GetJWTConfig() struct {
	SecretKey string
	ExpiresIn time.Duration
	Issuer    string
} {
	return struct {
		SecretKey string
		ExpiresIn time.Duration
		Issuer    string
	}{
		SecretKey: c.JWTSecretKey,
		ExpiresIn: c.JWTExpiresIn,
		Issuer:    c.JWTIssuer,
	}
}

func (c *Config) GetRedisConfig() struct {
	Host        string
	Port        int
	Password    string
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
} {
	return struct {
		Host        string
		Port        int
		Password    string
		User        string
		DialTimeout time.Duration
		ReadTimeout time.Duration
	}{
		Host:        c.RedisHost,
		Port:        c.RedisPort,
		Password:    c.RedisPassword,
		User:        c.RedisUser,
		DialTimeout: c.RedisDialTimeout,
		ReadTimeout: c.RedisReadTimeout,
	}
}
