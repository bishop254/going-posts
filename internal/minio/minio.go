package minio

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(endpoint, accessKeyID, secretAccessKey string, useSsl bool) (*minio.Client, error) {

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSsl,
	})
	if err != nil {
		return nil, err
	}

	return minioClient, nil
}
