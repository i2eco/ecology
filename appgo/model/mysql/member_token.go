package mysql

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goecology/ecology/appgo/pkg/conf"
)

type MemberToken struct {
	TokenId   int       `gorm:"not null;primary_key;AUTO_INCREMENT"json:"tokenId"`
	MemberId  int       `gorm:"not null;"json:"memberId"`
	Token     string    `gorm:"not null;"json:"token"`
	Email     string    `gorm:"not null;"json:"email"`
	IsValid   bool      `gorm:"not null;"json:"isValid"`
	ValidTime time.Time `gorm:"not null;"json:"validTime"`
	SendTime  time.Time `gorm:"not null;"json:"sendTime"`
}

// TableName 获取对应数据库表名.
func (m *MemberToken) TableName() string {
	return "member_token"
}

// TableEngine 获取数据使用的引擎.
func (m *MemberToken) TableEngine() string {
	return "INNODB"
}

func (m *MemberToken) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

func NewMemberToken() *MemberToken {
	return &MemberToken{}
}

func (m *MemberToken) InsertOrUpdate() (*MemberToken, error) {
	o := orm.NewOrm()

	if m.TokenId > 0 {
		_, err := o.Update(m)
		return m, err
	}
	_, err := o.Insert(m)

	return m, err
}

func (m *MemberToken) FindByFieldFirst(field string, value interface{}) (*MemberToken, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter(field, value).OrderBy("-token_id").One(m)

	return m, err
}

func (m *MemberToken) FindSendCount(mail string, startTime time.Time, endTime time.Time) (int, error) {
	o := orm.NewOrm()

	c, err := o.QueryTable(m.TableNameWithPrefix()).Filter("send_time__gte", startTime.Format("2006-01-02 15:04:05")).Filter("send_time__lte", endTime.Format("2006-01-02 15:04:05")).Count()

	if err != nil {
		return 0, err
	}
	return int(c), nil
}
