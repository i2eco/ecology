package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/TruthHun/BookStack/conf"
	"github.com/TruthHun/BookStack/mysql"
	"github.com/TruthHun/BookStack/utils"
	"github.com/astaxie/beego"
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/pkg/mus"
)

//搜索首页
func (this *SearchController) Search() {
	if wd := strings.TrimSpace(this.GetString("wd")); wd != "" {
		this.Redirect(beego.URLFor("LabelController.Index", ":key", wd), 302)
		return
	}
	c.Tpl().Data["SeoTitle"] = "搜索 - " + dao.Global.GetSiteName()
	c.Tpl().Data["IsSearch"] = true
	this.TplName = "search/search.html"
}

// 搜索结果页
func (this *SearchController) Result() {

	totalRows := 0

	var ids []int

	wd := this.GetString("wd")
	if wd == "" {
		this.Redirect(beego.URLFor("SearchController.Search"), 302)
		return
	}

	now := time.Now()

	tab := this.GetString("tab", mysql.GetOptionValue("DEFAULT_SEARCH", "book"))
	isSearchDoc := false
	if tab == "doc" {
		isSearchDoc = true
	}

	page, _ := this.GetInt("page", 1)
	size := 10

	if page < 1 {
		page = 1
	}

	client := mysql.NewElasticSearchClient()

	if client.On { // elasticsearch 进行全文搜索
		result, err := mysql.NewElasticSearchClient().Search(wd, page, size, isSearchDoc)
		if err != nil {
			mus.Logger.Error(err.Error())
		} else { // 搜索结果处理
			totalRows = result.Hits.Total
			for _, item := range result.Hits.Hits {
				ids = append(ids, item.Source.Id)
			}
		}
	} else { //MySQL like 查询
		if isSearchDoc { //搜索文档
			docs, count, err := mysql.NewDocumentSearchResult().SearchDocument(wd, 0, page, size)
			totalRows = count
			if err != nil {
				mus.Logger.Error(err.Error())
			} else {
				for _, doc := range docs {
					ids = append(ids, doc.DocumentId)
				}
			}
		} else { //搜索书籍
			books, count, err := dao.Book.SearchBook(wd, page, size)
			totalRows = count
			if err != nil {
				mus.Logger.Error(err.Error())
			} else {
				for _, book := range books {
					ids = append(ids, book.BookId)
				}
			}
		}
	}
	if len(ids) > 0 {
		if isSearchDoc {
			c.Tpl().Data["Docs"], _ = mysql.NewDocumentSearchResult().GetDocsById(ids)
		} else {
			c.Tpl().Data["Books"], _ = dao.Book.GetBooksById(ids)
		}
		c.Tpl().Data["Words"] = client.SegWords(wd)
	}

	c.Tpl().Data["TotalRows"] = totalRows
	if totalRows > size {
		if totalRows > 1000 {
			totalRows = 1000
		}
		urlSuffix := fmt.Sprintf("&tab=%v&wd=%v", tab, wd)
		html := utils.NewPaginations(conf.RollPage, totalRows, size, page, beego.URLFor("SearchController.Result"), urlSuffix)
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}
	c.Tpl().Data["SpendTime"] = fmt.Sprintf("%.3f", time.Since(now).Seconds())
	c.Tpl().Data["Wd"] = wd
	c.Tpl().Data["Tab"] = tab
	c.Tpl().Data["IsSearch"] = true
	this.TplName = "search/result.html"
}
