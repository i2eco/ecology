package mysql

import (
	"time"
)

// Document struct.
type Document struct {
	DocumentId   int           `gorm:"not null;primary_key;AUTO_INCREMENT"json:"documentId"`
	DocumentName string        `gorm:"not null;"json:"documentName"`
	Identify     string        `gorm:"not null;"json:"identify"` // Identify 文档唯一标识
	BookId       int           `gorm:"not null;"json:"bookId"`
	ParentId     int           `gorm:"not null;"json:"parentId"`
	OrderSort    int           `gorm:"not null;"json:"orderSort"`
	Release      string        `gorm:"not null;type:longtext"json:"release"` // Release 发布后的Html格式内容.
	CreateTime   time.Time     `gorm:"not null;"json:"createTime"`
	MemberId     int           `gorm:"not null;"json:"memberId"`
	ModifyTime   time.Time     `gorm:"not null;"json:"modifyTime"`
	ModifyAt     int           `gorm:"not null;"json:"modifyAt"`
	Version      int64         `gorm:"not null;"json:"version"`
	AttachList   []*Attachment `gorm:"-"json:"attachList"`
	Vcnt         int           `gorm:"not null;"json:"vcnt"` //文档项目被浏览次数
	Markdown     string        `gorm:"-"json:"markdown"`
}

// TableName 获取对应数据库表名.
func (m Document) TableName() string {
	return "document"
}

func NewDocument() *Document {
	return &Document{
		Version: time.Now().Unix(),
	}
}
