package account

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/model"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/code"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/goecology/ecology/appgo/router/core"
)

func LoginApi(c *core.Context) {
	var (
		remember model.CookieRemember
		//captchaOn bool //是否开启了验证码
	)

	account, flag1 := c.GetPostForm("account")
	if !flag1 {
		c.String(401, "error3")
		return
	}
	password, flag2 := c.GetPostForm("password")
	if !flag2 {
		c.String(401, "error3")
		return
	}

	//if captchaOn && !cpt.VerifyReq(c.Context.Request) {
	//	c.JSONErrStr(1, "验证码不正确")
	//}

	member, err := dao.Member.Login(account, password)

	//如果没有数据
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	err = dao.Member.UpdateX(c.Context, mus.Db, mysql.Conds{
		"member_id": member.MemberId,
	}, mysql.Ups{
		"last_login_time": time.Now(),
	})
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	err = c.UpdateUser(member)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	remember.MemberId = member.MemberId
	remember.Account = member.Account
	remember.Time = time.Now()
	v, err := utils.Encode(remember)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	SetSecureCookie(c.Context, conf.Conf.App.AppKey, "login", v, 24*3600*365)
	c.JSONOK()
}

//用户注册.[移除用户注册，直接叫用户绑定]
//注意：如果用户输入的账号密码跟现有的账号密码相一致，则表示绑定账号，否则表示注册新账号。
func BindApi(c *core.Context) {

	options := dao.Global.AllOptions()
	var err error
	account := c.GetString("account")
	nickname := strings.TrimSpace(c.GetString("nickname"))
	password1 := c.GetString("password1")
	password2 := c.GetString("password2")
	email := c.GetString("email")
	oauthType := c.GetString("oauth")
	oauthId := c.GetString("id")
	avatar := c.GetString("avatar") //用户头像
	isbind := c.GetInt("isbind")

	ibind := func(oauthType string, oauthId, memberId interface{}) (err error) {
		//注册成功，绑定用户
		switch oauthType {
		case "gitee":
			//err = mysql.ModelGitee.Bind(oauthId, memberId)
		case "github":
			//err = mysql.ModelGithub.Bind(oauthId, memberId)
		case "qq":
			err = mysql.ModelQQ.Bind(oauthId, memberId)
		}
		return
	}

	if oauthType != "email" {
		//if auth, ok := this.GetSession("auth").(string); !ok || fmt.Sprintf("%v-%v", oauthType, oauthId) != auth {
		//	c.JSONErr(6005,errors.New("绑定信息有误，授权类型不符"))
		//	return
		//}
	} else { //邮箱登录，如果开启了验证码，则对验证码进行校验
		if v, ok := options["ENABLED_CAPTCHA"]; ok && strings.EqualFold(v, "true") {
			//if !cpt.VerifyReq(c.Context.Request) {
			//	c.JSONErrStr(1, "验证码不正确")
			//}
		}
	}

	member := mysql.NewMember()

	if isbind == 1 {
		if member, err = dao.Member.Login(account, password1); err != nil || member.MemberId == 0 {
			c.JSONErr(1, errors.New("绑定用户失败，用户名或密码不正确"))
			return
		}
	} else {
		if password1 != password2 {
			c.JSONErr(6003, errors.New("登录密码与确认密码不一致"))
			return
		}

		if ok, err := regexp.MatchString(conf.RegexpAccount, account); account == "" || !ok || err != nil {
			c.JSONErrStr(6001, "用户名只能由英文字母数字组成，且在3-50个字符")
			return
		}
		if l := strings.Count(password1, ""); password1 == "" || l > 50 || l < 6 {
			c.JSONErrStr(6002, "密码必须在6-50个字符之间")
			return
		}

		if ok, err := regexp.MatchString(conf.RegexpEmail, email); !ok || err != nil || email == "" {
			c.JSONErrStr(6004, "邮箱格式不正确")
			return
		}
		if l := strings.Count(nickname, "") - 1; l < 2 || l > 20 {
			c.JSONErrStr(6005, "用户昵称限制在2-20个字符")
			return
		}

		//出错或者用户不存在，则重新注册用户，否则直接登录
		member.Account = account
		member.Nickname = nickname
		member.Password = password1
		member.Role = conf.MemberGeneralRole
		member.Avatar = conf.GetDefaultAvatar()
		member.CreateAt = 0
		member.Email = email
		member.Status = 0
		if len(avatar) > 0 {
			member.Avatar = avatar
		}
		if err := dao.Member.CreateX(c.Context, member); err != nil {
			c.JSONErr(6006, err)
			return
		}
	}
	if err = loginByMemberId(c, member.MemberId); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	if err = ibind(oauthType, oauthId, member.MemberId); err != nil {
		c.JSONErrStr(code.MsgErr, "登录失败")
		return
	}

	if oauthType == "email" {
		c.JSONOK("注册成功")
		return
	}
	c.JSONOK("登录成功")
}

func loginByMemberId(c *core.Context, memberId int) (err error) {
	member, err := dao.Member.Find(memberId)
	if err != nil {
		return
	}

	err = dao.Member.UpdateX(c.Context, mus.Db, mysql.Conds{
		"member_id": member.MemberId,
	}, mysql.Ups{
		"last_login_time": time.Now(),
	})

	err = c.UpdateUser(member)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	var remember model.CookieRemember

	remember.MemberId = member.MemberId
	remember.Account = member.Account
	remember.Time = time.Now()
	var v string
	v, err = utils.Encode(remember)
	if err != nil {
		return
	}
	SetSecureCookie(c.Context, conf.Conf.App.AppKey, "login", v, 24*3600*365)
	return
}

func FindPasswordApi(c *core.Context) {
	email := c.GetString("email")
	mailConf := conf.GetMailConfig()

	if email == "" {
		c.JSONErrStr(6005, "邮箱地址不能为空")
		return
	}
	if !mailConf.EnableMail {
		c.JSONErrStr(6004, "未启用邮件服务")
		return
	}

	//captcha := this.GetString("code")
	//如果开启了验证码
	//if v, ok := this.Option["ENABLED_CAPTCHA"]; ok && strings.EqualFold(v, "true") {
	//	v, ok := this.GetSession(conf.CaptchaSessionName).(string)
	//	if !ok || !strings.EqualFold(v, captcha) {
	//		c.JSONErrStr(6001, "验证码不正确")
	//	}
	//}

	//if !cpt.VerifyReq(c.Context.Request) {
	//	c.JSONErrStr(6001, "验证码不正确")
	//}

	member, err := dao.Member.FindByFieldFirst("email", email)
	if err != nil {
		c.JSONErrStr(6006, "邮箱不存在")
		return
	}
	if member.Status != 0 {
		c.JSONErrStr(6007, "账号已被禁用")
		return
	}
	if member.AuthMethod == conf.AuthMethodLDAP {
		c.JSONErrStr(6011, "当前用户不支持找回密码")
		return
	}

	count, err := mysql.NewMemberToken().FindSendCount(email, time.Now().Add(-1*time.Hour), time.Now())

	if err != nil {
		c.JSONErrStr(6008, "发送邮件失败")
		return
	}
	if count > mailConf.MailNumber {
		c.JSONErrStr(6008, "发送次数太多，请稍候再试")
		return
	}

	memberToken := mysql.NewMemberToken()

	memberToken.Token = string(utils.Krand(32, utils.KC_RAND_KIND_ALL))
	memberToken.Email = email
	memberToken.MemberId = member.MemberId
	memberToken.IsValid = false
	if _, err := memberToken.InsertOrUpdate(); err != nil {
		c.JSONErrStr(6009, "邮件发送失败")
		return
	}

	data := map[string]interface{}{
		"SITE_NAME": dao.Global.GetSiteName(),
		//"url":       c.BaseUrl() + beego.URLFor("AccountController.FindPassword", "token", memberToken.Token, "mail", email),
		"url": c.BaseUrl() + "/find_password?token=" + memberToken.Token + "&mail=" + email,
	}

	body, err := c.ExecuteViewPathTemplate("account/mail_template", data)
	if err != nil {
		c.JSONErrStr(6003, "邮件发送失败")
		return
	}

	if err = utils.SendMail(mailConf, "找回密码", email, body); err != nil {
		c.JSONErrStr(6003, "邮件发送失败")
		return
	}
	c.JSONOK("/login")
}

//校验邮件并修改密码.
func ValidEmail(c *core.Context) {
	password1 := c.GetString("password1")
	password2 := c.GetString("password2")
	token := c.GetString("token")
	mail := c.GetString("mail")

	if password1 == "" {
		c.JSONErrStr(6001, "密码不能为空")
		return
	}
	if l := strings.Count(password1, ""); l < 6 || l > 50 {
		c.JSONErrStr(6001, "密码不能为空且必须在6-50个字符之间")
		return
	}
	if password2 == "" {
		c.JSONErrStr(6002, "确认密码不能为空")
		return
	}
	if password1 != password2 {
		c.JSONErrStr(6003, "确认密码输入不正确")
		return
	}

	//if !cpt.VerifyReq(c.Context.Request) {
	//	c.JSONErrStr(6001, "验证码不正确")
	//}

	mailConf := conf.GetMailConfig()
	memberToken, err := dao.MemberToken.FindByFieldFirst("token", token)

	if err != nil {
		c.JSONErrStr(6007, "邮件已失效")
		return
	}
	subTime := memberToken.SendTime.Sub(time.Now())

	if !strings.EqualFold(memberToken.Email, mail) || subTime.Minutes() > float64(mailConf.MailExpired) || !memberToken.ValidTime.IsZero() {
		c.JSONErrStr(6008, "验证码已过期，请重新操作。")
		return
	}
	member, err := dao.Member.Find(memberToken.MemberId)
	if err != nil {
		c.JSONErrStr(6005, "用户不存在")
		return
	}
	hash, err := utils.PasswordHash(password1)

	if err != nil {
		c.JSONErrStr(6006, "保存密码失败")
		return
	}

	musdb := mus.Db.Begin()
	err = dao.Member.UpdateX(c.Context, musdb, mysql.Conds{
		"member_id": member.MemberId,
	}, mysql.Ups{
		"password": hash,
	})

	if err != nil {
		musdb.Rollback()
		c.JSONErrStr(6006, "保存密码失败")
		return
	}

	memberToken.ValidTime = time.Now()
	memberToken.IsValid = true
	err = dao.MemberToken.InsertOrUpdate(musdb, memberToken)
	if err != nil {
		c.JSONErrStr(6006, "保存密码失败")
		return
	}
	c.JSONOK("/login")
}

// Logout 退出登录.
func Logout(c *core.Context) {
	err := c.Logout()
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	SetSecureCookie(c.Context, conf.Conf.App.AppKey, "login", "", -3600)
	c.Redirect(302, "/login")
}

//记录笔记
func Note(c *core.Context) {
	docid := c.GetInt("doc_id")
	fmt.Println(docid)
	if strings.ToLower(c.Context.Request.Method) == "post" {

	} else {
		c.Tpl().Data["SeoTitle"] = "笔记"
		c.Html("account/note")
	}
}

// SetSecureCookie Set Secure cookie for response.
func SetSecureCookie(c *gin.Context, Secret, name, value string, others ...interface{}) {
	vs := base64.URLEncoding.EncodeToString([]byte(value))
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	h := hmac.New(sha1.New, []byte(Secret))
	//fmt.Fprintf(h, "%s%s", vs, timestamp)
	sig := fmt.Sprintf("%02x", h.Sum(nil))
	cookie := strings.Join([]string{vs, timestamp, sig}, "|")
	//context.SetCookie("name", "Shimin Li", 10, "/", "localhost", false, true)

	c.SetCookie(name, cookie, 10, "/", "", false, true)
}
