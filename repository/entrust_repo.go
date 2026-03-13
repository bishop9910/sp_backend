package repository

import (
	"sp_backend/models"

	"gorm.io/gorm"
)

type EntrustRepository struct {
	db *gorm.DB
}

func NewEntrustRepository(db *gorm.DB) *EntrustRepository {
	return &EntrustRepository{db: db}
}

func (r *EntrustRepository) Create(entrust *models.CommunityEntrust) error {
	return r.db.Create(entrust).Error
}

func (r *EntrustRepository) GetByID(id uint64) (*models.CommunityEntrust, error) {
	var e models.CommunityEntrust
	err := r.db.Where("id = ?", id).First(&e).Error
	return &e, err
}

// GetWithImages 预加载附图
func (r *EntrustRepository) GetWithImages(id uint64) (*models.CommunityEntrust, error) {
	var e models.CommunityEntrust
	err := r.db.Where("id = ?", id).
		Preload("Images").
		First(&e).Error
	return &e, err
}

func (r *EntrustRepository) Update(e *models.CommunityEntrust) error {
	return r.db.Model(e).Updates(e).Error
}

func (r *EntrustRepository) Delete(id uint64) error {
	return r.db.Delete(&models.CommunityEntrust{}, id).Error
}

// ListByUser 查询某用户发布的委托
func (r *EntrustRepository) ListByUser(userID uint64, page, pageSize int) ([]models.CommunityEntrust, int64, error) {
	var list []models.CommunityEntrust
	var total int64

	db := r.db.Model(&models.CommunityEntrust{}).Where("user_id = ?", userID)
	db.Count(&total)

	err := db.Order("create_time desc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

// UpdateStatus 更新委托状态
func (r *EntrustRepository) UpdateStatus(id uint64, progressing, over bool) error {
	return r.db.Model(&models.CommunityEntrust{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_progressing": progressing,
			"is_over":        over,
		}).Error
}
