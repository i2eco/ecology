package book

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/TruthHun/gotil/filetil"
	"github.com/TruthHun/gotil/mdtil"
	"github.com/TruthHun/gotil/util"
	"github.com/TruthHun/gotil/ziptil"
	"github.com/TruthHun/html2md"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/code"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/ecology/appgo/router/core"
	"github.com/jinzhu/gorm"
	"github.com/russross/blackfriday"
	"go.uber.org/zap"
)

// 替换字符串
func Replace(c *core.Context) {
	identify := c.Param("key")
	src, _ := c.GetQuery("src")
	dst, _ := c.GetQuery("dst")

	member := c.Member()
	uid := member.MemberId
	if uid <= 0 {
		c.JSONErrStr(code.MsgErr, "给文档打分失败，请先登录再操作")
		return
	}

	book, err := dao.Book.ResultFindByIdentify(identify, member.MemberId)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSONErr(code.MsgErr, err)
		return
	}

	if err == gorm.ErrRecordNotFound {
		c.JSONErrStr(code.MsgErr, "内容不存在")
		return
	}

	dao.Book.Replace(book.BookId, src, dst)
	c.JSONOK("替换成功")
}

//收藏书籍
func Star(c *core.Context) {
	member := c.Member()
	uid := member.MemberId
	if uid <= 0 {
		c.JSONErrStr(code.MsgErr, "member id error")
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSONErrStr(code.MsgErr, "id error")
		return
	}

	cancel, err := dao.Star.Star(uid, id)
	//data := map[string]bool{"IsCancel": cancel}
	if err != nil {
		if cancel {
			c.JSONErrStr(code.MsgErr, "取消收藏失败")
			return
		}
		c.JSONErrStr(code.MsgErr, "添加收藏失败")
		return
	}

	if cancel {
		c.JSONOK(code.MsgErr, "取消收藏成功")
		return
	}
	c.JSONOK(code.MsgErr, "添加收藏成功")
}

//给文档项目打分
func Score(c *core.Context) {
	member := c.Member()
	uid := member.MemberId
	if uid <= 0 {
		c.JSONErrStr(code.MsgErr, "给文档打分失败，请先登录再操作")
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSONErrStr(code.MsgErr, "id error")
		return
	}

	scoreStr, _ := c.GetPostForm("score")
	score, _ := strconv.Atoi(scoreStr)

	if err := dao.Score.AddScore(c.Context, uid, id, score); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	c.JSONOK("感谢您给当前文档打分")
}

//添加评论
func Comment(c *core.Context) {
	member := c.Member()
	uid := member.MemberId
	if uid <= 0 {
		c.JSONErrStr(code.MsgErr, "请先登录在评论")
		return
	}
	content, _ := c.GetPostForm("content")
	if l := len(content); l < 5 || l > 256 {
		c.JSONErrStr(code.MsgErr, "评论内容限 5 - 256 个字符")
		return
	}
	bookId, _ := strconv.Atoi(c.Param("id"))
	if bookId <= 0 {
		c.JSONErrStr(code.MsgErr, "文档项目不存在")
		return
	}
	err := dao.Comments.AddComments(member.MemberId, bookId, content)
	if err != nil {
		c.JSONErr(1, err)
		return
	}
	c.JSONOK("评论成功")
}

//上传项目
func UploadProject(c *core.Context) {
	member := c.Member()
	uid := member.MemberId
	if uid <= 0 {
		c.JSONErrStr(code.MsgErr, "请先登录在上传")
		return
	}

	//处理步骤
	//1、接受上传上来的zip文件，并存放到store/temp目录下
	//2、解压zip到当前目录，然后移除非图片文件
	//3、将文件夹移动到uploads目录下

	identify, _ := c.GetPostForm("identify")

	if !dao.Book.HasProjectAccess(identify, member.MemberId, conf.Conf.Info.BookEditor) {
		c.JSONErrStr(code.MsgErr, "无操作权限")
		return
	}

	book, _ := dao.Book.ResultFindByIdentify(identify, member.MemberId)
	if book.BookId == 0 {
		c.JSONErrStr(code.MsgErr, "项目不存在")
		return
	}

	fh, err := c.FormFile("zipfile")
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	if strings.ToLower(filepath.Ext(fh.Filename)) != ".zip" && strings.ToLower(filepath.Ext(fh.Filename)) != ".epub" {
		c.JSONErrStr(code.MsgErr, "请上传指定格式文件")
		return
	}
	tmpFile := "store/" + identify + ".zip" //保存的文件名
	err = c.SaveToFile("zipfile", tmpFile)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
	}
	go unzipToData(book.BookId, member.MemberId, identify, tmpFile, fh.Filename)
	c.JSONOK("上传成功")
}

//从github等拉取下载markdown项目
func DownloadProject(c *core.Context) {

	//处理步骤
	//1、接受上传上来的zip文件，并存放到store/temp目录下
	//2、解压zip到当前目录，然后移除非图片文件
	//3、将文件夹移动到uploads目录下
	member := c.Member()
	if _, err := isPermission(c); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	//普通用户没有权限
	if member.Role > 1 {
		c.JSONErrStr(1, "您没有操作权限")
		return
	}

	identify := c.Param("identify")
	book, _ := dao.Book.ResultFindByIdentify(identify, member.MemberId)
	if book.BookId == 0 {
		c.JSONErrStr(1, "导入失败，只有项目创建人才有权限导入项目")
		return
	}
	//GitHub项目链接
	link, _ := c.GetQuery("link")
	if strings.ToLower(filepath.Ext(link)) != ".zip" {
		c.JSONErrStr(1, "只支持拉取zip压缩的markdown项目")
		return
	}
	go func() {
		if file, err := util.CrawlFile(link, "store", 60); err != nil {
			mus.Logger.Error("crawl file", zap.Error(err))
		} else {
			unzipToData(book.BookId, member.MemberId, identify, file, filepath.Base(file))
		}
	}()

	c.JSONOK("提交成功。下载任务已交由后台执行")
}

//发布项目.
func Release(c *core.Context) {
	identify := c.Param("key")
	bookId := 0
	member := c.Member()

	if identify == "" {
		c.JSONCode(code.MsgErr)
		return
	}

	if member.IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			mus.Logger.Error("release error", zap.Error(err))
		}
		bookId = book.BookId
	} else {
		book, err := dao.Book.ResultFindByIdentify(identify, member.MemberId)
		if err != nil {
			c.JSONErrStr(6003, "未知错误")
			return
		}
		if book.RoleId != conf.BookAdmin && book.RoleId != conf.BookFounder && book.RoleId != conf.BookEditor {
			c.JSONErrStr(6003, "权限不足")
			return
		}
		bookId = book.BookId
	}

	if exist := utils.BooksRelease.Exist(bookId); exist {
		c.JSONErrStr(1, "上次内容发布正在执行中，请稍后再操作")
		return
	}

	go func(identify string) {
		dao.Document.ReleaseContent(bookId, c.BaseUrl())
	}(identify)

	c.JSONOK("发布任务已推送到任务队列，稍后将在后台执行。")
}

//文档排序.
func SaveSort(c *core.Context) {
	identify := c.Param("key")
	if identify == "" {
		c.JSONErrStr(code.MsgErr, "err")
		return
	}

	member := c.Member()

	bookId := 0
	if member.IsAdministrator() {
		book, err := dao.Book.FindByFieldFirst("identify", identify)
		if err != nil {
			mus.Logger.Error(err.Error())
			return
		}
		bookId = book.BookId
	} else {
		bookResult, err := dao.Book.ResultFindByIdentify(identify, member.MemberId)
		if err != nil {
			mus.Logger.Error(err.Error())
			c.JSONErr(code.MsgErr, err)
			return
		}

		if bookResult.RoleId == conf.BookObserver {
			c.JSONErrStr(6002, "项目不存在或权限不足")
			return
		}
		bookId = bookResult.BookId
	}

	body, _ := ioutil.ReadAll(c.Request.Body)
	var docs []struct {
		Id     int `json:"id"`
		Sort   int `json:"sort"`
		Parent int `json:"parent"`
	}

	err := json.Unmarshal(body, &docs)
	if err != nil {
		c.JSONErrStr(6003, "数据错误")
		return
	}
	qs := mus.Db.Model(mysql.Document{}).Where("book_id = ?", bookId)
	now := time.Now()
	for _, item := range docs {
		err = qs.Where("document_id = ?", item.Id).Updates(mysql.Ups{
			"parent_id":   item.Parent,
			"order_sort":  item.Sort,
			"modify_time": now,
		}).Error
		if err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
	}
	c.JSONOK()
}

// 从Git仓库拉取项目
func GitPull(c *core.Context) {
	//处理步骤
	//1、接受上传上来的zip文件，并存放到store/temp目录下
	//2、解压zip到当前目录，然后移除非图片文件
	//3、将文件夹移动到uploads目录下

	identify, _ := c.GetPostForm("identify")
	member := c.Member()
	if !dao.Book.HasProjectAccess(identify, member.MemberId, conf.BookEditor) {
		c.JSONErrStr(1, "无操作权限")
		return
	}

	book, _ := dao.Book.ResultFindByIdentify(identify, member.MemberId)
	if book.BookId == 0 {
		c.JSONErrStr(1, "导入失败，只有项目创建人才有权限导入项目")
		return
	}
	//GitHub项目链接
	link, _ := c.GetPostForm("link")
	go func() {
		folder := "store/" + identify
		mus.Logger.Info("git clone", zap.String("link", link), zap.String("folder", folder))
		err := utils.GitClone(link, folder)

		if err != nil {
			mus.Logger.Error("git clone error", zap.Error(err))
		} else {
			loadByFolder(book.BookId, member.MemberId, identify, folder)
		}
	}()

	c.JSONOK("提交成功，请耐心等待。")
}

//将zip压缩文件解压并录入数据库
//@param            book_id             项目id(其实有想不标识了可以不要这个的，但是这里的项目标识只做目录)
//@param            identify            项目标识
//@param            zipfile             压缩文件
//@param            originFilename      上传文件的原始文件名
func unzipToData(bookId, memberId int, identify, zipFile, originFilename string) {

	//说明：
	//OSS中的图片存储规则为"projects/$identify/项目中图片原路径"
	//本地存储规则为"uploads/projects/$identify/项目中图片原路径"
	var err error
	projectRoot := "" //项目根目录

	//解压目录
	unzipPath := "store/" + identify

	//如果存在相同目录，则率先移除
	err = os.RemoveAll(unzipPath)

	if err != nil {
		// todo log
		return
	}
	err = os.MkdirAll(unzipPath, os.ModePerm)
	if err != nil {
		return
	}

	imgMap := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".svg": true, ".webp": true}

	defer func() {
		err = os.Remove(zipFile) //最后删除上传的临时文件
		if err != nil {
			// todo log
			return
		}
		err = os.RemoveAll(unzipPath) //删除解压后的文件夹
		if err != nil {
			// todo log
			return
		}
	}()

	//注意：这里的prefix必须是判断是否是GitHub之前的prefix
	err = ziptil.Unzip(zipFile, unzipPath)
	if err != nil {
		mus.Logger.Error("解压失败", zap.String("zipfile", zipFile), zap.Error(err))
		return
	}

	//读取文件，把图片文档录入oss
	if files, err := filetil.ScanFiles(unzipPath); err == nil {
		projectRoot = getProjectRoot(files)
		replaceToAbs(projectRoot, identify)

		//文档对应的标识
		for _, file := range files {
			if !file.IsDir {
				ext := strings.ToLower(filepath.Ext(file.Path))
				imgUrl := "eco-doc-img/book/" + identify + "/" + strings.TrimPrefix(file.Path, projectRoot)

				if ok, _ := imgMap[ext]; ok { //图片，录入oss
					err = mus.Oss.PutObjectFromFile(imgUrl, file.Path)
					if err != nil {
						mus.Logger.Error("unzip file img put to oss error", zap.String("filepath", file.Path), zap.Error(err))
					}
				} else if ext == ".md" || ext == ".markdown" || ext == ".html" { //markdown文档，提取文档内容，录入数据库
					doc := new(mysql.Document)
					var (
						mdcont  string
						htmlStr string
						b       []byte
						err     error
						docId   int
					)
					b, err = ioutil.ReadFile(file.Path)
					if err != nil {
						mus.Logger.Error("读取文档失败", zap.String("path", file.Path), zap.Error(err))
						continue
					}

					if ext == ".md" || ext == ".markdown" {
						mdcont = strings.TrimSpace(string(b))
						htmlStr = mdtil.Md2html(mdcont)
					} else {
						htmlStr = string(b)
						mdcont = html2md.Convert(htmlStr)
					}
					if !strings.HasPrefix(mdcont, "[TOC]") {
						mdcont = "[TOC]\r\n\r\n" + mdcont
					}
					doc.DocumentName = utils.ParseTitleFromMdHtml(htmlStr)
					doc.BookId = bookId
					//文档标识
					doc.Identify = strings.Replace(strings.Trim(strings.TrimPrefix(file.Path, projectRoot), "/"), "/", "-", -1)
					doc.Identify = strings.Replace(doc.Identify, ")", "", -1)
					doc.MemberId = memberId
					doc.OrderSort = 1
					if strings.HasSuffix(strings.ToLower(file.Name), "summary.md") {
						doc.OrderSort = 0
					}
					if strings.HasSuffix(strings.ToLower(file.Name), "summary.html") {
						mdcont += "<bookstack-summary></bookstack-summary>"
						// 生成带$的文档标识，阅读BaseController.go代码可知，
						// 要使用summary.md的排序功能，必须在链接中带上符号$
						mdcont = strings.Replace(mdcont, "](", "]($", -1)
						// 去掉可能存在的url编码的右括号，否则在url译码后会与markdown语法混淆
						mdcont = strings.Replace(mdcont, "%29", "", -1)
						mdcont, _ = url.QueryUnescape(mdcont)
						doc.OrderSort = 0
						doc.Identify = "summary.md"
					}

					docId, err = dao.Document.InsertOrUpdate(mus.Db, doc)
					if err != nil {
						mus.Logger.Error("doc insert or update error", zap.Error(err))
						continue
					}

					err = dao.DocumentStore.InsertOrUpdate(mus.Db, &mysql.DocumentStore{
						DocumentId: int(docId),
						Markdown:   mdcont,
					})
					if err != nil {
						mus.Logger.Error("store doc insert or update error", zap.Error(err))
					}

				}
			}
		}
	}
}

//获取文档项目的根目录
func getProjectRoot(fl []filetil.FileList) (root string) {
	//获取项目的根目录(感觉这个函数封装的不是很好，有更好的方法，请通过issue告知我，谢谢。)
	i := 1000
	for _, f := range fl {
		if !f.IsDir {
			if cnt := strings.Count(f.Path, "/"); cnt < i {
				root = filepath.Dir(f.Path)
				i = cnt
			}
		}
	}
	return
}

//查找并替换markdown文件中的路径，把图片链接替换成url的相对路径，把文档间的链接替换成【$+文档标识链接】
func replaceToAbs(folder string, identify string) {
	files, _ := filetil.ScanFiles(folder)
	for _, file := range files {
		if ext := strings.ToLower(filepath.Ext(file.Path)); ext == ".md" || ext == ".markdown" {
			//mdb ==> markdown byte
			mdb, _ := ioutil.ReadFile(file.Path)
			mdCont := string(mdb)
			basePath := filepath.Dir(file.Path)
			basePath = strings.Trim(strings.Replace(basePath, "\\", "/", -1), "/")
			basePathSlice := strings.Split(basePath, "/")
			l := len(basePathSlice)
			b, _ := ioutil.ReadFile(file.Path)
			output := blackfriday.Run(b)
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(output)))
			imgUrl := "eco-doc-img/book/" + identify + "/" + strings.TrimPrefix(file.Path, folder)

			//图片链接处理
			doc.Find("img").Each(func(i int, selection *goquery.Selection) {
				//非http开头的图片地址，即是相对地址
				if src, ok := selection.Attr("src"); ok && !strings.HasPrefix(strings.ToLower(src), "http") {
					newSrc := src                                  //默认为旧地址
					if cnt := strings.Count(src, "../"); cnt < l { //以或者"../"开头的路径
						newSrc = strings.Join(basePathSlice[0:l-cnt], "/") + "/" + strings.TrimLeft(src, "./")
					}
					newSrc = imgUrl
					mdCont = strings.Replace(mdCont, src, newSrc, -1)
				}
			})

			//a标签链接处理。要注意判断有锚点的情况
			doc.Find("a").Each(func(i int, selection *goquery.Selection) {
				if href, ok := selection.Attr("href"); ok && !strings.HasPrefix(strings.ToLower(href), "http") && !strings.HasPrefix(href, "#") {
					newHref := href //默认
					if cnt := strings.Count(href, "../"); cnt < l {
						newHref = strings.Join(basePathSlice[0:l-cnt], "/") + "/" + strings.TrimLeft(href, "./")
					}
					newHref = strings.TrimPrefix(strings.Trim(newHref, "/"), folder)
					if !strings.HasPrefix(href, "$") { //原链接不包含$符开头，否则表示已经替换过了。
						newHref = "$" + strings.Replace(strings.Trim(newHref, "/"), "/", "-", -1)
						slice := strings.Split(newHref, "$")
						if ll := len(slice); ll > 0 {
							newHref = "$" + slice[ll-1]
						}
						mdCont = strings.Replace(mdCont, "]("+href, "]("+newHref, -1)
					}
				}
			})
			ioutil.WriteFile(file.Path, []byte(mdCont), os.ModePerm)
		}
	}
}

// 判断是否具有管理员或管理员以上权限
func isPermission(c *core.Context) (*mysql.BookResult, error) {
	member := c.Member()
	identify := c.Param("identify")

	book, err := dao.Book.ResultFindByIdentify(identify, member.MemberId)
	if err != nil {
		return book, err
	}

	if book.RoleId != conf.BookAdmin && book.RoleId != conf.BookFounder {
		return book, errors.New("权限不足")
	}
	return book, nil
}

func isFormPermission(c *core.Context) (*mysql.BookResult, error) {
	member := c.Member()
	identify, _ := c.GetPostForm("identify")
	book, err := dao.Book.ResultFindByIdentify(identify, member.MemberId)
	if err != nil {
		return book, err
	}

	if book.RoleId != conf.BookAdmin && book.RoleId != conf.BookFounder {
		return book, errors.New("权限不足")
	}
	return book, nil
}

func loadByFolder(bookId int, memberId int, identify, folder string) {
	mus.Logger.Info("load folder start", zap.Any("bookId", bookId), zap.Int("memberId", memberId), zap.Any("identify", identify), zap.String("folder", folder))

	//说明：

	imgMap := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".svg": true, ".webp": true}

	defer func() {
		err := os.RemoveAll(folder) //删除解压后的文件夹
		mus.Logger.Error("load folder remove error", zap.Error(err), zap.String("folder", folder))
	}()

	//注意：这里的prefix必须是判断是否是GitHub之前的prefix

	//读取文件，把图片文档录入oss
	files, err := filetil.ScanFiles(folder)
	if err != nil {
		mus.Logger.Error("load folder ScanFiles error", zap.Error(err), zap.String("folder", folder))
		return
	}

	mus.Logger.Info("load folder files", zap.Any("files", files), zap.String("folder", folder))

	replaceToAbs(folder, identify)

	//文档对应的标识
	for _, file := range files {
		if !file.IsDir {
			ext := strings.ToLower(filepath.Ext(file.Path))
			imgUrl := "eco-doc-img/book/" + identify + "/" + strings.TrimPrefix(file.Path, folder)

			if ok, _ := imgMap[ext]; ok { //图片，录入oss

				err = mus.Oss.PutObjectFromFile(imgUrl, file.Path)
				if err != nil {
					mus.Logger.Error("file img put to oss error", zap.String("filepath", file.Path), zap.Error(err))
				}
			} else if ext == ".md" || ext == ".markdown" { //markdown文档，提取文档内容，录入数据库
				doc := new(mysql.Document)
				if b, err := ioutil.ReadFile(file.Path); err == nil {
					mdCont := strings.TrimSpace(string(b))
					if !strings.HasPrefix(mdCont, "[TOC]") {
						mdCont = "[TOC]\r\n\r\n" + mdCont
					}
					htmlStr := mdtil.Md2html(mdCont)
					doc.DocumentName = utils.ParseTitleFromMdHtml(htmlStr)
					doc.BookId = bookId
					//文档标识
					doc.Identify = strings.Replace(strings.Trim(strings.TrimPrefix(file.Path, folder), "/"), "/", "-", -1)
					doc.MemberId = memberId
					doc.OrderSort = 1
					if strings.HasSuffix(strings.ToLower(file.Name), "summary.md") {
						doc.OrderSort = 0
					}
					if docId, err := dao.Document.InsertOrUpdate(mus.Db, doc); err == nil {
						if err := dao.DocumentStore.InsertOrUpdate(mus.Db, &mysql.DocumentStore{
							DocumentId: int(docId),
							Markdown:   mdCont,
							Content:    "",
						}); err != nil {
							mus.Logger.Error("loadByFolder error1", zap.Error(err))
						}
					} else {
						mus.Logger.Error("loadByFolder error2", zap.Error(err))
					}

				} else {
					mus.Logger.Error("loadByFolder error3", zap.Error(err))
				}

			}
		}
	}
}
