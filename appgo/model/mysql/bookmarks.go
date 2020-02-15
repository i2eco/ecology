package mysql

import (
	"time"
)

//书签
type Bookmark struct {
	Id       int `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"`
	BookId   int `gorm:"not null;index"json:"bookId"` //书籍id，主要是为了方便根据书籍id查询书签
	Uid      int `gorm:"not null;"json:"uid"`         //用户id
	DocId    int `gorm:"not null;"json:"docId"`       //文档id
	CreateAt int `gorm:"not null;"json:"createAt"`    //创建时间
}

func (Bookmark) TableName() string {
	return "bookmark"
}

//书签列表
type bookmarkList struct {
	Id           int       `json:"id,omitempty"`
	Title        string    `json:"title"`
	Identify     string    `json:"identify"`
	BookId       int       `json:"book_id"`
	Uid          int       `json:"uid"`
	DocId        int       `json:"doc_id"`
	CreateAt     int       `json:"-"`
	CreateAtTime time.Time `json:"created_at"`
}

var tableBookmark = "md_bookmark"

// 多字段唯一键
func (m *Bookmark) TableUnique() [][]string {
	return [][]string{
		[]string{"Uid", "DocId"},
	}
}

func NewBookmark() *Bookmark {
	return &Bookmark{}
}
