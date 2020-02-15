// Package mysql .
package mysql

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/utils"
)

// member
type Member struct {
	MemberId                   int       `gorm:"not null;primary_key;AUTO_INCREMENT"json:"memberId"`
	Account                    string    `gorm:"not null;"json:"account"`
	Nickname                   string    `gorm:"not null;"json:"nickname"` //昵称
	Password                   string    `gorm:"not null;"json:"-"`
	AuthMethod                 string    `gorm:"not null;"json:"authMethod"` //认证方式: local 本地数据库 /ldap LDAP
	Description                string    `gorm:"not null;"json:"description"`
	Email                      string    `gorm:"not null;"json:"email"`
	Phone                      string    `gorm:"not null;"json:"phone"`
	Avatar                     string    `gorm:"not null;"json:"avatar"`
	Role                       int       `gorm:"not null;"json:"role"` //用户角色：0 超级管理员 /1 管理员/ 2 普通用户 .
	RoleName                   string    `gorm:"-"json:"roleName"`
	Status                     int       `gorm:"not null;"json:"status"` //用户状态：0 正常/1 禁用
	CreateTime                 time.Time `gorm:""json:"createTime"`
	CreateAt                   int       `gorm:"not null;"json:"createAt"`
	LastLoginTime              time.Time `gorm:""json:"lastLoginTime"`
	Wxpay                      string    `gorm:"not null;"json:"wxpay"`                      // 微信支付的收款二维码
	Alipay                     string    `gorm:"not null;"json:"alipay"`                     // 支付宝支付的收款二维码
	TotalReadingTime           int       `gorm:"not null;"json:"totalReadingTime"`           // 总阅读时长
	TotalSign                  int       `gorm:"not null;"json:"totalSign"`                  // 总签到天数
	TotalContinuousSign        int       `gorm:"not null;"json:"totalContinuousSign"`        // 总连续签到天数
	HistoryTotalContinuousSign int       `gorm:"not null;"json:"historyTotalContinuousSign"` // 历史最高连续签到天数
}

// TableName 获取对应数据库表名.
func (m Member) TableName() string {
	return "member"
}

// TableEngine 获取数据使用的引擎.
func (m *Member) TableEngine() string {
	return "INNODB"
}

func (m *Member) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

func NewMember() *Member {
	return &Member{}
}

// Add 添加一个用户.
func (m *Member) Add() error {
	o := orm.NewOrm()

	if ok, err := regexp.MatchString(conf.RegexpAccount, m.Account); m.Account == "" || !ok || err != nil {
		return errors.New("用户名只能由英文字母数字组成，且在3-50个字符")
	}
	if m.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if ok, err := regexp.MatchString(conf.RegexpEmail, m.Email); !ok || err != nil || m.Email == "" {
		return errors.New("邮箱格式不正确")
	}

	if l := strings.Count(m.Password, ""); l < 7 || l >= 50 {
		return errors.New("密码不能为空且必须在6-50个字符之间")
	}

	cond := orm.NewCondition().Or("email", m.Email).Or("nickname", m.Nickname).Or("account", m.Account)
	var one Member
	if o.QueryTable(m.TableNameWithPrefix()).SetCond(cond).One(&one, "member_id", "nickname", "account", "email"); one.MemberId > 0 {
		if one.Nickname == m.Nickname {
			return errors.New("昵称已存在，请更换昵称")
		}
		if one.Email == m.Email {
			return errors.New("邮箱已被注册，请更换邮箱")
		}
		if one.Account == m.Account {
			return errors.New("用户名已存在，请更换用户名")
		}
	}

	// 这里必需设置为读者，避免采坑：普通用户注册的时候注册成了管理员...
	if m.Account == "admin" {
		m.Role = conf.MemberSuperRole
	} else {
		m.Role = conf.MemberGeneralRole
	}

	hash, err := utils.PasswordHash(m.Password)

	if err != nil {
		return err
	}

	m.Password = hash
	if m.AuthMethod == "" {
		m.AuthMethod = "local"
	}
	_, err = o.Insert(m)

	if err != nil {
		return err
	}
	m.ResolveRoleName()
	return nil
}

// Update 更新用户信息.
func (m *Member) Update(cols ...string) error {
	o := orm.NewOrm()

	if m.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if _, err := o.Update(m, cols...); err != nil {
		return err
	}
	return nil
}

func (m *Member) ResolveRoleName() {
	switch m.Role {
	case conf.MemberSuperRole:
		m.RoleName = "超级管理员"
	case conf.MemberAdminRole:
		m.RoleName = "管理员"
	case conf.MemberGeneralRole:
		m.RoleName = "读者"
	case conf.MemberEditorRole:
		m.RoleName = "作者"
	}
}

//根据账号查找用户.
func (m *Member) FindByAccount(account string) (*Member, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("account", account).One(m)

	if err == nil {
		m.ResolveRoleName()
	}
	return m, err
}

func (m *Member) IsAdministrator() bool {
	if m == nil || m.MemberId <= 0 {
		return false
	}
	return m.Role == 0 || m.Role == 1
}

//根据指定字段查找用户.
func (m *Member) FindByFieldFirst(field string, value interface{}) (*Member, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter(field, value).OrderBy("-member_id").One(m)

	return m, err
}

// 获取用户信息，根据用户名或邮箱
func (this *Member) GetByUsername(username string) (member Member, err error) {
	q := orm.NewOrm().QueryTable("md_members")
	if strings.Contains(username, "@") { //存在 @ 符号的表示邮箱，因为用户名只有数字和字母
		err = q.Filter("email", username).One(&member)
	}
	if err != nil || member.MemberId == 0 {
		err = q.Filter("account", username).One(&member)
	}
	return
}
