package main

import (
	"log"
	"os"
	"sp_backend/config"
	"sp_backend/frontend"
	"sp_backend/handlers"
	"sp_backend/middleware"
	"sp_backend/models"
	"sp_backend/repository"
	"sp_backend/utils"

	_ "sp_backend/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           sp_backend API
// @version         1.0
// @description     社区平台 API 文档，包含用户、帖子、委托等功能
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  bishop9910@163.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /app

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                请输入 "Bearer <token>"，例如：Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
// @schemes  http https

func main() {
	db, err := config.InitDB("./schoolPlatform.db")
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	_ = db.AutoMigrate(
		&models.AppUser{},
		&models.CommunityEntrust{},
		&models.CommunityPost{},
		&models.CommunityEntrustImage{},
		&models.CommunityPostImage{},
		&models.PostComment{},
		&models.EntrustComment{},
	)

	secretKey := os.Getenv("JWT_SECRET")

	if secretKey == "" {
		secretKey = "048311b545983cbcf3d2306c4d806568"
	}

	jwtConfig := utils.NewJWTConfig(secretKey)

	server := gin.Default()

	frontend.WebInit(server)

	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userRepo := repository.NewUserRepository(db)
	authHandler := handlers.NewAuthHandler(userRepo, jwtConfig)
	appRouter := server.Group("/app")

	authRouter := appRouter.Group("/auth")
	{
		authRouter.POST("/login", authHandler.Login)       // @Tags 认证
		authRouter.POST("/register", authHandler.Register) // @Tags 认证
		authRouter.POST("/refresh", authHandler.Refresh)   // @Tags 认证
	}

	protectedRouter := appRouter.Group("")
	protectedRouter.Use(middleware.OptionalAuth(jwtConfig))

	server.Run(":8080")
}
