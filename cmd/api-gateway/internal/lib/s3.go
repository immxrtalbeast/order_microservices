package lib

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func UploadToSupabase(tempPath, fileName, contentType string) (string, error) {
	accessKeyID := firstNonEmpty(os.Getenv("SUPABASE_S3_ACCESS_KEY_ID"), os.Getenv("SUPABASE_ANON_KEY"))
	secretAccessKey := firstNonEmpty(os.Getenv("SUPABASE_S3_SECRET_ACCESS_KEY"), os.Getenv("SUPABASE_SECRET_KEY"))
	endpoint := strings.TrimSpace(os.Getenv("SUPABASE_STORAGE_ENDPOINT"))
	bucket := strings.TrimSpace(os.Getenv("SUPABASE_BUCKET_NAME"))
	region := firstNonEmpty(os.Getenv("SUPABASE_S3_REGION"), os.Getenv("SUPABASE_REGION"), "eu-north-1")
	if accessKeyID == "" || secretAccessKey == "" || endpoint == "" || bucket == "" {
		return "", fmt.Errorf("missing Supabase S3 configuration")
	}

	credsProvider := credentials.NewStaticCredentialsProvider(
		accessKeyID,
		secretAccessKey,
		"",
	)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credsProvider),
		config.WithRegion(region),
	)
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	uploader := manager.NewUploader(client)

	file, err := os.Open(tempPath)
	if err != nil {
		return "", fmt.Errorf("unable to open file %s: %w", tempPath, err)
	}
	defer file.Close()

	objectKey := fmt.Sprintf("product-images/%s%s", uuid.NewString(), imageExtension(fileName, contentType))

	_, err = uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	publicURL := fmt.Sprintf("%s/object/public/%s/%s",
		strings.TrimSuffix(endpoint, "/s3"),
		bucket,
		objectKey,
	)

	return publicURL, nil
}

func imageExtension(fileName, contentType string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return ext
	}
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
