package models

import (
	"sp_backend/enums"
	"time"

	"gorm.io/gorm"
)

// CommunityPost 对应 community_post 表
type CommunityPost struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"not null;index" json:"user_id"`
	Title      string    `gorm:"type:varchar(255);not null;default:'未命名标题'" json:"title"`
	Content    string    `gorm:"type:text" json:"content"`
	CreateTime time.Time `gorm:"not null;autoCreateTime" json:"create_time"`
	LikeCount  uint64    `gorm:"not null;default:0" json:"like_count"`

	Images   []CommunityPostImage `gorm:"foreignKey:PostID" json:"images,omitempty"`
	Comments []PostComment        `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE" json:"comments,omitempty" swaggerignore:"true"`
}

func (c *CommunityPost) TableName() string {
	return "community_post"
}

// 实现 LikeTarget 接口
func (c *CommunityPost) GetTargetType() enums.TargetType {
	return enums.TargetTypePost
}

func (c *CommunityPost) GetTableName() string {
	return "community_post"
}

func (c *CommunityPost) GetTargetID() uint64 {
	return c.ID
}

// 实现 LikeCountUpdater 接口（如果需要更新计数）
func (c *CommunityPost) IncrementLikeCount(tx *gorm.DB) error {
	return tx.Model(c).Where("id = ?", c.ID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

func (c *CommunityPost) DecrementLikeCount(tx *gorm.DB) error {
	// 逻辑：如果 like_count > 0 则减 1，否则保持 0
	return tx.Model(c).Where("id = ?", c.ID).
		UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).Error
}
