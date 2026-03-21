package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sp_backend/enums"
	"sp_backend/models"
	"sp_backend/repository"
	"sp_backend/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EntrustHandler struct {
	userRepo            *repository.UserRepository
	entrustRepo         *repository.EntrustRepository
	entrustImageRepo    *repository.EntrustImageRepository
	entrustQRCodeRepo   *repository.EntrustQRCodeRepository
	entrustCommentRepo  *repository.EntrustCommentRepository
	likeRepo            *repository.LikeRepository
	jwtConfig           *utils.JWTConfig
	entrustImageConfig  *EntrustImageConfig
	entrustQRCodeConfig *EntrustQRCodeConfig
}

type EntrustHandlerConfig struct {
	UserRepo           *repository.UserRepository
	EntrustRepo        *repository.EntrustRepository
	EntrustImageRepo   *repository.EntrustImageRepository
	EntrustQRCodeRepo  *repository.EntrustQRCodeRepository
	EntrustCommentRepo *repository.EntrustCommentRepository
	LikeRepo           *repository.LikeRepository
	JwtConfig          *utils.JWTConfig
}

type EntrustImageConfig struct {
	imageDir        string
	maxUploadSize   int64
	allowedMimeType map[string]bool
}

type EntrustQRCodeConfig struct {
	codeDir string
}

func NewEntrustHandler(config *EntrustHandlerConfig) *EntrustHandler {
	return &EntrustHandler{
		userRepo:           config.UserRepo,
		entrustRepo:        config.EntrustRepo,
		entrustImageRepo:   config.EntrustImageRepo,
		entrustCommentRepo: config.EntrustCommentRepo,
		jwtConfig:          config.JwtConfig,
		likeRepo:           config.LikeRepo,
		entrustImageConfig: &EntrustImageConfig{
			imageDir:      "./uploads/entrust_images",
			maxUploadSize: 10 * 1024 * 1024,
			allowedMimeType: map[string]bool{
				"image/jpeg": true,
				"image/png":  true,
				"image/webp": true,
			},
		},
		entrustQRCodeConfig: &EntrustQRCodeConfig{
			codeDir: "./generate/qr_codes",
		},
	}
}

// NewEntrustRequest
type NewEntrustRequest struct {
	Title                   string                 `json:"title"`
	Content                 string                 `json:"content"`
	AllowedCreditScoreLevel enums.CreditScoreLevel `json:"allowed_credit_score_level"`
	CreditCoin              int                    `json:"coin"`
}

// NewEntrustResponse
type NewEntrustResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewEnturst 发布委托
// @Summary      发布委托
// @Description  发布委托（只能文字，图片有单独上传api，到时候拿文件列表遍历访问那个api）
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        request  body      NewEntrustRequest  true  "发布委托请求"
// @Success      200      {object}  NewEntrustResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/entrust/new [post]
func (h *EntrustHandler) NewEnturst(c *gin.Context) {
	var req NewEntrustRequest
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

	if req.Title == "" {
		req.Title = "未命名标题"
	}
	if len(req.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "委托内容不能为空",
		})
		return
	}
	if len(req.Content) > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "委托内容过长",
		})
		return
	}

	newEntrust := &models.CommunityEntrust{
		UserID:                  UserID,
		Title:                   req.Title,
		Content:                 req.Content,
		AllowedCreditScoreLevel: req.AllowedCreditScoreLevel,
		CreditCoin:              req.CreditCoin,
	}

	if err := h.entrustRepo.Create(newEntrust); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server error",
		})
		return
	}

	c.JSON(http.StatusCreated, NewEntrustResponse{
		Success: true,
		Message: "create entrust successfully",
	})

}

// AddEntrustImageRequest 上传图片请求
type AddEntrustImageRequest struct {
	// 图像文件
	// @in formData
	// @type file
	// @required
	Image string `form:"image"`

	// 委托ID
	EntrustID uint64 `form:"entrust_id"`
}

// AddEntrustImageResponse 上传图片响应
type AddEntrustImageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"上传成功"`
}

// AddEntrustImage 给委托添加图片
// @Summary      添加图片
// @Description  拿图片文件列表遍历访问我 注意！！那个image是string类型是错的应该为file文件
// @Tags         委托
// @Accept       multipart/form-data
// @Produce      json
// @Param        request  body      AddEntrustImageRequest  true  "上传自定义图片表单"
// @Success      200      {object}  AddEntrustImageResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/uploads/entrust [post]
func (h *EntrustHandler) AddEntrustImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.entrustImageConfig.maxUploadSize)

	_userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		c.Abort()
		return
	}

	UserID := _userID.(uint64)

	entrustIDStr := c.PostForm("entrust_id")
	entrustID, err := strconv.ParseUint(entrustIDStr, 10, 64)
	if err != nil || entrustID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown ID format"})
		c.Abort()
		return
	}

	entrust, err := h.entrustRepo.GetByID(entrustID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "entrust not found"})
		c.Abort()
		return
	}

	user, err := h.userRepo.GetByID(UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		c.Abort()
		return
	}

	if entrust.UserID != user.ID && user.Permission != enums.AdminPermission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot edit other's entrust"})
		c.Abort()
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	if !utils.ValidateFileType(file, h.entrustImageConfig.allowedMimeType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only allowed 'jpg, png, gif, webp' format"})
		c.Abort()
		return
	}

	if file.Size > h.entrustImageConfig.maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size cannot bigger than 5MB"})
		c.Abort()
		return
	}

	if err := os.MkdirAll(h.entrustImageConfig.imageDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(h.entrustImageConfig.imageDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	entrustImageURL := fmt.Sprintf("/files/entrust/%s", filename)

	newEntrustImage := models.CommunityEntrustImage{
		EntrustID: entrust.ID,
		ImageURL:  entrustImageURL,
	}

	err = h.entrustImageRepo.Create(&newEntrustImage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, AddPostImageResponse{
		Success: true,
		Message: "upload image successfully",
	})
}

// GetEntrustsRequest 获取委托列表请求
type GetEntrustsRequest struct {
	Page     uint16 `json:"page" form:"page"`           // 页码，默认1
	PageSize uint16 `json:"page_size" form:"page_size"` // 每页数量，默认20
}

// GetEntrustsResponse 获取委托列表响应
type GetEntrustsResponse struct {
	Success bool                      `json:"success"`
	Message string                    `json:"message"`
	Data    []models.CommunityEntrust `json:"data"`
	Total   int64                     `json:"total"`
	Page    uint16                    `json:"page"`
}

// GetEntrusts 获取委托列表（分页 + 预加载图片）
// @Summary      获取委托列表
// @Description  分页获取社区委托列表，默认按创建时间倒序
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        page      query  uint16  false  "页码"     default(1)
// @Param        page_size query uint16  false  "每页数量"  default(20)
// @Success      200       {object}  GetEntrustsResponse
// @Failure      400       {object}  ErrorResponse
// @Router       /app/entrust/list [get]
func (h *EntrustHandler) GetEntrusts(c *gin.Context) {
	var req GetEntrustsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 || req.PageSize > 100 {
		req.PageSize = 20 // 限制最大100，防止恶意拉取
	}

	posts, total, err := h.entrustRepo.ListEntrustsWithPreload(int(req.Page), int(req.PageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询帖子列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, GetEntrustsResponse{
		Success: true,
		Message: "success",
		Data:    posts,
		Total:   total,
		Page:    req.Page,
	})
}

// GetEntrustByUserRequest 获取用户委托请求
type GetEntrustByUserRequest struct {
	UserID   uint64 `json:"user_id"`
	Page     uint16 `json:"page" form:"page"`
	PageSize uint16 `json:"page_size" form:"page_size"`
}

// GetEntrustByUserResponse 获取用户委托响应
type GetEntrustByUserResponse = GetEntrustsResponse

// GetEntrustByUser 获取指定用户的委托列表
// @Summary      获取用户委托
// @Description  分页获取指定用户发布的委托列表
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        user_id  path   uint64  true  "用户ID"
// @Param        page     query  uint16  false  "页码"     default(1)
// @Param        page_size query uint16  false  "每页数量"  default(20)
// @Success      200      {object}  GetEntrustByUserRequest
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Router       /app/user/{user_id}/entrusts [get]
func (h *EntrustHandler) GetEntrustByUser(c *gin.Context) {
	var req GetEntrustByUserRequest
	if userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64); err == nil {
		req.UserID = userID
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	if req.UserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "user_id 不能为空",
		})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	if _, err := h.userRepo.GetByID(req.UserID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}

	entrusts, total, err := h.entrustRepo.ListByUserWithPreload(req.UserID, int(req.Page), int(req.PageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询用户委托失败",
		})
		return
	}

	c.JSON(http.StatusOK, GetEntrustByUserResponse{
		Success: true,
		Message: "success",
		Data:    entrusts,
		Total:   total,
		Page:    req.Page,
	})
}

type GetEntrustByIDResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Data    models.CommunityEntrust `json:"data"`
}

// GetEntrustByID 获取指定ID的委托
// @Summary      获取委托
// @Description  给ID拿委托
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        entrust_id  path   uint64  true  "委托ID"
// @Success      200      {object}  GetEntrustByIDResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Router       /app/entrust/{entrust_id} [get]
func (h *EntrustHandler) GetEntrustByID(c *gin.Context) {
	var EntrustID uint64
	if entrustID, err := strconv.ParseUint(c.Param("entrust_id"), 10, 64); err == nil {
		EntrustID = entrustID
	}

	entrust, err := h.entrustRepo.GetByID(EntrustID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询帖子失败",
		})
		return
	}

	c.JSON(http.StatusOK, GetEntrustByIDResponse{
		Success: true,
		Message: "ok",
		Data:    *entrust,
	})

}

// DeleteEntrustRequest 委托删除请求
type DeleteEntrustRequest struct {
	EntrustID uint64 `json:"entrust_id"`
}

// DeleteEntrustResponse 委托删除响应
type DeleteEntrustResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeleteEntrust 删委托
// @Summary      删委托
// @Description  删委托
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        request  body      DeleteEntrustRequest  true  "委托删除请求"
// @Success      200      {object}  DeleteEntrustResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/entrust/delete [post]
func (h *EntrustHandler) DeleteEntrust(c *gin.Context) {
	var req DeleteEntrustRequest

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

	entrust, err := h.entrustRepo.GetByID(req.EntrustID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}

	user, err := h.userRepo.GetByID(UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}

	if entrust.UserID != user.ID && user.Permission != enums.AdminPermission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete other's entrust"})
		c.Abort()
		return
	}

	for i := range entrust.Images {
		parts := strings.Split(entrust.Images[i].ImageURL, "/")
		filename := parts[len(parts)-1]
		err := utils.DeleteImageFile(filename, h.entrustImageConfig.imageDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "fail to delete images",
			})
			c.Abort()
			return
		}
		err = h.entrustImageRepo.Delete(entrust.Images[i].ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "fail to delete images",
			})
			c.Abort()
			return
		}
	}

	if entrust.QRCode != nil && entrust.QRCode.QRCodeURL != "" {
		qr_parts := strings.Split(entrust.QRCode.QRCodeURL, "/")
		qr_code_filename := qr_parts[len(qr_parts)-1]

		err := utils.DeleteImageFile(qr_code_filename, h.entrustQRCodeConfig.codeDir)
		if err != nil && !errors.Is(err, utils.ErrFileNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "fail to delete qr code",
			})
			c.Abort()
			return
		}
	}

	err = h.entrustRepo.Delete(entrust.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "fail to delete post",
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, DeleteEntrustResponse{
		Success: true,
		Message: "delete successfully",
	})
}

// LikeEntrustRequest 点赞请求
type LikeEntrustRequest struct {
	EntrustID uint64 `json:"entrust_id" binding:"required"` // 委托ID
}

// LikeEntrustResponse 点赞响应
type LikeEntrustResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"删除成功"`
}

// LikeEntrust 点赞委托
// @Summary      点赞委托
// @Description  点赞委托,需要登陆
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        request  body      LikeEntrustRequest  true  "点赞请求"
// @Success      200      {object}  LikeEntrustResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Router       /app/entrust/like [post]
func (h *EntrustHandler) LikeEntrust(c *gin.Context) {
	var req LikeEntrustRequest
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

	target := &models.CommunityEntrust{ID: req.EntrustID}

	likeErr := h.likeRepo.Like(UserID, target)

	if likeErr != nil {
		if errors.Is(likeErr, repository.ErrAlreadyLiked) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "已经点过赞了",
			})
			return
		}
		if errors.Is(likeErr, repository.ErrNotLiked) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "已经点过赞了",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, LikeEntrustResponse{
		Success: true,
		Message: "点赞成功",
	})
}

type UnlikeEntrustRequest = LikeEntrustRequest
type UnlikeEntrustResponse = LikeEntrustResponse

// UnlikeEntrust 取消点赞委托
// @Summary      取消点赞委托
// @Description  取消委托的点赞,需要登陆
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        request  body      UnlikeEntrustRequest  true  "取消点赞请求"
// @Success      200      {object}  UnlikeEntrustResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse      "未点赞，无法取消"
// @Router       /app/entrust/unlike [post]
func (h *EntrustHandler) UnlikeEntrust(c *gin.Context) {
	var req UnlikeEntrustRequest
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

	target := &models.CommunityEntrust{ID: req.EntrustID}

	unlikeErr := h.likeRepo.Unlike(UserID, target)

	if unlikeErr != nil {
		if errors.Is(unlikeErr, repository.ErrNotLiked) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "尚未点赞，无法取消",
			})
			return
		}
		if errors.Is(unlikeErr, repository.ErrAlreadyLiked) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "尚未点赞，无法取消",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, UnlikeEntrustResponse{
		Success: true,
		Message: "取消点赞成功",
	})
}

// HandleEntrustImage 安全的委托图片访问路由
// @Summary 获取委托图片
// @Description 通过文件名访问委托图片，禁止路径遍历和目录列表
// @Tags 委托
// @Produce image/png,image/jpeg,image/gif,image/webp
// @Param filename path string true "委托文件名"
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /app/files/entrust/{filename} [get]
func (h *EntrustHandler) HandleEntrustImage(c *gin.Context) {
	filename := c.Param("filename")
	cleanPath, ext, err := utils.ValidateAndResolveImagePath(filename, h.entrustImageConfig.imageDir)

	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := "server error"

		switch {
		case errors.Is(err, utils.ErrEmptyFilename), errors.Is(err, utils.ErrUnallowedFilename),
			errors.Is(err, utils.ErrUnallowedExt), errors.Is(err, utils.ErrUnallowedPath), errors.Is(err, utils.ErrIsDirectory):
			statusCode = http.StatusBadRequest
			errMsg = err.Error()
		case errors.Is(err, utils.ErrFileNotFound):
			statusCode = http.StatusNotFound
			errMsg = err.Error()
		}

		c.JSON(statusCode, gin.H{"error": errMsg})
		return
	}

	contentType := utils.GetContentType(ext)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename=\""+filename+"\"")
	c.Header("Cache-Control", "public, max-age=31536000") // 缓存 1 年

	c.File(cleanPath)
}

// CreateEntrustCommentRequest 创建评论请求体
type CreateEntrustCommentRequest struct {
	EntrustID uint64 `json:"entrust_id" binding:"required"` // 帖子ID
	Content   string `json:"content" binding:"required"`    // 评论内容
}

// CreateEntrustCommentResponse 创建评论响应
type CreateEntrustCommentResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"评论成功"`
}

// CreateEntrustComment 创建评论
// @Summary      创建评论
// @Description  为指定委托创建一条新评论，需要用户登录认证
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        request  body      CreateEntrustCommentRequest  true  "评论创建请求"
// @Success      201      {object}  CreateEntrustCommentResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/entrust/comment [post]
func (h *EntrustHandler) CreateEntrustComment(c *gin.Context) {
	var req CreateEntrustCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if len(req.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空"})
		return
	}
	if len(req.Content) > 2048 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容过长，最多2048字符"})
		return
	}

	_userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		c.Abort()
		return
	}

	UserID := _userID.(uint64)

	post, err := h.entrustRepo.GetByID(req.EntrustID)
	if err != nil || post.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Entrust not found"})
		return
	}

	newComment := &models.EntrustComment{
		UserID:    UserID,
		EntrustID: req.EntrustID,
		Content:   req.Content,
		LikeCount: 0,
	}
	if err := h.entrustCommentRepo.Create(newComment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, CreateEntrustCommentResponse{
		Success: true,
		Message: "comment created successfully",
	})
}

// GetEntrustCommentsRequest 获取评论列表请求
type GetEntrustCommentsRequest struct {
	EntrustID uint64 `json:"entrust_id" form:"entrust_id" binding:"required"` // 帖子ID
	Page      uint16 `json:"page" form:"page"`                                // 页码，默认1
	PageSize  uint16 `json:"page_size" form:"page_size"`                      // 每页数量，默认20
}

// GetEntrustCommentsResponse 获取评论列表响应
type GetEntrustCommentsResponse struct {
	Success bool                    `json:"success"`
	Message string                  `json:"message"`
	Data    []models.EntrustComment `json:"data"`
	Total   int64                   `json:"total"`
	Page    uint16                  `json:"page"`
}

// GetEntrustComments 获取委托评论列表
// @Summary      获取评论列表
// @Description  分页获取指定委托的评论列表，按创建时间倒序
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        entrust_id  query  uint64  true   "委托ID"
// @Param        page     query  uint16  false  "页码"     default(1)
// @Param        page_size query uint16  false  "每页数量"  default(20)
// @Success      200      {object}  GetEntrustCommentsResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Router       /app/entrust/comment [get]
func (h *EntrustHandler) GetEntrustComments(c *gin.Context) {
	var req GetEntrustCommentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	_, err := h.entrustRepo.GetByID(req.EntrustID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Entrust not found",
		})
		return
	}

	comments, total, err := h.entrustCommentRepo.ListByEntrustID(req.EntrustID, int(req.Page), int(req.PageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to fetch comments",
		})
		return
	}

	c.JSON(http.StatusOK, GetEntrustCommentsResponse{
		Success: true,
		Message: "success",
		Data:    comments,
		Total:   total,
		Page:    req.Page,
	})
}

// DeleteEntrustCommentRequest 删除评论请求
type DeleteEntrustCommentRequest struct {
	CommentID uint64 `json:"comment_id" binding:"required"` // 评论ID
}

// DeleteEntrustCommentResponse 删除评论响应
type DeleteEntrustCommentResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"删除成功"`
}

// DeleteEntrustComment 删除评论
// @Summary      删除评论
// @Description  删除指定评论，仅评论作者或管理员可操作
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        request  body      DeleteEntrustCommentRequest  true  "评论删除请求"
// @Success      200      {object}  DeleteEntrustCommentResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Router       /app/entrust/comment/delete [post]
func (h *EntrustHandler) DeleteEntrustComment(c *gin.Context) {
	var req DeleteEntrustCommentRequest
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

	comment, err := h.entrustCommentRepo.GetByID(req.CommentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	user, err := h.userRepo.GetByID(UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		return
	}

	if comment.UserID != user.ID && user.Permission != enums.AdminPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	if err := h.entrustCommentRepo.Delete(req.CommentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, DeleteEntrustCommentResponse{
		Success: true,
		Message: "comment deleted successfully",
	})
}

// LikeEntrustCommentRequest 点赞请求
type LikeEntrustCommentRequest struct {
	CommentID uint64 `json:"comment_id" binding:"required"` // 评论ID
}

// LikeEntrustCommentResponse 点赞响应
type LikeEntrustCommentResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"删除成功"`
}

// LikeEntrustComment 点赞评论
// @Summary      点赞评论
// @Description  点赞评论,需要登陆
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        request  body      LikeEntrustCommentRequest  true  "点赞请求"
// @Success      200      {object}  LikeEntrustCommentResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Router       /app/entrust/comment/like [post]
func (h *EntrustHandler) LikeEntrustComment(c *gin.Context) {
	var req LikeEntrustCommentRequest
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

	target := &models.EntrustComment{ID: req.CommentID}

	likeErr := h.likeRepo.Like(UserID, target)

	if likeErr != nil {
		if errors.Is(likeErr, repository.ErrAlreadyLiked) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "已经点过赞了",
			})
			return
		}
		if errors.Is(likeErr, repository.ErrNotLiked) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "已经点过赞了",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, LikeEntrustCommentResponse{
		Success: true,
		Message: "点赞成功",
	})
}

type UnlikeEntrustCommentRequest = LikePostCommentRequest
type UnlikeEntrustCommentResponse = LikePostCommentResponse

// UnlikeEntrustComment 取消点赞评论
// @Summary      取消点赞评论
// @Description  取消对评论的点赞,需要登陆
// @Tags         委托
// @Accept       json
// @Produce      json
// @Param        request  body      UnlikeEntrustCommentRequest  true  "取消点赞请求"
// @Success      200      {object}  UnlikeEntrustCommentResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse      "未点赞，无法取消"
// @Router       /app/entrust/comment/unlike [post]
func (h *EntrustHandler) UnlikeEntrustComment(c *gin.Context) {
	var req UnlikeEntrustCommentRequest
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

	target := &models.EntrustComment{ID: req.CommentID}

	unlikeErr := h.likeRepo.Unlike(UserID, target)

	if unlikeErr != nil {
		if errors.Is(unlikeErr, repository.ErrNotLiked) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "尚未点赞，无法取消",
			})
			return
		}
		if errors.Is(unlikeErr, repository.ErrAlreadyLiked) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "尚未点赞，无法取消",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器内部错误",
		})
		return
	}

	c.JSON(http.StatusOK, UnlikeEntrustCommentResponse{
		Success: true,
		Message: "取消点赞成功",
	})
}
