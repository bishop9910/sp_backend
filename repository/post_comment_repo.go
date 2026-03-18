package repository

import (
	"sp_backend/models"

	"gorm.io/gorm"
)

type PostCommentRepository struct {
	db *gorm.DB
}

func NewPostCommentRepository(db *gorm.DB) *PostCommentRepository {
	return &PostCommentRepository{db: db}
}

func (r *PostCommentRepository) Create(comment *models.PostComment) error {
	return r.db.Create(comment).Error
}

func (r *PostCommentRepository) ListByPostID(postID uint64, page, pageSize int) ([]models.PostComment, int64, error) {
	var list []models.PostComment
	var total int64

	db := r.db.Model(&models.PostComment{}).Where("post_id = ?", postID)
	db.Count(&total)

	err := db.Order("id desc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *PostCommentRepository) Delete(id uint64) error {
	return r.db.Delete(&models.PostComment{}, id).Error
}

func (r *PostCommentRepository) DeleteByPostID(postID uint64) error {
	return r.db.Where("post_id = ?", postID).Delete(&models.PostComment{}).Error
}

func (r *PostCommentRepository) GetByID(id uint64) (*models.PostComment, error) {
	var comment models.PostComment
	err := r.db.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
