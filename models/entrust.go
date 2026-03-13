package models

import "time"

// CommunityEntrust 对应 community_entrust 表
type CommunityEntrust struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"not null;index" json:"user_id"`
	Title         string    `gorm:"type:varchar(255);not null" json:"title"`
	Content       string    `gorm:"type:text" json:"content"`
	CreditCoin    int       `gorm:"not null" json:"credit_coin"`
	CreateTime    time.Time `gorm:"not null;autoCreateTime" json:"create_time"`
	IsProgressing bool      `gorm:"not null;default:false" json:"is_progressing"`
	IsOver        bool      `gorm:"not null;default:false" json:"is_over"`

	Images   []CommunityEntrustImage `gorm:"foreignKey:EntrustID" json:"images,omitempty"`
	Comments []EntrustComment        `gorm:"foreignKey:EntrustID" json:"comments,omitempty"`
}

func (CommunityEntrust) TableName() string {
	return "community_entrust"
}
