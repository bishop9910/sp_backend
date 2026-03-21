package models

// CommunityEntrustQRCode 对应 community_entrust_qr_code 表
type CommunityEntrustQRCode struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	EntrustID uint64 `gorm:"not null;index" json:"entrust_id"`
	QRCodeURL string `gorm:"type:varchar(500);not null" json:"qr_code_url"`
}

func (CommunityEntrustQRCode) TableName() string {
	return "community_entrust_qr_code"
}
