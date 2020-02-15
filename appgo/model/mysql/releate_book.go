package mysql

type RelateBook struct {
	Id      int
	BookId  int `orm:"unique"`
	BookIds string
	Expire  int
}

func NewRelateBook() *RelateBook {
	return &RelateBook{}
}

// Get the related books for a given book
func (r *RelateBook) Lists(bookId int, limit ...int) (books []Book) {
	//day, _ := strconv.Atoi(GetOptionValue("RELATE_BOOK", "0"))
	//if day <= 0 {
	//	return
	//}
	//
	//length := 12
	//if len(limit) > 0 && limit[0] > 0 {
	//	length = limit[0]
	//}
	//
	//var rb RelateBook
	//var ids []int
	//
	//now := int(time.Now().Unix())
	//
	//o := orm.NewOrm()
	//o.QueryTable(r).Filter("book_id", bookId).One(&rb)
	//bookModel := NewBook()
	//
	//fields := []string{"book_id", "book_name", "cover", "identify"}
	//
	//if rb.BookId > 0 && rb.Expire > now {
	//	bookIds := rb.BookIds
	//	if !strings.HasPrefix(bookIds, "[") {
	//		bookIds = "[" + bookIds + "]"
	//	}
	//
	//	err := json.Unmarshal([]byte(bookIds), &ids)
	//	if err == nil && len(ids) > 0 {
	//		books, _ = bookModel.GetBooksById(ids, fields...)
	//		return
	//	}
	//}
	//
	//book, err := bookModel.Find(bookId)
	//if err != nil {
	//	return
	//}
	//
	//if GetOptionValue("ELASTICSEARCH_ON", "false") == "true" {
	//	ids = listByES(book, length)
	//} else {
	//	ids = listByDBWithLabel(book, length)
	//}
	//
	//books, _ = bookModel.GetBooksById(ids, fields...)
	//rb.BookId = bookId
	//if ids == nil {
	//	ids = []int{}
	//}
	//relatedIdBytes, _ := json.Marshal(ids)
	//rb.BookIds = string(relatedIdBytes)
	//rb.Expire = now + day*24*3600
	//if rb.Id > 0 {
	//	o.Update(&rb)
	//} else {
	//	o.Insert(&rb)
	//}
	return
}
