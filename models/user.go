package models

import (
	"sp_backend/enums"
	"time"

	"gorm.io/gorm"
)

// AppUser 对应 app_user 表
type AppUser struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Username    string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"username"`
	Email       string         `gorm:"type:varchar(255);index" json:"email"`
	Avatar      string         `gorm:"type:varchar(500)" json:"avatar"`
	Password    string         `gorm:"type:varchar(255);not null" json:"-"` // 不返回密码
	CreditCoin  int64          `gorm:"not null;default:0" json:"credit_coin"`
	CreditScore int            `gorm:"not null;default:100" json:"credit_score"`
	Gender      enums.Gender   `gorm:"type:int;not null;default:0" json:"gender"` // 0=unknown,1=male,2=female
	Birth       *time.Time     `gorm:"type:datetime" json:"birth"`
	NickName    string         `gorm:"type:varchar(100)" json:"nick_name"`
	Signature   string         `gorm:"type:varchar(255)" json:"signature"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"` // 软删除
}

// TableName 指定表名
func (AppUser) TableName() string {
	return "app_user"
}
