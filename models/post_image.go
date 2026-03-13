package models

// CommunityPostImage 对应 community_post_image 表
type CommunityPostImage struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID   uint64 `gorm:"not null;index" json:"post_id"`
	ImageURL string `gorm:"type:varchar(500);not null" json:"image_url"`
}

func (CommunityPostImage) TableName() string {
	return "community_post_image"
}
