// @title Enderchest API
// @version 1.0
// @description This is the API server for Enderchest.
// @BasePath /api/v1
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	docs "github.com/kalvin807/enderchest/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var mongoClient *mongo.Client
var logger *zap.Logger

func setupLogger(env string) *zap.Logger {
	var logger *zap.Logger
	var err error

	if env == "prod" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(fmt.Sprintf("Can't initialize zap logger: %v", err))
	}

	return logger
}

func setupMongoClient(ctx context.Context) *mongo.Client {
	connectionString := os.Getenv("MONGODB_URI")
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(connectionString).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		panic(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("Failed to ping MongoDB", zap.Error(err))
		panic(err)
	}
	logger.Info("Successfully connected to MongoDB")
	return client
}

func setupRouter(mongoClient *mongo.Client) *gin.Engine {
	router := gin.New()
	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))
	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := router.Group("/api/v1")
	{
		v1.PUT("/image", func(c *gin.Context) {
			UploadImage(c, mongoClient)
		})
		v1.GET("/ping", Ping)
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return router
}

// @Summary Check server status
// @Description Checks if the server is alive
// @Produce json
// @Success 200 {string} string "pong"
// @Router /ping [get]
func Ping(c *gin.Context) {
	c.JSON(200, "pong")
}

func main() {
	env := os.Getenv("APP_ENV")
	logger = setupLogger(env)
	defer logger.Sync()

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient = setupMongoClient(ctx)
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		}
	}()

	router := setupRouter(mongoClient)
	port := func() string {
		if p := os.Getenv("PORT"); p != "" {
			return p
		}
		return "8080"
	}()
	router.Run(":" + port)
}
