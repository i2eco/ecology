package bookMember

import (
	"errors"
	"strconv"

	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/router/core"
)

// AddMember 参加参与用户.
func AddMember(c *core.Context) {
	identify := c.PostForm("identify")
	account := c.PostForm("account")
	// todo fix
	//roleId, _ := this.GetInt("role_id", 3)
	roleId, _ := strconv.Atoi(c.PostForm("role_id"))

	if identify == "" || account == "" {
		c.JSONErrStr(6001, "参数错误")
		return
	}
	book, err := isPermission(c)

	if err != nil {
		c.JSONErrStr(6001, err.Error())
		return
	}

	var member *mysql.Member

	if member, err = dao.Member.FindByAccount(account); err != nil {
		c.JSONErrStr(404, "用户不存在")
		return
	}
	if member.Status == 1 {
		c.JSONErrStr(6003, "用户已被禁用")
		return
	}

	if _, err := dao.Relationship.FindForRoleId(book.BookId, member.MemberId); err == nil {
		c.JSONErrStr(6003, "用户已存在该项目中")
		return
	}

	relationship := mysql.Relationship{}
	relationship.BookId = book.BookId
	relationship.MemberId = member.MemberId
	relationship.RoleId = roleId

	if err := dao.Relationship.Insert(mus.Db, &relationship); err != nil {
		c.JSONErrStr(500, err.Error())
		return
	}

	result := mysql.NewMemberRelationshipResult().FromMember(member)
	result.RoleId = roleId
	result.RelationshipId = relationship.RelationshipId
	result.BookId = book.BookId
	result.ResolveRoleName()
	c.JSONOK(result)
}

// 变更指定用户在指定项目中的权限
func ChangeRole(c *core.Context) {
	identify := c.GetPostFormString("identify")
	memberId, _ := strconv.Atoi(c.GetPostFormString("member_id"))
	role, _ := strconv.Atoi(c.GetPostFormString("role_id"))
	userMember := c.Member()
	if identify == "" || memberId <= 0 {
		c.JSONErrStr(6001, "参数错误")
		return
	}
	if memberId == userMember.MemberId {
		c.JSONErrStr(6006, "不能变更自己的权限")
		return
	}

	book, err := dao.Book.ResultFindByIdentify(identify, userMember.MemberId)
	if err != nil {
		c.JSONErrStr(6002, err.Error())
		return
	}

	if book.RoleId != 0 && book.RoleId != 1 {
		c.JSONErrStr(403, "权限不足")
		return
	}

	var member *mysql.Member

	if member, err = dao.Member.Find(memberId); err != nil {
		c.JSONErrStr(6003, "用户不存在")
		return
	}

	if member.Status == 1 {
		c.JSONErrStr(6004, "用户已被禁用")
		return
	}

	relationship, err := dao.Relationship.UpdateRoleId(book.BookId, memberId, role)

	if err != nil {
		c.JSONErrStr(6005, err.Error())
		return
	}

	result := mysql.NewMemberRelationshipResult().FromMember(member)
	result.RoleId = relationship.RoleId
	result.RelationshipId = relationship.RelationshipId
	result.BookId = book.BookId
	result.ResolveRoleName()

	c.JSONOK(result)
}

// 删除参与者.
func RemoveMember(c *core.Context) {
	identify := c.GetPostFormString("identify")
	memberId, _ := strconv.Atoi(c.GetPostFormString("member_id"))
	userMember := c.Member()

	if identify == "" || memberId <= 0 {
		c.JSONErrStr(6001, "参数错误")
		return
	}
	if memberId == userMember.MemberId {
		c.JSONErrStr(6006, "不能删除自己")
		return
	}
	book, err := dao.Book.ResultFindByIdentify(identify, userMember.MemberId)

	if err != nil {
		c.JSONErrStr(6002, err.Error())
		return
	}
	//如果不是创始人也不是管理员则不能操作
	if book.RoleId != conf.Conf.Info.BookFounder && book.RoleId != conf.BookAdmin {
		c.JSONErrStr(403, "权限不足")
		return
	}

	err = dao.Relationship.DeleteByBookIdAndMemberId(book.BookId, memberId)
	if err != nil {
		c.JSONErrStr(6007, err.Error())
		return
	}
	c.JSONOK()
}

func isPermission(c *core.Context) (*mysql.BookResult, error) {
	identify := c.GetPostFormString("identify")
	member := c.Member()
	book, err := dao.Book.ResultFindByIdentify(identify, member.MemberId)

	if err != nil {
		return book, err
	}
	if book.RoleId != conf.BookAdmin && book.RoleId != conf.BookFounder {
		return book, errors.New("权限不足")
	}
	return book, nil
}
