package mysql

import (
	"time"
)

type MemberToken struct {
	TokenId   int       `gorm:"not null;primary_key;AUTO_INCREMENT"json:"tokenId"`
	MemberId  int       `gorm:"not null;"json:"memberId"`
	Token     string    `gorm:"not null;"json:"token"`
	Email     string    `gorm:"not null;"json:"email"`
	IsValid   bool      `gorm:"not null;"json:"isValid"`
	ValidTime time.Time `gorm:""json:"validTime"`
	SendTime  time.Time `gorm:""json:"sendTime"`
}

// TableName 获取对应数据库表名.
func (m MemberToken) TableName() string {
	return "member_token"
}

func NewMemberToken() *MemberToken {
	return &MemberToken{}
}
