package core

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/constx"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/code"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/muses/pkg/tpl/tplbeego"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type HandlerFunc func(c *Context)

func Handle(h HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &Context{
			c,
		}
		h(ctx)
	}
}

type Context struct {
	*gin.Context
}

func (c *Context) Tpl() *tplbeego.Tmpl {
	return c.MustGet(TPL).(*tplbeego.Tmpl)
}

//根据页面获取seo
//@param			page			页面标识
//@param			defSeo			默认的seo的map，必须有title、keywords和description字段
func (c *Context) GetSeoByPage(page string, defSeo map[string]string) {
	var seo mysql.Seo

	defSeo["sitename"] = dao.Global.Get(constx.SITE_NAME)
	if seo.Id > 0 {
		for k, v := range defSeo {
			seo.Title = strings.Replace(seo.Title, fmt.Sprintf("{%v}", k), v, -1)
			seo.Keywords = strings.Replace(seo.Keywords, fmt.Sprintf("{%v}", k), v, -1)
			seo.Description = strings.Replace(seo.Description, fmt.Sprintf("{%v}", k), v, -1)
		}
	}
	c.Tpl().Data["SeoTitle"] = seo.Title
	c.Tpl().Data["SeoKeywords"] = seo.Keywords
	c.Tpl().Data["SeoDescription"] = seo.Description
}

func (c *Context) GetSeoPage() {
	var seo mysql.Seo
	mus.Db.Where("page = ?", c.Request.URL.Path).Find(&seo)
	if seo.Id > 0 {
		c.Tpl().Data["SeoTitle"] = seo.Title
		c.Tpl().Data["SeoKeywords"] = seo.Keywords
		c.Tpl().Data["SeoDescription"] = seo.Description
	}

}

func (c *Context) Html(path string) {
	c.Tpl().SetTplPath(path)
	c.GetSeoPage()
	rb, err := c.Tpl().RenderBytes()
	if err != nil {
		mus.Logger.Error("html error", zap.Error(err))
		c.String(401, "error4")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Writer.Write(rb)
}

func (c *Context) Html404() {
	c.Tpl().SetTplPath("errors/404")
	rb, err := c.Tpl().RenderBytes()
	if err != nil {
		c.String(401, "error3")
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Writer.Write(rb)
}

// JSONResult json
type JSONResult struct {
	Code    int         `json:"code"`
	Message string      `json:"msg"`
	Data    interface{} `json:"data"`
}

// JSON 提供了系统标准JSON输出方法。
func (c *Context) JSONCode(Code int, data ...interface{}) {
	result := new(JSONResult)
	result.Code = Code
	info, ok := code.CodeMap[Code]
	if ok {
		result.Message = info
	} else {
		result.Message = "error"
	}

	if len(data) > 0 {
		result.Data = data[0]
	} else {
		result.Data = ""
	}
	c.JSON(http.StatusOK, result)
	return
}

func (c *Context) JSONOK(result ...interface{}) {
	j := new(JSONResult)
	j.Code = 0
	j.Message = "成功"
	if len(result) > 0 {
		j.Data = result[0]
	} else {
		j.Data = ""
	}
	c.JSON(http.StatusOK, j)
	return
}

// JSON 提供了系统标准JSON输出方法。
func (c *Context) JSONErr(Code int, err error) {
	result := new(JSONResult)
	result.Code = Code
	info, ok := code.CodeMap[Code]
	if ok {
		result.Message = info
	} else {
		result.Message = "error"
	}
	if err != nil {
		fmt.Println("code is", Code, "info is", result.Message, "============== err is", err.Error())
	}

	c.JSON(http.StatusOK, result)
	return
}

func (c *Context) JSONErrTips(msg string, err error) {
	result := new(JSONResult)
	result.Code = code.MsgErr
	if err != nil {
		fmt.Println("info is", result.Message, "============== err is", err.Error())
	}
	result.Message = msg
	c.JSON(http.StatusOK, result)
	return
}

func (c *Context) JSONErrStr(Code int, err string) {
	c.JSONErr(Code, errors.New(err))
	return
}

func (c *Context) BaseUrl() string {
	return BaseUrl(c.Context)
}

// ExecuteViewPathTemplate 执行指定的模板并返回执行结果.
func (c *Context) ExecuteViewPathTemplate(tplName string, data interface{}) (string, error) {
	var buf bytes.Buffer
	viewPath := "views"
	if err := tplbeego.ExecuteViewPathTemplate(&buf, tplName, viewPath, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// UpdateUser updates the User object stored in the session. This is useful incase a change
// is made to the user model that needs to persist across requests.
func (c *Context) UpdateUser(a *mysql.Member) error {
	s := sessions.Default(c.Context)
	s.Options(sessions.Options{
		Path:     "/",
		MaxAge:   24 * 3600 * 30,
		Secure:   false,
		HttpOnly: true,
	})
	s.Set(FrontSessionKey, a)
	return s.Save()
}

// Logout will clear out the session and call the Logout() user function.
func (c *Context) Logout() error {
	s := sessions.Default(c.Context)
	s.Options(sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		Secure:   false,
		HttpOnly: true,
	})

	s.Delete(FrontSessionKey)
	return s.Save()
}

func (c *Context) Member() *mysql.Member {
	var resp *mysql.Member
	respI, flag := c.Get(FrontContextKey)
	if flag {
		resp = respI.(*mysql.Member)
	}
	return resp

}

// SaveToFile saves uploaded file to new path.
// it only operates the first one of mutil-upload form file field.
func (c *Context) SaveToFile(fromfile, tofile string) error {
	fileHeader, err := c.FormFile(fromfile)
	if err != nil {
		return err
	}
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	f, err := os.OpenFile(tofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, file)
	return nil
}

func (c *Context) GetPostFormString(key string) string {
	value, _ := c.GetPostForm(key)
	return value
}

func (c *Context) ForbidGeneralRole() bool {
	// 如果只有作者和管理员才能写作的话，那么已创建了项目的普通用户无法将项目转为公开或者是私密分享
	if c.Member().Role == conf.MemberGeneralRole && dao.Global.GetOptionValue("ALL_CAN_WRITE_BOOK", "true") != "true" {
		return true
	}
	return false
}

func (c *Context) SetSession(key, value string) error {
	s := sessions.Default(c.Context)
	s.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		Secure:   false,
		HttpOnly: true,
	})

	s.Set(key, value)
	return s.Save()
}

func (c *Context) GetSession(key string) interface{} {
	return sessions.Default(c.Context).Get(key)
}

func (c *Context) Download(file string, filename ...string) {
	// check get file error, file not found or other error.
	if _, err := os.Stat(file); err != nil {
		http.ServeFile(c.Context.Writer, c.Context.Request, file)
		return
	}

	var fName string
	if len(filename) > 0 && filename[0] != "" {
		fName = filename[0]
	} else {
		fName = filepath.Base(file)
	}
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(fName))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	http.ServeFile(c.Context.Writer, c.Context.Request, file)
}

//替换链接
//如果是summary，则根据这个进行排序调整
func (c *Context) ReplaceLinks(bookIdentify string, docHtml string, isSummary ...bool) string {
	var (
		book mysql.Book
		docs []mysql.Document
	)

	mus.Db.Select("book_id").Where("identify = ?", bookIdentify).Where(&book)
	if book.BookId > 0 {
		mus.Db.Select("identify, document_id").Where("book_id = ?", book.BookId).Limit(5000).Find(&docs)
		if len(docs) > 0 {
			Links := make(map[string]string)
			for _, doc := range docs {
				idStr := strconv.Itoa(doc.DocumentId)
				if len(doc.Identify) > 0 {
					Links["$"+strings.ToLower(doc.Identify)] = "/read/" + bookIdentify + "/" + doc.Identify + "||" + idStr
				}
				if doc.DocumentId > 0 {
					Links["$"+strconv.Itoa(doc.DocumentId)] = "/read/" + bookIdentify + "/" + strconv.Itoa(doc.DocumentId) + "||" + idStr
				}
			}

			//替换文档内容中的链接
			if gq, err := goquery.NewDocumentFromReader(strings.NewReader(docHtml)); err == nil {

				gq.Find("a").Each(func(i int, selection *goquery.Selection) {
					if href, ok := selection.Attr("href"); ok && strings.HasPrefix(href, "$") {
						if slice := strings.Split(href, "#"); len(slice) > 1 {
							if newHref, ok := Links[strings.ToLower(slice[0])]; ok {
								arr := strings.Split(newHref, "||") //整理的arr数组长度，肯定为2，所以不做数组长度判断
								selection.SetAttr("href", arr[0]+"#"+strings.Join(slice[1:], "#"))
								selection.SetAttr("data-pid", arr[1])
							}
						} else {
							if newHref, ok := Links[strings.ToLower(href)]; ok {
								arr := strings.Split(newHref, "||") //整理的arr数组长度，肯定为2，所以不做数组长度判断
								selection.SetAttr("href", arr[0])
								selection.SetAttr("data-pid", arr[1])
							}
						}
					}
				})

				if newHtml, err := gq.Find("body").Html(); err == nil {
					docHtml = newHtml
					if len(isSummary) > 0 && isSummary[0] == true { //更新排序
						c.sortBySummary(bookIdentify, docHtml, book.BookId) //更新排序
					}
				}
			} else {
				mus.Logger.Error(err.Error())
			}
		}
	}

	return docHtml
}

//在markdown头部加上<bookstack></bookstack>或者<bookstack/>，即解析markdown中的ul>li>a链接作为目录
func (c *Context) sortBySummary(bookIdentify, htmlStr string, bookId int) {
	//debug := beego.AppConfig.String("runmod") != "prod"
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		mus.Logger.Error(err.Error())
		return
	}
	idx := 1
	//if debug {
	//	beego.Info("根据summary文件进行排序")
	//}

	//查找ul>li下的所有a标签，并提取text和href，查询数据库，如果标识不存在，则把这些新的数据录入数据库
	var hrefs = make(map[string]string)
	var hrefSlice []interface{}
	var docs []mysql.Document
	doc.Find("li>a").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("href"); ok && strings.HasPrefix(href, "$") {
			href = strings.TrimLeft(strings.Replace(href, "/", "-", -1), "$")
			hrefs[href] = selection.Text()
			hrefSlice = append(hrefSlice, href)
		}
	})
	//if debug {
	//	beego.Info(hrefs)
	//}
	if len(hrefSlice) > 0 {

		err = mus.Db.Select("identify").Where("book_id = ? and identify in (?)", bookId, hrefSlice).Limit(len(hrefSlice)).Find(&docs).Error
		if err != nil {
			mus.Logger.Error(err.Error())
		} else {
			for _, doc := range docs {
				//删除存在的标识
				delete(hrefs, doc.Identify)
			}
		}
	}
	if len(hrefs) > 0 { //存在未创建的文档，先创建
		for identify, docName := range hrefs {
			doc := mysql.Document{
				BookId:       bookId,
				Identify:     identify,
				DocumentName: docName,
				CreateTime:   time.Now(),
				ModifyTime:   time.Now(),
			}
			if docId, err := dao.Document.InsertOrUpdate(mus.Db, &doc); err == nil {
				err = dao.DocumentStore.InsertOrUpdate(mus.Db, &mysql.DocumentStore{
					DocumentId: int(docId),
					Markdown:   "[TOC]\n\r\n\r",
				})

			}
		}

	}

	// 重置所有之前的文档排序
	mus.Db.Model(mysql.Document{}).Where("book_id = ?", bookId).Update(map[string]interface{}{
		"order_sort": 100000,
	})

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		docName := selection.Text()
		pid := 0
		if docId, exist := selection.Attr("data-pid"); exist {
			did, _ := strconv.Atoi(docId)
			eleParent := selection.Parent().Parent().Parent()
			if eleParent.Is("li") {
				fst := eleParent.Find("a").First()
				pidstr, _ := fst.Attr("data-pid")
				//如果这里的pid为0，表示数据库还没存在这个标识，需要创建
				pid, _ = strconv.Atoi(pidstr)
			}
			if did > 0 {
				mus.Db.Table("md_documents").Where("book_id = ? and document_id = ?", bookId, did).Update(map[string]interface{}{
					"parent_id": pid, "document_name": docName,
					"order_sort": idx, "modify_time": time.Now(),
				})
			}
		} else if href, ok := selection.Attr("href"); ok && strings.HasPrefix(href, "$") {
			identify := strings.TrimPrefix(href, "$") //文档标识
			eleParent := selection.Parent().Parent().Parent()
			if eleParent.Is("li") {
				if parentHref, ok := eleParent.Find("a").First().Attr("href"); ok {
					var one mysql.Document
					mus.Db.Select("document_id").Where("book_id = ? and identify = ?", bookId, strings.Split(strings.TrimPrefix(parentHref, "$"), "#")[0])
					pid = one.DocumentId
				}
			}

			err = mus.Db.Table("md_documents").Where("book_id = ? and identify = ?", bookId, identify).Update(map[string]interface{}{
				"parent_id": pid, "document_name": docName,
				"order_sort": idx, "modify_time": time.Now(),
			}).Error

			if err != nil {
				mus.Logger.Error(err.Error())
			}
		}
		idx++
	})

	if len(hrefs) > 0 { //如果有新创建的文档，则再调用一遍，用于处理排序
		c.ReplaceLinks(bookIdentify, htmlStr, true)
	}
}

func (c *Context) SaveToFileImg(fromfile string) (tofile string, name string, err error) {
	var fileHeader *multipart.FileHeader
	fileHeader, err = c.FormFile(fromfile)
	if err != nil {
		return
	}
	var file multipart.File
	file, err = fileHeader.Open()
	if err != nil {
		return
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)

	if !strings.EqualFold(ext, ".png") && !strings.EqualFold(ext, ".jpg") && !strings.EqualFold(ext, ".gif") && !strings.EqualFold(ext, ".jpeg") {
		err = errors.New("img type error")
		return
	}

	fileName := strconv.FormatInt(time.Now().UnixNano(), 16)
	name = fileName + ext
	tofile = filepath.Join("uploads", time.Now().Format("200601"), name)

	path := filepath.Dir(tofile)
	// todo 优化
	err = os.MkdirAll(path, os.ModePerm)
	f, err := os.OpenFile(tofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(f, file)
	return
}

func (c *Context) JSONList(data interface{}, current, pageSize, total int) {
	j := new(JSONResult)
	j.Code = 0
	j.Message = "ok"
	j.Data = RespList{
		List: data,
		Pagination: struct {
			Current  int `json:"current"`
			PageSize int `json:"pageSize"`
			Total    int `json:"total"`
		}{
			Current:  current,
			PageSize: pageSize,
			Total:    total,
		},
	}
	c.JSON(http.StatusOK, j)
	return
}
