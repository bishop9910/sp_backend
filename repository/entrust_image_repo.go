package repository

import (
	"sp_backend/models"

	"gorm.io/gorm"
)

type EntrustImageRepository struct {
	db *gorm.DB
}

func NewEntrustImageRepository(db *gorm.DB) *EntrustImageRepository {
	return &EntrustImageRepository{db: db}
}

func (r *EntrustImageRepository) Create(img *models.CommunityEntrustImage) error {
	return r.db.Create(img).Error
}

// CreateBatch 批量添加附图
func (r *EntrustImageRepository) CreateBatch(images []*models.CommunityEntrustImage) error {
	return r.db.CreateInBatches(images, len(images)).Error
}

func (r *EntrustImageRepository) GetByID(id uint64) (*models.CommunityEntrustImage, error) {
	var img models.CommunityEntrustImage
	err := r.db.Where("id = ?", id).First(&img).Error
	return &img, err
}

func (r *EntrustImageRepository) ListByEntrustID(entrustID uint64) ([]models.CommunityEntrustImage, error) {
	var imgs []models.CommunityEntrustImage
	err := r.db.Where("entrust_id = ?", entrustID).Find(&imgs).Error
	return imgs, err
}

func (r *EntrustImageRepository) Delete(id uint64) error {
	return r.db.Delete(&models.CommunityEntrustImage{}, id).Error
}

// DeleteByEntrustID 删除某委托的所有附图
func (r *EntrustImageRepository) DeleteByEntrustID(entrustID uint64) error {
	return r.db.Where("entrust_id = ?", entrustID).Delete(&models.CommunityEntrustImage{}).Error
}
