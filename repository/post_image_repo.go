package repository

import (
	"sp_backend/models"

	"gorm.io/gorm"
)

type PostImageRepository struct {
	db *gorm.DB
}

func NewPostImageRepository(db *gorm.DB) *PostImageRepository {
	return &PostImageRepository{db: db}
}

func (r *PostImageRepository) Create(img *models.CommunityPostImage) error {
	return r.db.Create(img).Error
}

func (r *PostImageRepository) CreateBatch(images []*models.CommunityPostImage) error {
	return r.db.CreateInBatches(images, len(images)).Error
}

func (r *PostImageRepository) ListByPostID(postID uint64) ([]models.CommunityPostImage, error) {
	var imgs []models.CommunityPostImage
	err := r.db.Where("post_id = ?", postID).Find(&imgs).Error
	return imgs, err
}

func (r *PostImageRepository) Delete(id uint64) error {
	return r.db.Delete(&models.CommunityPostImage{}, id).Error
}

func (r *PostImageRepository) DeleteByPostID(postID uint64) error {
	return r.db.Where("post_id = ?", postID).Delete(&models.CommunityPostImage{}).Error
}
