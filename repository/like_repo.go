// repository/like_repository.go

package repository

import (
	"errors"
	"strings"

	"sp_backend/models"

	"gorm.io/gorm"
)

var (
	ErrAlreadyLiked = errors.New("already liked")
	ErrNotLiked     = errors.New("not liked yet")
)

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

// ==================== 公共方法（自动事务）====================

// Like 点赞（公共方法：自动开启事务）
func (r *LikeRepository) Like(userID uint64, target LikeTarget) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.doLike(tx, userID, target)
	})
}

// Unlike 取消点赞（公共方法：自动开启事务）
func (r *LikeRepository) Unlike(userID uint64, target LikeTarget) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.doUnlike(tx, userID, target)
	})
}

// ==================== 私有方法（内部逻辑）====================

// doLike 内部点赞逻辑（需要传入事务 tx）
func (r *LikeRepository) doLike(tx *gorm.DB, userID uint64, target LikeTarget) error {
	like := models.UserLike{
		UserID:     userID,
		TargetType: target.GetTargetType(),
		TargetID:   target.GetTargetID(),
	}

	if err := tx.Create(&like).Error; err != nil {
		// ✅ 修复：兼容 SQLite/MySQL 的唯一约束错误判断
		if r.isDuplicateKeyError(err) {
			return ErrAlreadyLiked
		}
		return err
	}

	// 如果目标支持计数更新，同步 +1
	if counter, ok := target.(LikeCountUpdater); ok {
		if err := counter.IncrementLikeCount(tx); err != nil {
			return err
		}
	}
	return nil
}

// doUnlike 内部取消点赞逻辑（直接用条件删除，避免主键问题）
func (r *LikeRepository) doUnlike(tx *gorm.DB, userID uint64, target LikeTarget) error {
	// 直接用 WHERE 条件删除，避免主键为 0 导致 "WHERE conditions required"
	result := tx.Where("user_id = ? AND target_type = ? AND target_id = ?",
		userID, target.GetTargetType(), target.GetTargetID()).
		Delete(&models.UserLike{})

	if result.Error != nil {
		return result.Error
	}

	// 没删除任何行 = 没赞过
	if result.RowsAffected == 0 {
		return ErrNotLiked
	}

	// 如果目标支持计数更新，同步 -1
	if counter, ok := target.(LikeCountUpdater); ok {
		if err := counter.DecrementLikeCount(tx); err != nil {
			return err
		}
	}

	return nil
}

// ==================== 辅助方法 ====================

// isDuplicateKeyError 判断是否唯一约束冲突（兼容 SQLite/MySQL/PostgreSQL）
func (r *LikeRepository) isDuplicateKeyError(err error) bool {
	// 1. 标准 GORM 错误（主要兼容 MySQL）
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	// 2. SQLite: 错误信息包含 "UNIQUE constraint failed"
	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return true
	}

	// 3. MySQL: 错误信息包含 "Duplicate entry" 或错误码 1062
	if strings.Contains(err.Error(), "Duplicate entry") {
		return true
	}

	// 4. 其他数据库可以在此扩展...

	return false
}

// ==================== 查询方法（不需要事务）====================

// IsLiked 检查是否已点赞
func (r *LikeRepository) IsLiked(userID uint64, target LikeTarget) (bool, error) {
	var count int64
	err := r.db.Model(&models.UserLike{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?",
			userID, target.GetTargetType(), target.GetTargetID()).
		Count(&count).Error
	return count > 0, err
}

// GetLikeCount 获取点赞总数（建议生产环境读冗余字段或 Redis）
func (r *LikeRepository) GetLikeCount(target LikeTarget) (int64, error) {
	var count int64
	err := r.db.Model(&models.UserLike{}).
		Where("target_type = ? AND target_id = ?",
			target.GetTargetType(), target.GetTargetID()).
		Count(&count).Error
	return count, err
}
