package mysql

import (
	"time"
)

type DocumentSearchResult struct {
	DocumentId   int       `json:"doc_id"`
	BookId       int       `json:"book_id"`
	DocumentName string    `json:"doc_name"`
	Identify     string    `json:"identify"` // Identify 文档唯一标识
	Description  string    `json:"description"`
	Author       string    `json:"author"`
	BookName     string    `json:"book_name"`
	BookIdentify string    `json:"book_identify"`
	ModifyTime   time.Time `json:"modify_time"`
	CreateTime   time.Time `json:"create_time"`
}

// 文档结果
type DocResult struct {
	DocumentId   int       `json:"doc_id"`
	DocumentName string    `json:"doc_name"`
	Identify     string    `json:"identify"` // Identify 文档唯一标识
	Release      string    `json:"release"`  // Release 发布后的Html格式内容.
	Vcnt         int       `json:"vcnt"`     //文档项目被浏览次数
	CreateTime   time.Time `json:"create_time"`
	BookId       int       `json:"book_id"`
	BookIdentify string    `json:"book_identify"`
	BookName     string    `json:"book_name"`
}
