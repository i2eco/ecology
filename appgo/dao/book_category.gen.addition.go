package dao

import (
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"strconv"
)

//处理书籍分类
func (this *bookCategory) SetBookCates(bookId int, cids []string) {
	if len(cids) == 0 {
		return
	}

	var (
		cates []mysql.Category
	)
	mus.Db.Where("id in (?)", cids).Find(&cates)

	cidMap := make(map[string]bool)
	for _, cate := range cates {
		cidMap[strconv.Itoa(cate.Pid)] = true
		cidMap[strconv.Itoa(cate.Id)] = true
	}
	cids = []string{}
	for cid, _ := range cidMap {
		cids = append(cids, cid)
	}

	mus.Db.Where("book_id = ?", bookId).Delete(&mysql.BookCategory{})
	for _, cid := range cids {
		cidNum, _ := strconv.Atoi(cid)
		bookCate := mysql.BookCategory{
			CategoryId: cidNum,
			BookId:     bookId,
		}
		mus.Db.Create(&bookCate)

	}
	go mysql.CountCategory()
}

//根据书籍id查询分类id
func (this *bookCategory) GetByBookId(book_id int) (cates []mysql.Category, err error) {
	sql := "select c.* from category c left join book_category bc on c.id=bc.category_id where bc.book_id=?"
	err = mus.Db.Raw(sql, book_id).Scan(&cates).Error
	return
}
