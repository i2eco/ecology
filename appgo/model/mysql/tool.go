package mysql

import "time"

type Tool struct {
	Id        int        `gorm:"not null;"json:"id"`
	Name      string     `gorm:"not null;"json:"name"`
	Desc      string     `gorm:"not null;"json:"desc"`
	Identify  string     `gorm:"not null;UNIQUE_INDEX"json:"identify"`
	Cover     string     `gorm:"not null;"json:"cover"`
	CreatedAt time.Time  `gorm:""json:"createdAt"`      // 创建时间
	UpdatedAt time.Time  `gorm:""json:"updatedAt"`      // 更新时间
	DeletedAt *time.Time `gorm:"index"json:"deletedAt"` // 删除时间
	Uid       int        `gorm:"not null;"json:"uid"`
}
