package config

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceHost string
	ServicePort int

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

	viper.SetDefault("redis_host", "localhost")
	viper.SetDefault("redis_port", 6379)
	viper.SetDefault("redis_dial_timeout", "10s")
	viper.SetDefault("redis_read_timeout", "10s")

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		ServiceHost:      viper.GetString("ServiceHost"),
		ServicePort:      viper.GetInt("ServicePort"),
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
	logrus.Infof("Config loaded: Host=%s, Port=%d", cfg.ServiceHost, cfg.ServicePort)
	logrus.Infof("MinIO Config: endpoint=%s, access_key=%s, bucket=%s",
		cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOBucket)

	logrus.Info("config parsed")
	logrus.Info("config parsed")

	return cfg, nil

}

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
