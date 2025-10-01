package repository

import (
	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db          *gorm.DB
	minioClient *minio.Client
	bucketName  string
}

func New(dsn string, minioClient *minio.Client, bucketName string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:          db,
		minioClient: minioClient,
		bucketName:  bucketName,
	}, nil
}

////

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}
