package account

import (
	"strings"
	"time"

	"github.com/goecology/ecology/appgo/dao"

	"github.com/astaxie/beego"
	"github.com/goecology/ecology/appgo/model/constx"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/oauth"
	"github.com/goecology/ecology/appgo/router/core"
)

//第三方登录回调
//封装一个内部调用的函数，loginByMemberId
func OauthHtml(c *core.Context) {
	var (
		nickname  string //昵称
		avatar    string //头像的http链接地址
		email     string //邮箱地址
		username  string //用户名
		tips      string
		id        interface{} //第三方的用户id，唯一识别码
		IsEmail   bool        //是否是使用邮箱注册
		captchaOn bool        //是否开启了验证码
	)

	options := dao.Global.AllOptions()

	//如果开启了验证码
	if v, ok := options["ENABLED_CAPTCHA"]; ok && strings.EqualFold(v, "true") {
		captchaOn = true
		c.Tpl().Data["CaptchaOn"] = captchaOn
	}

	oauthLogin := false
	if v, ok := options["LOGIN_QQ"]; ok && strings.EqualFold(v, "true") {
		c.Tpl().Data["LoginQQ"] = true
		oauthLogin = true
	}
	if v, ok := options["LOGIN_GITHUB"]; ok && strings.EqualFold(v, "true") {
		c.Tpl().Data["LoginGitHub"] = true
		oauthLogin = true
	}
	if v, ok := options["LOGIN_GITEE"]; ok && strings.EqualFold(v, "true") {
		c.Tpl().Data["LoginGitee"] = true
		oauthLogin = true
	}
	c.Tpl().Data["OauthLogin"] = oauthLogin

	oa := c.GetString(":oauth")
	codeStr := c.GetString("code")
	switch oa {
	case "gitee":
		//tips = `您正在使用【码云】登录`
		//token, err := oauth.GetGiteeAccessToken(codeStr)
		//if err != nil {
		//	c.Html404()
		//	return
		//}
		//
		//info, err := oauth.GetGiteeUserInfo(token.AccessToken)
		//if err != nil {
		//	c.Html404()
		//	return
		//}

		//if info.Id > 0 {
		//	existInfo, _ := mysql.ModelGitee.GetUserByGiteeId(info.Id, "id", "member_id")
		//	if existInfo.MemberId > 0 { //直接登录
		//		err = this.loginByMemberId(existInfo.MemberId)
		//		if err != nil {
		//			mus.Logger.Error(err)
		//			c.Html404()
		//		}
		//		this.Redirect(beego.URLFor("HomeController.Index"), 302)
		//		return
		//	}
		//	if existInfo.Id == 0 { //原本不存在于数据库中的数据需要入库
		//		orm.NewOrm().Insert(&mysql.Gitee{GiteeUser: info})
		//	}
		//	nickname = info.Name
		//	username = info.Login
		//	avatar = info.AvatarURL
		//	email = info.Email
		//	id = info.Id
		//} else {
		//	err = errors.New("获取gitee用户数据失败")
		//	mus.Logger.Error(err)
		//	c.Html404()
		//}
	case "github":
		tips = `您正在使用【GitHub】登录`
		token, err := oauth.GetGithubAccessToken(codeStr)
		if err != nil {
			c.Html404()
			return
		}

		info, err := oauth.GetGithubUserInfo(token.AccessToken)
		if err != nil {
			c.Html404()
			return
		}

		if info.Id > 0 {
			//existInfo, _ := mysql.ModelGithub.GetUserByGithubId(info.Id, "id", "member_id")
			//if existInfo.MemberId > 0 { //直接登录
			//	err = this.loginByMemberId(existInfo.MemberId)
			//	if err != nil {
			//		c.Html404()
			//		return
			//	}
			//	this.Redirect(beego.URLFor("HomeController.Index"), 302)
			//	return
			//}
			//if existInfo.Id == 0 { //原本不存在于数据库中的数据需要入库
			//	orm.NewOrm().Insert(&mysql.Github{GithubUser: info})
			//}
			//nickname = info.Name
			//username = info.Login
			//avatar = info.AvatarURL
			//email = info.Email
			//id = info.Id
		} else {
			//err = errors.New("获取github用户数据失败")
			c.Html404()
			return
		}

	case "qq":
		tips = `您正在使用【QQ】登录`
		token, err := oauth.GetQQAccessToken(codeStr)
		if err != nil {
			c.Html404()
			return
		}

		openid, err := oauth.GetQQOpenId(token)
		if err != nil {
			c.Html404()
			return
		}

		info, err := oauth.GetQQUserInfo(token.AccessToken, openid)
		if err != nil {
			c.Html404()
			return
		}

		if info.Ret == 0 {
			//existInfo, _ := mysql.ModelQQ.GetUserByOpenid(openid, "id", "member_id")
			//if existInfo.MemberId > 0 { //直接登录
			//	err = this.loginByMemberId(existInfo.MemberId)
			//	if err != nil {
			//		c.Html404()
			//		return
			//	}
			//	c.Redirect(302,"/")
			//	return
			//}
			//
			//if existInfo.Id == 0 { //原本不存在于数据库中的数据需要入库
			//	orm.NewOrm().Insert(&mysql.QQ{
			//		OpenId:    openid,
			//		Name:      info.Name,
			//		Gender:    info.Gender,
			//		AvatarURL: info.AvatarURL,
			//	})
			//}
			//nickname = info.Name
			//username = ""
			//avatar = info.AvatarURL
			//email = ""
			//id = openid
		} else {
			c.Html404()
			return
		}
	default: //email
		IsEmail = true
	}

	c.Tpl().Data["IsEmail"] = IsEmail
	c.Tpl().Data["Nickname"] = nickname
	c.Tpl().Data["Avatar"] = avatar
	c.Tpl().Data["Email"] = email
	c.Tpl().Data["Username"] = username
	c.Tpl().Data["AuthType"] = oa
	c.Tpl().Data["SeoTitle"] = "完善信息"
	c.Tpl().Data["Tips"] = tips
	c.Tpl().Data["Id"] = id
	c.Tpl().Data["GiteeClientId"] = beego.AppConfig.String("oauth::giteeClientId")
	c.Tpl().Data["GiteeCallback"] = beego.AppConfig.String("oauth::giteeCallback")
	c.Tpl().Data["GithubClientId"] = beego.AppConfig.String("oauth::githubClientId")
	c.Tpl().Data["GithubCallback"] = beego.AppConfig.String("oauth::githubCallback")
	c.Tpl().Data["QQClientId"] = beego.AppConfig.String("oauth::qqClientId")
	c.Tpl().Data["QQCallback"] = beego.AppConfig.String("oauth::qqCallback")
	c.Tpl().Data["RandomStr"] = time.Now().Unix()
	//this.SetSession("auth", fmt.Sprintf("%v-%v", oa, id)) //存储标识，以标记是哪个用户，在完善用户信息的时候跟传递过来的auth和id进行校验
	c.Html("account/bind")

}

// Login 用户登录.
func LoginHtml(c *core.Context) {
	var (
		//remember  CookieRemember
		captchaOn bool //是否开启了验证码
	)
	options := dao.Global.AllOptions()

	//如果开启了验证码
	if v, ok := options["ENABLED_CAPTCHA"]; ok && strings.EqualFold(v, "true") {
		captchaOn = true
		c.Tpl().Data["CaptchaOn"] = captchaOn
	}

	oauthLogin := false
	if v, ok := options["LOGIN_QQ"]; ok && strings.EqualFold(v, "true") {
		c.Tpl().Data["LoginQQ"] = true
		oauthLogin = true
	}
	if v, ok := options["LOGIN_GITHUB"]; ok && strings.EqualFold(v, "true") {
		c.Tpl().Data["LoginGitHub"] = true
		oauthLogin = true
	}
	if v, ok := options["LOGIN_GITEE"]; ok && strings.EqualFold(v, "true") {
		c.Tpl().Data["LoginGitee"] = true
		oauthLogin = true
	}
	c.Tpl().Data["OauthLogin"] = oauthLogin

	//如果Cookie中存在登录信息
	//if cookie, ok := this.GetSecureCookie(conf.GetAppKey(), "login"); ok {
	//	if err := utils.Decode(cookie, &remember); err == nil {
	//		if err = this.loginByMemberId(remember.MemberId); err == nil {
	//			this.Redirect(beego.URLFor("HomeController.Index"), 302)
	//			return
	//		}
	//	}
	//}

	c.Tpl().Data["GiteeClientId"] = beego.AppConfig.String("oauth::giteeClientId")
	c.Tpl().Data["GiteeCallback"] = beego.AppConfig.String("oauth::giteeCallback")
	c.Tpl().Data["GithubClientId"] = beego.AppConfig.String("oauth::githubClientId")
	c.Tpl().Data["GithubCallback"] = beego.AppConfig.String("oauth::githubCallback")
	c.Tpl().Data["QQClientId"] = beego.AppConfig.String("oauth::qqClientId")
	c.Tpl().Data["QQCallback"] = beego.AppConfig.String("oauth::qqCallback")
	c.Tpl().Data["RandomStr"] = time.Now().Unix()
	c.GetSeoByPage("login", map[string]string{
		"title":       "登录 - " + dao.Global.Get(constx.SITE_NAME),
		"keywords":    "登录," + dao.Global.Get(constx.SITE_NAME),
		"description": dao.Global.Get(constx.SITE_NAME) + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Html("account/login")
}

//找回密码.
func FindPasswordHtml(c *core.Context) {
	mailConf := conf.GetMailConfig()
	c.GetSeoByPage("findpwd", map[string]string{
		"title":       "找回密码 - " + dao.Global.GetSiteName(),
		"keywords":    "找回密码",
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})

	token := c.GetString("token")
	mail := c.GetString("mail")

	if token != "" && mail != "" {
		memberToken, err := mysql.NewMemberToken().FindByFieldFirst("token", token)
		if err != nil {
			c.Tpl().Data["ErrorMessage"] = "邮件已失效"
			c.Html("errors/error")
			return
		}
		subTime := memberToken.SendTime.Sub(time.Now())

		if !strings.EqualFold(memberToken.Email, mail) || subTime.Minutes() > float64(mailConf.MailExpired) || !memberToken.ValidTime.IsZero() {
			c.Tpl().Data["ErrorMessage"] = "验证码已过期，请重新操作。"
			c.Html("errors/error")
			return
		}
		c.Tpl().Data["Email"] = memberToken.Email
		c.Tpl().Data["Token"] = memberToken.Token
		c.Html("account/find_password_setp2")
		return

	}
	c.Html("account/find_password_setp1")

}
