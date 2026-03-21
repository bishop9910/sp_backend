package models

import (
	"sp_backend/enums"

	"gorm.io/gorm"
)

// PostComment 对应 post_comments 表
type PostComment struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64 `gorm:"not null;index" json:"user_id"`
	PostID    uint64 `gorm:"not null;index" json:"post_id"`
	Content   string `gorm:"type:text;not null" json:"content"`
	LikeCount uint64 `gorm:"not null;default:0" json:"like_count"`
}

func (c *PostComment) TableName() string {
	return "post_comments"
}

// 实现 LikeTarget 接口
func (c *PostComment) GetTargetType() enums.TargetType {
	return enums.TargetTypePostComment
}

func (c *PostComment) GetTargetID() uint64 {
	return c.ID
}

func (c *PostComment) GetTableName() string {
	return "post_comments"
}

// 实现 LikeCountUpdater 接口（如果需要更新计数）
func (c *PostComment) IncrementLikeCount(tx *gorm.DB) error {
	return tx.Model(c).Where("id = ?", c.ID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

func (c *PostComment) DecrementLikeCount(tx *gorm.DB) error {
	// 逻辑：如果 like_count > 0 则减 1，否则保持 0
	return tx.Model(c).Where("id = ?", c.ID).
		UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).Error
}
