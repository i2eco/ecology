package dao

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"go.uber.org/zap"
	"regexp"
	"strings"
	"time"
)

// Login 用户登录.
func (*member) Login(account string, password string) (*mysql.Member, error) {
	one := &mysql.Member{}
	var err error
	err = mus.Db.Where("account = ? AND status = ?", account, 0).Find(one).Error
	if err != nil {
		return one, err
	}
	ok, err := utils.PasswordVerify(one.Password, password)
	if ok && err == nil {
		one.ResolveRoleName()
		return one, nil
	}
	return one, ErrorMemberPasswordError
}

// Create 添加一个用户.
func (m *member) CreateX(c *gin.Context, data *mysql.Member) error {
	if ok, err := regexp.MatchString(conf.Conf.Info.RegexpAccount, data.Account); data.Account == "" || !ok || err != nil {
		return errors.New("用户名只能由英文字母数字组成，且在3-50个字符")
	}
	if data.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if ok, err := regexp.MatchString(conf.Conf.Info.RegexpEmail, data.Email); !ok || err != nil || data.Email == "" {
		return errors.New("邮箱格式不正确")
	}
	if data.AuthMethod == "local" {
		if l := strings.Count(data.Password, ""); l < 6 || l >= 50 {
			return errors.New("密码不能为空且必须在6-50个字符之间")
		}
	}

	var one mysql.Member
	mus.Db.Select("member_id,nickname,account,email").Where("email = ? or nickname = ? or account =?", data.Email, data.Nickname, data.Account).Find(&one)
	if one.MemberId > 0 {
		if one.Nickname == data.Nickname {
			return errors.New("昵称已存在，请更换昵称")
		}
		if one.Email == data.Email {
			return errors.New("邮箱已被注册，请更换邮箱")
		}
		if one.Account == data.Account {
			return errors.New("用户名已存在，请更换用户名")
		}
	}

	// 这里必需设置为读者，避免采坑：普通用户注册的时候注册成了管理员...
	if data.Account == "admin" {
		data.Role = conf.MemberSuperRole
	} else {
		data.Role = conf.MemberGeneralRole
	}

	hash, err := utils.PasswordHash(data.Password)
	if err != nil {
		return err
	}

	data.Password = hash
	if data.AuthMethod == "" {
		data.AuthMethod = "local"
	}
	data.CreateTime = time.Now()
	data.LastLoginTime = time.Now()
	if err = mus.Db.Create(data).Error; err != nil {
		mus.Logger.Error("create mdMembers create error", zap.String("err", err.Error()))
		return err
	}
	data.ResolveRoleName()
	return nil
}

func (m *member) Find(id int) (*mysql.Member, error) {
	member := &mysql.Member{}
	err := mus.Db.Where("member_id = ?", id).Find(member).Error
	if err != nil {
		return member, err
	}
	member.ResolveRoleName()
	return member, nil
}

//根据指定字段查找用户.
func (m *member) FindByFieldFirst(field string, value interface{}) (*mysql.Member, error) {
	member := &mysql.Member{}
	var err error
	err = mus.Db.Where(field+" = ?", value).Order("member_id desc").Find(member).Error
	if err != nil {
		return member, err
	}
	return member, err
}

//获取昵称
func (m *member) GetNicknameByUid(id interface{}) string {
	var user mysql.Member
	mus.Db.Where("member_id = ?", id).Find(&user)
	return user.Nickname
}

//根据账号查找用户.
func (m *member) FindByAccount(account string) (*mysql.Member, error) {
	member := &mysql.Member{}
	err := mus.Db.Where("account = ?", account).Find(member).Error
	if err == nil {
		member.ResolveRoleName()
	}
	return member, err
}

//根据用户id获取二维码
func (m *member) GetQrcodeByUid(uid interface{}) (qrcode map[string]string) {
	var member mysql.Member
	mus.Db.Select("alipay,wxpay").Where("member_id = ?", uid).Find(&member)
	qrcode = make(map[string]string)
	qrcode["Alipay"] = member.Alipay
	qrcode["Wxpay"] = member.Wxpay
	return qrcode
}

//删除一个用户.

func (m *member) DeleteXX(c *gin.Context, oldId int, newId int) error {
	o := mus.Db.Begin()
	var err error

	err = o.Raw("DELETE FROM md_members WHERE member_id = ?", oldId).Error
	if err != nil {
		o.Rollback()
		return err
	}
	err = o.Raw("UPDATE md_attachment SET `create_at` = ? WHERE `create_at` = ?", newId, oldId).Error
	if err != nil {
		o.Rollback()
		return err
	}

	err = o.Raw("UPDATE md_books SET member_id = ? WHERE member_id = ?", newId, oldId).Error
	if err != nil {
		o.Rollback()
		return err
	}
	err = o.Raw("UPDATE md_document_history SET member_id=? WHERE member_id = ?", newId, oldId).Error
	if err != nil {
		o.Rollback()
		return err
	}
	err = o.Raw("UPDATE md_document_history SET modify_at=? WHERE modify_at = ?", newId, oldId).Error
	if err != nil {
		o.Rollback()
		return err
	}
	err = o.Raw("UPDATE md_documents SET member_id = ? WHERE member_id = ?;", newId, oldId).Error
	if err != nil {
		o.Rollback()
		return err
	}
	err = o.Raw("UPDATE md_documents SET modify_at = ? WHERE modify_at = ?", newId, oldId).Error
	if err != nil {
		o.Rollback()
		return err
	}
	//_,err = o.Raw("UPDATE md_relationship SET member_id = ? WHERE member_id = ?",newId,oldId).Error
	//if err != nil {
	//
	//	if err != nil {
	//		o.Rollback()
	//		return err
	//	}
	//}
	var relationshipList []mysql.Relationship

	relationshipList, err = Relationship.List(c, mysql.Conds{"member_id": oldId}, "")
	if err == nil {
		for _, relationship := range relationshipList {
			//如果存在创始人，则删除
			if relationship.RoleId == 0 {
				rel, err := Relationship.InfoX(c, mysql.Conds{"member_id": newId})
				if err == nil {
					if err := o.Delete(relationship); err != nil {
					}
					relationship.RelationshipId = rel.RelationshipId
				}
				relationship.MemberId = newId
				relationship.RoleId = 0
				if err := o.Update(relationship); err != nil {
				}
			} else {
				if err := o.Delete(relationship); err != nil {
				}
			}
		}
	}

	if err = o.Commit().Error; err != nil {
		o.Rollback()
		return err
	}
	return nil
}

//校验用户.
func (m *member) Valid(data *mysql.Member, isHashPassword bool) error {
	var err error
	//邮箱不能为空
	if data.Email == "" {
		return ErrMemberEmailEmpty
	}
	//用户描述必须小于500字
	if strings.Count(data.Description, "") > 500 {
		return ErrMemberDescriptionTooLong
	}
	if data.Role != conf.Conf.Info.MemberGeneralRole && data.Role != conf.Conf.Info.MemberSuperRole && data.Role != conf.Conf.Info.MemberAdminRole {
		return ErrMemberRoleError
	}
	if data.Status != 0 && data.Status != 1 {
		data.Status = 0
	}
	//邮箱格式校验
	if ok, err := regexp.MatchString(conf.Conf.Info.RegexpEmail, data.Email); !ok || err != nil || data.Email == "" {
		return ErrMemberEmailFormatError
	}
	//如果是未加密密码，需要校验密码格式
	if !isHashPassword {
		if l := strings.Count(data.Password, ""); data.Password == "" || l > 50 || l < 6 {
			return ErrMemberPasswordFormatError
		}
	}
	//校验邮箱是否呗使用

	var one *mysql.Member
	one, err = m.FindByFieldFirst("email", data.Account)

	if err == nil && one.MemberId > 0 {
		if data.MemberId > 0 && data.MemberId != one.MemberId {
			return ErrMemberEmailExist
		}
		if data.MemberId <= 0 {
			return ErrMemberEmailExist
		}
	}

	if data.MemberId > 0 {
		//校验用户是否存在
		if _, err := m.Find(data.MemberId); err != nil {
			return err
		}
	} else {
		//校验账号格式是否正确
		if ok, err := regexp.MatchString(conf.Conf.Info.RegexpAccount, data.Account); data.Account == "" || !ok || err != nil {
			return ErrMemberAccountFormatError
		}
		//校验账号是否被使用
		if member, err := m.FindByAccount(data.Account); err == nil && member.MemberId > 0 {
			return ErrMemberExist
		}
	}

	return nil
}

//获取用户名
func (m *member) GetUsernameByUid(id interface{}) string {
	var user mysql.Member
	mus.Db.Where("member_id = ?", id).Find(&user)
	return user.Account
}

//根据用户名获取用户信息
func (this *member) GetByUsername(username string) (user mysql.Member, err error) {
	mus.Db.Where("account = ?", username).Find(&user)
	return
}

//分页查找用户.
func (m *member) FindToPager(pageIndex, pageSize int, wd string, role ...int) ([]*mysql.Member, int, error) {
	var members []*mysql.Member
	var cnt int
	offset := (pageIndex - 1) * pageSize

	qs := mus.Db

	if len(role) > 0 && role[0] != -1 {
		qs = qs.Where("role = ?", role[0])
	}

	if wd != "" {
		qs = qs.Where("account like %?% or nickname like %?% or email like %?%", wd, wd, wd)
	}

	err := qs.Model(mysql.Member{}).Count(&cnt).Error
	if err != nil {
		return members, 0, err
	}

	err = qs.Order("member_id desc").Offset(offset).Limit(pageSize).Find(&members).Error

	if err != nil {
		return members, 0, err
	}

	for _, m := range members {
		m.ResolveRoleName()
	}
	return members, cnt, nil
}
