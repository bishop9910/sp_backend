package models

import (
	"sp_backend/enums"

	"gorm.io/gorm"
)

// EntrustComment 对应 entrust_comments 表
type EntrustComment struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64 `gorm:"not null;index" json:"user_id"`
	EntrustID uint64 `gorm:"not null;index" json:"entrust_id"`
	Content   string `gorm:"type:text;not null" json:"content"`
	LikeCount uint64 `gorm:"not null;default:0" json:"like_count"`
}

func (c *EntrustComment) TableName() string {
	return "entrust_comments"
}

// 实现 LikeTarget 接口
func (c *EntrustComment) GetTargetType() enums.TargetType {
	return enums.TargetTypeEntrustComment
}

func (c *EntrustComment) GetTargetID() uint64 {
	return c.ID
}

func (c *EntrustComment) GetTableName() string {
	return "entrust_comments"
}

// 实现 LikeCountUpdater 接口（如果需要更新计数）
func (c *EntrustComment) IncrementLikeCount(tx *gorm.DB) error {
	return tx.Model(c).Where("id = ?", c.ID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

func (c *EntrustComment) DecrementLikeCount(tx *gorm.DB) error {
	// 逻辑：如果 like_count > 0 则减 1，否则保持 0
	return tx.Model(c).Where("id = ?", c.ID).
		UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).Error
}
