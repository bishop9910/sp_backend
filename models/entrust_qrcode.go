package models

import "time"

// CommunityEntrustQRCode 对应 community_entrust_qr_code 表
type CommunityEntrustQRCode struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	EntrustID uint64 `gorm:"not null;index" json:"entrust_id"`
	Token     string `gorm:"type:varchar(500);not null" json:"token"`
	QRCodeURL string `gorm:"type:varchar(500);not null" json:"qr_code_url" swaggerignore:"true"`
	IsUsed    bool   `gorm:"not null;default:false" json:"is_used"`

	CreateTime time.Time `gorm:"autoCreateTime" json:"create_time"`
}

func (CommunityEntrustQRCode) TableName() string {
	return "community_entrust_qr_code"
}
