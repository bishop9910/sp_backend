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

type PostHandler struct {
	userRepo        *repository.UserRepository
	postRepo        *repository.PostRepository
	postImageRepo   *repository.PostImageRepository
	postCommentRepo *repository.PostCommentRepository
	jwtConfig       *utils.JWTConfig
	postImageConfig *PostImageConfig
}

type PostHandlerConfig struct {
	UserRepo        *repository.UserRepository
	PostRepo        *repository.PostRepository
	PostImageRepo   *repository.PostImageRepository
	PostCommentRepo *repository.PostCommentRepository
	JwtConfig       *utils.JWTConfig
}

type PostImageConfig struct {
	imageDir        string
	maxUploadSize   int64
	allowedMimeType map[string]bool
}

func NewPostHandler(config *PostHandlerConfig) *PostHandler {
	return &PostHandler{
		userRepo:        config.UserRepo,
		postRepo:        config.PostRepo,
		postImageRepo:   config.PostImageRepo,
		postCommentRepo: config.PostCommentRepo,
		jwtConfig:       config.JwtConfig,
		postImageConfig: &PostImageConfig{
			imageDir:      "./uploads/post_images",
			maxUploadSize: 10 * 1024 * 1024,
			allowedMimeType: map[string]bool{
				"image/jpeg": true,
				"image/png":  true,
				"image/webp": true,
			},
		},
	}
}

// NewPostRequest NewPost请求体
type NewPostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// NewPostResponse NewPost响应
type NewPostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AddPostImage 发帖
// @Summary      发帖
// @Description  发帖（只能文字，图片有单独上传api，到时候拿文件列表遍历访问那个api）
// @Tags         帖子
// @Accept       json
// @Produce      json
// @Param        request  body      NewPostRequest  true  "上传用户自己头像表单"
// @Success      200      {object}  NewPostResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/post/new [post]
func (h *PostHandler) NewPost(c *gin.Context) {
	var req NewPostRequest
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

	if req.Title == "" {
		req.Title = "未命名标题"
	}
	if len(req.Content) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "帖子内容不能为空",
		})
		return
	}
	if len(req.Content) > 65535 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "帖子内容过长",
		})
		return
	}

	newPost := &models.CommunityPost{
		UserID:  claims.UserID,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := h.postRepo.Create(newPost); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "server error",
		})
		return
	}

	c.JSON(http.StatusCreated, NewPostResponse{
		Success: true,
		Message: "create post successfully",
	})

}

// AddPostImageRequest 上传图片请求
type AddPostImageRequest struct {
	// 图像文件
	// @in formData
	// @type file
	// @required
	Image string `form:"image"`

	// 帖子ID
	PostID uint64 `form:"post_id"`
}

// AddPostImageResponse 上传图片响应
type AddPostImageResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"上传成功"`
}

// AddPostImage 给帖子添加图片
// @Summary      添加图片
// @Description  拿图片文件列表遍历访问我 注意！！那个image是string类型是错的应该为file文件
// @Tags         帖子
// @Accept       multipart/form-data
// @Produce      json
// @Param        request  body      AddPostImageRequest  true  "上传用户自己头像表单"
// @Success      200      {object}  AddPostImageResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/uploads/post [post]
func (h *PostHandler) AddPostImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.postImageConfig.maxUploadSize)

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
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}

	postIDStr := c.PostForm("post_id")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil || postID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown ID format"})
		c.Abort()
		return
	}

	post, err := h.postRepo.GetByID(postID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "post not found"})
		c.Abort()
		return
	}

	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		c.Abort()
		return
	}

	if post.UserID != user.ID && user.Permission != enums.AdminPermission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot edit other's post"})
		c.Abort()
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	if !utils.ValidateFileType(file, h.postImageConfig.allowedMimeType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only allowed 'jpg, png, gif, webp' format"})
		c.Abort()
		return
	}

	if file.Size > h.postImageConfig.maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size cannot bigger than 5MB"})
		c.Abort()
		return
	}

	if err := os.MkdirAll(h.postImageConfig.imageDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(h.postImageConfig.imageDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	postImageURL := fmt.Sprintf("/files/post/%s", filename)

	newPostImage := models.CommunityPostImage{
		PostID:   post.ID,
		ImageURL: postImageURL,
	}

	err = h.postImageRepo.Create(&newPostImage)
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

// func (h *PostHandler) Comment(c *gin.Context) {

// }

// GetPostsRequest 获取帖子列表请求
type GetPostsRequest struct {
	Page     uint16 `json:"page" form:"page"`           // 页码，默认1
	PageSize uint16 `json:"page_size" form:"page_size"` // 每页数量，默认20
}

// GetPostsResponse 获取帖子列表响应
type GetPostsResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    []models.CommunityPost `json:"data"`
	Total   int64                  `json:"total"`
	Page    uint16                 `json:"page"`
}

// GetPosts 获取帖子列表（分页 + 预加载图片）
// @Summary      获取帖子列表
// @Description  分页获取社区帖子列表，默认按创建时间倒序
// @Tags         帖子
// @Accept       json
// @Produce      json
// @Param        page     query  uint16  false  "页码"     default(1)
// @Param        page_size query uint16  false  "每页数量"  default(20)
// @Success      200      {object}  GetPostsResponse
// @Failure      400      {object}  ErrorResponse
// @Router       /app/post/list [get]
func (h *PostHandler) GetPosts(c *gin.Context) {
	var req GetPostsRequest
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

	posts, total, err := h.postRepo.ListPostsWithPreload(int(req.Page), int(req.PageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询帖子列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, GetPostsResponse{
		Success: true,
		Message: "success",
		Data:    posts,
		Total:   total,
		Page:    req.Page,
	})
}

// GetPostByUserRequest 获取用户帖子请求
type GetPostByUserRequest struct {
	UserID   uint64 `json:"user_id"`
	Page     uint16 `json:"page" form:"page"`
	PageSize uint16 `json:"page_size" form:"page_size"`
}

// GetPostByUserResponse 获取用户帖子响应
type GetPostByUserResponse = GetPostsResponse

// GetPostByUser 获取指定用户的帖子列表
// @Summary      获取用户帖子
// @Description  分页获取指定用户发布的帖子列表
// @Tags         帖子
// @Accept       json
// @Produce      json
// @Param        user_id  path   uint64  true  "用户ID"
// @Param        page     query  uint16  false  "页码"     default(1)
// @Param        page_size query uint16  false  "每页数量"  default(20)
// @Success      200      {object}  GetPostByUserResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Router       /app/users/{user_id}/posts [get]
func (h *PostHandler) GetPostByUser(c *gin.Context) {
	var req GetPostByUserRequest
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

	posts, total, err := h.postRepo.ListByUserWithPreload(req.UserID, int(req.Page), int(req.PageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询用户帖子失败",
		})
		return
	}

	c.JSON(http.StatusOK, GetPostByUserResponse{
		Success: true,
		Message: "success",
		Data:    posts,
		Total:   total,
		Page:    req.Page,
	})
}

type GetPostByIDResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Data    models.CommunityPost `json:"data"`
}

// GetPostByUser 获取指定ID的帖子
// @Summary      获取帖子
// @Description  给ID拿帖子
// @Tags         帖子
// @Accept       json
// @Produce      json
// @Param        post_id  path   uint64  true  "帖子ID"
// @Success      200      {object}  GetPostByIDResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Router       /app/post/{post_id} [get]
func (h *PostHandler) GetPostByID(c *gin.Context) {
	var PostID uint64
	if postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64); err == nil {
		PostID = postID
	}

	post, err := h.postRepo.GetByID(PostID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "查询帖子失败",
		})
		return
	}

	c.JSON(http.StatusOK, GetPostByIDResponse{
		Success: true,
		Message: "ok",
		Data:    *post,
	})

}

// DeletePostRequest 帖子删除请求
type DeletePostRequest struct {
	PostID uint64 `json:"post_id"`
}

// DeletePostResponse 帖子删除响应
type DeletePostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeletePost 删帖
// @Summary      删帖
// @Description  删帖
// @Tags         帖子
// @Accept       json
// @Produce      json
// @Param        request  body      DeletePostRequest  true  "帖子删除请求"
// @Success      200      {object}  DeletePostResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/post/delete [post]
func (h *PostHandler) DeletePost(c *gin.Context) {
	var req DeletePostRequest

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

	post, err := h.postRepo.GetByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("%v", err)})
		c.Abort()
		return
	}

	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		c.Abort()
		return
	}

	if post.UserID != user.ID && user.Permission != enums.AdminPermission {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete other's post"})
		c.Abort()
		return
	}

	for i := range post.Images {
		parts := strings.Split(post.Images[i].ImageURL, "/")
		filename := parts[len(parts)-1]
		err := utils.DeleteImageFile(filename, h.postImageConfig.imageDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "fail to delete images",
			})
		}
		h.postImageRepo.Delete(post.Images[i].ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "fail to delete images",
			})
		}
	}

	err = h.postRepo.Delete(post.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "fail to delete post",
		})
	}

	c.JSON(http.StatusOK, DeletePostResponse{
		Success: true,
		Message: "delete successfully",
	})
}

// HandlePostImage 安全的帖子图片访问路由
// @Summary 获取帖子图片
// @Description 通过文件名访问帖子图片，禁止路径遍历和目录列表
// @Tags 帖子
// @Produce image/png,image/jpeg,image/gif,image/webp
// @Param filename path string true "帖子文件名"
// @Success 200 {file} binary
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /app/files/post/{filename} [get]
func (h *PostHandler) HandlePostImage(c *gin.Context) {
	filename := c.Param("filename")
	cleanPath, ext, err := utils.ValidateAndResolveImagePath(filename, h.postImageConfig.imageDir)

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

// CreateCommentRequest 创建评论请求体
type CreateCommentRequest struct {
	PostID  uint64 `json:"post_id" binding:"required"` // 帖子ID
	Content string `json:"content" binding:"required"` // 评论内容
}

// CreateCommentResponse 创建评论响应
type CreateCommentResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"评论成功"`
}

// CreateComment 创建评论
// @Summary      创建评论
// @Description  为指定帖子创建一条新评论，需要用户登录认证
// @Tags         帖子
// @Accept       json
// @Produce      json
// @Param        request  body      CreateCommentRequest  true  "评论创建请求"
// @Success      201      {object}  CreateCommentResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Router       /app/post/comment [post]
func (h *PostHandler) CreateComment(c *gin.Context) {
	var req CreateCommentRequest
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

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		return
	}
	token := authHeader[len(bearerPrefix):]
	claims, err := h.jwtConfig.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		return
	}

	post, err := h.postRepo.GetByID(req.PostID)
	if err != nil || post.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}

	newComment := &models.PostComment{
		UserID:  claims.UserID,
		PostID:  req.PostID,
		Content: req.Content,
		Like:    0,
	}
	if err := h.postCommentRepo.Create(newComment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, CreateCommentResponse{
		Success: true,
		Message: "comment created successfully",
	})
}

// GetCommentsRequest 获取评论列表请求
type GetCommentsRequest struct {
	PostID   uint64 `json:"post_id" form:"post_id" binding:"required"` // 帖子ID
	Page     uint16 `json:"page" form:"page"`                          // 页码，默认1
	PageSize uint16 `json:"page_size" form:"page_size"`                // 每页数量，默认20
}

// GetCommentsResponse 获取评论列表响应
type GetCommentsResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message"`
	Data    []models.PostComment `json:"data"`
	Total   int64                `json:"total"`
	Page    uint16               `json:"page"`
}

// GetComments 获取帖子评论列表
// @Summary      获取评论列表
// @Description  分页获取指定帖子的评论列表，按创建时间倒序
// @Tags         帖子
// @Accept       json
// @Produce      json
// @Param        post_id  query  uint64  true   "帖子ID"
// @Param        page     query  uint16  false  "页码"     default(1)
// @Param        page_size query uint16  false  "每页数量"  default(20)
// @Success      200      {object}  GetCommentsResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Router       /app/post/comment [get]
func (h *PostHandler) GetComments(c *gin.Context) {
	var req GetCommentsRequest
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

	_, err := h.postRepo.GetByID(req.PostID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "post not found",
		})
		return
	}

	comments, total, err := h.postCommentRepo.ListByPostID(req.PostID, int(req.Page), int(req.PageSize))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "failed to fetch comments",
		})
		return
	}

	c.JSON(http.StatusOK, GetCommentsResponse{
		Success: true,
		Message: "success",
		Data:    comments,
		Total:   total,
		Page:    req.Page,
	})
}

// DeleteCommentRequest 删除评论请求
type DeleteCommentRequest struct {
	CommentID uint64 `json:"comment_id" binding:"required"` // 评论ID
}

// DeleteCommentResponse 删除评论响应
type DeleteCommentResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message" example:"删除成功"`
}

// DeleteComment 删除评论
// @Summary      删除评论
// @Description  删除指定评论，仅评论作者或管理员可操作
// @Tags         帖子
// @Accept       json
// @Produce      json
// @Param        request  body      DeleteCommentRequest  true  "评论删除请求"
// @Success      200      {object}  DeleteCommentResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      403      {object}  ErrorResponse
// @Router       /app/post/comment/delete [post]
func (h *PostHandler) DeleteComment(c *gin.Context) {
	var req DeleteCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		return
	}
	token := authHeader[len(bearerPrefix):]
	claims, err := h.jwtConfig.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		return
	}

	comment, err := h.postCommentRepo.GetByID(req.CommentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization"})
		return
	}

	if comment.UserID != user.ID && user.Permission != enums.AdminPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	if err := h.postCommentRepo.Delete(req.CommentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete comment"})
		return
	}

	c.JSON(http.StatusOK, DeleteCommentResponse{
		Success: true,
		Message: "comment deleted successfully",
	})
}
