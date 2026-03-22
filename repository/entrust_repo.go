package repository

import (
	"errors"
	"fmt"
	"sp_backend/enums"
	"sp_backend/models"

	"gorm.io/gorm"
)

type EntrustRepository struct {
	db *gorm.DB
}

// EntrustStatus 表示委托的受理状态
type EntrustStatus struct {
	UserID        uint64                 //发布委托的人
	CreditLevel   enums.CreditScoreLevel //信用需求等级
	IsAccepted    bool                   // 是否已被受理
	AcceptorID    *uint64                // 受理人 ID（nil 表示无人受理）
	IsOver        bool                   // 委托是否已结束
	IsProgressing bool                   // 是否正在进行中
}

func NewEntrustRepository(db *gorm.DB) *EntrustRepository {
	return &EntrustRepository{db: db}
}

func (r *EntrustRepository) Create(entrust *models.CommunityEntrust) error {
	return r.db.Create(entrust).Error
}

func (r *EntrustRepository) GetByID(id uint64) (*models.CommunityEntrust, error) {
	var e models.CommunityEntrust
	err := r.db.Preload("Images").Where("id = ?", id).First(&e).Error
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

func (r *EntrustRepository) ListEntrustsWithPreload(page, pageSize int) ([]models.CommunityEntrust, int64, error) {
	var entrusts []models.CommunityEntrust
	var total int64

	r.db.Model(&models.CommunityEntrust{}).Count(&total)

	err := r.db.
		Preload("Images").
		Order("create_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&entrusts).Error

	return entrusts, total, err
}

func (r *EntrustRepository) ListByUserWithPreload(userID uint64, page, pageSize int) ([]models.CommunityEntrust, int64, error) {
	var entrusts []models.CommunityEntrust
	var total int64

	db := r.db.Model(&models.CommunityEntrust{}).Where("user_id = ?", userID)
	db.Count(&total)

	err := db.
		Preload("Images").
		Order("create_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&entrusts).Error

	return entrusts, total, err
}

// CheckEntrustAcceptStatus 检查委托的受理状态
// 返回：状态信息 + 错误
func (r *EntrustRepository) CheckEntrustAcceptStatus(entrustID uint64) (*EntrustStatus, error) {
	var entrust models.CommunityEntrust

	err := r.db.
		Select("acceptor_id", "is_over", "is_progressing").
		First(&entrust, entrustID).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("委托不存在")
		}
		return nil, fmt.Errorf("查询委托失败：%w", err)
	}

	return &EntrustStatus{
		UserID:        entrust.UserID,
		CreditLevel:   entrust.AllowedCreditScoreLevel,
		IsAccepted:    entrust.AcceptorID != nil,
		AcceptorID:    entrust.AcceptorID,
		IsOver:        entrust.IsOver,
		IsProgressing: entrust.IsProgressing,
	}, nil
}

// IsEntrustAvailable 检查委托是否可被受理（未被接受且未结束）
// 返回：是否可用 + 错误
func (r *EntrustRepository) IsEntrustAvailable(entrustID uint64) (bool, error) {
	status, err := r.CheckEntrustAcceptStatus(entrustID)
	if err != nil {
		return false, err
	}

	// 可受理条件：未被接受 + 未结束
	return !status.IsAccepted && !status.IsOver, nil
}

// TryAcceptEntrust 尝试受理委托（原子操作，防止并发抢单）
// 返回：是否成功 + 错误
func (r *EntrustRepository) TryAcceptEntrust(entrustID, userID uint64) (bool, error) {
	result := r.db.
		Model(&models.CommunityEntrust{}).
		Where("id = ? AND acceptor_id IS NULL AND is_over = ?", entrustID, false).
		Updates(map[string]interface{}{
			"acceptor_id":    userID,
			"is_progressing": true,
			"updated_at":     gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		return false, fmt.Errorf("受理委托失败：%w", result.Error)
	}

	// RowsAffected == 0 表示条件不满足（已被受理或已结束）
	return result.RowsAffected > 0, nil
}

// GetAcceptedEntrusts 获取某用户已受理的委托列表
func (r *EntrustRepository) GetAcceptedEntrusts(userID uint64, page, pageSize int) ([]models.CommunityEntrust, int64, error) {
	var entrusts []models.CommunityEntrust

	// 查询列表
	err := r.db.
		Preload("Images").
		Where("acceptor_id = ?", userID).
		Order("create_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&entrusts).
		Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询已受理委托失败：%w", err)
	}

	// 查询总数
	var total int64
	err = r.db.
		Model(&models.CommunityEntrust{}).
		Where("acceptor_id = ?", userID).
		Count(&total).
		Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询总数失败：%w", err)
	}

	return entrusts, total, nil
}

func (r *EntrustRepository) CompleteEntrust(entrustID, userID uint64) (bool, error) {
	result := r.db.
		Model(&models.CommunityEntrust{}).
		// 核心条件：
		// 1. ID 匹配
		// 2. 必须是当前的受理人 (acceptor_id = userID)
		// 3. 必须处于进行中状态 (is_progressing = true)
		// 4. 必须未结束 (is_over = false)
		Where("id = ? AND acceptor_id = ? AND is_progressing = ? AND is_over = ?",
			entrustID, userID, true, false).
		Updates(map[string]interface{}{
			"is_progressing": false, // 停止进行中
			"is_over":        true,  // 标记为已结束
			"updated_at":     gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		return false, fmt.Errorf("完成委托失败：%w", result.Error)
	}

	// 如果影响行数为 0，说明条件不满足
	// 可能原因：
	// 1. 委托已结束
	// 2. 操作者不是受理人
	// 3. 委托不在进行中状态
	if result.RowsAffected == 0 {
		return false, fmt.Errorf("无法完成委托：委托可能已结束、未开始或非本人受理")
	}

	return true, nil
}
