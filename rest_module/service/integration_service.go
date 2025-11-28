package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	. "rest_module/model"
	. "rest_module/utils"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type IntegrationService struct {
	client        *minio.Client
	defaultBucket string
}

func NewIntegrationService() *IntegrationService {
	var endpoint = GetEnv("MINIO_ENDPOINT", "localhost:9000")
	var accessKey = GetEnv("MINIO_ACCESS_KEY", "minioadmin")
	var secretKey = GetEnv("MINIO_SECRET_KEY", "minioadmin")
	var bucket = GetEnv("MINIO_BUCKET", "users")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})

	if err != nil {
		go log.Printf("Unable minio connect %s: %v", bucket, err)
		return nil
	}

	svc := &IntegrationService{
		client:        client,
		defaultBucket: bucket,
	}

	if bucket != "" {
		if err := svc.ensureBucket(context.Background(), bucket); err != nil {
			go log.Printf("Unable to ensure bucket %s: %v", bucket, err)
		}
	}

	return svc
}

func (service *IntegrationService) GetBucket() string {
	return service.defaultBucket
}

func (service *IntegrationService) ensureBucket(ctx context.Context, bucket string) error {
	if bucket == "" {
		return fmt.Errorf("Bucket name is required")
	}

	exists, err := service.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return service.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
}

func (service *IntegrationService) bucketOrDefault(bucket string) (string, error) {
	if bucket != "" {
		return bucket, nil
	}

	if service.defaultBucket != "" {
		return service.defaultBucket, nil
	}

	return "", fmt.Errorf("Bucket name is required")
}

func (service *IntegrationService) UploadObject(ctx context.Context, bucket, objectName string, content []byte, contentType string) (*minio.UploadInfo, error) {
	targetBucket, err := service.bucketOrDefault(bucket)
	if err != nil {
		return nil, err
	}

	if objectName == "" {
		return nil, fmt.Errorf("Object name is required")
	}

	if err := service.ensureBucket(ctx, targetBucket); err != nil {
		return nil, err
	}

	reader := bytes.NewReader(content)
	info, err := service.client.PutObject(ctx, targetBucket, objectName, reader, reader.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (service *IntegrationService) PresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	targetBucket, err := service.bucketOrDefault(bucket)
	if err != nil {
		return "", err
	}

	if objectName == "" {
		return "", fmt.Errorf("Object name is required")
	}

	if expiry <= 0 {
		expiry = 15 * time.Minute
	}

	if err := service.ensureBucket(ctx, targetBucket); err != nil {
		return "", err
	}

	url, err := service.client.PresignedGetObject(ctx, targetBucket, objectName, expiry, nil)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (service *IntegrationService) ExportUserSnapshot(ctx context.Context, bucket string, user *User) (string, error) {
	if user == nil {
		return "", fmt.Errorf("User can not be null")
	}

	payload, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	objectName := fmt.Sprintf("users/user-%d-%d.json", user.ID, time.Now().UnixNano())
	if _, err := service.UploadObject(ctx, bucket, objectName, payload, "application/json"); err != nil {
		return "", err
	}

	return objectName, nil
}
