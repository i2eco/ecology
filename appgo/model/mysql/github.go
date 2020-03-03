package mysql

import (
	"time"
)

//GitHub 用户数据
//用户使用GitHub登录的时候，直接根据GitHub的id获取数据
type GithubUser struct {
	Id        int       `gorm:"not null;primary_key"json:"id"` //用户id
	MemberId  int       `gorm:"not null;"json:"memberId"`      //绑定的用户id
	UpdatedAt time.Time `gorm:"not null;"json:"updatedAt"`     //用户资料更新时间
	AvatarURL string    `gorm:"not null;"json:"avatarUrl"`     //用户头像链接
	Email     string    `gorm:"not null;"json:"email"`         //电子邮箱
	Login     string    `gorm:"not null;"json:"login"`         //用户名
	Name      string    `gorm:"not null;"json:"name"`          //昵称
	HtmlURL   string    `gorm:"not null;"json:"htmlUrl"`       //github主页
}

// TableName 获取对应数据库表名.
func (m GithubUser) TableName() string {
	return "github_user"
}
