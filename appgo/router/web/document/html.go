package document

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"html/template"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/TruthHun/BookStack/commands"
	"github.com/TruthHun/html2md"
	"github.com/astaxie/beego"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/model/mysql/store"
	"github.com/goecology/ecology/appgo/pkg/code"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/goecology/ecology/appgo/router/core"
	"github.com/spf13/viper"
)

// 解析并提取版本控制的commit内容
func parseGitCommit(str string) (cont, commit string) {
	var slice []string
	arr := strings.Split(str, "<bookstack-git>")
	if len(arr) > 1 {
		slice = append(slice, arr[0])
		str = strings.Join(arr[1:], "")
	}
	arr = strings.Split(str, "</bookstack-git>")
	if len(arr) > 1 {
		slice = append(slice, arr[1:]...)
		commit = arr[0]
	}
	if len(slice) > 0 {
		cont = strings.Join(slice, "")
	} else {
		cont = str
	}
	return
}

//判断用户是否可以阅读文档.
func isReadable(c *core.Context, identify, token string) (resp *mysql.BookResult) {
	book, err := dao.Book.FindByFieldFirst("identify", identify)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.Html404()
		return
	}

	//如果文档是私有的
	if book.PrivatelyOwned == 1 && !c.Member().IsAdministrator() {
		isOk := false
		if c.Member() != nil {
			_, err := dao.Relationship.FindForRoleId(book.BookId, c.Member().MemberId)
			if err == nil {
				isOk = true
			}
		}

		if book.PrivateToken != "" && !isOk {
			//如果有访问的Token，并且该项目设置了访问Token，并且和用户提供的相匹配，则记录到Session中.
			//如果用户未提供Token且用户登录了，则判断用户是否参与了该项目.
			//如果用户未登录，则从Session中读取Token.
			if token != "" && strings.EqualFold(token, book.PrivateToken) {
				c.SetSession(identify, token)
			} else if token, ok := c.GetSession(identify).(string); !ok || !strings.EqualFold(token, book.PrivateToken) {
				hasErr := ""
				if c.Context.Request.Method == "POST" {
					hasErr = "true"
				}
				c.Redirect(302, beego.URLFor("DocumentController.Index", ":key", identify)+"?with-password=true&err="+hasErr)
				return
			}
		} else if !isOk {
			c.Html404()
			return
		}
	}

	bookResult := book.ToBookResult()
	if c.Member() != nil {
		rel, err := dao.Relationship.FindByBookIdAndMemberId(bookResult.BookId, c.Member().MemberId)
		if err == nil {
			bookResult.MemberId = book.MemberId
			bookResult.RoleId = rel.RoleId
			bookResult.RelationshipId = rel.RelationshipId
		}
	}
	//判断是否需要显示评论框
	switch bookResult.CommentStatus {
	case "closed":
		bookResult.IsDisplayComment = false
	case "open":
		bookResult.IsDisplayComment = true
	case "group_only":
		bookResult.IsDisplayComment = bookResult.RelationshipId > 0
	case "registered_only":
		bookResult.IsDisplayComment = true
	}
	return bookResult
}

//文档首页.
func Index(c *core.Context) {
	identify := c.Param("key")
	if identify == "" {
		c.Html404()
		return
	}

	token, _ := c.GetQuery("token")
	withPwd, _ := c.GetQuery("with-password")
	tab, _ := c.GetQuery("tab")
	if len(strings.TrimSpace(withPwd)) > 0 {
		indexWithPassword(c)
		return
	}

	tab = strings.ToLower(tab)

	bookResult := isReadable(c, identify, token)
	if bookResult.BookId == 0 { //没有阅读权限
		c.Redirect(302, "/")
		return
	}

	bookResult.Lang = utils.GetLang(bookResult.Lang)
	c.Tpl().Data["Book"] = bookResult

	switch tab {
	case "comment", "score":
	default:
		tab = "default"
	}
	c.Tpl().Data["Qrcode"] = dao.Member.GetQrcodeByUid(bookResult.MemberId)
	c.Tpl().Data["MyScore"] = dao.Score.BookScoreByUid(c.Member().MemberId, bookResult.BookId)
	c.Tpl().Data["Tab"] = tab
	if beego.AppConfig.DefaultBool("showWechatCode", false) && bookResult.PrivatelyOwned == 0 {
		wechatCode := mysql.NewWechatCode()
		go wechatCode.CreateWechatCode(bookResult.BookId) //如果已经生成了小程序码，则不会再生成
		c.Tpl().Data["Wxacode"] = wechatCode.GetCode(bookResult.BookId)
	}

	//当前默认展示100条评论
	c.Tpl().Data["Comments"], _ = dao.Comments.Comments(1, 100, bookResult.BookId, 1)
	c.Tpl().Data["Menu"], _ = dao.Document.GetMenuTop(bookResult.BookId)
	title := "《" + bookResult.BookName + "》"
	if tab == "comment" {
		title = "点评 - " + title
	}
	c.GetSeoByPage("book_info", map[string]string{
		"title":       title,
		"keywords":    bookResult.Label,
		"description": bookResult.Description,
	})
	c.Tpl().Data["RelateBooks"] = mysql.NewRelateBook().Lists(bookResult.BookId)
	c.Html("document/intro")

}

//文档首页.
func indexWithPassword(c *core.Context) {
	identify := c.Param("key")
	if identify == "" {
		c.Html404()
		return
	}
	c.GetSeoByPage("book_info", map[string]string{
		"title":       "密码访问",
		"keywords":    "密码访问",
		"description": "密码访问",
	})
	c.Tpl().Data["ShowErrTips"] = c.GetString("err") != ""
	c.Tpl().Data["Identify"] = identify
	c.Html("document/read-with-password")
	return
}

//阅读文档.
func ReadHtml(c *core.Context) {
	identify := c.Param("key")
	id := c.Param("id")

	token, _ := c.GetQuery("token")

	if identify == "" || id == "" {
		c.Html404()
		return
	}
	//如果没有开启你们匿名则跳转到登录
	if !dao.Global.IsEnableAnonymous() && c.Member() == nil {
		c.Redirect(302, "/login")
		return
	}

	bookResult := isReadable(c, identify, token)

	var err error

	doc := mysql.NewDocument()
	if docId, _ := strconv.Atoi(id); docId > 0 {
		doc, err = dao.Document.Find(docId) //文档id
		if err != nil {
			mus.Logger.Error(err.Error())
			c.Html404()
			return
		}
	} else {
		//此处的id是字符串，标识文档标识，根据文档标识和文档所属的书的id作为key去查询
		doc, err = dao.Document.FindByBookIdAndDocIdentify(bookResult.BookId, id) //文档标识
		if err != nil {
			// todo log
			c.Html404()
			return
		}
	}

	if doc.BookId != bookResult.BookId {
		c.Html404()
		return
	}

	bodyText := ""
	authHTTPS := strings.ToLower(dao.Global.GetOptionValue("AUTO_HTTPS", "false")) == "true"
	if doc.Release != "" {
		query, err := goquery.NewDocumentFromReader(bytes.NewBufferString(doc.Release))
		if err != nil {
			mus.Logger.Error(err.Error())
		} else {
			query.Find("img").Each(func(i int, contentSelection *goquery.Selection) {
				src, ok := contentSelection.Attr("src")
				if ok {
					if utils.StoreType == utils.StoreOss && !(strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "http://")) {
						src = viper.GetString("app.ossDomain") + "/" + strings.TrimLeft(src, "./")
					}
				}
				if authHTTPS {
					if srcArr := strings.Split(src, "://"); len(srcArr) > 1 {
						src = "https://" + strings.Join(srcArr[1:], "://")
					}
				}
				contentSelection.SetAttr("src", src)
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
		bodyText = query.Find(".markdown-toc").Text()
	}

	attach, err := dao.Attachment.FindListByDocumentId(doc.DocumentId)
	if err == nil {
		doc.AttachList = attach
	}

	//文档阅读人次+1
	if err := mysql.SetIncreAndDecre(mysql.Document{}.TableName(), "vcnt",
		fmt.Sprintf("document_id=%v", doc.DocumentId),
		true, 1,
	); err != nil {
		mus.Logger.Error(err.Error())
	}
	//项目阅读人次+1
	if err := mysql.SetIncreAndDecre(mysql.Book{}.TableName(), "vcnt",
		fmt.Sprintf("book_id=%v", doc.BookId),
		true, 1,
	); err != nil {
		mus.Logger.Error(err.Error())
	}

	if c.Member().MemberId > 0 { //增加用户阅读记录
		if err := new(mysql.ReadRecord).Add(doc.DocumentId, c.Member().MemberId); err != nil {
			mus.Logger.Error(err.Error())
		}
	}
	parentTitle := dao.Document.GetParentTitle(doc.ParentId)
	seo := map[string]string{
		"title":       doc.DocumentName + " - 《" + bookResult.BookName + "》",
		"keywords":    bookResult.Label,
		"description": beego.Substr(bodyText+" "+bookResult.Description, 0, 200),
	}

	if len(parentTitle) > 0 {
		seo["title"] = parentTitle + " - " + doc.DocumentName + " - 《" + bookResult.BookName + "》"
	}

	//SEO
	c.GetSeoByPage("book_read", seo)

	existBookmark := dao.Bookmark.Exist(c.Member().MemberId, doc.DocumentId)

	doc.Vcnt = doc.Vcnt + 1

	mysql.NewBookCounter().Increase(bookResult.BookId, true)

	tree, err := dao.Document.CreateDocumentTreeForHtml(bookResult.BookId, doc.DocumentId)

	if err != nil {
		mus.Logger.Error(err.Error())
		c.Html404()
	}

	// 查询用户哪些文档阅读了
	if c.Member().MemberId > 0 {
		modelRecord := new(mysql.ReadRecord)
		lists, cnt, _ := modelRecord.List(c.Member().MemberId, bookResult.BookId)
		if cnt > 0 {
			var readMap = make(map[string]bool)
			for _, item := range lists {
				readMap[strconv.Itoa(item.DocId)] = true
			}
			if doc, err := goquery.NewDocumentFromReader(strings.NewReader(tree)); err == nil {
				doc.Find("li").Each(func(i int, selection *goquery.Selection) {
					if id, exist := selection.Attr("id"); exist {
						if _, ok := readMap[id]; ok {
							selection.AddClass("readed")
						}
					}
				})
				tree, _ = doc.Find("body").Html()
			}
		}
	}

	if beego.AppConfig.DefaultBool("showWechatCode", false) && bookResult.PrivatelyOwned == 0 {
		wechatCode := mysql.NewWechatCode()
		go wechatCode.CreateWechatCode(bookResult.BookId) //如果已经生成了小程序码，则不会再生成
		c.Tpl().Data["Wxacode"] = wechatCode.GetCode(bookResult.BookId)
	}

	if wd, _ := c.GetQuery("wd"); strings.TrimSpace(wd) != "" {
		c.Tpl().Data["Keywords"] = dao.NewElasticSearchClient().SegWords(wd)
	}
	c.Tpl().Data["Bookmark"] = existBookmark
	c.Tpl().Data["Model"] = bookResult
	c.Tpl().Data["Book"] = bookResult //文档下载需要用到Book变量
	c.Tpl().Data["Result"] = template.HTML(tree)
	c.Tpl().Data["Title"] = doc.DocumentName
	c.Tpl().Data["DocId"] = doc.DocumentId
	c.Tpl().Data["Content"] = template.HTML(doc.Release)
	c.Tpl().Data["View"] = doc.Vcnt
	c.Tpl().Data["UpdatedAt"] = doc.ModifyTime.Format("2006-01-02 15:04:05")
	c.Html("document/" + bookResult.Theme + "_read")
}

//编辑文档.
func Edit(c *core.Context) {
	docId := 0 // 文档id

	identify := c.Param("key")
	if identify == "" {
		c.Html404()
		return
	}

	bookResult := mysql.NewBookResult()

	var err error
	//如果是超级管理者，则不判断权限
	if c.Member().IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			c.JSONErrStr(6002, "项目不存在或权限不足")
			return
		}
		bookResult = book.ToBookResult()
	} else {
		bookResult, err = dao.Book.ResultFindByIdentify(identify, c.Member().MemberId)
		if err != nil {
			mus.Logger.Error(err.Error())
			c.Html404()
			return
		}

		if bookResult.RoleId == conf.BookObserver {
			c.JSONErrStr(6002, "项目不存在或权限不足")
			return
		}
	}

	c.Tpl().Data["Model"] = bookResult
	r, _ := json.Marshal(bookResult)

	c.Tpl().Data["ModelResult"] = template.JS(string(r))

	c.Tpl().Data["Result"] = template.JS("[]")

	// 编辑的文档
	if id := c.Param("id"); id != "" {
		if num, _ := strconv.Atoi(id); num > 0 {
			docId = num
		} else { //字符串
			var doc = mysql.NewDocument()
			mus.Db.Where("identify=? and book_id = ?", id, bookResult.BookId).Find(doc)
			docId = doc.DocumentId
		}
	}

	trees, err := dao.Document.FindDocumentTree(bookResult.BookId, docId, true)
	if err != nil {
		mus.Logger.Error(err.Error())
	} else {
		if len(trees) > 0 {
			if jsTree, err := json.Marshal(trees); err == nil {
				c.Tpl().Data["Result"] = template.JS(string(jsTree))
			}
		} else {
			c.Tpl().Data["Result"] = template.JS("[]")
		}
	}
	c.Tpl().Data["BaiDuMapKey"] = beego.AppConfig.DefaultString("baidumapkey", "")
	//根据不同编辑器类型加载编辑器【注：现在只支持markdown】
	c.Html("document/markdown_edit_template")

}

//批量创建文档
func CreateMulti(c *core.Context) {
	bookIdStr, _ := c.GetQuery("book_id")
	bookId, _ := strconv.Atoi(bookIdStr)

	if !(c.Member().MemberId > 0 && bookId > 0) {
		c.JSONErrStr(1, "操作失败：只有项目创始人才能批量添加")
		return
	}

	var book mysql.Book
	mus.Db.Where("book_id = ? and member_id = ?", bookId, c.Member().MemberId).Find(&book)

	if book.BookId > 0 {
		content, _ := c.GetQuery("content")
		slice := strings.Split(content, "\n")
		if len(slice) > 0 {
			for _, row := range slice {
				if chapter := strings.Split(strings.TrimSpace(row), " "); len(chapter) > 1 {
					if ok, err := regexp.MatchString(`^[a-zA-Z0-9_\-\.]*$`, chapter[0]); ok && err == nil {
						i, _ := strconv.Atoi(chapter[0])
						if chapter[0] != "0" && strconv.Itoa(i) != chapter[0] { //不为纯数字
							doc := mysql.Document{
								DocumentName: strings.Join(chapter[1:], " "),
								Identify:     chapter[0],
								BookId:       bookId,
								//Markdown:     "[TOC]\n\r",
								MemberId: c.Member().MemberId,
							}
							if docId, err := dao.Document.InsertOrUpdate(mus.Db, &doc); err == nil {
								if err := dao.DocumentStore.InsertOrUpdate(mus.Db, &mysql.DocumentStore{DocumentId: int(docId), Markdown: "[TOC]\n\r\n\r"}); err != nil {
									mus.Logger.Error(err.Error())
								}
							} else {
								mus.Logger.Error(err.Error())
							}
						}

					}
				}
			}
		}
	}
	c.JSONOK()
}

//上传附件或图片.
func Upload(c *core.Context) {
	identify := c.GetString("identify")
	docId := c.GetInt("doc_id")
	isAttach := true

	if identify == "" {
		c.JSONErrStr(6001, "参数错误")
	}

	name := "editormd-file-file"
	//
	//file, moreFile, err := c.(name)
	//if err == http.ErrMissingFile {
	//	name = "editormd-image-file"
	//	file, moreFile, err = this.GetFile(name)
	//	if err == http.ErrMissingFile {
	//		c.JSONErrStr(6003, "没有发现需要上传的文件")
	//	}
	//}

	fileHeader, err := c.FormFile(name)
	if err == http.ErrMissingFile {
		name = "editormd-image-file"
		fileHeader, err = c.FormFile(name)
		if err == http.ErrMissingFile {
			c.JSONErrStr(6003, "没有发现需要上传的文件")
			return
		}
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSONErr(code.MsgErr, err)
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

	filePath := filepath.Join(commands.WorkingDirectory, "uploads", time.Now().Format("200601"), fileName+ext)

	path := filepath.Dir(filePath)

	os.MkdirAll(path, os.ModePerm)

	err = c.SaveToFile(name, filePath)

	if err != nil {
		c.JSONErrStr(6005, "保存文件失败")
		return
	}
	attachment := mysql.NewAttachment()
	attachment.BookId = bookId
	attachment.FileName = fileHeader.Filename
	attachment.CreateAt = c.Member().MemberId
	attachment.FileExt = ext
	attachment.FilePath = strings.TrimPrefix(filePath, commands.WorkingDirectory)
	attachment.DocumentId = docId

	if fileInfo, err := os.Stat(filePath); err == nil {
		attachment.FileSize = float64(fileInfo.Size())
	}
	if docId > 0 {
		attachment.DocumentId = docId
	}

	if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") || strings.EqualFold(ext, ".png") || strings.EqualFold(ext, ".gif") {

		attachment.HttpPath = "/" + strings.Replace(strings.TrimPrefix(filePath, commands.WorkingDirectory), "\\", "/", -1)
		if strings.HasPrefix(attachment.HttpPath, "//") {
			attachment.HttpPath = string(attachment.HttpPath[1:])
		}
		isAttach = false
	}

	err = mus.Db.Create(&attachment).Error

	if err != nil {
		os.Remove(filePath)
		c.JSONErrStr(6006, "文件保存失败")
		return
	}
	if attachment.HttpPath == "" {
		attachment.HttpPath = beego.URLFor("DocumentController.DownloadAttachment", ":key", identify, ":attach_id", attachment.AttachmentId)

		if err := mus.Db.UpdateColumns(attachment).Error; err != nil {
			c.JSONErrStr(6005, "保存文件失败")
			return
		}
	}
	osspath := fmt.Sprintf("projects/%v/%v", identify, fileName+filepath.Ext(attachment.HttpPath))
	switch utils.StoreType {
	case utils.StoreOss:
		if err := store.ModelStoreOss.MoveToOss("."+attachment.HttpPath, osspath, true, false); err != nil {
			mus.Logger.Error(err.Error())
		}
		//attachment.HttpPath = this.OssDomain + "/" + osspath
		attachment.HttpPath = "/" + osspath
	case utils.StoreLocal:
		osspath = "uploads/" + osspath
		if err := store.ModelStoreLocal.MoveToStore("."+attachment.HttpPath, osspath); err != nil {
			mus.Logger.Error(err.Error())
		}
		attachment.HttpPath = "/" + osspath
	}

	result := map[string]interface{}{
		"errcode":   0,
		"success":   1,
		"message":   "ok",
		"url":       attachment.HttpPath,
		"alt":       attachment.FileName,
		"is_attach": isAttach,
		"attach":    attachment,
	}
	c.JSONOK(result)
}

//DownloadAttachment 下载附件.
func DownloadAttachment(c *core.Context) {
	identify := c.Param(":key")
	attachId, _ := strconv.Atoi(c.Param(":attach_id"))
	token := c.GetString("token")

	memberId := 0

	if c.Member() != nil {
		memberId = c.Member().MemberId
	}
	bookId := 0

	//判断用户是否参与了项目
	bookResult, err := dao.Book.ResultFindByIdentify(identify, memberId)

	if err != nil {
		//判断项目公开状态
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			c.Html404()
			return
		}
		//如果不是超级管理员则判断权限
		if c.Member() == nil || c.Member().Role != conf.MemberSuperRole {
			//如果项目是私有的，并且token不正确
			if (book.PrivatelyOwned == 1 && token == "") || (book.PrivatelyOwned == 1 && book.PrivateToken != token) {
				c.Html404()
				return
			}
		}

		bookId = book.BookId
	} else {
		bookId = bookResult.BookId
	}
	//查找附件
	attachment, err := dao.Attachment.Find(attachId)

	if err != nil {
		c.Html404()
		return
	}
	if attachment.BookId != bookId {
		c.Html404()
		return
	}

	c.Download(filepath.Join(commands.WorkingDirectory, attachment.FilePath), attachment.FileName)
}

//删除附件.
func RemoveAttachment(c *core.Context) {
	attachIdStr, _ := c.GetQuery("attach_id")
	attachId, _ := strconv.Atoi(attachIdStr)
	if attachId <= 0 {
		c.JSONErrStr(6001, "参数错误")
		return
	}

	attach, err := dao.Attachment.Find(attachId)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.JSONErrStr(6002, "附件不存在")
		return
	}

	document, err := dao.Document.Find(attach.DocumentId)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.JSONErrStr(6003, "文档不存在")
		return
	}

	if c.Member().Role != conf.MemberSuperRole {
		rel, err := dao.Relationship.FindByBookIdAndMemberId(document.BookId, c.Member().MemberId)
		if err != nil {
			mus.Logger.Error(err.Error())
			c.JSONErrStr(6004, "权限不足")
			return
		}
		if rel.RoleId == conf.BookObserver {
			c.JSONErrStr(6004, "权限不足")
			return
		}
	}

	if err = dao.Attachment.Delete(c.Context, mus.Db, attachId); err != nil {
		mus.Logger.Error(err.Error())
		c.JSONErrStr(6005, "删除失败")
	}

	os.Remove(filepath.Join(commands.WorkingDirectory, attach.FilePath))
	c.JSONOK(attach)
}

//删除文档.
func Delete(c *core.Context) {

	identify := c.GetString("identify")
	docId := c.GetInt("doc_id")

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

//获取或更新文档内容.
func ContentGet(c *core.Context) {
	identify := c.Param("key")
	_, flag := c.GetQuery("doc_id")
	var docId int
	if !flag {
		docId, _ = strconv.Atoi(c.Param("id"))
	}
	currentMember := c.Member()
	if currentMember.IsAdministrator() {
		_, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
	} else {
		bookResult, err := dao.Book.ResultFindByIdentify(identify, currentMember.MemberId)
		if err != nil || bookResult.RoleId == conf.Conf.Info.BookObserver {
			c.JSONCode(code.MsgErr)
			return
		}
	}

	if docId <= 0 {
		c.JSONCode(code.MsgErr)
		return

	}

	doc, err := dao.Document.Find(docId)

	if err != nil {
		c.JSONCode(code.MsgErr)
		return
	}
	attach, err := dao.Attachment.FindListByDocumentId(doc.DocumentId)
	if err == nil {
		doc.AttachList = attach
	}

	//为了减少数据的传输量，这里Release和Content的内容置空，前端会根据markdown文本自动渲染
	//doc.Release = ""
	//doc.Content = ""
	doc.Markdown = dao.DocumentStore.GetFiledById(doc.DocumentId, "markdown")
	c.JSONOK(doc)

}

//获取或更新文档内容.
func ContentPost(c *core.Context) {
	identify := c.Param("key")
	_, flag := c.GetQuery("doc_id")
	errMsg := code.MsgOk
	var docId int
	if !flag {
		docId, _ = strconv.Atoi(c.Param("id"))
	}
	bookId := 0
	currentMember := c.Member()
	//如果是超级管理员，则忽略权限
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
			c.JSONOK(code.MsgErr)
			return
		}
		bookId = bookResult.BookId
	}

	if docId <= 0 {
		c.JSONCode(code.MsgErr)
		return

	}

	//更新文档内容
	markdown := strings.TrimSpace(c.PostForm("markdown"))
	content := c.PostForm("html")

	// 文档拆分
	gq, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err == nil {
		seg := gq.Find("bookstack-split").Text()
		if strings.Contains(seg, "#") {
			markdown = strings.Replace(markdown, fmt.Sprintf("<bookstack-split>%v</bookstack-split>", seg), "", -1)
			err := dao.Document.SplitMarkdownAndStore(seg, markdown, docId)
			if err != nil {
				c.JSONErr(code.MsgErr, err)
				return
			}
			c.JSONOK(code.MsgOk)
			return
		}
	}

	version, _ := strconv.Atoi(c.PostForm("version"))
	isCover := c.PostForm("cover")

	doc, err := dao.Document.Find(docId)

	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	if doc.BookId != bookId {
		c.JSONCode(code.MsgErr)
		return
	}
	if doc.Version != int64(version) && !strings.EqualFold(isCover, "yes") {
		c.JSONCode(code.MsgErr)
		return
	}

	isSummary := false
	isAuto := false
	//替换文档中的url链接
	if strings.ToLower(doc.Identify) == "summary.md" && (strings.Contains(markdown, "<bookstack-summary></bookstack-summary>") || strings.Contains(doc.Markdown, "<bookstack-summary/>")) {
		//如果标识是summary.md，并且带有bookstack的标签，则表示更新目录
		isSummary = true
		//要清除，避免每次保存的时候都要重新排序
		replaces := []string{"<bookstack-summary></bookstack-summary>", "<bookstack-summary/>"}
		for _, r := range replaces {
			markdown = strings.Replace(markdown, r, "", -1)
		}
	}

	//爬虫采集
	access := currentMember.IsAdministrator()
	if op, err := dao.Global.FindByKey("SPIDER"); err == nil {
		access = access && op.OptionValue == "true"
	}
	if access && strings.ToLower(doc.Identify) == "summary.md" && (strings.Contains(markdown, "<spider></spider>") || strings.Contains(doc.Markdown, "<spider/>")) {
		//如果标识是summary.md，并且带有bookstack的标签，则表示更新目录
		isSummary = true
		//要清除，避免每次保存的时候都要重新排序
		replaces := []string{"<spider></spider>", "<spider/>"}
		for _, r := range replaces {
			markdown = strings.Replace(markdown, r, "", -1)
		}
		content, markdown, _ = dao.Document.BookStackCrawl(content, markdown, bookId, currentMember.MemberId)
	}

	if strings.Contains(markdown, "<bookstack-auto></bookstack-auto>") || strings.Contains(doc.Markdown, "<bookstack-auto/>") {
		//自动生成文档内容

		var imd, icont string
		if strings.ToLower(doc.Identify) == "summary.md" {
			icont, _ = dao.Document.CreateDocumentTreeForHtml(doc.BookId, doc.DocumentId)
			imd = html2md.Convert(icont)
			imd = strings.Replace(imd, "(/read/"+identify+"/", "($", -1)
		} else {
			imd, icont = dao.Document.BookStackAuto(bookId, docId)
		}

		markdown = strings.Replace(markdown, "<bookstack-auto></bookstack-auto>", imd, -1)
		content = strings.Replace(content, "<bookstack-auto></bookstack-auto>", icont, -1)
		isAuto = true
	}
	content = c.ReplaceLinks(identify, content, isSummary)

	var ds = mysql.DocumentStore{}
	var actionName string

	// 替换掉<git></git>标签内容
	if markdown == "" && content != "" {
		ds.Markdown = content
	} else {
		ds.Markdown = markdown
	}

	ds.Markdown, actionName = parseGitCommit(ds.Markdown)
	ds.Content, _ = parseGitCommit(content)

	if actionName == "" {
		actionName = "--"
	} else {
		isAuto = true
	}
	fmt.Println("ds------>", ds)

	doc.Version = time.Now().Unix()
	if docId, err := dao.Document.InsertOrUpdate(mus.Db, doc); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	} else {
		ds.DocumentId = int(docId)
		if err := dao.DocumentStore.InsertOrUpdate(mus.Db, &ds); err != nil {
			mus.Logger.Error(err.Error())
		}
	}

	//如果启用了文档历史，则添加历史文档
	if enableDocumentHistory() > 0 {
		if len(strings.TrimSpace(ds.Markdown)) > 0 { //空内容不存储版本
			history := mysql.DocumentHistory{}
			history.DocumentId = docId
			history.DocumentName = doc.DocumentName
			history.ModifyAt = currentMember.MemberId
			history.MemberId = doc.MemberId
			history.ParentId = doc.ParentId
			history.Version = time.Now().Unix()
			history.Action = "modify"
			history.ActionName = actionName
			// todo fix

			//_, err = service.MdDocumentHistory.InsertOrUpdate()
			//if err != nil {
			//	mus.Logger.Error("DocumentHistory InsertOrUpdate => " + err.Error())
			//} else {
			//	vc := service.NewVersionControl(docId, history.Version)
			//	vc.SaveVersion(ds.Content, ds.Markdown)
			//	service.MdDocumentHistory.DeleteByLimit(docId, enableDocumentHistory())
			//}
		}

	}

	if isAuto {
		errMsg = code.DocumentContentAuto
	} else if isSummary {
		errMsg = code.DocumentContentTrue
	}

	doc.Release = ""
	//注意：如果errMsg的值是true，则表示更新了目录排序，需要刷新，否则不刷新
	c.JSONCode(errMsg, doc)

}

//导出文件
func Export(c *core.Context) {
	if c.Member() == nil || c.Member().MemberId == 0 {
		if tips := dao.Global.Get("DOWNLOAD_LIMIT"); tips != "" {
			tips = strings.TrimSpace(tips)
			if len(tips) > 0 {
				c.JSONErrStr(1, tips)
				return
			}
		}
	}

	//this.TplName = "document/export.html"
	identify := c.Param(":key")
	ext := strings.ToLower(c.GetString("output"))
	switch ext {
	case "pdf", "epub", "mobi":
		ext = "." + ext
	default:
		ext = ".pdf"
	}
	if identify == "" {
		c.JSONErrStr(1, "下载失败，无法识别您要下载的文档")
		return
	}
	book, err := dao.Book.FindByIdentify(identify)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.JSONErrStr(1, "下载失败，您要下载的文档当前并未生成可下载文档。")
		return
	}
	if book.PrivatelyOwned == 1 && c.Member().MemberId != book.MemberId {
		c.JSONErrStr(1, "私有文档，只有文档创建人可导出")
		return
	}
	//查询文档是否存在
	obj := fmt.Sprintf("projects/%v/books/%v%v", book.Identify, book.GenerateTime.Unix(), ext)
	switch utils.StoreType {
	case utils.StoreOss:
		if err := store.ModelStoreOss.IsObjectExist(obj); err != nil {
			mus.Logger.Error(err.Error())
			c.JSONErrStr(1, "下载失败，您要下载的文档当前并未生成可下载文档。")
			return
		}
		c.JSONOK(map[string]interface{}{"url": viper.GetString("app.OssDomain") + "/" + obj})
		return

	case utils.StoreLocal:
		obj = "uploads/" + obj
		if err := store.ModelStoreLocal.IsObjectExist(obj); err != nil {
			mus.Logger.Error(err.Error())
			c.JSONErrStr(1, "下载失败，您要下载的文档当前并未生成可下载文档。")
			return
		}
		c.JSONOK(map[string]interface{}{"url": "/" + obj})
	}
	c.JSONErrStr(1, "下载失败，您要下载的文档当前并未生成可下载文档。")
}

//生成项目访问的二维码.

func QrCode(c *core.Context) {
	identify := c.GetString(":key")

	book, err := dao.Book.FindByIdentify(identify)

	if err != nil || book.BookId <= 0 {
		c.Html404()
		return
	}

	uri := c.BaseUrl() + beego.URLFor("DocumentController.Index", ":key", identify)
	code, err := qr.Encode(uri, qr.L, qr.Unicode)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.Html404()
		return
	}
	code, err = barcode.Scale(code, 150, 150)

	if err != nil {
		mus.Logger.Error(err.Error())
		c.Html404()
	}
	c.Header("Content-Type", "image/png")

	//imgpath := filepath.Join("cache","qrcode",identify + ".png")

	err = png.Encode(c.Context.Writer, code)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.Html404()
		return
	}
}

//项目内搜索.
func Search(c *core.Context) {
	identify := c.Param(":key")
	token := c.GetString("token")
	keyword := strings.TrimSpace(c.GetString("keyword"))

	if identify == "" {
		c.JSONErrStr(6001, "参数错误")
	}
	if !dao.Global.IsEnableAnonymous() && c.Member() == nil {
		c.Redirect(302, "/login")
		return
	}
	bookResult := isReadable(c, identify, token)

	client := dao.NewElasticSearchClient()
	if client.On { // 全文搜索
		result, err := client.Search(keyword, 1, 10000, true, bookResult.BookId)
		if err != nil {
			mus.Logger.Error(err.Error())
			c.JSONErrStr(6002, "搜索结果错误")
			return
		}

		var ids []int
		for _, item := range result.Hits.Hits {
			ids = append(ids, item.Source.Id)
		}
		docs, err := dao.DocumentSearchResult.GetDocsById(ids, true)
		if err != nil {
			mus.Logger.Error(err.Error())
			return
		}

		// 如果全文搜索查询不到结果，用 MySQL like 再查询一次
		if len(docs) == 0 {
			if docsMySQL, _, err := dao.DocumentSearchResult.SearchDocument(keyword, bookResult.BookId, 1, 10000); err != nil {
				mus.Logger.Error(err.Error())
				c.JSONErrStr(6002, "搜索结果错误")
				return
			} else {
				c.JSONOK(client.SegWords(keyword), docsMySQL)
			}
		} else {
			c.JSONOK(client.SegWords(keyword), docs)
		}

	} else {
		docs, _, err := dao.DocumentSearchResult.SearchDocument(keyword, bookResult.BookId, 1, 10000)
		if err != nil {
			mus.Logger.Error(err.Error())
			c.JSONErrStr(6002, "搜索结果错误")
			return
		}
		c.JSONOK(keyword, docs)
	}
}

//文档历史列表.
func History(c *core.Context) {

	identify := c.GetString("identify")
	docId := c.GetInt("doc_id")
	//pageIndex := c.GetInt("page")

	bookId := 0
	//如果是超级管理员则忽略权限判断
	if c.Member().IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			c.Tpl().Data["ErrorMessage"] = "项目不存在或权限不足"
			c.Html("document/history")
			return
		}
		bookId = book.BookId
		c.Tpl().Data["Model"] = book
	} else {
		bookResult, err := dao.Book.ResultFindByIdentify(identify, c.Member().MemberId)

		if err != nil || bookResult.RoleId == conf.BookObserver {
			c.Tpl().Data["ErrorMessage"] = "项目不存在或权限不足"
			c.Html("document/history")

			return
		}
		bookId = bookResult.BookId
		c.Tpl().Data["Model"] = bookResult
	}

	if docId <= 0 {
		c.Tpl().Data["ErrorMessage"] = "参数错误"
		c.Html("document/history")
		return
	}

	doc, err := dao.Document.Find(docId)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.Tpl().Data["ErrorMessage"] = "获取历史失败"
		c.Html("document/history")
		return
	}
	//如果文档所属项目错误
	if doc.BookId != bookId {
		c.Tpl().Data["ErrorMessage"] = "参数错误"
		c.Html("document/history")
		return
	}

	// todo fix

	//histories, totalCount, err := mysql.NewDocumentHistory().FindToPager(docId, pageIndex, conf.PageSize)
	//if err != nil {
	//	mus.Logger.Error("FindToPager => ", err)
	//	c.Tpl().Data["ErrorMessage"] = "获取历史失败"
	//	return
	//}

	//c.Tpl().Data["List"] = histories
	c.Tpl().Data["PageHtml"] = ""
	c.Tpl().Data["Document"] = doc

	//if totalCount > 0 {
	//	html := utils.GetPagerHtml(c.Context.Request.RequestURI, pageIndex, conf.PageSize, totalCount)
	//	c.Tpl().Data["PageHtml"] = html
	//}
	c.Html("document/history")
}

func DeleteHistory(c *core.Context) {
	identify := c.GetString("identify")
	docId := c.GetInt("doc_id")
	historyId := c.GetInt("history_id")

	if historyId <= 0 {
		c.JSONErrStr(6001, "参数错误")
		return
	}
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
		c.JSONErrStr(6001, "获取历史失败")
		return
	}

	//如果文档所属项目错误
	if doc.BookId != bookId {
		c.JSONErrStr(6001, "参数错误")
		return
	}

	//err = mysql.NewDocumentHistory().Delete(history_id, doc_id)
	// todo fix
	//err = mysql.NewDocumentHistory().DeleteByHistoryId(historyId)
	//if err != nil {
	//	mus.Logger.Error(err)
	//	c.JSONErrStr(6002, "删除失败")
	//}
	c.JSONOK()
}

func RestoreHistory(c *core.Context) {

	identify := c.GetString("identify")
	docId := c.GetInt("doc_id")

	historyId := c.GetInt("history_id")
	if historyId <= 0 {
		c.JSONErrStr(6001, "参数错误")
		return
	}

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
		c.JSONErrStr(6001, "获取历史失败")
		return
	}
	//如果文档所属项目错误
	if doc.BookId != bookId {
		c.JSONErrStr(6001, "参数错误")
	}

	// todo fix
	//err = mysql.NewDocumentHistory().Restore(historyId, docId, c.Member().MemberId)
	//if err != nil {
	//	mus.Logger.Error(err)
	//	c.JSONErrStr(6002, "删除失败")
	//}
	c.JSONOK()
}

func CompareHtml(c *core.Context) {
	historyId, _ := strconv.Atoi(c.Param(":id"))
	identify := c.Param(":key")

	bookId := 0
	//如果是超级管理员则忽略权限判断
	if c.Member().IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			mus.Logger.Error("CompareHtml error ", zap.Error(err))
			c.Html404()
			return
		}
		bookId = book.BookId
		c.Tpl().Data["Model"] = book
	} else {
		bookResult, err := dao.Book.ResultFindByIdentify(identify, c.Member().MemberId)

		if err != nil || bookResult.RoleId == conf.BookObserver {
			c.Html404()
			return
		}
		bookId = bookResult.BookId
		c.Tpl().Data["Model"] = bookResult
	}

	if historyId <= 0 {
		c.JSONErrStr(60002, "参数错误")
		return
	}

	// todo fix
	//history, err := mysql.NewDocumentHistory().Find(historyId)
	//if err != nil {
	//	mus.Logger.Error("DocumentController.Compare => ", err)
	//	this.ShowErrorPage(60003, err.Error())
	//}
	//doc, err := dao.Document.Find(history.DocumentId)
	//
	//if doc.BookId != bookId {
	//	this.ShowErrorPage(60002, "参数错误")
	//}
	//vc := mysql.NewVersionControl(doc.DocumentId, history.Version)
	c.Tpl().Data["HistoryId"] = historyId
	//c.Tpl().Data["DocumentId"] = doc.DocumentId
	//ModelStore := new(mysql.DocumentStore)
	//c.Tpl().Data["HistoryContent"] = vc.GetVersionContent(false)
	fmt.Println("bookId------>", bookId)
	//c.Tpl().Data["Content"] = ModelStore.GetFiledById(doc.DocumentId, "markdown")
	c.Html("document/compare")
}

func enableDocumentHistory() int {
	option, err := dao.Global.FindByKey("ENABLE_DOCUMENT_HISTORY")
	if err != nil {
		return 0
	}
	verNum, _ := strconv.Atoi(option.OptionValue)
	return verNum
}
