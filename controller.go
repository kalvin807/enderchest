package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	s3Bucket = "enderchest"
)

func UploadImage(c *gin.Context, mongoClient *mongo.Client) {
	client := setupS3Client()
	file, err := c.FormFile("image")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	// Upload to S3
	fileReader, _ := file.Open()

	_, err = client.PutObject(c, &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(file.Filename),
		Body:   fileReader,
	})

	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("failed to upload file: %v", err))
		return
	}

	// // Save metadata to MongoDB
	// metadata := c.PostForm("metadata")
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// collection := mongoClient.Database("test").Collection("images")
	// _, err = collection.InsertOne(ctx, bson.M{"id": id, "metadata": metadata})

	// if err != nil {
	// 	c.String(http.StatusInternalServerError, fmt.Sprintf("failed to save metadata: %v", err))
	// 	return
	// }

	c.String(http.StatusOK, "Image uploaded successfully")
}

func setupS3Client() *s3.Client {
	var accountId = "9468e198bc32c085d2a864d0d5627fb0"
	var accessKeyId = os.Getenv("ACCESS_KEY_ID")
	var accessKeySecret = os.Getenv("ACCESS_KEY_SECRET")
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
	)
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg)
	return client
}
