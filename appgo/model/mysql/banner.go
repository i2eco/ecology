package mysql

import (
	"time"
)

type Banner struct {
	Id        int       `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"`
	Type      string    `gorm:"not null;"json:"type"`
	Title     string    `gorm:"not null;"json:"title"`
	Link      string    `gorm:"not null;"json:"link"`
	Image     string    `gorm:"not null;"json:"image"`
	Sort      int       `gorm:"not null;"json:"sort"`
	Status    bool      `gorm:"not null;"json:"status"`
	CreatedAt time.Time `gorm:"not null;"json:"createdAt"`
}

// TableName 获取对应数据库表名.
func (m *Banner) TableName() string {
	return "banner"
}

func NewBanner() *Banner {
	return &Banner{}
}
