package repository

import (
	"sp_backend/models"

	"gorm.io/gorm"
)

type EntrustCommentRepository struct {
	db *gorm.DB
}

func NewEntrustCommentRepository(db *gorm.DB) *EntrustCommentRepository {
	return &EntrustCommentRepository{db: db}
}

func (r *EntrustCommentRepository) Create(comment *models.EntrustComment) error {
	return r.db.Create(comment).Error
}

func (r *EntrustCommentRepository) ListByEntrustID(entrustID uint64, page, pageSize int) ([]models.EntrustComment, int64, error) {
	var list []models.EntrustComment
	var total int64

	db := r.db.Model(&models.EntrustComment{}).Where("entrust_id = ?", entrustID)
	db.Count(&total)

	err := db.Order("id desc").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *EntrustCommentRepository) Delete(id uint64) error {
	return r.db.Delete(&models.EntrustComment{}, id).Error
}

func (r *EntrustCommentRepository) DeleteByEntrustID(entrustID uint64) error {
	return r.db.Where("entrust_id = ?", entrustID).Delete(&models.EntrustComment{}).Error
}

func (r *EntrustCommentRepository) GetByID(id uint64) (*models.EntrustComment, error) {
	var comment models.EntrustComment
	err := r.db.First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
