package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Provider defines the interface for file storage operations.
type Provider interface {
	Upload(ctx context.Context, file io.Reader, fileSize int64, contentType string, fileName string) (string, error)
	GetPresignedURL(ctx context.Context, fileName string) (string, error)
}

// Config holds the configuration for the storage provider.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	Location        string
}

type minioProvider struct {
	client     *minio.Client
	bucketName string
}

// NewMinioProvider creates a new instance of a MinIO storage provider.
func NewMinioProvider(cfg Config) (Provider, error) {
	// Initialize minio client object.
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	provider := &minioProvider{
		client:     minioClient,
		bucketName: cfg.BucketName,
	}

	// Ensure bucket exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{Region: cfg.Location})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		// Set bucket policy to public read (optional, depending on requirements)
		// For now, we'll stick to presigned URLs or direct access if public.
		// policy := `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::` + cfg.BucketName + `/*"],"Sid": ""}]}`
		// err = minioClient.SetBucketPolicy(ctx, cfg.BucketName, policy)
		// if err != nil {
		// 	 return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		// }
	}

	return provider, nil
}

func (p *minioProvider) Upload(ctx context.Context, file io.Reader, fileSize int64, contentType string, fileName string) (string, error) {
	info, err := p.client.PutObject(ctx, p.bucketName, fileName, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the object name or a full URL depending on needs.
	// For now, returning the fileName (key) is often best, and we generate URLs on retrieval.
	// But to make it easy for the frontend, we can return a presigned URL immediately or the key.
	// Let's return the key for storage in DB.
	return info.Key, nil
}

func (p *minioProvider) GetPresignedURL(ctx context.Context, fileName string) (string, error) {
	// Set request parameters for content-disposition.
	reqParams := make(url.Values)
	// reqParams.Set("response-content-disposition", "attachment; filename=\""+fileName+"\"")

	// Generates a presigned url which expires in a day.
	presignedURL, err := p.client.PresignedGetObject(ctx, p.bucketName, fileName, time.Hour*24, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned url: %w", err)
	}
	return presignedURL.String(), nil
}
