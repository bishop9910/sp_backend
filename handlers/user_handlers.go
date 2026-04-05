package handlers

import (
	"fmt"
	"net/http"
	"sp_backend/config"
	"sp_backend/enums"
	"sp_backend/repository"
	"sp_backend/utils"
	"time"

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
	ID           uint64 `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	CreditCoin   int64  `json:"credit_coin"`
	CreditScore  int    `json:"credit_score"`
	Gender       string `json:"gender"`
	Permission   string `json:"permission"`
	Birth        string `json:"birth"`
	NickName     string `json:"nickname"`
	Signature    string `json:"signature"`
	IsProhibited bool   `json:"is_prohibited"`
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

	_userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		c.Abort()
		return
	}

	UserID := _userID.(uint64)

	user, err := h.userRepo.GetByID(UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
	}

	data := GetInfoResponseData{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		NickName:     user.NickName,
		Avatar:       user.Avatar,
		CreditCoin:   user.CreditCoin,
		CreditScore:  user.CreditScore,
		Gender:       user.Gender.String(),
		Permission:   user.Permission.String(),
		Birth:        utils.TimeToString(user.Birth),
		Signature:    user.Signature,
		IsProhibited: user.IsProhibited,
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

	_userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		c.Abort()
		return
	}

	UserID := _userID.(uint64)

	user, err := h.userRepo.GetByID(UserID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		c.Abort()
		return
	}

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
// @Description  除了ID和用户名和密码和头像和权限和信用分和信用金币(不在这里改)改不了其他都能改
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

	_userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		c.Abort()
		return
	}

	UserID := _userID.(uint64)

	adminUser, err := h.userRepo.GetByID(UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		c.Abort()
		return
	}
	targetUser, err := h.userRepo.GetByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		c.Abort()
		return
	}

	if adminUser.Permission != enums.AdminPermission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission denied"})
		c.Abort()
		return
	}

	if targetUser.Permission == enums.AdminPermission && !config.IsSuperAdmin(adminUser.Username) {
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

// CheckInResponse 签到响应
type CheckInResponse struct {
	Success bool                `json:"success" example:"true"`
	Message string              `json:"message"`
	Data    CheckInResponseData `json:"data,omitempty"`
}

type CheckInResponseData struct {
	IsNewCheckIn bool  `json:"is_new_check_in" example:"true"` // 是否本次为新签到
	CreditCoin   int64 `json:"credit_coin" example:"101"`      // 签到后金币数量
}

// CheckIn 用户签到
// @Summary      用户每日签到
// @Description  用户每日签到，每人每天限签一次，签到成功金币+1
// @Tags         用户
// @Accept       json
// @Produce      json
// @Success      200  {object}  CheckInResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /app/user/check-in [get]
func (h *UserHandler) CheckIn(c *gin.Context) {
	_userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		c.Abort()
		return
	}
	UserID := _userID.(uint64)

	user, err := h.userRepo.GetByID(UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		c.Abort()
		return
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if user.LastCheckInTime != nil {
		lastDate := time.Date(user.LastCheckInTime.Year(), user.LastCheckInTime.Month(),
			user.LastCheckInTime.Day(), 0, 0, 0, 0, user.LastCheckInTime.Location())
		if lastDate.Equal(today) {
			c.JSON(http.StatusOK, CheckInResponse{
				Success: true,
				Message: "already checked in today",
				Data: CheckInResponseData{
					IsNewCheckIn: false,
					CreditCoin:   user.CreditCoin,
				},
			})
			return
		}
	}

	newCoin := user.CreditCoin + 1
	err = h.userRepo.UpdateFields(user.ID, map[string]interface{}{
		"credit_coin":        newCoin,
		"last_check_in_time": time.Now(),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("update failed: %v", err)})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, CheckInResponse{
		Success: true,
		Message: "check in successfully, +1 credit coin",
		Data: CheckInResponseData{
			IsNewCheckIn: true,
			CreditCoin:   newCoin,
		},
	})
}

// BanRequest 封禁/解封请求
type BanRequest struct {
	UserID       uint64 `json:"user_id" binding:"required" example:"123"`
	IsProhibited bool   `json:"is_prohibited" binding:"required" example:"true"` // true=封禁, false=解封
	Reason       string `json:"reason,omitempty" example:"违反社区规则"`               // 封禁原因（可选，可扩展日志用）
}

// BanResponse 封禁操作响应
type BanResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"`
	Data    struct {
		UserID       uint64 `json:"user_id"`
		IsProhibited bool   `json:"is_prohibited"`
		UpdatedAt    string `json:"updated_at"`
	} `json:"data,omitempty"`
}

// BanUser 封禁/解封用户
// @Summary      管理员封禁或解封用户
// @Description  仅管理员可调用，设置目标用户的封禁状态。注意：不能封禁其他管理员（除非是超级管理员）
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        request  body      BanRequest  true  "封禁操作请求"
// @Success      200  {object}  BanResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Router       /app/user/ban [post]
func (h *UserHandler) BanUser(c *gin.Context) {
	var req BanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 1. 获取当前操作者身份
	operatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: user not found"})
		c.Abort()
		return
	}
	operatorUID := operatorID.(uint64)

	// 2. 查询操作者信息并校验管理员权限
	operator, err := h.userRepo.GetByID(operatorUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operator user not found"})
		c.Abort()
		return
	}
	if operator.Permission != enums.AdminPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied: admin required"})
		c.Abort()
		return
	}

	// 3. 查询目标用户
	targetUser, err := h.userRepo.GetByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target user not found"})
		c.Abort()
		return
	}

	// 4. 安全校验：普通管理员不能操作其他管理员（超级管理员除外）
	if targetUser.Permission == enums.AdminPermission && !config.IsSuperAdmin(operator.Username) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied: cannot ban other admins"})
		c.Abort()
		return
	}

	// 5. 不能封禁/解封自己（避免逻辑冲突，如需自锁请移除此校验）
	if targetUser.ID == operator.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot ban/unban yourself via this endpoint"})
		c.Abort()
		return
	}

	err = h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
		"is_prohibited": req.IsProhibited,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("update failed: %v", err)})
		c.Abort()
		return
	}

	// 7. 返回响应
	action := "unbanned"
	if req.IsProhibited {
		action = "banned"
	}

	c.JSON(http.StatusOK, BanResponse{
		Success: true,
		Message: fmt.Sprintf("user %d has been %s successfully", req.UserID, action),
		Data: struct {
			UserID       uint64 `json:"user_id"`
			IsProhibited bool   `json:"is_prohibited"`
			UpdatedAt    string `json:"updated_at"`
		}{
			UserID:       req.UserID,
			IsProhibited: req.IsProhibited,
			UpdatedAt:    time.Now().Format(time.RFC3339),
		},
	})
}

// CheckBanStatus 中间件：检查用户是否被封禁
func (h *UserHandler) CheckBanStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		isProhibited, err := h.userRepo.IsProhibited(userID.(uint64))
		if err != nil {
			// 用户不存在等错误，交由后续逻辑处理
			c.Next()
			return
		}

		if isProhibited {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account prohibited",
				"message": "您的账号已被封禁，请联系管理员",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdjustCreditScoreRequest 信用分调整请求
type AdjustCreditScoreRequest struct {
	UserID uint64 `json:"user_id" binding:"required" example:"123"`
	Amount int    `json:"amount" binding:"required" example:"10"` // 正数=加分，负数=减分
}

// AdjustCreditScoreResponse 信用分调整响应
type AdjustCreditScoreResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"`
	Data    struct {
		UserID    uint64 `json:"user_id"`
		OldScore  int    `json:"old_score"`
		NewScore  int    `json:"new_score"`
		ChangedBy uint64 `json:"changed_by"`
		UpdatedAt string `json:"updated_at"`
	} `json:"data,omitempty"`
}

// AdjustCreditScore 调整用户信用分
// @Summary      管理员调整用户信用分
// @Description  仅管理员可调用，amount>0 加分，amount<0 减分，信用分最低为0
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        request  body      AdjustCreditScoreRequest  true  "信用分调整请求"
// @Success      200  {object}  AdjustCreditScoreResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Router       /app/user/adjust-credit [post]
func (h *UserHandler) AdjustCreditScore(c *gin.Context) {
	var req AdjustCreditScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 1. 获取当前操作者
	operatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}
	operatorUID := operatorID.(uint64)

	// 2. 校验管理员权限
	operator, err := h.userRepo.GetByID(operatorUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operator not found"})
		c.Abort()
		return
	}
	if operator.Permission != enums.AdminPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied: admin required"})
		c.Abort()
		return
	}

	// 3. 查询目标用户
	targetUser, err := h.userRepo.GetByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target user not found"})
		c.Abort()
		return
	}

	// 4. 安全校验：不能操作其他管理员（超级管理员除外）
	if targetUser.Permission == enums.AdminPermission && !config.IsSuperAdmin(operator.Username) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied: cannot modify other admins"})
		c.Abort()
		return
	}

	// 5. 计算新信用分（最低为0）
	oldScore := targetUser.CreditScore
	newScore := oldScore + req.Amount
	if newScore < 0 {
		newScore = 0
	}
	if newScore > 100 { // 可选：设置上限
		newScore = 100
	}

	err = h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
		"credit_score": newScore,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("update failed: %v", err)})
		c.Abort()
		return
	}

	if newScore == 0 {
		err = h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
			"is_prohibited": true,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("update failed: %v", err)})
			c.Abort()
			return
		}
	}

	c.JSON(http.StatusOK, AdjustCreditScoreResponse{
		Success: true,
		Message: fmt.Sprintf("credit score adjusted: %d -> %d", oldScore, newScore),
		Data: struct {
			UserID    uint64 `json:"user_id"`
			OldScore  int    `json:"old_score"`
			NewScore  int    `json:"new_score"`
			ChangedBy uint64 `json:"changed_by"`
			UpdatedAt string `json:"updated_at"`
		}{
			UserID:    req.UserID,
			OldScore:  oldScore,
			NewScore:  newScore,
			ChangedBy: operatorUID,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	})
}

// UpdatePermissionRequest 权限修改请求
type UpdatePermissionRequest struct {
	UserID     uint64           `json:"user_id" binding:"required" example:"123"`
	Permission enums.Permission `json:"permission" binding:"required" example:"1"` // 0=普通用户, 1=管理员
}

// UpdatePermissionResponse 权限修改响应
type UpdatePermissionResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"`
	Data    struct {
		UserID    uint64 `json:"user_id"`
		OldPerm   string `json:"old_permission"`
		NewPerm   string `json:"new_permission"`
		ChangedBy string `json:"changed_by"`
		UpdatedAt string `json:"updated_at"`
	} `json:"data,omitempty"`
}

// UpdateUserPermission 修改用户权限
// @Summary      超级管理员修改用户权限
// @Description  仅超级管理员可调用，用于授予或撤销管理员权限
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        request  body      UpdatePermissionRequest  true  "权限修改请求"
// @Success      200  {object}  UpdatePermissionResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Router       /app/user/update-permission [post]
func (h *UserHandler) UpdateUserPermission(c *gin.Context) {
	var req UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 1. 获取当前操作者
	operatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}
	operatorUID := operatorID.(uint64)

	operator, err := h.userRepo.GetByID(operatorUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operator not found"})
		c.Abort()
		return
	}
	if !config.IsSuperAdmin(operator.Username) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied: super admin required"})
		c.Abort()
		return
	}

	// 3. 查询目标用户
	targetUser, err := h.userRepo.GetByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target user not found"})
		c.Abort()
		return
	}

	// 4. 安全校验：不能修改自己的权限（避免自锁）
	if targetUser.ID == operator.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot modify your own permission via this endpoint"})
		c.Abort()
		return
	}

	// 5. 记录旧权限用于响应
	oldPerm := targetUser.Permission.String()
	newPerm := req.Permission.String()

	// 6. 执行权限更新
	err = h.userRepo.UpdatePermission(targetUser.ID, req.Permission)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("update failed: %v", err)})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, UpdatePermissionResponse{
		Success: true,
		Message: fmt.Sprintf("permission updated: %s -> %s", oldPerm, newPerm),
		Data: struct {
			UserID    uint64 `json:"user_id"`
			OldPerm   string `json:"old_permission"`
			NewPerm   string `json:"new_permission"`
			ChangedBy string `json:"changed_by"`
			UpdatedAt string `json:"updated_at"`
		}{
			UserID:    req.UserID,
			OldPerm:   oldPerm,
			NewPerm:   newPerm,
			ChangedBy: operator.Username,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	})
}

// AdjustCreditCoinRequest 信用金币调整请求
type AdjustCreditCoinRequest struct {
	UserID uint64 `json:"user_id" binding:"required" example:"123"`
	Amount int    `json:"amount" binding:"required" example:"50"` // 正数=增加，负数=减少
	Reason string `json:"reason,omitempty" example:"活动奖励/违规扣除"`   // 调整原因（可选，用于日志）
}

// AdjustCreditCoinResponse 信用金币调整响应
type AdjustCreditCoinResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message"`
	Data    struct {
		UserID    uint64 `json:"user_id"`
		OldCoin   int64  `json:"old_coin"`
		NewCoin   int64  `json:"new_coin"`
		Changed   int    `json:"changed"` // 实际变动值（可能因余额不足被截断）
		Operator  uint64 `json:"operator_id"`
		UpdatedAt string `json:"updated_at"`
	} `json:"data,omitempty"`
}

// AdjustCreditCoin 管理员调整用户信用金币
// @Summary      管理员调整用户信用金币
// @Description  仅管理员可调用，amount>0 增加金币，amount<0 扣除金币
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        request  body      AdjustCreditCoinRequest  true  "信用金币调整请求"
// @Success      200  {object}  AdjustCreditCoinResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Failure      403  {object}  ErrorResponse
// @Router       /app/user/adjust-coin [post]
func (h *UserHandler) AdjustCreditCoin(c *gin.Context) {
	var req AdjustCreditCoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 1. 获取当前操作者身份
	operatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: user not found"})
		c.Abort()
		return
	}
	operatorUID := operatorID.(uint64)

	// 2. 校验管理员权限
	operator, err := h.userRepo.GetByID(operatorUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "operator not found"})
		c.Abort()
		return
	}
	if operator.Permission != enums.AdminPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied: admin required"})
		c.Abort()
		return
	}

	// 3. 查询目标用户
	targetUser, err := h.userRepo.GetByID(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target user not found"})
		c.Abort()
		return
	}

	// 4. 安全校验：普通管理员不能操作其他管理员（超级管理员除外）
	if targetUser.Permission == enums.AdminPermission && !config.IsSuperAdmin(operator.Username) {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied: cannot modify other admins"})
		c.Abort()
		return
	}

	oldCoin := targetUser.CreditCoin

	// 6. 执行原子更新
	if req.Amount > 0 {
		err = h.userRepo.AddCreditCoin(targetUser.ID, req.Amount)
	} else if req.Amount < 0 {
		err = h.userRepo.DivCreditCoin(targetUser.ID, -req.Amount)
		if err != nil {
			newCoin, err := h.userRepo.GetCreditCoin(targetUser.ID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
				c.Abort()
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("insufficient balance: current=%d, request=%d", newCoin, req.Amount),
			})
			c.Abort()
			return
		}
	} else {
		// amount == 0，无操作
		c.JSON(http.StatusOK, AdjustCreditCoinResponse{
			Success: true,
			Message: "no change: amount is 0",
			Data: struct {
				UserID    uint64 `json:"user_id"`
				OldCoin   int64  `json:"old_coin"`
				NewCoin   int64  `json:"new_coin"`
				Changed   int    `json:"changed"`
				Operator  uint64 `json:"operator_id"`
				UpdatedAt string `json:"updated_at"`
			}{
				UserID:    req.UserID,
				OldCoin:   oldCoin,
				NewCoin:   oldCoin,
				Changed:   0,
				Operator:  operatorUID,
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("update failed: %v", err)})
		c.Abort()
		return
	}

	newCoin, err := h.userRepo.GetCreditCoin(targetUser.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}

	// 9. 返回响应
	action := "adjusted"
	if req.Amount > 0 {
		action = "added"
	} else if req.Amount < 0 {
		action = "deducted"
	}

	c.JSON(http.StatusOK, AdjustCreditCoinResponse{
		Success: true,
		Message: fmt.Sprintf("credit coin %s: %d -> %d (%s %d)",
			action, oldCoin, newCoin, action, utils.Abs(req.Amount)),
		Data: struct {
			UserID    uint64 `json:"user_id"`
			OldCoin   int64  `json:"old_coin"`
			NewCoin   int64  `json:"new_coin"`
			Changed   int    `json:"changed"`
			Operator  uint64 `json:"operator_id"`
			UpdatedAt string `json:"updated_at"`
		}{
			UserID:    req.UserID,
			OldCoin:   oldCoin,
			NewCoin:   newCoin,
			Changed:   utils.Abs(req.Amount),
			Operator:  operatorUID,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	})
}
