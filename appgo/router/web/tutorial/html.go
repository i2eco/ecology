package tutorial

import (
	"math"

	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/constx"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/ecology/appgo/router/core"
)

func Index(c *core.Context) {
	tutorialType := c.Param("key")
	// todo 判断条件，现在先实现
	if tutorialType == "" {
		tutorialType = "go"
	}

	req := ReqIndex{}
	err := c.Bind(&req)
	if err != nil {
		// 首页就用默认值
	}
	pageIndex := req.Page
	if pageIndex == 0 {
		pageIndex = 1
	}

	c.Tpl().Data["IsTutorial"] = true

	//每页显示24个，为了兼容Pad、mobile、PC
	pageSize := 24
	books, totalCount, err := dao.Book.HomeData(pageIndex, pageSize, mysql.OrderLatest, "tutorial-"+tutorialType, 0)
	if err != nil {
		c.Html404()
		return
	}
	if totalCount > 0 {
		html := utils.NewPaginations(conf.RollPage, totalCount, pageSize, pageIndex, "/tutorial", "")
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}

	c.Tpl().Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))

	for _, book := range books {
		book.DealAll()
	}

	c.Tpl().Data["TutorialType"] = "tutorial-" + tutorialType

	c.Tpl().Data["Lists"] = books
	title := dao.Global.Get(constx.SITE_NAME)

	title = "探索，发现新世界，畅想新知识 - " + dao.Global.Get(constx.SITE_NAME)
	c.GetSeoByPage("index", map[string]string{
		"title":       title,
		"keywords":    "文档托管,在线创作,文档在线管理,在线知识管理,文档托管平台,在线写书,文档在线转换,在线编辑,在线阅读,开发手册,api手册,文档在线学习,技术文档,在线编辑",
		"description": dao.Global.Get(constx.SITE_NAME) + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Html("tutorial/index")
}
