//go:generate goversioninfo
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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	// swaggerFiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           sp_backend API
// @version         1.3.6
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
		&models.CommunityEntrustQRCode{},
		&models.CommunityPostImage{},
		&models.PostComment{},
		&models.EntrustComment{},
		&models.UserLike{},
	)

	secretKey := os.Getenv("JWT_SECRET")

	if secretKey == "" {
		secretKey = "048311b545983cbcf3d2306c4d806568"
	}

	jwtConfig := utils.NewJWTConfig(secretKey)

	server := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"authorization",
		"istoken",
		"X-Request-ID",
	}

	server.Use(cors.New(config))

	// server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	userRepo := repository.NewUserRepository(db)

	postRepo := repository.NewPostRepository(db)
	postImageRepo := repository.NewPostImageRepository(db)
	postCommentRepo := repository.NewPostCommentRepository(db)

	entrustRepo := repository.NewEntrustRepository(db)
	entrustImageRepo := repository.NewEntrustImageRepository(db)
	entrustCommentRepo := repository.NewEntrustCommentRepository(db)
	entrustQRCodeRepo := repository.NewEntrustQRCodeRepository(db)

	likeRepo := repository.NewLikeRepository(db)

	authHandler := handlers.NewAuthHandler(userRepo, jwtConfig)
	userHandler := handlers.NewUserHandler(userRepo, jwtConfig)
	avatarHandler := handlers.NewAvatarHandler(userRepo, jwtConfig)
	postHandler := handlers.NewPostHandler(&handlers.PostHandlerConfig{
		UserRepo:        userRepo,
		PostRepo:        postRepo,
		PostImageRepo:   postImageRepo,
		PostCommentRepo: postCommentRepo,
		LikeRepo:        likeRepo,
		JwtConfig:       jwtConfig,
	})
	entrustHandler := handlers.NewEntrustHandler(&handlers.EntrustHandlerConfig{
		UserRepo:           userRepo,
		EntrustRepo:        entrustRepo,
		EntrustImageRepo:   entrustImageRepo,
		EntrustQRCodeRepo:  entrustQRCodeRepo,
		EntrustCommentRepo: entrustCommentRepo,
		LikeRepo:           likeRepo,
		JwtConfig:          jwtConfig,
	})

	frontend.WebInit(server)

	appRouter := server.Group("/app")

	authRouter := appRouter.Group("/auth")
	{
		authRouter.POST("/login", authHandler.Login)       // @Tags 认证
		authRouter.POST("/register", authHandler.Register) // @Tags 认证
		authRouter.POST("/refresh", authHandler.Refresh)   // @Tags 认证
	}

	protectedRouter := appRouter.Group("")
	protectedRouter.Use(middleware.RequiredAuth(jwtConfig))

	userRouter := protectedRouter.Group("/user")
	{
		userRouter.GET("/get-info", userHandler.GetInfo)                                       // @Tags 用户
		userRouter.POST("/edit", userHandler.Edit)                                             // @Tags 用户
		userRouter.POST("/edit-other", userHandler.EditOther)                                  // @Tags 用户
		userRouter.GET("/:user_id/posts", postHandler.GetPostByUser)                           // @Tags 帖子
		userRouter.GET("/:user_id/entrusts", entrustHandler.GetEntrustByUser)                  // @Tags 委托
		userRouter.GET("/:user_id/entrusts/accepted", entrustHandler.GetAcceptedEntrustByUser) // @Tags 委托
	}

	postRouter := protectedRouter.Group("/post")
	{
		postRouter.GET("/list", postHandler.GetPosts)                                  // @Tags 帖子
		postRouter.POST("/new", postHandler.NewPost)                                   // @Tags 帖子
		postRouter.POST("/delete", postHandler.DeletePost)                             // @Tags 帖子
		postRouter.POST("/comment", postHandler.CreatePostComment)                     // @Tags 帖子
		postRouter.GET("/comment", postHandler.GetPostComments)                        // @Tags 帖子
		postRouter.POST("/comment/delete", postHandler.DeletePostComment)              // @Tags 帖子
		postRouter.GET("/:post_id", postHandler.GetPostByID)                           // @Tags 帖子
		postRouter.POST("/like", postHandler.LikePost)                                 // @Tags 帖子
		postRouter.POST("/unlike", postHandler.UnlikePost)                             // @Tags 帖子
		postRouter.GET("/like/status", postHandler.CheckPostLikeStatus)                // @Tags 帖子
		postRouter.POST("/comment/like", postHandler.LikePostComment)                  // @Tags 帖子
		postRouter.POST("/comment/unlike", postHandler.UnlikePostComment)              // @Tags 帖子
		postRouter.GET("/comment/like/status", postHandler.CheckPostCommentLikeStatus) // @Tags 帖子
	}

	entrustRouter := protectedRouter.Group("/entrust")
	{
		entrustRouter.GET("/list", entrustHandler.GetEntrusts)                                  // @Tags 委托
		entrustRouter.POST("/new", entrustHandler.NewEnturst)                                   // @Tags 委托
		entrustRouter.POST("/delete", entrustHandler.DeleteEntrust)                             // @Tags 委托
		entrustRouter.POST("/comment", entrustHandler.CreateEntrustComment)                     // @Tags 委托
		entrustRouter.GET("/comment", entrustHandler.GetEntrustComments)                        // @Tags 委托
		entrustRouter.POST("/comment/delete", entrustHandler.DeleteEntrustComment)              // @Tags 委托
		entrustRouter.GET("/:entrust_id", entrustHandler.GetEntrustByID)                        // @Tags 委托
		entrustRouter.POST("/like", entrustHandler.LikeEntrust)                                 // @Tags 委托
		entrustRouter.POST("/unlike", entrustHandler.UnlikeEntrust)                             // @Tags 委托
		entrustRouter.GET("/like/status", entrustHandler.CheckEntrustLikeStatus)                // @Tags 委托
		entrustRouter.POST("/comment/like", entrustHandler.LikeEntrustComment)                  // @Tags 委托
		entrustRouter.POST("/comment/unlike", entrustHandler.UnlikeEntrustComment)              // @Tags 委托
		entrustRouter.GET("/comment/like/status", entrustHandler.CheckEntrustCommentLikeStatus) // @Tags 委托
		entrustRouter.POST("/accept", entrustHandler.AcceptEntrust)                             // @Tags 委托
		entrustRouter.POST("/get-qrcode", entrustHandler.GetQRCodeInfo)                         // @Tags 委托二维码
		entrustRouter.GET("/verify", entrustHandler.VerifyQRCode)                               // @Tags 委托二维码
	}

	uploadRouter := protectedRouter.Group("/uploads")
	{
		uploadRouter.POST("/avatar", avatarHandler.UploadAvatar)            // @Tags 头像
		uploadRouter.POST("/avatar-other", avatarHandler.UploadOtherAvatar) // @Tags 头像
		uploadRouter.POST("/post", postHandler.AddPostImage)                // @Tags 帖子
		uploadRouter.POST("/entrust", entrustHandler.AddEntrustImage)       // @Tags 委托
	}

	fileRouter := appRouter.Group("/files")
	{
		fileRouter.GET("/avatar/:filename", avatarHandler.AvatarsHandler)       // @Tags 头像
		fileRouter.GET("/post/:filename", postHandler.HandlePostImage)          // @Tags 帖子
		fileRouter.GET("/entrust/:filename", entrustHandler.HandleEntrustImage) // @Tags 委托
	}

	server.RunTLS(":443", "./ssl/cert.pem", "./ssl/key.key")
}
