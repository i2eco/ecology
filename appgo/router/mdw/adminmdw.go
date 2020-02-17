package mdw

import (
	"encoding/gob"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/router/types"
)

func init() {
	gob.Register(&mysql.Member{})
}

// 后台取用户
func LoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		one := &mysql.Member{}
		// 从session中获取用户信息
		if user, ok := DefaultSessionUser(c); ok && user.MemberId > 0 {
			one = user
		} else {

		}
		c.Set(types.ContextUser, one)
		c.Next()
	}
}

// 后台取用户
func DefaultSessionUser(c *gin.Context) (*mysql.Member, bool) {
	resp, flag := sessions.Default(c).Get(types.SessionDefaultKey).(*mysql.Member)
	return resp, flag
}

// 后台取用户
func DefaultContextUser(c *gin.Context) *mysql.Member {
	var resp *mysql.Member
	respI, flag := c.Get(types.ContextUser)
	if flag {
		resp = respI.(*mysql.Member)
	}
	return resp

}

func BaseUrl(c *gin.Context) string {
	host := conf.Conf.App.SitemapHost
	if len(host) > 0 {
		if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
			return host
		}
		return c.Request.URL.Scheme + "://" + host
	}
	return c.Request.URL.Scheme + "://" + c.Request.URL.Host
}
