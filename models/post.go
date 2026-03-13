package models

import "time"

// CommunityPost 对应 community_post 表
type CommunityPost struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"not null;index" json:"user_id"`
	Title      string    `gorm:"type:varchar(255);not null;default:'未命名标题'" json:"title"`
	Content    string    `gorm:"type:text" json:"content"`
	CreateTime time.Time `gorm:"not null;autoCreateTime" json:"create_time"`

	Images   []CommunityPostImage `gorm:"foreignKey:PostID" json:"images,omitempty"`
	Comments []PostComment        `gorm:"foreignKey:PostID" json:"comments,omitempty"`
}

func (CommunityPost) TableName() string {
	return "community_post"
}
