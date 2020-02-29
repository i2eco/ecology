package mysql

import "time"

// SEO struct .
type Seo struct {
	Id          int        `gorm:"not null;"json:"id"`         //自增主键
	Page        string     `gorm:"not null;"json:"page"`       //页面
	Statement   string     `gorm:"not null;"json:"statement"`  //页面说明
	Title       string     `gorm:"not null;"json:"title"`      //SEO标题
	Keywords    string     `gorm:"not null;"json:"keywords"`   //SEO关键字
	Description string     `gorm:"not null"json:"description"` //SEO摘要
	CreatedAt   time.Time  `gorm:""json:"createdAt"`           // 创建时间
	UpdatedAt   time.Time  `gorm:""json:"updatedAt"`           // 更新时间
	DeletedAt   *time.Time `gorm:"index"json:"deletedAt"`      // 删除时间
}

func (Seo) TableName() string {
	return "seo"
}
