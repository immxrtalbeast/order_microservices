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
	credsProvider := credentials.NewStaticCredentialsProvider(
		os.Getenv("SUPABASE_ANON_KEY"),
		os.Getenv("SUPABASE_SECRET_KEY"),
		"",
	)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credsProvider),
		config.WithRegion("eu-north-1"),
	)
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("SUPABASE_STORAGE_ENDPOINT"))
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
		Bucket:      aws.String(os.Getenv("SUPABASE_BUCKET_NAME")),
		Key:         aws.String(hashedString),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	publicURL := fmt.Sprintf("%s/object/public/%s/%s",
		strings.TrimSuffix(os.Getenv("SUPABASE_STORAGE_ENDPOINT"), "/s3"),
		os.Getenv("SUPABASE_BUCKET_NAME"),
		hashedString,
	)

	return publicURL, nil
}
