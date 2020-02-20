package setting

import (
	"strconv"

	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/ecology/appgo/router/core"
)

//基本信息
func Index(c *core.Context) {
	c.Tpl().Data["SeoTitle"] = "基本信息 - " + dao.Global.GetSiteName()
	c.Tpl().Data["SettingBasic"] = true
	c.Html("setting/index")
}

//修改密码
func Password(c *core.Context) {
	c.Tpl().Data["SettingPwd"] = true
	c.Tpl().Data["SeoTitle"] = "修改密码 - " + dao.Global.GetSiteName()
	c.Html("setting/password")
}

//收藏
func Star(c *core.Context) {
	pageStr, _ := c.GetQuery("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	sort, _ := c.GetQuery("sort")
	if sort == "" {
		sort = "read"
	}

	cnt, books, _ := dao.Star.ListXX(c.Member().MemberId, page, conf.PageSize, sort)
	if cnt > 1 {
		c.Tpl().Data["PageHtml"] = utils.NewPaginations(conf.RollPage, int(cnt), conf.PageSize, page, "/setting/star", "")
	}
	c.Tpl().Data["Books"] = books
	c.Tpl().Data["Sort"] = sort
	c.Tpl().Data["SettingStar"] = true
	c.Tpl().Data["SeoTitle"] = "我的收藏 - " + dao.Global.GetSiteName()
	c.Html("setting/star")
}

//二维码
func Qrcode(c *core.Context) {
	c.Tpl().Data["SeoTitle"] = "二维码管理 - " + dao.Global.GetSiteName()
	c.Tpl().Data["Qrcode"] = dao.Member.GetQrcodeByUid(c.Member().MemberId)
	c.Tpl().Data["SettingQrcode"] = true
	c.Html("setting/qrcode")
}
