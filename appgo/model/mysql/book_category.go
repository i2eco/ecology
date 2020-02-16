package mysql

//文档项目与分类关联表，一个文档项目可以属于多个分类
type BookCategory struct {
	Id         int `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"`      //自增主键
	BookId     int `gorm:"not null;unique_index:unique_idx"json:"bookId"`     //书籍id
	CategoryId int `gorm:"not null;unique_index:unique_idx"json:"categoryId"` //分类id
}

// TableName 获取对应数据库表名.
func (m BookCategory) TableName() string {
	return "book_category"
}
