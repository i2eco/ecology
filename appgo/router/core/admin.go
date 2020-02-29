package core

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/i2eco/ecology/appgo/model/mysql"
)

// 后台取用户
func AdminLoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		one := &mysql.User{}
		// 从session中获取用户信息
		if user, ok := AdminSessionUser(c); ok && user.Id > 0 {
			one = user
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": 401,
				"data": "",
				"msg":  "user not login",
			})
			c.Abort()
			return
		}
		c.Set(AdminContextKey, one)
		c.Next()
	}
}

// 后台取用户
func AdminSessionUser(c *gin.Context) (*mysql.User, bool) {
	resp, flag := sessions.Default(c).Get(AdminSessionKey).(*mysql.User)
	return resp, flag
}

// 后台取用户
func AdminContextUser(c *gin.Context) *mysql.User {
	resp := &mysql.User{}
	respI, flag := c.Get(AdminContextKey)
	if flag {
		resp = respI.(*mysql.User)
	}
	return resp
}

// Authed 鉴权通过
func (c *Context) AdminAuthed() bool {
	if user, ok := AdminSessionUser(c.Context); ok && user.Id > 0 {
		return true
	}
	return false
}

// 后台 Uid 返回uid
func (c *Context) AdminUid() int {
	return AdminContextUser(c.Context).Id
}

// UpdateUser updates the User object stored in the session. This is useful incase a change
// is made to the user model that needs to persist across requests.
func (c *Context) AdminUpdateUser(a *mysql.User) error {
	s := sessions.Default(c.Context)
	s.Options(sessions.Options{
		Path:     "/",
		MaxAge:   24 * 3600,
		Secure:   false,
		HttpOnly: true,
	})
	s.Set(AdminSessionKey, a)
	return s.Save()
}

// Logout will clear out the session and call the Logout() user function.
func (c *Context) AdminLogout() error {
	s := sessions.Default(c.Context)
	s.Options(sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		Secure:   false,
		HttpOnly: true,
	})

	s.Delete(AdminSessionKey)
	return s.Save()
}
