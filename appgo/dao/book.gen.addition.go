package dao

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/model/mysql/store"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/core/errors"
	"go.uber.org/zap"
)

type TotalResult struct {
	Cnt int
}

//分页查询指定用户的项目
//按照最新的进行排序
func (m *book) FindToPager(pageIndex, pageSize, memberId int, PrivatelyOwned ...int) (books []*mysql.BookResult, totalCount int, err error) {
	sql1 := "SELECT COUNT(book.book_id) AS cnt FROM " + mysql.Book{}.TableName() + " AS book LEFT JOIN " +
		mysql.Relationship{}.TableName() + " AS rel ON book.book_id=rel.book_id AND rel.member_id = ? WHERE rel.relationship_id > 0 "
	if len(PrivatelyOwned) > 0 {
		sql1 = sql1 + " and book.privately_owned=" + strconv.Itoa(PrivatelyOwned[0])
	}
	var result TotalResult
	err = mus.Db.Raw(sql1, memberId).Scan(&result).Error
	if err != nil {
		return
	}
	totalCount = result.Cnt

	offset := (pageIndex - 1) * pageSize
	sql2 := "SELECT book.*,rel.member_id,rel.role_id,m.account as create_name FROM " + mysql.Book{}.TableName() + " AS book" +
		" LEFT JOIN " + mysql.Relationship{}.TableName() + " AS rel ON book.book_id=rel.book_id AND rel.member_id = ?" +
		" LEFT JOIN " + mysql.Relationship{}.TableName() + " AS rel1 ON book.book_id=rel1.book_id  AND rel1.role_id=0" +
		" LEFT JOIN " + mysql.Member{}.TableName() + " AS m ON rel1.member_id=m.member_id " +
		" WHERE rel.relationship_id > 0 "

	if len(PrivatelyOwned) > 0 {
		sql2 = sql2 + " and book.privately_owned=" + strconv.Itoa(PrivatelyOwned[0])
	}

	sql2 = sql2 + " ORDER BY book.book_id DESC LIMIT " + fmt.Sprintf("%d,%d", offset, pageSize)

	err = mus.Db.Raw(sql2, memberId).Scan(&books).Error
	if err != nil {
		return
	}

	if len(books) > 0 {
		sql := "SELECT m.account,doc.modify_time FROM " + mysql.Document{}.TableName() + " AS doc LEFT JOIN " + mysql.Member{}.TableName() + " AS m ON doc.modify_at=m.member_id WHERE book_id = ? ORDER BY doc.modify_time DESC LIMIT 1 "

		for index, book := range books {
			var text struct {
				Account    string
				ModifyTime time.Time
			}

			err = mus.Db.Raw(sql, book.BookId).Scan(&text).Error
			if err == nil {
				books[index].LastModifyText = text.Account + " 于 " + text.ModifyTime.Format("2006-01-02 15:04:05")
			}

			if book.RoleId == 0 {
				book.RoleName = "创始人"
			} else if book.RoleId == 1 {
				book.RoleName = "管理员"
			} else if book.RoleId == 2 {
				book.RoleName = "编辑者"
			} else if book.RoleId == 3 {
				book.RoleName = "观察者"
			}
		}
	}
	return
}

// 根据项目标识查询项目以及指定用户权限的信息.
func (m *book) ResultFindByIdentify(identify string, memberId int) (result *mysql.BookResult, err error) {
	if identify == "" || memberId <= 0 {
		return result, ErrInvalidParameter
	}

	book := mysql.NewBook()

	err = mus.Db.Where("identify=?", identify).Find(book).Error
	if err != nil {
		return
	}

	relationship := mysql.NewRelationship()
	err = mus.Db.Where("book_id = ? and member_id = ?", book.BookId, book.MemberId).Find(relationship).Error
	if err != nil {
		return
	}

	relationship2 := mysql.NewRelationship()
	err = mus.Db.Where("book_id = ? and role_id = ?", book.BookId, conf.Conf.Info.BookFounder).Find(relationship2).Error
	if err != nil {
		return
	}

	member := mysql.NewMember()
	err = mus.Db.Where("member_id = ?", relationship2.MemberId).Find(member).Error
	if err != nil {
		return result, err
	}

	result = book.ToBookResult()

	result.CreateName = member.Account
	result.MemberId = relationship.MemberId
	result.RoleId = relationship.RoleId
	result.RelationshipId = relationship.RelationshipId

	switch result.RoleId {
	case conf.BookFounder:
		result.RoleName = "创始人"
	case conf.BookAdmin:
		result.RoleName = "管理员"
	case conf.BookEditor:
		result.RoleName = "编辑者"
	case conf.BookObserver:
		result.RoleName = "观察者"
	}

	doc := mysql.NewDocument()
	err = mus.Db.Where("book_id =?", book.BookId).Order("modify_time desc").Find(doc).Error
	if err != nil {
		return
	}

	var member2 *mysql.Member
	member2, err = Member.Find(doc.ModifyAt)
	if err != nil {
		return
	}
	result.LastModifyText = member2.Account + " 于 " + doc.ModifyTime.Format("2006-01-02 15:04:05")
	return
}

// 内容替换
func (m *book) Replace(bookId int, src, dst string) {
	var docs []mysql.Document

	err := mus.Db.Where("book_id = ?", bookId).Limit(10000).Find(&docs).Error
	if err != nil {
		return
	}

	if len(docs) > 0 {
		for _, doc := range docs {
			var docStore mysql.DocumentStore
			err = mus.Db.Where("document_id = ?", doc.DocumentId).Find(&docStore).Error
			if err != nil {
				// todo log
				continue
			}
			if docStore.DocumentId > 0 {
				docStore.Markdown = strings.Replace(docStore.Markdown, src, dst, -1)
				docStore.Content = strings.Replace(docStore.Content, src, dst, -1)
				err = mus.Db.Table(docStore.TableName()).Updates(map[string]interface{}{
					"markdown": docStore.Markdown,
					"content":  docStore.Content,
				}).Error
				if err != nil {
					// todo log
					continue
				}
			}
		}
	}
}

// minRole 最小的角色权限
//conf.BookFounder
//conf.BookAdmin
//conf.BookEditor
//conf.BookObserver
func (m *book) HasProjectAccess(identify string, memberId int, minRole int) bool {
	var oneBook mysql.Book
	err := mus.Db.Where("identify = ?", identify).Find(&oneBook).Error
	if err != nil {
		return false
	}
	var rel mysql.Relationship

	err = mus.Db.Where("book_id = ?", oneBook.BookId).Find(&rel).Error
	if err != nil {
		return false
	}

	return rel.RoleId <= minRole
}

func (m *book) Sorted(limit int, orderField string) (books []mysql.Book) {
	fields := []string{"book_id", "book_name", "identify", "cover", "vcnt", "star", "cnt_comment"}
	mus.Db.Select(fields).Where("order_index > ? and privately_owned=? ", 0, 0).Order(orderField + " desc").Limit(limit).Find(&books)
	return
}

func (m *book) Insert(db *gorm.DB, create *mysql.Book) (err error) {
	err = db.Create(create).Error
	if err != nil {
		return
	}

	if create.Label != "" {
		Label.InsertOrUpdateMulti(create.Label)
	}

	relationship := mysql.NewRelationship()
	relationship.BookId = create.BookId
	relationship.RoleId = 0
	relationship.MemberId = create.MemberId

	err = Relationship.Insert(db, relationship)
	if err != nil {
		return
	}

	document := mysql.NewDocument()
	document.BookId = create.BookId
	document.DocumentName = "空白文档"
	document.Identify = "blank"
	document.MemberId = create.MemberId
	document.ModifyAt = create.MemberId

	var id int
	id, err = Document.InsertOrUpdate(db, document)
	if err != nil {
		return
	}
	var ds = &mysql.DocumentStore{
		DocumentId: int(id),
		Markdown:   "[TOC]\n\r\n\r", //默认内容
	}
	err = DocumentStore.InsertOrUpdate(db, ds)
	return
}

func (m *book) Find(id int, cols ...string) (book *mysql.Book, err error) {
	if id <= 0 {
		err = errors.New("id is error")
		return
	}
	book = mysql.NewBook()
	// todo cols
	err = mus.Db.Where("book_id = ?", id).Find(book).Error
	return
}

func (m *book) UpdateXX(oneBook *mysql.Book) (err error) {
	var tempBook mysql.Book
	err = mus.Db.Where("book_id = ?", oneBook.BookId).Find(&tempBook).Error
	if err != nil {
		return
	}

	if (oneBook.Label + tempBook.Label) != "" {
		go Label.InsertOrUpdateMulti(oneBook.Label + "," + tempBook.Label)
	}

	err = mus.Db.UpdateColumns(oneBook).Error
	return err
}

//根据指定字段查询结果集.
func (m *book) FindByField(field string, value interface{}) (books []*mysql.Book, err error) {
	err = mus.Db.Where(field+"=?", value).Find(&books).Error
	return
}

//根据指定字段查询一个结果.
func (m *book) FindByFieldFirst(field string, value interface{}) (book *mysql.Book, err error) {
	book = &mysql.Book{}
	err = mus.Db.Where(field+"=?", value).Find(book).Error
	return
}

func (m *book) FindByIdentify(identify string, cols ...string) (book *mysql.Book, err error) {
	err = mus.Db.Where("identify=?", identify).Find(book).Error
	return
}

// 彻底删除项目.
func (m *book) ThoroughDeleteBook(id int) (err error) {
	if id <= 0 {
		return ErrInvalidParameter
	}

	var one mysql.Book
	err = mus.Db.Where("book_id = ?", id).Find(&one).Error

	var (
		docs  []mysql.Document
		docId []string
	)

	mus.Db.Select("document_id").Where("book_id = ?", id).Limit(10000).Find(&docs)
	if len(docs) > 0 {
		for _, doc := range docs {
			docId = append(docId, strconv.Itoa(doc.DocumentId))
		}
	}

	db := mus.Db.Begin()

	//删除md_document_store中的文档
	if len(docId) > 0 {
		sql1 := fmt.Sprintf("delete from "+mysql.DocumentStore{}.TableName()+" where document_id in(%v)", strings.Join(docId, ","))
		if err1 := db.Raw(sql1).Error; err1 != nil {
			db.Rollback()
			return err1
		}
	}

	sql2 := "DELETE FROM " + mysql.Document{}.TableName() + " WHERE book_id = ?"
	err = db.Raw(sql2, one.BookId).Error
	if err != nil {
		db.Rollback()
		return err
	}
	sql3 := "DELETE FROM " + mysql.Document{}.TableName() + " WHERE book_id = ?"

	err = db.Raw(sql3, one.BookId).Error
	if err != nil {
		db.Rollback()
		return err
	}

	sql4 := "DELETE FROM " + mysql.Relationship{}.TableName() + " WHERE book_id = ?"
	err = db.Raw(sql4, one.BookId).Error

	if err != nil {
		db.Rollback()
		return err
	}

	if one.Label != "" {
		Label.InsertOrUpdateMulti(one.Label)
	}

	if err = db.Commit().Error; err != nil {
		return err
	}
	//删除oss中项目对应的文件夹
	switch utils.StoreType {
	case utils.StoreLocal: //删除本地存储，记得加上uploads
		os.Remove(strings.TrimLeft(one.Cover, "/ ")) //删除封面
		go store.ModelStoreLocal.DelFromFolder("uploads/projects/" + one.Identify)
	case utils.StoreOss:
		go store.ModelStoreOss.DelOssFolder("projects/" + one.Identify)
	}

	// 删除历史记录
	// todo doc history
	go func() {
		//for _, id := range docId {
		//	idInt, _ := strconv.Atoi(id)
		//	DocumentHistory.DeleteByDocumentId(idInt)
		//}
	}()

	return
}

//首页数据
//完善根据分类查询数据
//orderType:排序条件，可选值：recommend(推荐)、latest（）
func (m *book) HomeData(pageIndex, pageSize int, orderType mysql.BookOrder, lang string, cid int, fields ...string) (books []*mysql.Book, totalCount int, err error) {
	if cid > 0 { //针对cid>0
		return m.homeData(pageIndex, pageSize, orderType, lang, cid, fields...)
	}
	order := ""   //排序
	condStr := "" //查询条件
	cond := []string{"privately_owned=0"}
	if len(fields) == 0 {
		fields = append(fields, "book_id", "book_name", "identify", "cover", "order_index", "pin")
	} else {
		fields = append(fields, "pin")
	}
	switch orderType {
	case mysql.OrderRecommend: //推荐
		cond = append(cond, "order_index>0")
		order = "order_index desc"
	case mysql.OrderLatestRecommend: //最新推荐
		cond = append(cond, "order_index>0")
		order = "book_id desc"
	case mysql.OrderPopular: //受欢迎
		order = "star desc,vcnt desc"
	case mysql.OrderLatest: //最新发布
		order = "release_time desc"
	case mysql.OrderScore: //评分
		order = "score desc"
	case mysql.OrderComment: //评论
		order = "cnt_comment desc"
	case mysql.OrderStar: //收藏
		order = "star desc"
	case mysql.OrderView: //收藏
		order = "vcnt desc"
	}
	if len(cond) > 0 {
		condStr = " where " + strings.Join(cond, " and ")
	}

	lang = strings.ToLower(lang)
	switch lang {
	case "zh", "en", "other":
	default:
		lang = ""
	}
	if strings.TrimSpace(lang) != "" {
		condStr = condStr + " and `lang` = '" + lang + "'"
	}
	sqlFmt := "select %v from " + mysql.Book{}.TableName() + condStr
	fieldStr := strings.Join(fields, ",")
	sql := fmt.Sprintf(sqlFmt, fieldStr) + " order by " + order + fmt.Sprintf(" limit %v offset %v", pageSize, (pageIndex-1)*pageSize)
	sqlCount := fmt.Sprintf(sqlFmt, "count(*) cnt")
	var result TotalResult
	if err := mus.Db.Raw(sqlCount).Scan(&result).Error; err == nil {
		totalCount = result.Cnt
	}
	err = mus.Db.Raw(sql).Scan(&books).Error
	return
}

//针对cid大于0
func (m *book) homeData(pageIndex, pageSize int, orderType mysql.BookOrder, lang string, cid int, fields ...string) (books []*mysql.Book, totalCount int, err error) {
	order := ""   //排序
	condStr := "" //查询条件
	cond := []string{"b.privately_owned=0"}
	if len(fields) == 0 {
		fields = append(fields, "book_id", "book_name", "identify", "cover", "order_index")
	}
	switch orderType {
	case mysql.OrderRecommend: //推荐
		cond = append(cond, "b.order_index>0")
		order = "b.order_index desc"
	case mysql.OrderPopular: //受欢迎
		order = "b.star desc,b.vcnt desc"
	case mysql.OrderLatest: //最新发布
		order = "b.release_time desc"
	case mysql.OrderScore: //评分
		order = "b.score desc"
	case mysql.OrderComment: //评论
		order = "b.cnt_comment desc"
	case mysql.OrderStar: //收藏
		order = "b.star desc"
	case mysql.OrderView: //收藏
		order = "b.vcnt desc"
	}
	if cid > 0 {
		cond = append(cond, "c.category_id="+strconv.Itoa(cid))
	}
	if len(cond) > 0 {
		condStr = " where " + strings.Join(cond, " and ")
	}
	lang = strings.ToLower(lang)
	switch lang {
	case "zh", "en", "other":
	default:
		lang = ""
	}
	if strings.TrimSpace(lang) != "" {
		condStr = condStr + " and `lang` = '" + lang + "'"
	}
	sqlFmt := "select %v from " + mysql.Book{}.TableName() + " b left join book_category c on b.book_id=c.book_id" + condStr
	fieldStr := "b." + strings.Join(fields, ",b.")
	sql := fmt.Sprintf(sqlFmt, fieldStr) + " order by " + order + fmt.Sprintf(" limit %v offset %v", pageSize, (pageIndex-1)*pageSize)
	sqlCount := fmt.Sprintf(sqlFmt, "count(*) cnt")

	var result TotalResult
	if err := mus.Db.Raw(sqlCount).Scan(&result).Error; err == nil {
		totalCount = result.Cnt
	}
	err = mus.Db.Raw(sql).Scan(&books).Error
	return
}

//分页查找系统首页数据.
func (m *book) FindForHomeToPager(pageIndex, pageSize, member_id int, orderType string) (books []*mysql.BookResult, totalCount int, err error) {
	var count int
	offset := (pageIndex - 1) * pageSize
	//如果是登录用户
	if member_id > 0 {
		sql1 := "SELECT COUNT(*) FROM " + mysql.Book{}.TableName() + " AS book LEFT JOIN relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ? WHERE relationship_id > 0 OR book.privately_owned = 0"
		err = mus.Db.Raw(sql1, member_id).Scan(&totalCount).Error
		if err != nil {
			return
		}

		sql2 := `SELECT book.*,rel1.*,member.account AS create_name FROM book AS book
			LEFT JOIN relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ?
			LEFT JOIN relationship AS rel1 ON rel1.book_id = book.book_id AND rel1.role_id = 0
			LEFT JOIN member AS member ON rel1.member_id = member.member_id
			WHERE rel.relationship_id > 0 OR book.privately_owned = 0 ORDER BY order_index DESC ,book.book_id DESC LIMIT ?,?`

		err = mus.Db.Raw(sql2, member_id, offset, pageSize).Scan(&books).Error
		return
	}
	err = mus.Db.Where("privately_owned = ?", 0).Count(&count).Error
	if err != nil {
		return
	}
	totalCount = int(count)

	sql := `SELECT book.*,rel.*,member.account AS create_name FROM book
			LEFT JOIN relationship AS rel ON rel.book_id = book.book_id AND rel.role_id = 0
			LEFT JOIN member AS member ON rel.member_id = member.member_id
			WHERE book.privately_owned = 0 ORDER BY order_index DESC ,book.book_id DESC LIMIT ?,?`

	err = mus.Db.Raw(sql, offset, pageSize).Scan(&books).Error
	return
}

//分页全局搜索.
func (m *book) FindForLabelToPager(keyword string, pageIndex, pageSize, memberId int) (books []*mysql.BookResult, totalCount int, err error) {
	keyword = "%" + keyword + "%"
	offset := (pageIndex - 1) * pageSize
	//如果是登录用户
	if memberId > 0 {
		sql1 := "SELECT COUNT(*) FROM book AS book LEFT JOIN relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ? WHERE (relationship_id > 0 OR book.privately_owned = 0) AND (book.label LIKE ? or book.book_name like ?) limit 1"
		if err = mus.Db.Raw(sql1, memberId, keyword, keyword).Scan(&totalCount).Error; err != nil {
			return
		}

		sql2 := `SELECT book.*,rel1.*,member.account AS create_name FROM book AS book
			LEFT JOIN relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ?
			LEFT JOIN relationship AS rel1 ON rel1.book_id = book.book_id AND rel1.role_id = 0
			LEFT JOIN member AS member ON rel1.member_id = member.member_id
			WHERE (rel.relationship_id > 0 OR book.privately_owned = 0) AND  (book.label LIKE ? or book.book_name like ?) ORDER BY order_index DESC ,book.book_id DESC LIMIT ?,?`

		err = mus.Db.Raw(sql2, memberId, keyword, keyword, offset, pageSize).Scan(&books).Error
		return
	}

	sql1 := "select COUNT(*) from book where privately_owned=0 and (label LIKE ? or book_name like ?) limit 1"
	if err = mus.Db.Raw(sql1, keyword, keyword).Scan(&totalCount).Error; err != nil {
		return
	}

	sql := `SELECT book.*,rel.*,member.account AS create_name FROM book AS book
			LEFT JOIN relationship AS rel ON rel.book_id = book.book_id AND rel.role_id = 0
			LEFT JOIN member AS member ON rel.member_id = member.member_id
			WHERE book.privately_owned = 0 AND (book.label LIKE ? or book.book_name LIKE ?) ORDER BY order_index DESC ,book.book_id DESC LIMIT ?,?`

	err = mus.Db.Raw(sql, keyword, keyword, offset, pageSize).Scan(&books).Error
	return
}

//重置文档数量
func (m *book) ResetDocumentNumber(bookId int) {
	var cnt int
	err := mus.Db.Model(mysql.Document{}).Where("book_id=?", bookId).Count(&cnt).Error
	if err != nil {
		return
	}
	// todo
	mus.Db.Raw("UPDATE book SET doc_count = ? WHERE book_id = ?", int(cnt), bookId).Row()
}

// 根据书籍id获取(公开的)书籍
func (m *book) GetBooksById(id []int, fields ...string) (books []mysql.Book, err error) {

	var bs []mysql.Book
	var idArr []interface{}

	if len(id) == 0 {
		return
	}

	for _, i := range id {
		idArr = append(idArr, i)
	}

	err = mus.Db.Where("book_id in (?) and privately_owned = ?", idArr, 0).Find(&bs).Error
	if err != nil {
		return
	}

	if len(bs) > 0 {
		bookMap := make(map[interface{}]mysql.Book)
		for _, book := range bs {
			bookMap[book.BookId] = book
		}
		for _, i := range id {
			if book, ok := bookMap[i]; ok {
				books = append(books, book)
			}
		}
	}

	return
}

// 搜索书籍，这里只返回book_id
func (n *book) SearchBook(wd string, page, size int) (books []mysql.Book, cnt int, err error) {
	sqlFmt := "select %v from book where privately_owned=0 and (book_name like ? or label like ? or description like ?) order by star desc"
	sqlCount := fmt.Sprintf(sqlFmt, "count(book_id) cnt")
	sql := fmt.Sprintf(sqlFmt, "book_id")

	var count struct{ Cnt int }
	wd = "%" + wd + "%"

	err = mus.Db.Raw(sqlCount, wd, wd, wd).Scan(&count).Error
	if err != nil {
		return
	}
	if count.Cnt <= 0 {
		return
	}

	cnt = count.Cnt
	err = mus.Db.Raw(sql+" limit ? offset ?", wd, wd, wd, size, (page-1)*size).Scan(&books).Error
	return
}

// search books with labels
func (b *book) SearchBookByLabel(labels []string, limit int, excludeIds []int) (bookIds []int, err error) {
	bookIds = []int{}
	if len(labels) == 0 {
		return
	}

	rawRegex := strings.Join(labels, "|")

	excludeClause := ""
	if len(excludeIds) == 1 {
		excludeClause = fmt.Sprintf("book_id != %d AND", excludeIds[0])
	} else if len(excludeIds) > 1 {
		excludeVal := strings.Replace(strings.Trim(fmt.Sprint(excludeIds), "[]"), " ", ",", -1)
		excludeClause = fmt.Sprintf("book_id NOT IN (%s) AND", excludeVal)
	}

	sql := fmt.Sprintf("SELECT book_id FROM book WHERE %v label REGEXP ? ORDER BY star DESC LIMIT ?", excludeClause)
	err = mus.Db.Raw(sql, rawRegex, limit).Scan(&bookIds).Error
	if err != nil {
		mus.Logger.Error("failed to execute sql", zap.String("sql", sql), zap.Error(err))
	}
	return
}

//分页查询指定用户的项目
//按照最新的进行排序
func (m *book) ResultFindToPager(pageIndex, pageSize, memberId int, PrivatelyOwned ...int) (books []*mysql.BookResult, totalCount int, err error) {
	sql1 := "SELECT COUNT(book.book_id) AS cnt FROM " + mysql.Book{}.TableName() + " AS book LEFT JOIN " +
		mysql.Relationship{}.TableName() + " AS rel ON book.book_id=rel.book_id AND rel.member_id = ? WHERE rel.relationship_id > 0 "
	if len(PrivatelyOwned) > 0 {
		sql1 = sql1 + " and book.privately_owned=" + strconv.Itoa(PrivatelyOwned[0])
	}
	var result TotalResult
	err = mus.Db.Raw(sql1, memberId).Scan(&result).Error
	if err != nil {
		return
	}
	totalCount = result.Cnt

	offset := (pageIndex - 1) * pageSize
	sql2 := "SELECT book.*,rel.member_id,rel.role_id,m.account as create_name FROM " + mysql.Book{}.TableName() + " AS book" +
		" LEFT JOIN " + mysql.Relationship{}.TableName() + " AS rel ON book.book_id=rel.book_id AND rel.member_id = ?" +
		" LEFT JOIN " + mysql.Relationship{}.TableName() + " AS rel1 ON book.book_id=rel1.book_id  AND rel1.role_id=0" +
		" LEFT JOIN " + mysql.Member{}.TableName() + " AS m ON rel1.member_id=m.member_id " +
		" WHERE rel.relationship_id > 0 "

	if len(PrivatelyOwned) > 0 {
		sql2 = sql2 + " and book.privately_owned=" + strconv.Itoa(PrivatelyOwned[0])
	}

	sql2 = sql2 + " ORDER BY book.book_id DESC LIMIT " + fmt.Sprintf("%d,%d", offset, pageSize)

	err = mus.Db.Raw(sql2, memberId).Scan(&books).Error
	if err != nil {
		return
	}

	if len(books) > 0 {
		sql := "SELECT m.account,doc.modify_time FROM document AS doc LEFT JOIN member AS m ON doc.modify_at=m.member_id WHERE book_id = ? ORDER BY doc.modify_time DESC LIMIT 1 "

		for index, book := range books {
			var text struct {
				Account    string
				ModifyTime time.Time
			}

			err = mus.Db.Raw(sql, book.BookId).Scan(&text).Error
			if err == nil {
				books[index].LastModifyText = text.Account + " 于 " + text.ModifyTime.Format("2006-01-02 15:04:05")
			}

			if book.RoleId == 0 {
				book.RoleName = "创始人"
			} else if book.RoleId == 1 {
				book.RoleName = "管理员"
			} else if book.RoleId == 2 {
				book.RoleName = "编辑者"
			} else if book.RoleId == 3 {
				book.RoleName = "观察者"
			}
		}
	}
	return
}
