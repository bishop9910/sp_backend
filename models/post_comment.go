package models

// PostComment 对应 post_comments 表
type PostComment struct {
	ID      uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID  uint64 `gorm:"not null;index" json:"user_id"`
	PostID  uint64 `gorm:"not null;index" json:"post_id"`
	Content string `gorm:"type:text;not null" json:"content"`
	Like    uint64 `gorm:"not null;default:0" json:"like"`
}

func (PostComment) TableName() string {
	return "post_comments"
}
