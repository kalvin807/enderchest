package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	s3Bucket = "enderchest"
	testUser = "kalvin807"
)

var (
	once     sync.Once
	s3Client *s3.Client
)

func GetObjectPath(user string, filename string) string {
	return fmt.Sprintf("%s/%s", user, filename)
}

func GetPresignedURL(ctx context.Context, objectKey string, client *s3.Client) string {
	presignClient := s3.NewPresignClient(client)

	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		logger.Error("Failed to get form file", zap.Error(err))
		return ""
	}

	logger.Info("Presigned URL", zap.String("URL", presignResult.URL))
	return presignResult.URL
}

func UploadImage(c *gin.Context, mongoClient *mongo.Client) {
	file, err := c.FormFile("image")
	if err != nil {
		logger.Error("Failed to get form file", zap.Error(err))
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	fileReader, _ := file.Open()

	// Upload to S3
	client := getS3Client(context.Background())

	objectKey := GetObjectPath(testUser, file.Filename)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(objectKey),
		Body:   fileReader,
	})
	if err != nil {
		logger.Error("Failed to upload file to S3", zap.Error(err))
		c.String(http.StatusInternalServerError, fmt.Sprintf("failed to upload file: %v", err))
		return
	}
	logger.Info("Successfully uploaded file to S3")

	fetchUrl := GetPresignedURL(ctx, objectKey, client)

	// Save metadata to MongoDB
	metadata := c.PostForm("metadata")
	collection := mongoClient.Database("enderchest").Collection("images")
	_, err = collection.InsertOne(ctx, bson.M{"key": objectKey, "url": fetchUrl, "metadata": metadata})
	if err != nil {
		logger.Error("Failed to save metadata to MongoDB", zap.Error(err))
		c.String(http.StatusInternalServerError, fmt.Sprintf("failed to save metadata: %v", err))
		return
	}
	logger.Info("Successfully saved metadata to MongoDB")

	c.String(http.StatusOK, "Image uploaded successfully")
}

func getS3Client(ctx context.Context) *s3.Client {
	once.Do(func() {
		accessKeyId := os.Getenv("ACCESS_KEY_ID")
		if accessKeyId == "" {
			logger.Error("ACCESS_KEY_ID environment variable is missing")
			panic(errors.New("ACCESS_KEY_ID environment variable is missing"))
		}

		accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")
		if accessKeySecret == "" {
			logger.Error("ACCESS_KEY_SECRET environment variable is missing")
			panic(errors.New("ACCESS_KEY_SECRET environment variable is missing"))
		}

		var accountId = "9468e198bc32c085d2a864d0d5627fb0"
		r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
			}, nil
		})

		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithEndpointResolverWithOptions(r2Resolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		)
		if err != nil {
			logger.Error("Failed to load AWS config", zap.Error(err))
			panic(err)
		}

		s3Client = s3.NewFromConfig(cfg)
	})

	return s3Client
}
