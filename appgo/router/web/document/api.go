package document

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/code"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/goecology/ecology/appgo/router/core"
	"github.com/kataras/iris/core/errors"
)

//阅读文档.
func ReadApi(c *core.Context) {
	identify := c.Param("key")
	token := c.Query("token")
	id := c.Param("id")

	if identify == "" || id == "" {
		c.JSONErr(code.MsgErr, errors.New("err"))
		return
	}

	//如果没有开启匿名访问则跳转到登录
	if dao.Global.IsEnableAnonymous() && c.Member() == nil {
		c.Redirect(302, "/login")
		return
	}
	var err error
	bookResult := isReadable(c, identify, token)

	var doc *mysql.Document
	if docId, _ := strconv.Atoi(id); docId > 0 {
		doc, err = dao.Document.Find(docId) //文档id
		if err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
	} else {
		//此处的id是字符串，标识文档标识，根据文档标识和文档所属的书的id作为key去查询
		doc, err = dao.Document.FindByBookIdAndDocIdentify(bookResult.BookId, id) //文档标识
		if err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
	}

	if doc.BookId != bookResult.BookId {
		c.JSONErr(code.MsgErr, err)
		return
	}

	if doc.Release != "" {
		query, err := goquery.NewDocumentFromReader(bytes.NewBufferString(doc.Release))
		if err != nil {
			mus.Logger.Error(err.Error())
		} else {
			query.Find("img").Each(func(i int, contentSelection *goquery.Selection) {
				if src, ok := contentSelection.Attr("src"); ok {
					if utils.StoreType == utils.StoreOss && !(strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "http://")) {
						src = conf.Conf.Oss.Domain + "/" + strings.TrimLeft(src, "./")
						contentSelection.SetAttr("src", src)
					}
				}
				if alt, _ := contentSelection.Attr("alt"); alt == "" {
					contentSelection.SetAttr("alt", doc.DocumentName+" - 图"+fmt.Sprint(i+1))
				}
			})
			html, err := query.Find("body").Html()
			if err != nil {
				mus.Logger.Error(err.Error())
			} else {
				doc.Release = html
			}
		}
		//bodyText = query.Find(".markdown-toc").Text()
	}

	attach, err := dao.Attachment.FindListByDocumentId(doc.DocumentId)
	if err == nil {
		doc.AttachList = attach
	}

	//文档阅读人次+1
	if err := dao.SetIncreAndDecre("md_documents", "vcnt",
		fmt.Sprintf("document_id=%v", doc.DocumentId),
		true, 1,
	); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	//项目阅读人次+1
	if err := dao.SetIncreAndDecre("md_books", "vcnt",
		fmt.Sprintf("book_id=%v", doc.BookId),
		true, 1,
	); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	if c.Member().MemberId > 0 { //增加用户阅读记录
		if err := mysql.NewReadRecord().Add(doc.DocumentId, c.Member().MemberId); err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
	}

	existBookmark := dao.Bookmark.Exist(c.Member().MemberId, doc.DocumentId)
	doc.Vcnt = doc.Vcnt + 1

	var data struct {
		Id        int    `json:"docId"`
		DocTitle  string `json:"docTitle"`
		Body      string `json:"body"`
		Title     string `json:"title"`
		Bookmark  bool   `json:"bookmark"` //是否已经添加了书签
		View      int    `json:"view"`
		UpdatedAt string `json:"updatedAt"`
	}
	data.DocTitle = doc.DocumentName
	data.Body = doc.Release
	data.Id = doc.DocumentId
	//data.Title = tpl.Data["SeoTitle"].(string)
	data.Bookmark = existBookmark
	data.View = doc.Vcnt
	data.UpdatedAt = doc.ModifyTime.Format("2006-01-02 15:04:05")
	//data.Body = doc.Markdown
	c.JSONOK(data)
}

//创建一个文档.
/*
identify: test6
doc_id: 0
parent_id: 0
doc_name: 书本试卷
doc_identify:
*/

func CreateApi(c *core.Context) {
	identify := c.PostForm("identify") //书籍项目标识
	docName := c.PostForm("doc_name")
	parentId, _ := strconv.Atoi(c.PostForm("parent_id"))

	if identify == "" {
		c.JSONCode(code.MsgErr)
		return
	}
	if docName == "" {
		c.JSONCode(code.MsgErr)
		return
	}

	currentMember := c.Member()
	docIdentify := fmt.Sprintf("date-%v", time.Now().Format("2006.01.02.15.04.05"))

	bookId := 0
	//如果是超级管理员则不判断权限
	if currentMember.IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
		bookId = book.BookId
	} else {
		bookResult, err := dao.Book.ResultFindByIdentify(identify, currentMember.MemberId)

		if err != nil || bookResult.RoleId == conf.Conf.Info.BookObserver {
			c.JSONCode(code.MsgErr)
			return
		}
		bookId = bookResult.BookId
	}

	if parentId > 0 {
		doc, err := dao.Document.Find(parentId)
		if err != nil || doc.BookId != bookId {
			c.JSONCode(code.MsgErr)
			return
		}
	}
	var document mysql.Document

	document.MemberId = currentMember.MemberId
	document.BookId = bookId
	if docIdentify != "" {
		document.Identify = docIdentify
	}
	document.Version = time.Now().Unix()
	document.DocumentName = docName
	document.ParentId = parentId

	docIdInt64, err := dao.Document.InsertOrUpdate(mus.Db, &document)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	if dao.DocumentStore.GetFiledById(docIdInt64, "markdown") == "" {
		//因为创建和更新文档基本信息都调用的这个接口，先判断markdown是否有内容，没有内容则添加默认内容
		if err := dao.DocumentStore.InsertOrUpdate(mus.Db, &mysql.DocumentStore{DocumentId: int(docIdInt64), Markdown: "[TOC]\n\r\n\r"}); err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
	}
	c.JSONOK(document)
}

func UpdateApi(c *core.Context) {
	docId, _ := strconv.Atoi(c.PostForm("doc_id"))
	docIdentify := c.PostForm("doc_identify") //新建的文档标识

	if docIdentify == "" {
		c.JSONCode(code.MsgErr)
		return
	}

	identify := c.PostForm("identify") //书籍项目标识

	docName := c.PostForm("doc_name")
	parentId, _ := strconv.Atoi(c.PostForm("parent_id"))

	bookIdentify := strings.TrimSpace(c.Param("key"))

	if identify == "" {
		c.JSONCode(code.MsgErr)
		return
	}
	if docName == "" {
		c.JSONCode(code.MsgErr)
		return
	}

	currentMember := c.Member()

	if ok, err := regexp.MatchString(`^[a-zA-Z0-9_\-\.]*$`, docIdentify); !ok || err != nil {
		c.JSONCode(code.MsgErr)
		return
	}
	if num, _ := strconv.Atoi(docIdentify); docIdentify == "0" || strconv.Itoa(num) == docIdentify { //不能是纯数字
		c.JSONCode(code.MsgErr)
		return
	}

	if bookIdentify == "" {
		c.JSONCode(code.MsgErr)
		return
	}

	var book mysql.Book

	mus.Db.Select("book_id").Where("identify = ?", bookIdentify).Find(&book)
	if book.BookId == 0 {
		c.JSONCode(code.MsgErr)
		return
	}

	d, _ := dao.Document.FindByBookIdAndDocIdentify(book.BookId, docIdentify)
	if d.DocumentId > 0 && d.DocumentId != docId {
		c.JSONCode(code.MsgErr)
		return
	}

	bookId := 0
	//如果是超级管理员则不判断权限
	if currentMember.IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
		bookId = book.BookId
	} else {
		bookResult, err := dao.Book.ResultFindByIdentify(identify, currentMember.MemberId)

		if err != nil || bookResult.RoleId == conf.Conf.Info.BookObserver {
			c.JSONCode(code.MsgErr)
			return
		}
		bookId = bookResult.BookId
	}

	if parentId > 0 {
		doc, err := dao.Document.Find(parentId)
		if err != nil || doc.BookId != bookId {
			c.JSONCode(code.MsgErr)
			return
		}
	}

	var document mysql.Document

	document.DocumentId = docId
	document.MemberId = currentMember.MemberId
	document.BookId = bookId
	if docIdentify != "" {
		document.Identify = docIdentify
	}
	document.Version = time.Now().Unix()
	document.DocumentName = docName
	document.ParentId = parentId

	docIdInt64, err := dao.Document.InsertOrUpdate(mus.Db, &document)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	if dao.DocumentStore.GetFiledById(docIdInt64, "markdown") == "" {
		//因为创建和更新文档基本信息都调用的这个接口，先判断markdown是否有内容，没有内容则添加默认内容
		if err := dao.DocumentStore.InsertOrUpdate(mus.Db, &mysql.DocumentStore{DocumentId: int(docIdInt64), Markdown: "[TOC]\n\r\n\r"}); err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
	}
	c.JSONOK(document)
}
