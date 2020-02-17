package localhost

import (
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/orm"
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/router/core"
	"go.uber.org/zap"
)

//渲染markdown.
//根据文档id来。
func RenderMarkdownHtml(c *core.Context) {
	if c.Context.Request.Host != "localhost:9011" {
		c.Html404()
		return
	}

	idStr, _ := c.GetQuery("id")
	id, _ := strconv.Atoi(idStr)
	if id > 0 {
		c.Tpl().Data["Markdown"] = dao.DocumentStore.GetFiledById(id, "markdown")
		c.Html("widgets/render")
		return
	}
}

func RenderMarkdownApi(c *core.Context) {
	if c.Context.Request.Host != "localhost:9011" {
		c.Html404()
		return
	}

	idStr, _ := c.GetQuery("id")
	id, _ := strconv.Atoi(idStr)
	if id <= 0 {
		c.Html404()
		return
	}
	var doc mysql.Document
	mus.Db.Where("document_id = ?", id).Find(&doc)

	var book mysql.Book
	mus.Db.Where("book_id = ?", doc.BookId).Find(&book)
	content, _ := c.GetPostForm("content")

	docQuery, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err == nil {
		docQuery.Find("br").Each(func(i int, selection *goquery.Selection) {
			selection.Remove()
		})
		content, _ = docQuery.Find("body").Html()
	}

	content = c.ReplaceLinks(book.Identify, content)
	mus.Db.Model(mysql.Document{}).Where("document_id = ?", id).Update(orm.Params{
		"release":     content,
		"modify_time": time.Now(),
	})
	//这里要指定更新字段，否则markdown内容会被置空
	err = dao.DocumentStore.InsertOrUpdate(mus.Db, &mysql.DocumentStore{DocumentId: id, Content: content})
	if err != nil {
		mus.Logger.Error("render markdown error", zap.Error(err))
		return
	}
	c.JSONOK()

}

// 渲染生成封面截图
func RenderCover(c *core.Context) {
	if c.Context.Request.Host != "localhost:9011" {
		c.Html404()
		return
	}

	identify, _ := c.GetQuery("id")
	id, err := strconv.Atoi(identify)
	if identify == "" && err != nil {
		c.Html404()
		return
	}
	book := mysql.Book{}
	if id > 0 {
		err = mus.Db.Where("book_id = ?", id).Find(&book).Error
	} else {
		err = mus.Db.Where("identify = ?", identify).Find(&book).Error
	}
	if err != nil {
		mus.Logger.Error(err.Error())
		return
	}
	if book.BookId == 0 {
		c.Html404()
		return
	}
	c.Tpl().Data["Book"] = book
	c.Html("ebook/cover")
}
