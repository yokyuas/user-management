package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"strconv"
	"time"
	"user-management/internal/handler"
	"user-management/internal/middleware"
	mongorepo "user-management/internal/repository/mongodb"
	"user-management/internal/service"
)

func main() {
	port := env("PORT", "8080")
	uri := env("MONGODB_URI", "mongodb://localhost:27017")
	dbName := env("MONGODB_DATABASE", "user_management")
	secret := env("JWT_SECRET", "")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	repo := mongorepo.NewUserRepository(client.Database(dbName))
	if err = repo.EnsureIndexes(ctx); err != nil {
		log.Fatal(err)
	}

	ttlHours, _ := strconv.Atoi(env("JWT_EXPIRE_HOURS", "24"))
	svc := service.NewUserService(repo, secret, time.Duration(ttlHours)*time.Hour)
	
	ah := handler.NewAuthHandler(svc)
	uh := handler.NewUserHandler(svc)
	
	r := gin.New()
	r.Use(gin.Recovery(), middleware.Logger())
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/auth/register", ah.Register)
	r.POST("/auth/login", ah.Login)
	
	users := r.Group("/users", middleware.JWT(secret))
	users.POST("", ah.Register)
	users.GET("", uh.List)
	users.GET("/:id", uh.Get)
	users.PUT("/:id", uh.Update)
	users.DELETE("/:id", uh.Delete)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			c, done := context.WithTimeout(context.Background(), 5*time.Second)
			n, err := repo.Count(c)
			done()
			if err != nil {
				log.Printf("user_count_error: %v", err)
			} else {
				log.Printf("total_users=%d", n)
			}
		}
	}()
	log.Printf("API listening on :%s", port)
	if err = r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
