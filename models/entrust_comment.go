package models

// EntrustComment 对应 entrust_comments 表
type EntrustComment struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64 `gorm:"not null;index" json:"user_id"`
	EntrustID uint64 `gorm:"not null;index" json:"entrust_id"`
	Content   string `gorm:"type:text;not null" json:"content"`
}

func (EntrustComment) TableName() string {
	return "entrust_comments"
}
