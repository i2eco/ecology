//数据库模型.
package mysql

import (
	"time"
)

// Attachment struct .
type Attachment struct {
	AttachmentId int       `gorm:"not null;primary_key;AUTO_INCREMENT"json:"attachmentId"`
	BookId       int       `gorm:"not null;"json:"bookId"`
	DocumentId   int       `gorm:"not null;"json:"documentId"`
	FileName     string    `gorm:"not null;"json:"fileName"`
	FilePath     string    `gorm:"not null;"json:"filePath"`
	FileSize     float64   `gorm:"not null;"json:"fileSize"`
	HttpPath     string    `gorm:"not null;"json:"httpPath"`
	FileExt      string    `gorm:"not null;"json:"fileExt"`
	CreateTime   time.Time `gorm:""json:"createTime"`
	CreateAt     int       `gorm:"not null;"json:"createAt"`
}

// TableName 获取对应数据库表名.
func (m *Attachment) TableName() string {
	return "attachment"
}

func NewAttachment() *Attachment {
	return &Attachment{}
}
