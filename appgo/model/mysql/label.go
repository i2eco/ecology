package mysql

import (
	"github.com/goecology/ecology/appgo/pkg/conf"
)

type Label struct {
	LabelId    int    `gorm:"not null;primary_key;AUTO_INCREMENT"json:"labelId"`
	LabelName  string `gorm:"not null;"json:"labelName"`
	BookNumber int    `gorm:"not null;"json:"bookNumber"`
}

// TableName 获取对应数据库表名.
func (m *Label) TableName() string {
	return "label"
}

// TableEngine 获取数据使用的引擎.
func (m *Label) TableEngine() string {
	return "INNODB"
}

func (m *Label) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

func NewLabel() *Label {
	return &Label{}
}
