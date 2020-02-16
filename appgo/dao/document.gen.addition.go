package dao

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/goecology/ecology/appgo/model/mysql"
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
func (m *document) InsertOrUpdate(db *gorm.DB, documents *mysql.Document) (id int, err error) {
	id = documents.DocumentId
	documents.ModifyTime = time.Now()
	if documents.DocumentId > 0 { //文档id存在，则更新
		err = db.Model(documents).Where("document_id = ?", documents.DocumentId).UpdateColumns(documents).Error
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
		return
	}
	//identify存在，则执行更新
	err = db.Model(mysql.Document{}).Where("document_id = ?", documents.DocumentId).UpdateColumns(documents).Error
	if err != nil {
		return
	}
	id = mm.DocumentId
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
