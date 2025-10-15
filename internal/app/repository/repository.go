package repository

import (
	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"repback/internal/app/redis"
)

type Repository struct {
	db          *gorm.DB
	minioClient *minio.Client
	bucketName  string
	RedisClient *redis.Client
}

func New(dsn string, minioClient *minio.Client, bucketName string, redisClient *redis.Client) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:          db,
		minioClient: minioClient,
		bucketName:  bucketName,
		RedisClient: redisClient,
	}, nil
}

////

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}
