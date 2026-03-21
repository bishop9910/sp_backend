package models

import (
	"sp_backend/enums"
)

// UserLike 点赞记录表
type UserLike struct {
	// UserID 点赞的用户 ID
	UserID uint64 `gorm:"not null;uniqueIndex:idx_user_target;comment:点赞用户 ID"`

	// TargetType 目标类型 (枚举)
	TargetType enums.TargetType `gorm:"type:varchar(20);not null;uniqueIndex:idx_user_target;comment:目标类型 (post/entrust/...)"`

	// TargetID 目标对象 ID (帖子 ID 或评论 ID 等)
	TargetID uint64 `gorm:"not null;uniqueIndex:idx_user_target;comment:目标对象 ID"`
}

func (c *UserLike) TableName() string {
	return "user_likes"
}
