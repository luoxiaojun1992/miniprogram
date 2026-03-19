package entity

import "time"

// Follow represents the follows table.
type Follow struct {
	ID         uint64    `gorm:"primarykey" json:"id"`
	FollowerID uint64    `gorm:"comment:关注者ID" json:"follower_id"`
	FollowedID uint64    `gorm:"comment:被关注者ID" json:"followed_id"`
	CreatedAt  time.Time `gorm:"comment:创建时间" json:"created_at"`
}
