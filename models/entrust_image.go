package models

// CommunityEntrustImage 对应 community_entrust_image 表
type CommunityEntrustImage struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	EntrustID uint64 `gorm:"not null;index" json:"entrust_id"`
	ImageURL  string `gorm:"type:varchar(500);not null" json:"image_url"`
}

func (CommunityEntrustImage) TableName() string {
	return "community_entrust_image"
}
