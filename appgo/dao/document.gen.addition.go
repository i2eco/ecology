package dao

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/TruthHun/converter/converter"
	"github.com/TruthHun/gotil/cryptil"
	"github.com/TruthHun/gotil/util"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/orm"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/model/mysql/store"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/md"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

//根据文档ID查询指定文档.
func (m *document) Find(id int) (doc *mysql.Document, err error) {
	if id <= 0 {
		return doc, ErrInvalidParameter
	}
	var one mysql.Document
	err = mus.Db.Where("document_id = ?", id).Find(&one).Error
	return &one, err
}

//插入和更新文档.
//存在文档id或者文档标识，则表示更新文档内容
func (m *document) InsertOrUpdate(db *gorm.DB, documents *mysql.Document, cols ...string) (id int, err error) {
	id = documents.DocumentId
	if documents.DocumentId > 0 { //文档id存在，则更新
		err = db.Model(documents).Where("document_id = ?", documents.DocumentId).Update(map[string]interface{}{
			"modify_time":   time.Now(),
			"document_name": strings.TrimSpace(documents.DocumentName),
		}).Error
		if err != nil {
			return
		}
		return
	}

	var mm mysql.Document
	//直接查询一个字段，优化MySQL IO

	mus.Db.Where("identify = ? and book_id = ?", documents.Identify, documents.BookId).Find(&mm)

	if mm.DocumentId == 0 {
		documents.CreateTime = time.Now()
		documents.ModifyTime = time.Now()
		err = db.Create(&documents).Error
		if err != nil {
			return
		}
		id = documents.DocumentId
		Book.ResetDocumentNumber(documents.BookId)
	} else { //identify存在，则执行更新
		err = db.Model(mysql.Document{}).Where("document_id = ?", documents.DocumentId).Update(map[string]interface{}{
			"modify_time":   time.Now(),
			"document_name": strings.TrimSpace(documents.DocumentName),
		}).Error
		if err != nil {
			return
		}
		id = mm.DocumentId
	}
	return
}

//根据指定字段查询一条文档.
func (m *document) FindByFieldFirst(field string, v interface{}) (one *mysql.Document, err error) {
	err = mus.Db.Where(field+"=?", v).Find(one).Error
	return
}

//根据指定字段查询一条文档.
func (m *document) FindByBookIdAndDocIdentify(bookId, identify interface{}) (resp *mysql.Document, err error) {
	var one mysql.Document
	err = mus.Db.Where("book_id = ? and identify = ?", bookId, identify).Find(&one).Error
	return &one, err
}

//递归删除一个文档.
func (m *document) RecursiveDocument(docId int) (err error) {
	var doc *mysql.Document
	doc, err = m.Find(docId)
	if err != nil {
		return
	}

	err = mus.Db.Delete(doc).Error
	if err != nil {
		return
	}

	DocumentStore.DeleteById(docId)
	// todo
	//NewDocumentHistory().Clear(docId)

	var docs []*mysql.Document

	err = mus.Db.Where("parent_id = ?", docId).Find(&docs).Error
	if err != nil {
		return
	}

	for _, item := range docs {
		docId := item.DocumentId
		mus.Db.Where("document_id=?", docId).Delete(mysql.Document{})
		//删除document_store表的文档
		DocumentStore.DeleteById(docId)
		m.RecursiveDocument(docId)
	}
	return
}

//发布文档内容为HTML
func (m *document) ReleaseContent(bookId int, baseUrl string) {
	// 加锁
	utils.BooksRelease.Set(bookId)
	defer utils.BooksRelease.Delete(bookId)

	var (
		docs        []*mysql.Document
		book        mysql.Book
		releaseTime = time.Now() //发布的时间戳
	)

	mus.Db.Where("book_id = ?", bookId).Find(&book)

	//全部重新发布。查询该书籍的所有文档id

	err := mus.Db.Select("document_id").Where("book_id = ?", bookId).Limit(20000).Find(&docs).Error
	if err != nil {
		mus.Logger.Error(err.Error())
		return
	}

	for _, item := range docs {
		ds, err := DocumentStore.GetById(item.DocumentId)
		if err != nil {
			mus.Logger.Error("document release error", zap.Int("docid", item.DocumentId), zap.Error(err))
			continue
		}

		if strings.TrimSpace(utils.GetTextFromHtml(strings.Replace(ds.Markdown, "[TOC]", "", -1))) == "" {
			// 如果markdown内容为空，则查询下一级目录内容来填充
			ds.Markdown, ds.Content = Document.BookStackAuto(bookId, ds.DocumentId)
			ds.Markdown = "[TOC]\n\n" + ds.Markdown
		} else if len(utils.GetTextFromHtml(ds.Content)) == 0 {
			//内容为空，渲染一下文档，然后再重新获取
			utils.RenderDocumentById(item.DocumentId)
			ds, _ = DocumentStore.GetById(item.DocumentId)
		}

		item.Release = ds.Content

		attachList, err := Attachment.FindListByDocumentId(item.DocumentId)
		if err == nil && len(attachList) > 0 {
			content := bytes.NewBufferString("<div class=\"attach-list\"><strong>附件</strong><ul>")
			for _, attach := range attachList {
				li := fmt.Sprintf("<li><a href=\"%s\" target=\"_blank\" title=\"%s\">%s</a></li>", attach.HttpPath, attach.FileName, attach.FileName)
				content.WriteString(li)
			}
			content.WriteString("</ul></div>")
			item.Release += content.String()
		}

		ds.Content = item.Release
		err = DocumentStore.InsertOrUpdate(mus.Db, &ds)

		if err != nil {
			mus.Logger.Error(err.Error())
		}

		err = mus.Db.Model(mysql.Document{}).Where("document_id = ?", item.DocumentId).Update(map[string]interface{}{
			"release": item.Release,
		}).Error
		if err != nil {
			mus.Logger.Error(err.Error())
		}

	}

	//最后再更新时间戳
	if err = mus.Db.Model(mysql.Book{}).Where("book_id = ?", bookId).Update(map[string]interface{}{
		"release_time": releaseTime,
	}).Error; err != nil {
		mus.Logger.Error(err.Error())
	}
	client := NewElasticSearchClient()
	client.RebuildAllIndex(bookId)
}

//离线文档生成
func (m *document) GenerateBook(book *mysql.Book, baseUrl string) {
	//将书籍id加入进去，表示正在生成离线文档
	utils.BooksGenerate.Set(book.BookId)
	defer utils.BooksGenerate.Delete(book.BookId) //最后移除

	//公开文档，才生成文档文件
	debug := true
	//if beego.AppConfig.String("runmode") == "prod" {
	//	debug = false
	//}

	Nickname := Member.GetNicknameByUid(book.MemberId)

	docs, err := m.FindListByBookId(book.BookId)
	if err != nil {
		mus.Logger.Error("generate book file list error", zap.Error(err))
		return
	}

	var ExpCfg = converter.Config{
		Contributor: conf.Conf.App.ExportCreator,
		Cover:       "",
		Creator:     conf.Conf.App.ExportCreator,
		Timestamp:   book.ReleaseTime.Format("2006-01-02"),
		Description: book.Description,
		Header:      conf.Conf.App.ExportHeader,
		Footer:      conf.Conf.App.ExportFooter,
		Identifier:  "",
		Language:    "zh-CN",
		Publisher:   conf.Conf.App.ExportCreator,
		Title:       book.BookName,
		Format:      []string{"epub", "mobi", "pdf"},
		FontSize:    conf.Conf.App.ExportFontSize,
		PaperSize:   conf.Conf.App.ExportPagerSize,
		More: []string{
			"--pdf-page-margin-bottom", conf.Conf.App.ExportMarginBottom,
			"--pdf-page-margin-left", conf.Conf.App.ExportMarginLeft,
			"--pdf-page-margin-right", conf.Conf.App.ExportMarginRight,
			"--pdf-page-margin-top", conf.Conf.App.ExportMarginTop,
		},
	}

	folder := fmt.Sprintf("cache/books/%v/", book.Identify)
	os.MkdirAll(folder, os.ModePerm)
	if !debug {
		defer os.RemoveAll(folder)
	}

	//生成致谢信内容
	if htmlStr, err := utils.ExecuteViewPathTemplate("document/tpl_statement.html", map[string]interface{}{"Model": book, "Nickname": Nickname, "Date": ExpCfg.Timestamp}); err == nil {
		h1Title := "说明"
		if doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr)); err == nil {
			h1Title = doc.Find("h1").Text()
		}
		toc := converter.Toc{
			Id:    time.Now().Nanosecond(),
			Pid:   0,
			Title: h1Title,
			Link:  "statement.html",
		}
		htmlname := folder + toc.Link
		ioutil.WriteFile(htmlname, []byte(htmlStr), os.ModePerm)
		ExpCfg.Toc = append(ExpCfg.Toc, toc)
	}
	for _, doc := range docs {
		content := strings.TrimSpace(DocumentStore.GetFiledById(doc.DocumentId, "content"))
		if utils.GetTextFromHtml(content) == "" { //内容为空，渲染文档内容，并再重新获取文档内容
			utils.RenderDocumentById(doc.DocumentId)
			mus.Db.Where("document_id = ?", doc.DocumentId).Find(&doc)
		}

		//将图片链接更换成绝对链接
		toc := converter.Toc{
			Id:    doc.DocumentId,
			Pid:   doc.ParentId,
			Title: doc.DocumentName,
			Link:  fmt.Sprintf("%v.html", doc.DocumentId),
		}
		ExpCfg.Toc = append(ExpCfg.Toc, toc)
		//图片处理，如果图片路径不是http开头，则表示是相对路径的图片，加上BaseUrl.如果图片是以http开头的，下载下来
		if gq, err := goquery.NewDocumentFromReader(strings.NewReader(doc.Release)); err == nil {
			gq.Find("img").Each(func(i int, s *goquery.Selection) {
				pic := ""
				if src, ok := s.Attr("src"); ok {
					if srcLower := strings.ToLower(src); strings.HasPrefix(srcLower, "http://") || strings.HasPrefix(srcLower, "https://") {
						pic = src
					} else {
						if utils.StoreType == utils.StoreOss {
							pic = strings.TrimRight(beego.AppConfig.String("oss::Domain"), "/ ") + "/" + strings.TrimLeft(src, "./")
						} else {
							pic = baseUrl + src
						}
					}
					//下载图片，放到folder目录下
					ext := ""
					if picSlice := strings.Split(pic, "?"); len(picSlice) > 0 {
						ext = filepath.Ext(picSlice[0])
					}
					filename := cryptil.Md5Crypt(pic) + ext
					localPic := folder + filename
					req := httplib.Get(pic).SetTimeout(5*time.Second, 5*time.Second)
					if strings.HasPrefix(pic, "https") {
						req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
					}
					req.Header("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3298.4 Safari/537.36")
					if err := req.ToFile(localPic); err == nil { //成功下载图片
						s.SetAttr("src", filename)
					} else {
						beego.Error("错误:", err, filename, pic)
						s.SetAttr("src", pic)
					}

				}
			})
			gq.Find(".markdown-toc").Remove()
			doc.Release, _ = gq.Find("body").Html()
		}

		//生成html
		if htmlStr, err := utils.ExecuteViewPathTemplate("document/tpl_export.html", map[string]interface{}{"Model": book, "Doc": doc, "BaseUrl": baseUrl, "Nickname": Nickname, "Date": ExpCfg.Timestamp}); err == nil {
			htmlName := folder + toc.Link
			ioutil.WriteFile(htmlName, []byte(htmlStr), os.ModePerm)
		} else {
			mus.Logger.Error(err.Error())
		}

	}

	//复制css文件到目录
	if b, err := ioutil.ReadFile("static/editor.md/css/export-editormd.css"); err == nil {
		ioutil.WriteFile(folder+"editormd.css", b, os.ModePerm)
	} else {
		mus.Logger.Error(err.Error())
	}
	cfgFile := folder + "config.json"
	ioutil.WriteFile(cfgFile, []byte(util.InterfaceToJson(ExpCfg)), os.ModePerm)
	if Convert, err := converter.NewConverter(cfgFile, debug); err == nil {
		if err := Convert.Convert(); err != nil {
			mus.Logger.Error(err.Error())
		}
	} else {
		mus.Logger.Error(err.Error())
	}

	//将文档移动到oss
	//将PDF文档移动到oss
	oldBook := fmt.Sprintf("projects/%v/books/%v", book.Identify, book.GenerateTime.Unix()) //旧书籍的生成时间
	//最后再更新文档生成时间
	book.GenerateTime = time.Now()
	if _, err = orm.NewOrm().Update(book, "generate_time"); err != nil {
		mus.Logger.Error("generate book update error", zap.Error(err))
	}

	mus.Db.Where("book_id = ?", book.BookId).Find(&book)

	newBook := fmt.Sprintf("projects/%v/books/%v", book.Identify, book.GenerateTime.Unix())

	exts := []string{".pdf", ".epub", ".mobi"}
	for _, ext := range exts {
		switch utils.StoreType {
		case utils.StoreOss:
			//不要开启gzip压缩，否则会出现文件损坏的情况
			if err := store.ModelStoreOss.MoveToOss(folder+"output/book"+ext, newBook+ext, true, false); err != nil {
				mus.Logger.Error("generate book oss error", zap.Error(err))
			} else { //设置下载头
				store.ModelStoreOss.SetObjectMeta(newBook+ext, book.BookName+ext)
			}
		case utils.StoreLocal: //本地存储
			store.ModelStoreLocal.MoveToStore(folder+"output/book"+ext, "uploads/"+newBook+ext)
		}

	}

	//删除旧文件
	switch utils.StoreType {
	case utils.StoreOss:
		if err := store.ModelStoreOss.DelFromOss(oldBook+".pdf", oldBook+".epub", oldBook+".mobi"); err != nil { //删除旧版
			mus.Logger.Error("DelFromOss book oss error", zap.Error(err))
		}
	case utils.StoreLocal: //本地存储
		if err := store.ModelStoreLocal.DelFiles(oldBook+".pdf", oldBook+".epub", oldBook+".mobi"); err != nil { //删除旧版
			mus.Logger.Error("DelFiles book oss error", zap.Error(err))
		}
	}
}

//根据项目ID查询文档列表.
func (m *document) FindListByBookId(bookId int) (docs []*mysql.Document, err error) {
	err = mus.Db.Where("book_id = ?", bookId).Order("order_sort asc").Find(&docs).Error
	return
}

//根据项目ID查询文档一级目录.
func (m *document) GetMenuTop(bookId int) (docs []*mysql.Document, err error) {
	var docsAll []*mysql.Document
	mus.Db.Select("document_id,document_name,member_id,parent_id,book_id,identify").Where("book_id = ? and parent_id = ?", bookId, 0).Order("order_sort asc,document_id asc").Limit(5000).Find(&docsAll)
	//以"."开头的文档标识，不在阅读目录显示
	for _, doc := range docsAll {
		if !strings.HasPrefix(doc.Identify, ".") {
			docs = append(docs, doc)
		}
	}
	return
}

func (m *document) GetParentTitle(pid int) (title string) {
	var d mysql.Document
	mus.Db.Select("document_id,parent_id, document_name").Where("document_id = ?", pid).Find(&d)
	return d.DocumentName
}

//自动生成下一级的内容
func (m *document) BookStackAuto(bookId, docId int) (md, cont string) {
	//自动生成文档内容
	var docs []mysql.Document

	mus.Db.Select("document_id, document_name, identify").Where("book_id = ? and parent_id = ?", bookId, docId).Order("order_sort asc").Find(&docs)
	var newCont []string //新HTML内容
	var newMd []string   //新markdown内容
	for _, doc := range docs {
		newMd = append(newMd, fmt.Sprintf(`- [%v]($%v)`, doc.DocumentName, doc.Identify))
		newCont = append(newCont, fmt.Sprintf(`<li><a href="$%v">%v</a></li>`, doc.Identify, doc.DocumentName))
	}
	md = strings.Join(newMd, "\n")
	cont = "<ul>" + strings.Join(newCont, "") + "</ul>"
	return
}

//爬虫批量采集
//@param		html				html
//@param		md					markdown内容
//@return		content,markdown	把链接替换为标识后的内容
func (m *document) BookStackCrawl(html, md string, bookId, uid int) (content, markdown string, err error) {
	var gq *goquery.Document
	content = html
	markdown = md
	project := ""
	if book, err := Book.Find(bookId); err == nil {
		project = book.Identify
	}
	//执行采集
	if gq, err = goquery.NewDocumentFromReader(strings.NewReader(content)); err == nil {
		//采集模式mode
		CrawlByChrome := false
		if strings.ToLower(gq.Find("mode").Text()) == "chrome" {
			CrawlByChrome = true
		}
		//内容选择器selector
		selector := ""
		if selector = strings.TrimSpace(gq.Find("selector").Text()); selector == "" {
			err = errors.New("内容选择器不能为空")
			return
		}

		// 截屏选择器
		if screenshot := strings.TrimSpace(gq.Find("screenshot").Text()); screenshot != "" {
			utils.ScreenShotProjects.Store(project, screenshot)
			defer utils.DeleteScreenShot(project)
		}

		//排除的选择器
		var exclude []string
		if excludeStr := strings.TrimSpace(gq.Find("exclude").Text()); excludeStr != "" {
			slice := strings.Split(excludeStr, ",")
			for _, item := range slice {
				exclude = append(exclude, strings.TrimSpace(item))
			}
		}

		var links = make(map[string]string) //map[url]identify

		gq.Find("a").Each(func(i int, selection *goquery.Selection) {
			if href, ok := selection.Attr("href"); ok {
				if !strings.HasPrefix(href, "$") {
					identify := utils.MD5Sub16(href) + ".md"
					links[href] = identify
				}
			}
		})

		gq.Find("a").Each(func(i int, selection *goquery.Selection) {
			if href, ok := selection.Attr("href"); ok {
				hrefLower := strings.ToLower(href)
				//以http或者https开头
				if strings.HasPrefix(hrefLower, "http://") || strings.HasPrefix(hrefLower, "https://") {
					//采集文章内容成功，创建文档，填充内容，替换链接为标识
					if retMD, err := utils.CrawlHtml2Markdown(href, 0, CrawlByChrome, 2, selector, exclude, links, map[string]string{"project": project}); err == nil {
						var doc mysql.Document
						identify := utils.MD5Sub16(href) + ".md"
						doc.Identify = identify
						doc.BookId = bookId
						doc.Version = time.Now().Unix()
						doc.ModifyAt = int(time.Now().Unix())
						doc.DocumentName = selection.Text()
						doc.MemberId = uid

						if docId, err := m.InsertOrUpdate(mus.Db, &doc); err != nil {
							mus.Logger.Error("document err", zap.String("err", err.Error()))
						} else {
							var ds mysql.DocumentStore
							ds.DocumentId = int(docId)
							ds.Markdown = "[TOC]\n\r\n\r" + retMD
							if err := DocumentStore.InsertOrUpdate(mus.Db, &ds); err != nil {
								mus.Logger.Error("document err", zap.String("err", err.Error()))
							}
						}
						selection = selection.SetAttr("href", "$"+identify)
						if _, ok := links[href]; ok {
							markdown = strings.Replace(markdown, "("+href+")", "($"+identify+")", -1)
						}
					} else {
						mus.Logger.Error("document err", zap.String("err", err.Error()))
					}
				}
			}
		})
		content, _ = gq.Find("body").Html()
	}
	return
}

// markdown 文档拆分
func (m *document) SplitMarkdownAndStore(seg string, markdown string, docId int) (err error) {
	var mapReplace = map[string]string{
		"${7}$": "#######",
		"${6}$": "######",
		"${5}$": "#####",
		"${4}$": "####",
		"${3}$": "###",
		"${2}$": "##",
		"${1}$": "#",
	}
	var oneDoc *mysql.Document
	oneDoc, err = m.Find(docId)
	if err != nil {
		return
	}

	newIdentifyFmt := "spilt.%v." + oneDoc.Identify

	seg = fmt.Sprintf("${%v}$", strings.Count(seg, "#"))
	for i := 7; i > 0; i-- {
		slice := make([]string, i+1)
		k := "\n" + strings.Join(slice, "#")
		markdown = strings.Replace(markdown, k, fmt.Sprintf("\n${%v}$", i), -1)
	}
	contSlice := strings.Split(markdown, seg)

	for idx, val := range contSlice {
		doc := mysql.Document{}

		if idx != 0 {
			val = seg + val
		}
		for k, v := range mapReplace {
			val = strings.Replace(val, k, v, -1)
		}

		doc.Identify = fmt.Sprintf(newIdentifyFmt, idx)
		if idx == 0 { //不需要使用newIdentify
			doc = *oneDoc
		} else {
			doc.OrderSort = idx
			doc.ParentId = oneDoc.DocumentId
		}
		doc.Release = ""
		doc.BookId = oneDoc.BookId
		doc.Markdown = val
		doc.DocumentName = utils.ParseTitleFromMdHtml(md.MarkdownToHTML(val))
		doc.Version = time.Now().Unix()
		doc.MemberId = oneDoc.MemberId

		if !strings.Contains(doc.Markdown, "[TOC]") {
			doc.Markdown = "[TOC]\r\n" + doc.Markdown
		}

		if docId, err := m.InsertOrUpdate(mus.Db, &doc); err != nil {
			mus.Logger.Error(err.Error())
		} else {
			var ds = mysql.DocumentStore{
				DocumentId: int(docId),
				Markdown:   doc.Markdown,
			}
			//todo 这里有个bug
			if err := DocumentStore.InsertOrUpdate(mus.Db, &ds); err != nil {
				mus.Logger.Error(err.Error())
			}
		}

	}
	return
}
