package models

import (
	"sp_backend/enums"
	"time"

	"gorm.io/gorm"
)

// CommunityEntrust 对应 community_entrust 表
type CommunityEntrust struct {
	ID                      uint64                 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID                  uint64                 `gorm:"not null;index" json:"user_id"`
	AcceptorID              *uint64                `gorm:"index" json:"acceptor_id"`
	Title                   string                 `gorm:"type:varchar(255);not null" json:"title"`
	Content                 string                 `gorm:"type:text" json:"content"`
	AllowedCreditScoreLevel enums.CreditScoreLevel `gorm:"type:int;not null" json:"allowed_credit_score_level"`
	CreditCoin              int                    `gorm:"not null;" json:"credit_coin"`
	CreateTime              time.Time              `gorm:"not null;autoCreateTime" json:"create_time"`
	IsProgressing           bool                   `gorm:"not null;default:false" json:"is_progressing"`
	IsOver                  bool                   `gorm:"not null;default:false" json:"is_over"`
	Like_Count              uint64                 `gorm:"not null;default:0" json:"like_count"`

	Images   []CommunityEntrustImage `gorm:"foreignKey:EntrustID" json:"images,omitempty"`
	Comments []EntrustComment        `gorm:"foreignKey:EntrustID;constraint:OnDelete:CASCADE" json:"comments,omitempty" swaggerignore:"true"`
	QRCode   *CommunityEntrustQRCode `gorm:"foreignKey:EntrustID;" json:"qr_code,omitempty" swaggerignore:"true"`
}

func (CommunityEntrust) TableName() string {
	return "community_entrust"
}

// 实现 LikeTarget 接口
func (c *CommunityEntrust) GetTargetType() enums.TargetType {
	return enums.TargetTypeEntrust
}

func (c *CommunityEntrust) GetTableName() string {
	return "community_entrust"
}

func (c *CommunityEntrust) GetTargetID() uint64 {
	return c.ID
}

// 实现 LikeCountUpdater 接口（如果需要更新计数）
func (c *CommunityEntrust) IncrementLikeCount(tx *gorm.DB) error {
	return tx.Model(c).Where("id = ?", c.ID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

func (c *CommunityEntrust) DecrementLikeCount(tx *gorm.DB) error {
	// 逻辑：如果 like_count > 0 则减 1，否则保持 0
	return tx.Model(c).Where("id = ?", c.ID).
		UpdateColumn("like_count", gorm.Expr("CASE WHEN like_count > 0 THEN like_count - 1 ELSE 0 END")).Error
}
