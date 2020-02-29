package auth

import (
	"regexp"

	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/model/trans"
	"github.com/i2eco/ecology/appgo/router/core"
)

var emailRgx = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-z]{2,4}$`)

// {status: "error", type: "account", currentAuthority: "guest"}
func Login(c *core.Context) {
	var err error
	// 如果已经登录
	respView := trans.RespOauthLogin{
		CurrentAuthority: "admin",
	}

	if c.AdminAuthed() {
		c.JSONOK(respView)
		return
	}

	reqView := &trans.ReqOauthLogin{}
	err = c.Bind(reqView)
	if err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}

	// 对Identity进行校验，先判断是否是邮箱，若不是邮箱则当做用户名
	var oneUser *mysql.User
	if emailRgx.MatchString(reqView.Name) {
		oneUser, err = dao.User.GetBizByPwd("", reqView.Name, reqView.Pwd, c.ClientIP())
		if err != nil {
			c.JSONErrTips("邮箱错误", err)
			return
		}
	} else {
		oneUser, err = dao.User.GetBizByPwd(reqView.Name, "", reqView.Pwd, c.ClientIP())
		if err != nil {
			c.JSONErrTips("昵称错误", err)
			return
		}
	}
	err = c.AdminUpdateUser(oneUser)
	if err != nil {
		c.JSONErrTips("更新失败", err)
		return
	}
	c.JSONOK(respView)
	return
}

func Logout(c *core.Context) {
	err := c.Logout()
	if err != nil {
		c.JSONErrTips("登出失败", err)
		return
	}
	c.JSONOK()
	return

}

func Self(c *core.Context) {
	resp, err := dao.User.Info(c.Context, c.AdminUid())
	if err != nil {
		c.JSONErrTips("获取用户信息失败", err)
		return
	}
	c.JSONOK(resp)
	return
}
