package home

import (
	"math"
	"strconv"
	"strings"

	"github.com/goecology/ecology/appgo/dao"

	"github.com/goecology/ecology/appgo/model/constx"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/goecology/ecology/appgo/router/core"
)

func Index(c *core.Context) {
	//tab
	var (
		cid       int //分类，如果只是一级分类，则忽略，二级分类，则根据二级分类查找内容
		urlPrefix = "/"
		cate      mysql.Category
		lang      = c.GetString("lang")
		tabName   = map[string]string{"recommend": "站长推荐", "latest": "最新发布", "popular": "热门书籍"}
	)
	tab, _ := c.GetQuery("tab")
	tab = strings.ToLower(tab)
	switch tab {
	case "recommend", "popular", "latest":
	default:
		tab = "latest"
	}

	cates, _ := dao.Category.GetCates(c.Context, -1, 1)
	cid = c.GetInt("cid")
	pid := cid
	if cid > 0 {
		for _, item := range cates {
			if item.Id == cid {
				if item.Pid > 0 {
					pid = item.Pid
				}
				c.Tpl().Data["Cate"] = item
				cate = item
				break
			}
		}
	}
	c.Tpl().Data["Cates"] = cates
	c.Tpl().Data["Cid"] = cid
	c.Tpl().Data["Pid"] = pid
	c.Tpl().Data["IsHome"] = true

	pageIndex := c.GetInt("page")
	if pageIndex == 0 {
		pageIndex = 1
	}
	//每页显示24个，为了兼容Pad、mobile、PC
	pageSize := 24
	books, totalCount, err := dao.Book.HomeData(pageIndex, pageSize, mysql.BookOrder(tab), lang, cid)
	if err != nil {
		c.Html404()
		return
	}
	if totalCount > 0 {
		urlSuffix := "&tab=" + tab
		if cid > 0 {
			urlSuffix = urlSuffix + "&cid=" + strconv.Itoa(cid)
		}
		urlSuffix = urlSuffix + "&lang=" + lang
		html := utils.NewPaginations(conf.RollPage, totalCount, pageSize, pageIndex, urlPrefix, urlSuffix)
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}

	c.Tpl().Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))

	for _, book := range books {
		book.DealCover()
	}

	c.Tpl().Data["Lists"] = books
	c.Tpl().Data["Tab"] = tab
	c.Tpl().Data["Lang"] = lang
	title := dao.Global.Get(constx.SITE_NAME)

	if cid > 0 {
		title = "[发现] " + cate.Title + " - " + tabName[tab] + " - " + title
	} else {
		title = "探索，发现新世界，畅想新知识 - " + dao.Global.Get(constx.SITE_NAME)
	}
	c.GetSeoByPage("index", map[string]string{
		"title":       title,
		"keywords":    "文档托管,在线创作,文档在线管理,在线知识管理,文档托管平台,在线写书,文档在线转换,在线编辑,在线阅读,开发手册,api手册,文档在线学习,技术文档,在线编辑",
		"description": dao.Global.Get(constx.SITE_NAME) + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Html("home/index")

}
