package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"sp_backend/models"
	"sp_backend/repository"
	"sp_backend/utils"
)

type AuthHandler struct {
	userRepo  *repository.UserRepository
	jwtConfig *utils.JWTConfig
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtConfig *utils.JWTConfig) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool              `json:"success" example:"true"`
	Message string            `json:"message" example:"ok"`
	Data    LoginResponseData `json:"data"`
}

type LoginResponseData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // (多少秒token过期)
}

// Login 用户登录
// @Summary      用户登录
// @Description  使用用户名和密码获取访问令牌
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "登录信息"
// @Success      200      {object}  LoginResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 1. 查询用户
	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// 2. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// 3. 生成 Token
	accessToken, refreshToken, err := h.jwtConfig.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	data := LoginResponseData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(h.jwtConfig.AccessTokenExp.Seconds()),
	}

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "ok",
		Data:    data,
	})
}

// RefreshRequest 刷新令牌请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResponse 刷新令牌响应
type RefreshResponse struct {
	Success bool                `json:"success" example:"true"`
	Message string              `json:"message" example:"ok"`
	Data    RefreshResponseData `json:"data"`
}

type RefreshResponseData struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// Refresh 用户token刷新
// @Summary      用户token刷新
// @Description  刷新过时token
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request  body      RefreshRequest  true  "刷新信息"
// @Success      201      {object}  RefreshResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Router       /app/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	newAccessToken, err := h.jwtConfig.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	data := RefreshResponseData{
		AccessToken: newAccessToken,
		ExpiresIn:   int64(h.jwtConfig.AccessTokenExp.Seconds()),
	}

	c.JSON(http.StatusOK, RefreshResponse{
		Success: true,
		Message: "ok",
		Data:    data,
	})
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Success bool                 `json:"success" example:"true"`
	Message string               `json:"message" example:"ok"`
	Data    RegisterResponseData `json:"data"`
}

type RegisterResponseData struct {
	UserID       uint64 `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register 用户注册
// @Summary      用户注册
// @Description  创建新用户账号
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "注册信息"
// @Success      201      {object}  RegisterResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Router       /app/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 检查用户名是否已存在
	_, err := h.userRepo.GetByUsername(req.Username)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	// 2. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password hashing failed"})
		return
	}

	// 3. 创建用户
	newUser := &models.AppUser{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := h.userRepo.Create(newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}

	// 4. 生成 Token
	accessToken, refreshToken, _ := h.jwtConfig.GenerateToken(newUser.ID, newUser.Username)

	data := RegisterResponseData{
		UserID:       newUser.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		Success: true,
		Message: "registration successful",
		Data:    data,
	})
}
