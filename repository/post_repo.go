package repository

import (
	"sp_backend/models"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *models.CommunityPost) error {
	return r.db.Create(post).Error
}

func (r *PostRepository) GetByID(id uint64) (*models.CommunityPost, error) {
	var p models.CommunityPost
	err := r.db.Where("id = ?", id).First(&p).Error
	return &p, err
}

func (r *PostRepository) GetWithImages(id uint64) (*models.CommunityPost, error) {
	var p models.CommunityPost
	err := r.db.Where("id = ?", id).
		Preload("Images").
		First(&p).Error
	return &p, err
}

func (r *PostRepository) Update(p *models.CommunityPost) error {
	return r.db.Model(p).Updates(p).Error
}

func (r *PostRepository) UpdateFields(id uint64, updates map[string]interface{}) error {
	return r.db.Model(&models.CommunityPost{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PostRepository) Delete(id uint64) error {
	return r.db.Delete(&models.CommunityPost{}, id).Error
}

func (r *PostRepository) List(page, pageSize int) ([]models.CommunityPost, int64, error) {
	var list []models.CommunityPost
	var total int64

	db := r.db.Model(&models.CommunityPost{})
	db.Count(&total)

	err := db.Order("create_time desc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

// ListByUser 查询某用户的帖子
func (r *PostRepository) ListByUser(userID uint64, page, pageSize int) ([]models.CommunityPost, int64, error) {
	var list []models.CommunityPost
	var total int64

	db := r.db.Model(&models.CommunityPost{}).Where("user_id = ?", userID)
	db.Count(&total)

	err := db.Order("create_time desc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *PostRepository) ListPostsWithPreload(page, pageSize int) ([]models.CommunityPost, int64, error) {
	var posts []models.CommunityPost
	var total int64

	r.db.Model(&models.CommunityPost{}).Count(&total)

	err := r.db.
		Preload("Images").
		Preload("Comments").
		Order("create_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&posts).Error

	return posts, total, err
}

func (r *PostRepository) ListByUserWithPreload(userID uint64, page, pageSize int) ([]models.CommunityPost, int64, error) {
	var posts []models.CommunityPost
	var total int64

	db := r.db.Model(&models.CommunityPost{}).Where("user_id = ?", userID)
	db.Count(&total)

	err := db.
		Preload("Images").
		Preload("Comments").
		Order("create_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&posts).Error

	return posts, total, err
}
