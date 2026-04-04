package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "todo-backend/docs"

	"todo-backend/internal/tasks"
	"todo-backend/internal/users"
	"todo-backend/internal/notifications"
)

// @title AirFlow
// @version 1.0
// @description AirFlow API documentation

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name X-User-ID
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	r := gin.New()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		// SECURITY: Измените это на конкретный домен для production!
		AllowOrigins:     []string{getEnv("ALLOWED_ORIGIN", "http://localhost:3000")}, // TODO: Установить на ["https://yourdomain.com"]
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	db, err := openDB()
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&users.User{}, &tasks.Task{}, &tasks.AILog{}, &notifications.Notification{}); err != nil {
		log.Fatalf("failed to migrate db: %v", err)
	}

	userRepo := users.NewGormRepository(db)
	taskRepo := tasks.NewGormRepository(db)

	userService := users.NewService(userRepo)
	taskService := tasks.NewService(taskRepo, userService)

	users.NewHandler(userService).RegisterRoutes(r)
	tasks.NewHandler(taskService).RegisterRoutes(r)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "8080")
	log.Printf("server running on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func openDB() (*gorm.DB, error) {
	dsn := os.Getenv("DB_DSN")
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
