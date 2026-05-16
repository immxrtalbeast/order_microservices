package lib

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	hash := sha256.Sum256([]byte(fileName))
	hashedString := hex.EncodeToString(hash[:])

	_, err = uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(hashedString),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	publicURL := fmt.Sprintf("%s/object/public/%s/%s",
		strings.TrimSuffix(endpoint, "/s3"),
		bucket,
		hashedString,
	)

	return publicURL, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
