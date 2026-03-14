package handlers

import (
	"fmt"
	"net/http"
	"sp_backend/enums"
	"sp_backend/repository"
	"sp_backend/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo  *repository.UserRepository
	jwtConfig *utils.JWTConfig
}

func NewUserHandler(userRepo *repository.UserRepository, jwtConfig *utils.JWTConfig) *UserHandler {
	return &UserHandler{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

// GetInfoResponse 个人信息响应
type GetInfoResponse struct {
	Success bool                `json:"success" example:"true"`
	Message string              `json:"message"`
	Data    GetInfoResponseData `json:"data"`
}

type GetInfoResponseData struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar"`
	CreditCoin  int64  `json:"credit_coin"`
	CreditScore int    `json:"credit_score"`
	Gender      string `json:"gender"`
	Permission  string `json:"permission"`
	Birth       string `json:"birth"`
	NickName    string `json:"nickname"`
	Signature   string `json:"signature"`
}

// GetInfo 获取用户信息
// @Summary      获取用户信息
// @Description  用于获取用户信息
// @Tags         用户
// @Accept       json
// @Produce      json
// @Success      200      {object}  GetInfoResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/user/get-info [get]
func (h *UserHandler) GetInfo(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		c.Abort()
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		c.Abort()
		return
	}

	token := authHeader[len(bearerPrefix):]

	claims, err := h.jwtConfig.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
	}

	user, err := h.userRepo.GetByUsername(claims.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
	}

	data := GetInfoResponseData{
		Username:    user.Username,
		Email:       user.Email,
		NickName:    user.NickName,
		Avatar:      user.Avatar,
		CreditCoin:  user.CreditCoin,
		CreditScore: user.CreditScore,
		Gender:      user.Gender.String(),
		Permission:  user.Permission.String(),
		Birth:       utils.TimeToString(user.Birth),
		Signature:   user.Signature,
	}

	c.JSON(http.StatusCreated, GetInfoResponse{
		Success: true,
		Message: "ok",
		Data:    data,
	})
}

// EditRequest 修改数据请求
type EditRequest struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// EditResponse 修改数据响应
type EditResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"`
}

// Edit 修改用户数值
// @Summary      修改用户自己的信息
// @Description  只能修改邮箱，昵称，性别(2女,1男,0未知)，生日，签名
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        request  body      EditRequest  true  "修改用户数值信息"
// @Success      200      {object}  EditResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/user/edit [post]
func (h *UserHandler) Edit(c *gin.Context) {
	var req EditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		c.Abort()
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		c.Abort()
		return
	}

	token := authHeader[len(bearerPrefix):]

	claims, err := h.jwtConfig.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		c.Abort()
		return
	}

	user, err := h.userRepo.GetByUsername(claims.Username)

	if req.Key != "email" && req.Key != "birth" && req.Key != "nickname" && req.Key != "gender" && req.Key != "signature" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key"})
		c.Abort()
		return
	}

	switch req.Key {
	case "email":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		err := h.userRepo.UpdateFields(user.ID, map[string]interface{}{
			"email": utils.AnyToStringSafe(req.Value),
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "birth":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		birthTime, err := utils.StringToTime(utils.AnyToStringSafe(req.Value))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}
		err = h.userRepo.UpdateFields(user.ID, map[string]interface{}{
			"birth": birthTime,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "nickname":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		err := h.userRepo.UpdateFields(user.ID, map[string]interface{}{
			"nick_name": utils.AnyToStringSafe(req.Value),
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "gender":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		gender := enums.GenderFromString(utils.AnyToStringSafe(req.Value))
		err := h.userRepo.UpdateFields(user.ID, map[string]interface{}{
			"gender": gender,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "signature":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		err := h.userRepo.UpdateFields(user.ID, map[string]interface{}{
			"signature": utils.AnyToStringSafe(req.Value),
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	}

	c.JSON(http.StatusOK, EditResponse{
		Success: true,
		Message: "edited successfully",
	})
}

// EditOtherRequest 修改他人数据请求
type EditOtherRequest struct {
	UserID uint64 `json:"user_id"`
	Key    string `json:"key"`
	Value  any    `json:"value"`
}

// EditOtherResponse 修改他人数据响应
type EditOtherResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"`
}

// EditOther 修改其他用户数值
// @Summary      修改别的用户的信息
// @Description  除了ID和用户名和密码和头像(不在这里改)改不了其他都能改
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        request  body      EditOtherRequest  true  "修改其他用户数值信息"
// @Success      200      {object}  EditOtherResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/user/edit-other [post]
func (h *UserHandler) EditOther(c *gin.Context) {
	var req EditOtherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		c.Abort()
		return
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		c.Abort()
		return
	}

	token := authHeader[len(bearerPrefix):]

	claims, err := h.jwtConfig.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		c.Abort()
		return
	}

	adminUser, err := h.userRepo.GetByID(claims.UserID)
	targetUser, err := h.userRepo.GetByID(req.UserID)

	if adminUser.Permission != enums.AdminPermission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission denied"})
		c.Abort()
		return
	}

	if targetUser.Permission == enums.AdminPermission && adminUser.Username != "bishop9910" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission denied"})
		c.Abort()
		return
	}

	if targetUser.Username == adminUser.Username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot edit self through this method"})
		c.Abort()
		return
	}

	switch req.Key {
	case "email":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		err := h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
			"email": utils.AnyToStringSafe(req.Value),
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "birth":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		birthTime, err := utils.StringToTime(utils.AnyToStringSafe(req.Value))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}
		err = h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
			"birth": birthTime,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "nickname":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		err := h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
			"nick_name": utils.AnyToStringSafe(req.Value),
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "gender":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		gender := enums.GenderFromString(utils.AnyToStringSafe(req.Value))
		err := h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
			"gender": gender,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "signature":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		err := h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
			"signature": utils.AnyToStringSafe(req.Value),
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "permission":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		if adminUser.Username != "bishop9910" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "permission denied"})
			c.Abort()
			return
		}
		permission := enums.PermissionFromString(utils.AnyToStringSafe(req.Value))
		err := h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
			"permission": permission,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "credit_coin":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		val := max(utils.AnyToIntSafe(req.Value), 0)
		err := h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
			"credit_coin": val,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	case "credit_score":
		if !(utils.IsType[string](req.Value)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value type"})
			c.Abort()
			return
		}
		val := max(utils.AnyToIntSafe(req.Value), 0)
		err := h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
			"credit_score": val,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
			c.Abort()
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key"})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, EditResponse{
		Success: true,
		Message: "edited successfully",
	})
}
