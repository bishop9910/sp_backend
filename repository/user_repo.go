package repository

import (
	"sp_backend/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建用户
func (r *UserRepository) Create(user *models.AppUser) error {
	return r.db.Create(user).Error
}

// GetByID 根据ID查询用户
func (r *UserRepository) GetByID(id uint64) (*models.AppUser, error) {
	var user models.AppUser
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名查询
func (r *UserRepository) GetByUsername(username string) (*models.AppUser, error) {
	var user models.AppUser
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户（只更新非零值字段）
func (r *UserRepository) Update(user *models.AppUser) error {
	return r.db.Model(user).Updates(user).Error
}

// UpdateFields 更新指定字段（避免零值覆盖）
func (r *UserRepository) UpdateFields(id uint64, updates map[string]interface{}) error {
	return r.db.Model(&models.AppUser{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 软删除用户
func (r *UserRepository) Delete(id uint64) error {
	return r.db.Delete(&models.AppUser{}, id).Error
}

// List 分页查询用户列表
func (r *UserRepository) List(page, pageSize int) ([]models.AppUser, int64, error) {
	var users []models.AppUser
	var total int64

	db := r.db.Model(&models.AppUser{})
	db.Count(&total)

	err := db.Limit(pageSize).Offset((page - 1) * pageSize).Find(&users).Error
	return users, total, err
}

// AddCreditCoin 增加信用金币
func (r *UserRepository) AddCreditCoin(userID uint64, amount int) error {
	return r.db.Model(&models.AppUser{}).Where("id = ?", userID).
		UpdateColumn("credit_coin", gorm.Expr("credit_coin + ?", amount)).Error
}

// DivCreditCoin 减少信用分
func (r *UserRepository) DivCreditCoin(userID uint64, amount int) error {
	return r.db.Model(&models.AppUser{}).Where("id = ?", userID).
		UpdateColumn("credit_score", gorm.Expr("credit_score - ?", amount)).Error
}
