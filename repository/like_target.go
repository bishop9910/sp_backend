// repository/like_target.go

package repository

import (
	"sp_backend/enums"

	"gorm.io/gorm"
)

// LikeTarget 所有支持点赞的对象都需要实现这个接口
type LikeTarget interface {
	GetTargetType() enums.TargetType // 返回目标类型枚举
	GetTargetID() uint64             // 返回目标 ID
	GetTableName() string            // 返回对应的表名（用于更新计数）
}

// LikeCountUpdater 支持更新点赞计数的接口（可选，如果不需要计数可忽略）
type LikeCountUpdater interface {
	LikeTarget
	IncrementLikeCount(tx *gorm.DB) error // 点赞数 +1
	DecrementLikeCount(tx *gorm.DB) error // 点赞数 -1
}
