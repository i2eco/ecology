package dao

import (
	"fmt"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"strconv"
	"strings"
)

type documentSearchResult struct {
}

func NewDocumentSearchResult() *documentSearchResult {
	return &documentSearchResult{}
}

//分页全局搜索.
func (m *documentSearchResult) FindToPager(keyword string, pageIndex, pageSize, memberId int) (searchResult []*mysql.DocumentSearchResult, totalCount int, err error) {
	offset := (pageIndex - 1) * pageSize
	keyword = "%" + keyword + "%"

	if memberId <= 0 {
		sql1 := `SELECT count(doc.document_id) as total_count FROM ` + mysql.Document{}.TableName() + ` AS doc
  LEFT JOIN ` + mysql.Book{}.TableName() + ` as book ON doc.book_id = book.book_id
WHERE book.privately_owned = 0 AND (doc.document_name LIKE ? OR doc.release LIKE ?) `

		sql2 := `SELECT doc.document_id,doc.modify_time,doc.create_time,doc.document_name,doc.identify,doc.release as description,doc.modify_time,book.identify as book_identify,book.book_name,rel.member_id,member.account AS author FROM md_documents AS doc
  LEFT JOIN ` + mysql.Book{}.TableName() + ` as book ON doc.book_id = book.book_id
  LEFT JOIN ` + mysql.Relationship{}.TableName() + ` AS rel ON book.book_id = rel.book_id AND rel.role_id = 0
  LEFT JOIN ` + mysql.Member{}.TableName() + ` as member ON rel.member_id = member.member_id
WHERE book.privately_owned = 0 AND (doc.document_name LIKE ? OR doc.release LIKE ?)
 ORDER BY doc.document_id DESC LIMIT ?,? `

		err = mus.Db.Raw(sql1, keyword, keyword).Scan(&totalCount).Error
		if err != nil {
			return
		}
		err = mus.Db.Raw(sql2, keyword, keyword, offset, pageSize).Scan(&searchResult).Error
		if err != nil {
			return
		}
	} else {
		sql1 := `SELECT count(doc.document_id) as total_count FROM ` + mysql.Document{}.TableName() + ` AS doc
  LEFT JOIN ` + mysql.Book{}.TableName() + ` as book ON doc.book_id = book.book_id
  LEFT JOIN ` + mysql.Relationship{}.TableName() + ` AS rel ON doc.book_id = rel.book_id AND rel.role_id = 0
  LEFT JOIN ` + mysql.Relationship{}.TableName() + ` AS rel1 ON doc.book_id = rel1.book_id AND rel1.member_id = ?
WHERE (book.privately_owned = 0 OR rel1.relationship_id > 0)  AND (doc.document_name LIKE ? OR doc.release LIKE ?) `

		sql2 := `SELECT doc.document_id,doc.modify_time,doc.create_time,doc.document_name,doc.identify,doc.release as description,doc.modify_time,book.identify as book_identify,book.book_name,rel.member_id,member.account AS author FROM md_documents AS doc
  LEFT JOIN ` + mysql.Book{}.TableName() + ` as book ON doc.book_id = book.book_id
  LEFT JOIN ` + mysql.Relationship{}.TableName() + ` AS rel ON book.book_id = rel.book_id AND rel.role_id = 0
  LEFT JOIN ` + mysql.Member{}.TableName() + ` as member ON rel.member_id = member.member_id
  LEFT JOIN ` + mysql.Relationship{}.TableName() + ` AS rel1 ON doc.book_id = rel1.book_id AND rel1.member_id = ?
WHERE (book.privately_owned = 0 OR rel1.relationship_id > 0)  AND (doc.document_name LIKE ? OR doc.release LIKE ?)
 ORDER BY doc.document_id DESC LIMIT ?,? `

		err = mus.Db.Raw(sql1, memberId, keyword, keyword).Scan(&totalCount).Error
		if err != nil {
			return
		}
		err = mus.Db.Raw(sql2, memberId, keyword, keyword, offset, pageSize).Scan(&searchResult).Error
		if err != nil {
			return
		}
	}
	return
}

//项目内搜索.
func (m *documentSearchResult) SearchDocument(keyword string, bookId int, page, size int) (docs []*mysql.DocumentSearchResult, cnt int, err error) {
	fields := []string{"document_id", "document_name", "identify", "book_id"}
	sql := "SELECT %v FROM " + mysql.Document{}.TableName() + " WHERE book_id = " + strconv.Itoa(bookId) + " AND (document_name LIKE ? OR `release` LIKE ?) "
	sqlCount := fmt.Sprintf(sql, "count(document_id) cnt")
	sql = fmt.Sprintf(sql, strings.Join(fields, ",")) + " order by vcnt desc"
	if bookId == 0 {
		// bookId 为 0 的时候，只搜索公开的书籍的文档
		sql = "SELECT %v FROM " + mysql.Document{}.TableName() + " d left join md_books b on d.book_id=b.book_id WHERE b.privately_owned=0 and (d.document_name LIKE ? OR d.`release` LIKE ? )"
		sqlCount = fmt.Sprintf(sql, "count(d.document_id) cnt")
		sql = fmt.Sprintf(sql, "d."+strings.Join(fields, ",d.")) + " order by d.vcnt desc"
	}

	keyword = "%" + keyword + "%"

	var count struct {
		Cnt int
	}

	err = mus.Db.Raw(sqlCount, keyword, keyword).Scan(&count).Error
	if err != nil {
		return
	}
	cnt = count.Cnt

	limit := fmt.Sprintf(" limit %v offset %v", size, (page-1)*size)
	if cnt > 0 {
		err = mus.Db.Raw(sql+limit, keyword, keyword).Scan(&docs).Error
	}
	return
}

// 根据id查询搜索结果
func (m *documentSearchResult) GetDocsById(id []int, withoutCont ...bool) (docs []mysql.DocResult, err error) {
	if len(id) == 0 {
		return
	}

	var idArr []string
	for _, i := range id {
		idArr = append(idArr, fmt.Sprint(i))
	}

	fields := []string{
		"d.document_id", "d.document_name", "d.identify", "d.vcnt", "d.create_time", "b.book_id",
	}

	// 不返回内容
	if len(withoutCont) == 0 || !withoutCont[0] {
		fields = append(fields, "b.identify book_identify", "d.release", "b.book_name")
	}

	sqlFmt := "select " + strings.Join(fields, ",") + " from " + mysql.Document{}.TableName() + " d left join " + mysql.Document{}.TableName() + " b on d.book_id=b.book_id where d.document_id in(%v)"
	sql := fmt.Sprintf(sqlFmt, strings.Join(idArr, ","))

	var rows []mysql.DocResult
	var cnt int64

	err = mus.Db.Raw(sql).Scan(&rows).Error
	if cnt > 0 {
		docMap := make(map[int]mysql.DocResult)
		for _, row := range rows {
			docMap[row.DocumentId] = row
		}
		client := NewElasticSearchClient()
		for _, i := range id {
			if doc, ok := docMap[i]; ok {
				doc.Release = client.html2Text(doc.Release)
				docs = append(docs, doc)
			}
		}
	}

	return
}
