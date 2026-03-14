package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sp_backend/enums"
	"sp_backend/repository"
	"sp_backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AvatarHandler struct {
	userRepo     *repository.UserRepository
	jwtConfig    *utils.JWTConfig
	avatarConfig AvatarConfig
}

type AvatarConfig struct {
	avatarDir       string
	maxUploadSize   int64
	allowedMimeType map[string]bool
}

func NewAvatarHandler(userRepo *repository.UserRepository, jwtConfig *utils.JWTConfig) *AvatarHandler {
	return &AvatarHandler{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
		avatarConfig: AvatarConfig{
			avatarDir:     "./uploads/avatars",
			maxUploadSize: 5 * 1024 * 1024,
			allowedMimeType: map[string]bool{
				"image/jpeg": true,
				"image/png":  true,
				"image/webp": true,
			},
		},
	}
}

// UploadAvatarRequest 头像上传请求
type UploadAvatarRequest struct {
	// 头像文件
	// @in formData
	// @type file
	// @required
	Avatar string `form:"avatar" swaggerignore:"true"`
}

// UploadAvatarResponse 上传响应
type UploadAvatarResponse struct {
	Success bool       `json:"success" example:"true"`
	Message string     `json:"message" example:"头像上传成功"`
	Data    AvatarData `json:"data"`
}

// AvatarData
type AvatarData struct {
	Url string `json:"url" example:"/avatars/xxxx.png"`
}

// UploadAvatar 上传用户头像
// @Summary      上传用户头像
// @Description  上传用户头像
// @Tags         头像
// @Accept       multipart/form-data
// @Produce      json
// @Param        request  body      UploadAvatarRequest  true  "上传用户自己头像表单"
// @Success      200      {object}  UploadAvatarResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/uploads/avatar [post]
func (h *AvatarHandler) UploadAvatar(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.avatarConfig.maxUploadSize)

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
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

	fmt.Println(token)
	claims, err := h.jwtConfig.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		c.Abort()
		return
	}

	if !utils.ValidateFileType(file, h.avatarConfig.allowedMimeType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only allowed 'jpg, png, gif, webp' format"})
		c.Abort()
		return
	}

	if file.Size > h.avatarConfig.maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size cannot bigger than 5MB"})
		c.Abort()
		return
	}

	if err := os.MkdirAll(h.avatarConfig.avatarDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(h.avatarConfig.avatarDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	avatarURL := fmt.Sprintf("/files/avatar/%s", filename)

	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}

	if user.Avatar != "" && user.Avatar != "/avatar/default_avatar.avif" {
		oldFilename := strings.TrimPrefix(user.Avatar, "/avatar/")
		oldFilePath := filepath.Join(h.avatarConfig.avatarDir, oldFilename)

		if strings.HasPrefix(oldFilePath, h.avatarConfig.avatarDir) {
			if err := os.Remove(oldFilePath); err != nil {
				fmt.Println("delete old avatar failed",
					"user_id", claims.UserID,
					"old_path", oldFilePath,
					"err", err)
			}
		}
	}

	err = h.userRepo.UpdateFields(claims.UserID, map[string]interface{}{
		"avatar": avatarURL,
	})
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, UploadAvatarResponse{
		Success: true,
		Message: "upload avatar successfully",
		Data: AvatarData{
			Url: avatarURL,
		},
	})
}

// UploadOtherAvatarRequest 头像上传请求
type UploadOtherAvatarRequest struct {
	// 头像文件
	// @in formData
	// @type file
	// @required
	Avatar string `form:"avatar" swaggerignore:"true"`

	// 用户ID
	// @required
	UserID uint `form:"user_id" example:"1"`
}

// UploadOtherAvatarResponse 上传响应
type UploadOtherAvatarResponse struct {
	Success bool       `json:"success" example:"true"`
	Message string     `json:"message" example:"头像上传成功"`
	Data    AvatarData `json:"data"`
}

// UploadOtherAvatar 修改别的用户的头像
// @Summary      上传用户头像
// @Description  上传用户头像
// @Tags         头像
// @Accept       multipart/form-data
// @Produce      json
// @Param        request  body      UploadOtherAvatarRequest  true  "上传其他用户头像表单"
// @Success      200      {object}  UploadOtherAvatarResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/uploads/avatar-other [post]
func (h *AvatarHandler) UploadOtherAvatar(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.avatarConfig.maxUploadSize)

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
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
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		c.Abort()
		return
	}

	if adminUser.Permission != enums.AdminPermission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "permission denied"})
		c.Abort()
		return
	}

	userIDStr := c.PostForm("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户 ID 格式错误"})
		return
	}
	targetUser, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}

	if targetUser.Username == adminUser.Username {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot edit self avatar through this method"})
		c.Abort()
		return
	}

	if !utils.ValidateFileType(file, h.avatarConfig.allowedMimeType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only allowed 'jpg, png, gif, webp' format"})
		c.Abort()
		return
	}

	if file.Size > h.avatarConfig.maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size cannot bigger than 5MB"})
		c.Abort()
		return
	}

	if err := os.MkdirAll(h.avatarConfig.avatarDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(h.avatarConfig.avatarDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	avatarURL := fmt.Sprintf("/files/avatar/%s", filename)

	if targetUser.Avatar != "" && targetUser.Avatar != "/avatar/default_avatar.avif" {
		oldFilename := strings.TrimPrefix(targetUser.Avatar, "/avatar/")
		oldFilePath := filepath.Join(h.avatarConfig.avatarDir, oldFilename)

		if strings.HasPrefix(oldFilePath, h.avatarConfig.avatarDir) {
			if err := os.Remove(oldFilePath); err != nil {
				fmt.Println("delete old avatar failed",
					"user_id", claims.UserID,
					"old_path", oldFilePath,
					"err", err)
			}
		}
	}

	err = h.userRepo.UpdateFields(targetUser.ID, map[string]interface{}{
		"avatar": avatarURL,
	})
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, UploadAvatarResponse{
		Success: true,
		Message: "upload avatar successfully",
		Data: AvatarData{
			Url: avatarURL,
		},
	})
}

// AvatarsHandler 安全的头像访问路由
// @Summary 获取用户头像
// @Description 通过文件名访问头像图片，禁止路径遍历和目录列表
// @Tags 头像
// @Produce image/png,image/jpeg,image/gif,image/webp
// @Param filename path string true "头像文件名"
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /app/files/avatar/{filename} [get]
func (h *AvatarHandler) AvatarsHandler(c *gin.Context) {
	filename := c.Param("filename")

	// 1. 基础验证：文件名不能为空
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "",
		})
		return
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unallowed file name",
		})
		return
	}

	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".avif": true,
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only allowed to get image files",
		})
		return
	}

	safePath := filepath.Join(h.avatarConfig.avatarDir, filename)

	cleanPath, err := filepath.Abs(safePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server error",
		})
		return
	}

	baseDir, err := filepath.Abs(h.avatarConfig.avatarDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server error",
		})
		return
	}

	if !strings.HasPrefix(cleanPath, baseDir) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unallowed path",
		})
		return
	}

	fileInfo, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "file not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server error",
		})
		return
	}
	if fileInfo.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "dist not allowed",
		})
		return
	}

	contentType := utils.GetContentType(ext)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename=\""+filename+"\"")
	c.Header("Cache-Control", "public, max-age=31536000") // 缓存 1 年

	c.File(cleanPath)
}
