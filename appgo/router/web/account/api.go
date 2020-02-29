package account

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/code"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/ecology/appgo/router/core"
	"github.com/i2eco/ecology/appgo/service"
	"github.com/spf13/viper"
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
	var req ReqBind
	err := c.Bind(&req)
	if err != nil {
		c.JSONErr(code.AccountBindErr1, err)
		return
	}
	options := dao.Global.AllOptions()

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

	if req.OauthType != "email" {
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

	if req.IsBind == 1 {
		if member, err = dao.Member.Login(req.Account, req.Password1); err != nil || member.MemberId == 0 {
			c.JSONErr(code.AccountBindErr2, err)
			return
		}
	} else {
		if req.Password1 != req.Password2 {
			c.JSONErr(code.AccountBindErr3, err)
			return
		}

		if ok, err := regexp.MatchString(conf.RegexpAccount, req.Account); req.Account == "" || !ok || err != nil {
			c.JSONErr(code.AccountBindErr4, err)
			return
		}
		if l := strings.Count(req.Password1, ""); req.Password1 == "" || l > 50 || l < 6 {
			c.JSONErr(code.AccountBindErr5, err)
			return
		}

		if ok, err := regexp.MatchString(conf.RegexpEmail, req.Email); !ok || err != nil || req.Email == "" {
			c.JSONErr(code.AccountBindErr6, err)
			return
		}
		if l := strings.Count(req.Nickname, "") - 1; l < 2 || l > 20 {
			c.JSONErr(code.AccountBindErr7, err)
			return
		}

		//出错或者用户不存在，则重新注册用户，否则直接登录
		member.Account = req.Account
		member.Nickname = req.Nickname
		member.Password = req.Password1
		member.Role = conf.MemberGeneralRole
		member.Avatar = conf.GetDefaultAvatar()
		member.CreateAt = 0
		member.Email = req.Email
		member.Status = 0
		if len(req.Avatar) > 0 {
			member.Avatar = req.Avatar
		}
		if err := dao.Member.CreateX(c.Context, member); err != nil {
			c.JSONErr(code.AccountBindErr8, err)
			return
		}
	}
	if err = loginByMemberId(c, member.MemberId); err != nil {
		c.JSONErr(code.AccountBindErr9, err)
		return
	}

	if err = ibind(req.OauthType, req.OauthId, member.MemberId); err != nil {
		c.JSONErr(code.AccountBindErr10, err)
		return
	}

	if req.OauthType == "email" {
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
	email := c.PostForm("email")
	if email == "" {
		c.JSONErrStr(6005, "邮箱地址不能为空")
		return
	}
	if !viper.GetBool("email.isEnable") {
		c.JSONErrStr(6004, "未启用邮件服务")
		return
	}
	captchaId := c.PostForm("captchaId")
	captchaValue := c.PostForm("captcha")

	// 如果开启了验证码
	if dao.Global.IsEnabledCaptcha() && !captcha.VerifyString(captchaId, captchaValue) {
		c.JSONErrStr(6001, "验证码不正确")
		return
	}

	member, err := dao.Member.FindByFieldFirst("email", email)
	if err != nil {
		c.JSONErrTips("邮箱不存在", err)
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

	count, err := dao.MemberToken.FindSendCount(email, time.Now().Add(-1*time.Hour), time.Now())

	if err != nil {
		c.JSONErrStr(6008, "发送邮件失败")
		return
	}
	if count > viper.GetInt("email.mailMaxNum") {
		c.JSONErrStr(6008, "发送次数太多，请稍候再试")
		return
	}

	memberToken := mysql.MemberToken{
		MemberId: member.MemberId,
		Token:    string(utils.Krand(32, utils.KC_RAND_KIND_ALL)),
		Email:    email,
		IsValid:  false,
		// todo fix
		ValidTime: time.Date(1970, 1, 1, 0, 0, 01, 0, time.Local),
		SendTime:  time.Now(),
	}

	if err := dao.MemberToken.InsertOrUpdate(mus.Db, &memberToken); err != nil {
		c.JSONErrTips("邮件发送失败", err)
		return
	}

	data := map[string]interface{}{
		"SITE_NAME": dao.Global.GetSiteName(),
		"url":       c.BaseUrl() + "/find_password?token=" + memberToken.Token + "&mail=" + email,
	}

	body, err := c.ExecuteViewPathTemplate("account/mail_template.html", data)
	if err != nil {
		c.JSONErrStr(6003, "邮件发送失败")
		return
	}

	err = service.Mailer.Send("找回密码", email, body, "")
	if err != nil {
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

	memberToken, err := dao.MemberToken.FindByFieldFirst("token", token)

	if err != nil {
		c.JSONErrStr(6007, "邮件已失效")
		return
	}
	subTime := memberToken.SendTime.Sub(time.Now())

	if !strings.EqualFold(memberToken.Email, mail) || subTime.Minutes() > float64(viper.GetInt("email.mailExpired")) || !memberToken.ValidTime.IsZero() {
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
