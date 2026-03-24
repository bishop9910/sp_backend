package repository

import (
	"sp_backend/models"

	"gorm.io/gorm"
)

type EntrustQRCodeRepository struct {
	db *gorm.DB
}

func NewEntrustQRCodeRepository(db *gorm.DB) *EntrustQRCodeRepository {
	return &EntrustQRCodeRepository{db: db}
}

func (r *EntrustQRCodeRepository) Create(qrcode_img *models.CommunityEntrustQRCode) error {
	return r.db.Create(qrcode_img).Error
}

func (r *EntrustQRCodeRepository) GetByID(id uint64) (*models.CommunityEntrustQRCode, error) {
	var qrcode_img models.CommunityEntrustQRCode
	err := r.db.Where("id = ?", id).First(&qrcode_img).Error
	return &qrcode_img, err
}

func (r *EntrustQRCodeRepository) GetByEntrustID(entrustID uint64) (*models.CommunityEntrustQRCode, error) {
	var qrcode_img models.CommunityEntrustQRCode
	err := r.db.Where("entrust_id = ?", entrustID).First(&qrcode_img).Error
	return &qrcode_img, err
}

func (r *EntrustQRCodeRepository) Delete(id uint64) error {
	return r.db.Delete(&models.CommunityEntrustQRCode{}, id).Error
}

// DeleteByEntrustID 删除某委托的QRCode
func (r *EntrustQRCodeRepository) DeleteByEntrustID(entrustID uint64) error {
	return r.db.Where("entrust_id = ?", entrustID).Delete(&models.CommunityEntrustQRCode{}).Error
}
