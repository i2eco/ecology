package document

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/code"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/ecology/appgo/router/core"
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
	if err := dao.SetIncreAndDecre(mysql.Document{}.TableName(), "vcnt",
		fmt.Sprintf("document_id=%v", doc.DocumentId),
		true, 1,
	); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	//项目阅读人次+1
	if err := dao.SetIncreAndDecre(mysql.Book{}.TableName(), "vcnt",
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
		c.JSONCode(code.DocUpdateErr1)
		return
	}

	identify := c.PostForm("identify") //书籍项目标识

	docName := c.PostForm("doc_name")
	parentId, _ := strconv.Atoi(c.PostForm("parent_id"))

	bookIdentify := strings.TrimSpace(c.Param("key"))

	if identify == "" {
		c.JSONCode(code.DocUpdateErr2)
		return
	}
	if docName == "" {
		c.JSONCode(code.DocUpdateErr3)
		return
	}

	currentMember := c.Member()

	if ok, err := regexp.MatchString(`^[a-zA-Z0-9_\-\.]*$`, docIdentify); !ok || err != nil {
		c.JSONCode(code.DocUpdateErr4)
		return
	}
	if num, _ := strconv.Atoi(docIdentify); docIdentify == "0" || strconv.Itoa(num) == docIdentify { //不能是纯数字
		c.JSONCode(code.DocUpdateErr5)
		return
	}

	if bookIdentify == "" {
		c.JSONCode(code.DocUpdateErr6)
		return
	}

	var book mysql.Book

	mus.Db.Select("book_id").Where("identify = ?", bookIdentify).Find(&book)
	if book.BookId == 0 {
		c.JSONCode(code.DocUpdateErr7)
		return
	}

	d, _ := dao.Document.FindByBookIdAndDocIdentify(book.BookId, docIdentify)
	if d.DocumentId > 0 && d.DocumentId != docId {
		c.JSONCode(code.DocUpdateErr8)
		return
	}

	bookId := 0
	//如果是超级管理员则不判断权限
	if currentMember.IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			c.JSONErr(code.DocUpdateErr9, err)
			return
		}
		bookId = book.BookId
	} else {
		bookResult, err := dao.Book.ResultFindByIdentify(identify, currentMember.MemberId)

		if err != nil || bookResult.RoleId == conf.Conf.Info.BookObserver {
			c.JSONCode(code.DocUpdateErr10)
			return
		}
		bookId = bookResult.BookId
	}

	if parentId > 0 {
		doc, err := dao.Document.Find(parentId)
		if err != nil || doc.BookId != bookId {
			c.JSONCode(code.DocUpdateErr11)
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
		c.JSONErr(code.DocUpdateErr11, err)
		return
	}

	if dao.DocumentStore.GetFiledById(docIdInt64, "markdown") == "" {
		//因为创建和更新文档基本信息都调用的这个接口，先判断markdown是否有内容，没有内容则添加默认内容
		if err := dao.DocumentStore.InsertOrUpdate(mus.Db, &mysql.DocumentStore{DocumentId: int(docIdInt64), Markdown: "[TOC]\n\r\n\r"}); err != nil {
			c.JSONErr(code.DocUpdateErr12, err)
			return
		}
	}
	c.JSONOK(document)
}

//删除文档.
func Delete(c *core.Context) {
	identify, _ := c.GetPostForm("identify")
	docIdStr, _ := c.GetPostForm("doc_id")
	docId, _ := strconv.Atoi(docIdStr)

	bookId := 0
	//如果是超级管理员则忽略权限判断
	if c.Member().IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			c.JSONErrStr(6002, "项目不存在或权限不足")
			return
		}
		bookId = book.BookId
	} else {
		bookResult, err := dao.Book.ResultFindByIdentify(identify, c.Member().MemberId)
		if err != nil || bookResult.RoleId == conf.BookObserver {
			c.JSONErrStr(6002, "项目不存在或权限不足")
			return
		}
		bookId = bookResult.BookId
	}

	if docId <= 0 {
		c.JSONErrStr(6001, "参数错误")
		return
	}

	doc, err := dao.Document.Find(docId)
	if err != nil {
		c.JSONErrStr(6003, "删除失败")
		return
	}

	//如果文档所属项目错误
	if doc.BookId != bookId {
		c.JSONErrStr(6004, "参数错误")
		return
	}
	//递归删除项目下的文档以及子文档
	err = dao.Document.RecursiveDocument(doc.DocumentId)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.JSONErrStr(6005, "删除失败")
		return
	}

	//重置文档数量统计
	dao.Book.ResetDocumentNumber(doc.BookId)

	go func() {
		// 删除文档的索引
		client := dao.NewElasticSearchClient()
		if errDel := client.DeleteIndex(docId, false); errDel != nil && client.On {
			mus.Logger.Error(errDel.Error())
		}
	}()

	c.JSONOK()
}

//上传附件或图片.
func Upload(c *core.Context) {
	identify := c.Param("key")
	docIdStr, _ := c.GetPostForm("doc_id")
	docId, _ := strconv.Atoi(docIdStr)
	isAttach := true

	if identify == "" {
		c.JSONErrStr(code.DodUploadErr1, "参数错误")
		return
	}

	name := "editormd-file-file"

	fileHeader, err := c.FormFile(name)
	if err == http.ErrMissingFile {
		name = "editormd-image-file"
		fileHeader, err = c.FormFile(name)
		if err == http.ErrMissingFile {
			c.JSONErrStr(code.DodUploadErr2, "没有发现需要上传的文件")
			return
		}
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSONErr(code.DodUploadErr3, err)
		return
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	if ext == "" {
		c.JSONErrStr(6003, "无法解析文件的格式")
		return
	}

	if !conf.IsAllowUploadFileExt(ext) {
		c.JSONErrStr(6004, "不允许的文件类型")
		return
	}

	bookId := 0
	//如果是超级管理员，则不判断权限
	if c.Member().IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			c.JSONErrStr(6006, "文档不存在或权限不足")
			return
		}
		bookId = book.BookId
	} else {
		book, err := dao.Book.ResultFindByIdentify(identify, c.Member().MemberId)
		if err != nil {
			mus.Logger.Error(err.Error())
			c.JSONErrStr(6001, err.Error())
			return
		}
		//如果没有编辑权限
		if book.RoleId != conf.BookEditor && book.RoleId != conf.BookAdmin && book.RoleId != conf.BookFounder {
			c.JSONErrStr(6006, "权限不足")
			return
		}
		bookId = book.BookId
	}

	if docId > 0 {
		doc, err := dao.Document.Find(docId)
		if err != nil {
			c.JSONErrStr(6007, "文档不存在")
			return
		}
		if doc.BookId != bookId {
			c.JSONErrStr(6008, "文档不属于指定的项目")
			return
		}
	}

	fileName := strconv.FormatInt(time.Now().UnixNano(), 16)

	filePath := filepath.Join("./", "uploads", time.Now().Format("200601"), fileName+ext)

	path := filepath.Dir(filePath)

	os.MkdirAll(path, os.ModePerm)

	err = c.SaveToFile(name, filePath)

	if err != nil {
		c.JSONErrStr(6005, "保存文件失败")
		return
	}
	attachment := mysql.Attachment{
		BookId:     bookId,
		DocumentId: docId,
		FileName:   fileHeader.Filename,
		FilePath:   strings.TrimPrefix(filePath, "./"),
		FileSize:   0,
		HttpPath:   "",
		FileExt:    ext,
		CreateTime: time.Now(),
		CreateAt:   c.Member().MemberId,
	}

	if fileInfo, err := os.Stat(filePath); err == nil {
		attachment.FileSize = float64(fileInfo.Size())
	}
	if docId > 0 {
		attachment.DocumentId = docId
	}

	var dstPath string

	if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") || strings.EqualFold(ext, ".png") || strings.EqualFold(ext, ".gif") {
		dstPath = mus.Oss.GenerateKey("eco-doc-img")
		attachment.HttpPath = dstPath
		isAttach = false
	} else {
		dstPath = mus.Oss.GenerateKey("eco-doc-file")
		attachment.HttpPath = "/document/download" + identify + "/" + strconv.Itoa(attachment.AttachmentId)

	}

	err = mus.Db.Create(&attachment).Error

	if err != nil {
		os.Remove(filePath)
		c.JSONErrStr(6006, "文件保存失败")
		return
	}

	err = mus.Oss.PutObjectFromFile(dstPath, filePath)
	if err != nil {
		c.JSONErr(code.UploadCoverErr10, err)
		return
	}

	result := map[string]interface{}{
		"code":      0,
		"success":   1,
		"message":   "ok",
		"url":       mus.Oss.ShowImg(attachment.HttpPath),
		"alt":       attachment.FileName,
		"is_attach": isAttach,
		"attach":    attachment,
	}
	c.JSONOK(result)
}
