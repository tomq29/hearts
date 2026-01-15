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
	}

	// Set bucket policy to public read
	policy := `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::` + cfg.BucketName + `/*"],"Sid": ""}]}`
	err = minioClient.SetBucketPolicy(ctx, cfg.BucketName, policy)
	if err != nil {
		// Log warning but don't fail, in case of permission issues or non-MinIO S3
		fmt.Printf("Warning: failed to set bucket policy: %v\n", err)
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
	// SANITY CHECK: If the fileName already looks like a URL, don't double-wrap it.
	// This happens because the frontend sends back the full URL in update requests,
	// and if the backend saves it blindly, we end up recursively wrapping URLs.
	if len(fileName) > 4 && fileName[:4] == "http" {
		// It's already a URL. Just return it, or ensure localhost.
		u, err := url.Parse(fileName)
		if err == nil {
			if u.Hostname() == "minio" || u.Hostname() == "minioadmin" {
				u.Host = "localhost:9000"
			}
			return u.String(), nil
		}
		// If parse fails, proceed to treat it as a key (unlikely but safe fallback)
	}

	// Generates a URL. Since the bucket is public, we don't strictly need a presigned URL with signature,
	// but we use the SDK to generate the correct object path and then strip the signature.
	u, err := p.client.PresignedGetObject(ctx, p.bucketName, fileName, time.Hour*24, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate url: %w", err)
	}

	// Hack for Docker: The API sees "minio:9000", but the browser needs "localhost:9000"
	if u.Hostname() == "minio" || u.Hostname() == "minioadmin" {
		u.Host = "localhost:9000"
	}

	// Since we made the bucket public, we can strip the query parameters (authentication signature).
	// This avoids 403 errors caused by Host header mismatch (internal 'minio' vs external 'localhost').
	u.RawQuery = ""

	return u.String(), nil
}
