package core

import (
	"github.com/astaxie/beego"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/muses/pkg/system"
	"github.com/i2eco/muses/pkg/tpl/tplbeego"
	"github.com/spf13/viper"
	"strings"
)

func FrontTplRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tpl, err := tplbeego.Caller()
		if err != nil {
			c.AbortWithStatus(401)
			return
		}
		options := dao.Global.AllOptions()
		for key, value := range options {
			tpl.Data[key] = value
		}
		tpl.Data["Member"] = FrontContextUser(c)
		tpl.Data["BaseUrl"] = BaseUrl(c)
		tpl.Data["BaiduTongji"] = viper.GetString("app.baidutongji")
		tpl.Data["Version"] = system.BuildInfo.Version
		isMobile := utils.IsMobile(c.Request.UserAgent())
		tpl.Data["IsMobile"] = isMobile
		//this.Member = mysql.NewMember() //初始化
		//this.EnableAnonymous = false
		//this.EnableDocumentHistory = 0
		tpl.Data["OssDomain"] = strings.TrimRight(beego.AppConfig.String("oss::Domain"), "/ ")
		tpl.Data["StaticDomain"] = strings.Trim(beego.AppConfig.DefaultString("static_domain", ""), "/")
		////从session中获取用户信息
		//if member, ok := this.GetSession(conf.LoginSessionName).(models.Member); ok && member.MemberId > 0 {
		//	m, _ := models.NewMember().Find(member.MemberId)
		//	this.Member = m
		//} else {
		//	//如果Cookie中存在登录信息，从cookie中获取用户信息
		//	if cookie, ok := this.GetSecureCookie(conf.GetAppKey(), "login"); ok {
		//		var remember CookieRemember
		//		err := utils.Decode(cookie, &remember)
		//		if err == nil {
		//			member, err := models.NewMember().Find(remember.MemberId)
		//			if err == nil {
		//				this.SetMember(*member)
		//				this.Member = member
		//			}
		//		}
		//	}
		//
		//}
		//if this.Member.RoleName == "" {
		//	this.Member.ResolveRoleName()
		//}
		//this.Data["Member"] = this.Member
		//this.Data["BaseUrl"] = this.BaseUrl()
		tpl.Data["IsSignedToday"] = false
		//if this.Member.MemberId > 0 {
		//	this.Data["IsSignedToday"] = models.NewSign().IsSignToday(this.Member.MemberId)
		//}
		//if options, err := models.NewOption().All(); err == nil {
		//	this.Option = make(map[string]string, len(options))
		//	for _, item := range options {
		//		if item.OptionName == "SITE_NAME" {
		//			this.Sitename = item.OptionValue
		//		}
		//		this.Data[item.OptionName] = item.OptionValue
		//		this.Option[item.OptionName] = item.OptionValue
		//		if strings.EqualFold(item.OptionName, "ENABLE_ANONYMOUS") && item.OptionValue == "true" {
		//			this.EnableAnonymous = true
		//		}
		//		if verNum, _ := strconv.Atoi(item.OptionValue); strings.EqualFold(item.OptionName, "ENABLE_DOCUMENT_HISTORY") && verNum > 0 {
		//			this.EnableDocumentHistory = verNum
		//		}
		//	}
		//}

		//this.Data["Friendlinks"] = new(models.FriendLink).GetList(false)

		if value := dao.Global.Get("CLOSE_OPEN_SOURCE_LINK"); value != "" {
			tpl.Data["CloseOpenSourceLink"] = value == "true"
		}

		if value := dao.Global.Get("HIDE_TAG"); value != "" {
			tpl.Data["HideTag"] = value == "true"
		}

		if value := dao.Global.Get("CLOSE_SUBMIT_ENTER"); value != "" {
			tpl.Data["CloseSubmitEnter"] = value == "true"
		}

		if value := dao.Global.Get("SITE_NAME"); value != "" {
			tpl.Data["SiteName"] = value
		}
		c.Set(Options, options)
		c.Set(TPL, tpl)
		c.Next()
	}
}

// 后台取用户
func FrontLoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		one := &mysql.Member{}
		// 从session中获取用户信息
		if user, ok := FrontSessionUser(c); ok && user.MemberId > 0 {
			one = user
		}
		c.Set(FrontContextKey, one)
		c.Next()
	}
}

// 后台取用户
func FrontSessionUser(c *gin.Context) (*mysql.Member, bool) {
	resp, flag := sessions.Default(c).Get(FrontSessionKey).(*mysql.Member)
	return resp, flag
}

// 后台取用户
func FrontContextUser(c *gin.Context) *mysql.Member {
	resp := &mysql.Member{}
	respI, flag := c.Get(FrontContextKey)
	if flag {
		resp = respI.(*mysql.Member)
	}
	return resp

}

func BaseUrl(c *gin.Context) string {
	return viper.GetString("app.scheme") + "://" + c.Request.Host
}
