package mysql

// 分类
type Category struct {
	Id     int    `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"` //自增主键
	Pid    int    `gorm:"not null;"json:"pid"`                          //分类id
	Title  string `gorm:"not null;"json:"title"`                        //分类名称
	Intro  string `gorm:"not null;"json:"intro"`                        //介绍
	Icon   string `gorm:"not null;"json:"icon"`                         //分类icon
	Cnt    int    `gorm:"not null;"json:"cnt"`                          //分类下的文档项目统计
	Sort   int    `gorm:"not null;"json:"sort"`                         //排序
	Status bool   `gorm:"not null;"json:"status"`                       //分类状态，true表示显示，否则表示隐藏
	//PrintBookCount int    `orm:"default(0)" json:"print_book_count"`
	//WikiCount      int    `orm:"default(0)" json:"wiki_count"`
	//ArticleCount   int    `orm:"default(0)" json:"article_count"`
}

func (m Category) TableName() string {
	return "category"
}

func NewCategory() *Category {
	return &Category{}
}
