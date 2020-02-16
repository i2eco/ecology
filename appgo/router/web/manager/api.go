package manager

import (
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/code"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/goecology/ecology/appgo/router/core"
	"regexp"
	"strconv"
	"strings"
)

// 添加用户.
func CreateMember(c *core.Context) {
	account, _ := c.GetPostForm("account")
	nickname, _ := c.GetPostForm("nickname")
	password1, _ := c.GetPostForm("password1")
	password2, _ := c.GetPostForm("password2")
	email, _ := c.GetPostForm("email")
	phone, _ := c.GetPostForm("phone")
	roleStr, _ := c.GetPostForm("role")
	role, _ := strconv.Atoi(roleStr)
	//status, _ := this.GetInt("status", 0)

	if ok, err := regexp.MatchString(conf.Conf.Info.RegexpAccount, account); account == "" || !ok || err != nil {
		c.JSONCode(code.UserAccountLengthErr)
		return
	}
	if l := strings.Count(nickname, "") - 1; l < 2 || l > 20 {
		c.JSONCode(code.UserAccountNicknameErr)
		return
	}
	if l := strings.Count(password1, ""); password1 == "" || l > 50 || l < 6 {
		c.JSONCode(code.UserAccountErr3)
		return
	}
	if password1 != password2 {
		c.JSONCode(code.UserAccountErr4)
		return
	}
	if ok, err := regexp.MatchString(conf.Conf.Info.RegexpEmail, email); !ok || err != nil || email == "" {
		c.JSONCode(code.UserAccountErr5)
		return
	}
	if role != 0 && role != 1 && role != 2 {
		role = 1
	}

	var member *mysql.Member
	var err error
	if member, err = dao.Member.FindByAccount(account); err != nil {
		c.JSONCode(code.UserAccountErr6)
		return
	}

	currentUser := c.Member()

	member.Account = account
	member.Password = password1
	member.Role = role
	member.Avatar = conf.Conf.Info.DefaultAvatar
	member.CreateAt = currentUser.MemberId
	member.Email = email
	member.Nickname = nickname
	if phone != "" {
		member.Phone = phone
	}

	if err := dao.Member.CreateX(c.Context, member); err != nil {
		c.JSONCode(code.UserAccountErr7)
		return
	}
	c.JSONOK()
}

//删除一个用户，并将该用户的所有信息转移到超级管理员上.
func DeleteMember(c *core.Context) {
	memberIdStr, _ := c.GetPostForm("id")
	memberId, _ := strconv.Atoi(memberIdStr)

	if memberId <= 0 {
		c.JSONCode(code.DeleteMemberErr1)
		return
	}

	member, err := dao.Member.Find(memberId)
	if err != nil {
		c.JSONErr(code.DeleteMemberErr2, err)
		return
	}
	if member.Role == conf.Conf.Info.MemberSuperRole {
		c.JSONCode(code.DeleteMemberErr3)
		return
	}
	superMember, err := dao.Member.FindByFieldFirst("role", 0)

	if err != nil {
		c.JSONErr(code.DeleteMemberErr4, err)
		return
	}

	err = dao.Member.DeleteXX(c.Context, memberId, superMember.MemberId)
	if err != nil {
		c.JSONErr(code.DeleteMemberErr5, err)
		return
	}
	c.JSONOK()
}

//更新用户状态.
func UpdateMemberStatus(c *core.Context) {
	var err error
	var member *mysql.Member

	memberIdStr, _ := c.GetPostForm("id")
	memberId, _ := strconv.Atoi(memberIdStr)

	statusStr, _ := c.GetPostForm("status")
	status, _ := strconv.Atoi(statusStr)

	if memberId <= 0 {
		c.JSONCode(code.MsgErr)
		return
	}
	if status != 0 && status != 1 {
		status = 0
	}

	if member, err = dao.Member.Find(memberId); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	currentMember := c.Member()
	if member.MemberId == currentMember.MemberId {
		c.JSONCode(code.MsgErr)
		return
	}
	if member.Role == conf.Conf.Info.MemberSuperRole {
		c.JSONCode(code.MsgErr)
		return
	}

	if err := dao.Member.UpdateX(c.Context, mus.Db, mysql.Conds{"member_id": memberId}, mysql.Ups{
		"status": status,
	}); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	c.JSONOK()
}

//变更用户权限.
func ChangeMemberRole(c *core.Context) {
	var err error
	var member *mysql.Member

	memberIdStr, _ := c.GetPostForm("id")
	memberId, _ := strconv.Atoi(memberIdStr)

	roleStr, _ := c.GetPostForm("role")
	role, _ := strconv.Atoi(roleStr)

	if role != conf.Conf.Info.MemberAdminRole && role != conf.Conf.Info.MemberGeneralRole {
		c.JSONCode(code.MsgErr)
		return
	}
	if member, err = dao.Member.Find(memberId); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	currentMember := c.Member()

	if member.MemberId == currentMember.MemberId {
		c.JSONCode(code.MsgErr)
		return
	}
	if member.Role == conf.Conf.Info.MemberSuperRole {
		c.JSONCode(code.MsgErr)
		return
	}
	member.Role = role

	if err := dao.Member.UpdateX(c.Context, mus.Db, mysql.Conds{"member_id": member.MemberId}, mysql.Ups{
		"role": role,
	}); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	member.ResolveRoleName()
	c.JSONOK(member)
}

//编辑用户信息.
func UpdateMember(c *core.Context) {
	memberIdStr, _ := c.GetPostForm("id")
	memberId, _ := strconv.Atoi(memberIdStr)

	if memberId <= 0 {
		c.JSONCode(code.UserUpdateErr1)
		return
	}

	member, err := dao.Member.Find(memberId)

	if err != nil {
		c.JSONErr(code.UserUpdateErr2, err)
		return
	}

	password1, _ := c.GetPostForm("password1")
	password2, _ := c.GetPostForm("password2")
	email, _ := c.GetPostForm("email")
	phone, _ := c.GetPostForm("phone")
	description, _ := c.GetPostForm("description")
	member.Email = email
	member.Phone = phone
	member.Description = description
	if password1 != "" && password2 != password1 {
		c.JSONCode(code.UserUpdateErr3)
		return
	}
	//if password1 != "" && member.AuthMethod != bootstrap.Conf.Info.AuthMethodLDAP {
	//	member.Password = password1
	//}
	if password1 != "" {
		member.Password = password1
	}
	if err := dao.Member.Valid(member, password1 == ""); err != nil {
		c.JSONCode(code.UserUpdateErr4)

		return
	}
	if password1 != "" {
		password, err := utils.PasswordHash(password1)
		if err != nil {
			c.JSONCode(code.UserUpdateErr5)
			return
		}
		member.Password = password
	}
	if err := dao.Member.UpdateX(c.Context, mus.Db, mysql.Conds{"member_id": memberId}, mysql.Ups{
		"email":       email,
		"phone":       phone,
		"password":    member.Password,
		"description": description,
	}); err != nil {
		c.JSONErr(code.UserUpdateErr6, err)
		return
	}
	c.JSONOK()
}

//编辑用户信息.
func EditMemberApi(c *core.Context) {
	memberId := c.GetInt(":id")
	if memberId <= 0 {
		c.Html404()
		return
	}

	member, err := dao.Member.Find(memberId)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.Html404()
		return
	}

	password1 := c.GetString("password1")
	password2 := c.GetString("password2")
	email := c.GetString("email")
	phone := c.GetString("phone")
	description := c.GetString("description")
	member.Email = email
	member.Phone = phone
	member.Description = description
	if password1 != "" && password2 != password1 {
		c.JSONErrStr(6001, "确认密码不正确")
		return
	}
	if password1 != "" && member.AuthMethod != conf.AuthMethodLDAP {
		member.Password = password1
	}
	if err := dao.Member.Valid(member, password1 == ""); err != nil {
		c.JSONErrStr(6002, err.Error())
		return
	}
	if password1 != "" {
		password, err := utils.PasswordHash(password1)
		if err != nil {
			mus.Logger.Error(err.Error())
			c.JSONErrStr(6003, "对用户密码加密时出错")
			return
		}
		member.Password = password
	}
	if err := mus.Db.UpdateColumns(member).Error; err != nil {
		mus.Logger.Error(err.Error())
		c.JSONErrStr(6004, "保存失败")
		return
	}
	c.JSONOK()
}

//编辑项目.
func EditBookApi(c *core.Context) {
	identify := c.GetString("key")
	if identify == "" {
		c.Html404()
		return
	}

	book, err := dao.Book.FindByFieldFirst("identify", identify)
	if err != nil {
		c.Html404()
		return
	}

	bookName := strings.TrimSpace(c.GetString("book_name"))
	description := strings.TrimSpace(c.GetString("description"))
	commentStatus := c.GetString("comment_status")
	tag := strings.TrimSpace(c.GetString("label"))
	orderIndex := c.GetInt("order_index")
	pin := c.GetInt("pin")

	if strings.Count(description, "") > 500 {
		c.JSONErrStr(6004, "项目描述不能大于500字")
		return
	}
	if commentStatus != "open" && commentStatus != "closed" && commentStatus != "group_only" && commentStatus != "registered_only" {
		commentStatus = "closed"
	}
	if tag != "" {
		tags := strings.Split(tag, ";")
		if len(tags) > 10 {
			c.JSONErrStr(6005, "最多允许添加10个标签")
			return
		}
	}

	book.BookName = bookName
	book.Description = description
	book.CommentStatus = commentStatus
	book.Label = tag
	book.OrderIndex = orderIndex
	book.Pin = pin

	if err := mus.Db.UpdateColumns(book); err != nil {
		c.JSONErrStr(6006, "保存失败")
		return
	}

	go func() {
		es := mysql.ElasticSearchData{
			Id:       book.BookId,
			BookId:   0,
			Title:    book.BookName,
			Keywords: book.Label,
			Content:  book.Description,
			Vcnt:     book.Vcnt,
			Private:  book.PrivatelyOwned,
		}
		client := dao.NewElasticSearchClient()
		if errSearch := client.BuildIndex(es); errSearch != nil && client.On {
			mus.Logger.Error(err.Error())
		}
	}()

	c.JSONErrStr(0, "ok")

}

func SettingApi(c *core.Context) {
	options := dao.Global.All()

	for _, item := range options {
		item.OptionValue = c.GetString(item.OptionName)
		// todo fix
		//item.InsertOrUpdate()
	}
	if err := dao.NewElasticSearchClient().Init(); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	// todo fix

	//mysql.NewSign().UpdateSignRule()
	//mysql.NewReadRecord().UpdateReadingRule()
	c.JSONOK()

}

func SeoApi(c *core.Context) {
	id, _ := c.GetPostForm("id")
	field, _ := c.GetPostForm("field")
	value, _ := c.GetPostForm("value")

	err := mus.Db.Where("id = ?", id).Updates(mysql.Ups{
		field: value,
	}).Error
	if err != nil {
		mus.Logger.Error(err.Error())
		c.JSONErrStr(1, "更新失败，请求错误")
		return
	}
	c.JSONOK()

}

//分类管理
func CategoryApi(c *core.Context) {
	var req CategoryCreate
	err := c.Bind(&req)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	//新增分类
	if err := dao.Category.AddCates(req.Pid, req.Cates); err != nil {
		c.JSONErr(1, err)
		return
	}
	c.JSONOK()
	return

}

//广告管理
func AdsApi(c *core.Context) {
	//pid, _ := this.GetInt("pid")
	//if pid <= 0 {
	//	c.JSONErrStr(1, "请选择广告位")
	//}
	//ads := &mysql.AdsCont{
	//	Title:  this.GetString("title"),
	//	Code:   this.GetString("code"),
	//	Status: true,
	//	Pid:    pid,
	//}
	//start, err := dateparse.ParseAny(this.GetString("start"))
	//if err != nil {
	//	start = time.Now()
	//}
	//end, err := dateparse.ParseAny(this.GetString("end"))
	//if err != nil {
	//	end = time.Now().Add(24 * time.Hour * 730)
	//}
	//ads.Start = int(start.Unix())
	//ads.End = int(end.Unix())
	//_, err = orm.NewOrm().Insert(ads)
	//if err != nil {
	//	c.JSONErrStr(1, err.Error())
	//}
	//go mysql.UpdateAdsCache()
	c.JSONErrStr(0, "新增广告成功")

}
