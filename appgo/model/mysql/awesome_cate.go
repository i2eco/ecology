package mysql

// 分类
type AwesomeCate struct {
	Id     int    `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"` //自增主键
	Pid    int    `gorm:"not null;"json:"pid"`                          //分类id
	Title  string `gorm:"not null;UNIQUE_INDEX"json:"title"`            //分类名称
	Intro  string `gorm:"not null;"json:"intro"`                        //介绍
	Icon   string `gorm:"not null;"json:"icon"`                         //分类icon
	Cnt    int    `gorm:"not null;"json:"cnt"`                          //分类下的文档项目统计
	Sort   int    `gorm:"not null;"json:"sort"`                         //排序
	Status bool   `gorm:"not null;"json:"status"`                       //分类状态，true表示显示，否则表示隐藏
}

func (m AwesomeCate) TableName() string {
	return "awesome_cate"
}
