package label

import (
	"math"

	"github.com/astaxie/beego"
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/goecology/ecology/appgo/router/core"
)

// 查看包含标签的文档列表.
func Index(c *core.Context) {
	c.Tpl().Data["IsLabel"] = true

	labelName := c.Param(":key")

	c.Redirect(302, beego.URLFor("SearchController.Result")+"?wd="+labelName)
	return
}

// 标签列表
func List(c *core.Context) {
	c.Tpl().Data["IsLabel"] = true

	pageIndex := c.GetInt("page")
	pageSize := 200

	labels, totalCount, err := dao.Label.FindToPager(pageIndex, pageSize)
	if err != nil {
		c.Html404()
		return
	}
	if totalCount > 0 {
		html := utils.NewPaginations(conf.Conf.Info.RollPage, totalCount, pageSize, pageIndex, beego.URLFor("LabelController.List"), "")
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}
	c.Tpl().Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))

	c.Tpl().Data["Labels"] = labels
	c.GetSeoByPage("label_list", map[string]string{
		"title":       "标签",
		"keywords":    "标签",
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Html("label/list")

}
