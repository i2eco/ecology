package account

import (
	"context"
	"github.com/astaxie/beego"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/router/core"
	"github.com/spf13/viper"
	"net/http"
	"strings"
	"time"
)

//第三方登录回调
//封装一个内部调用的函数，loginByMemberId
func OauthInfo(c *core.Context) {
	oa := c.Param("oauth")
	switch oa {
	case "gitee":
	case "github":
		c.Redirect(http.StatusFound, mus.GithubClient.Oauth.AuthCodeURL(""))
		return
	case "qq":

	default: //email

	}

	return

}

// Login 用户登录.
func LoginCallback(c *core.Context) {
	var (
		nickname string //昵称
		avatar   string //头像的http链接地址
		email    string //邮箱地址
		username string //用户名
		tips     string
		id       interface{} //第三方的用户id，唯一识别码
		IsEmail  bool        //是否是使用邮箱注册
	)

	oa := c.Param("oauth")
	code := c.Query("code")
	if code == "" {
		c.Html404()
		return
	}

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
		ctx := context.Background()
		token, err := mus.GithubClient.Oauth.Exchange(ctx, code)
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Html404()
			return
		}
		// 获取登录用户信息。
		user, err := mus.GithubClient.GetUserInfo(token.AccessToken)
		if err != nil {
			c.Html404()
			return
		}

		if user.ID <= 0 {
			c.Html404()
			return
		}

		userMysql, _ := dao.GithubUser.Info(c.Context, user.ID)
		if userMysql.Id > 0 && userMysql.MemberId > 0 { //直接登录
			err = c.LoginByMemberId(userMysql.MemberId)
			if err != nil {
				c.Html404()
				return
			}
			c.Redirect(302, "/")
			return
		}

		if userMysql.Id == 0 {
			// 说明用户不存在
			err = dao.GithubUser.Create(c.Context, mus.Db, &mysql.GithubUser{
				Id:        user.ID,
				MemberId:  0,
				UpdatedAt: time.Now(),
				AvatarURL: user.AvatarURL,
				Email:     user.Email,
				Login:     user.Login,
				Name:      user.Name,
				HtmlURL:   user.HTMLURL,
			})
			if err != nil {
				c.Html404()
				return
			}
		}
		nickname = user.Name
		username = user.Login
		avatar = user.AvatarURL
		email = user.Email
		id = user.ID
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

	c.Html("account/login")
}

//找回密码.
func FindPasswordHtml(c *core.Context) {
	token := c.Query("token")
	mail := c.Query("mail")

	if token != "" && mail != "" {
		memberToken, err := dao.MemberToken.FindByFieldFirst("token", token)
		if err != nil {
			c.Tpl().Data["ErrorMessage"] = "邮件已失效"
			c.Html("errors/error")
			return
		}
		subTime := memberToken.SendTime.Sub(time.Now())

		if !strings.EqualFold(memberToken.Email, mail) {
			c.Tpl().Data["ErrorMessage"] = "邮箱不正确。"
			c.Html("errors/error")
			return
		}

		if subTime.Minutes() > float64(viper.GetInt("email.mailExpired")) || !memberToken.ValidTime.IsZero() {
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
