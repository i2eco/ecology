package mysql

import (
	"github.com/i2eco/ecology/appgo/pkg/conf"
)

type Relationship struct {
	RelationshipId int `gorm:"not null;primary_key;AUTO_INCREMENT"json:"relationshipId"`
	MemberId       int `gorm:"not null;"json:"memberId"`
	BookId         int `gorm:"not null;"json:"bookId"`
	RoleId         int `gorm:"not null;"json:"roleId"` // RoleId 角色：0 创始人(创始人不能被移除) / 1 管理员/2 编辑者/3 观察者
}

// TableName 获取对应数据库表名.
func (m Relationship) TableName() string {
	return "relationship"
}
func (m *Relationship) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

// TableEngine 获取数据使用的引擎.
func (m *Relationship) TableEngine() string {
	return "INNODB"
}

// 联合唯一键
func (u *Relationship) TableUnique() [][]string {
	return [][]string{
		[]string{"MemberId", "BookId"},
	}
}

func NewRelationship() *Relationship {
	return &Relationship{}
}
