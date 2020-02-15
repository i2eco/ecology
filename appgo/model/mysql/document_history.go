package mysql

import (
	"time"
)

type DocumentHistory struct {
	HistoryId    int       `gorm:"not null;primary_key;AUTO_INCREMENT"json:"historyId"`
	Action       string    `gorm:"not null;"json:"action"`
	ActionName   string    `gorm:"not null;"json:"actionName"`
	DocumentId   int       `gorm:"not null;"json:"documentId"`
	DocumentName string    `gorm:"not null;"json:"documentName"`
	ParentId     int       `gorm:"not null;"json:"parentId"`
	MemberId     int       `gorm:"not null;"json:"memberId"`
	ModifyTime   time.Time `gorm:"not null;"json:"modifyTime"`
	ModifyAt     int       `gorm:"not null;"json:"modifyAt"`
	Version      int64     `gorm:"not null;"json:"version"`
}

type DocumentHistorySimpleResult struct {
	HistoryId  int       `json:"history_id"`
	ActionName string    `json:"action_name"`
	MemberId   int       `json:"member_id"`
	Account    string    `json:"account"`
	Nickname   string    `json:"nickname"`
	ModifyAt   int       `json:"modify_at"`
	ModifyName string    `json:"modify_name"`
	ModifyTime time.Time `json:"modify_time"`
	Version    int64     `json:"version"`
}

// TableName 获取对应数据库表名.
func (m *DocumentHistory) TableName() string {
	return "document_history"
}

func NewDocumentHistory() *DocumentHistory {
	return &DocumentHistory{}
}
