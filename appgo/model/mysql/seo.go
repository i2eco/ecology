package mysql

// SEO struct .
type Seo struct {
	Id          int    `gorm:"not null;"json:"id"`        //自增主键
	Page        string `gorm:"not null;"json:"page"`      //页面
	Statement   string `gorm:"not null;"json:"statement"` //页面说明
	Title       string `gorm:"not null;"json:"title"`     //SEO标题
	Keywords    string `gorm:"not null;"json:"keywords"`  //SEO关键字
	Description string `orm:"default({description})"`     //SEO摘要
}

func (Seo) TableName() string {
	return "seo"
}
