package mysql

//增加一个重新阅读的功能，即重置阅读，清空所有阅读记录

//阅读记录.用于记录阅读的文档，以及阅读进度统计
type ReadRecord struct {
	Id       int `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"`
	BookId   int `gorm:"not null;index"json:"bookId"` //书籍id
	DocId    int `gorm:"not null;"json:"docId"`       //文档id
	Uid      int `gorm:"not null;index"json:"uid"`    //用户id
	CreateAt int `gorm:"not null;"json:"createAt"`    //记录创建时间，也就是内容阅读时间
}

// 阅读统计
// 用来记录一本书（假设有100个章节），用户已经阅读了多少章节，以标识用户书籍的阅读进度
// 从而不用每次从阅读记录的表 read_record 表里面进行mysql 的 count 统计
type ReadCount struct {
	Id     int `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"` // 自增主键
	BookId int `gorm:"not null;"json:"bookId"`                       // 书籍
	Uid    int `gorm:"not null;"json:"uid"`                          // 用户id
	Cnt    int `gorm:"not null;"json:"cnt"`                          // 阅读的文档数
}

func (ReadRecord) TableName() string {
	return "read_record"
}

func (ReadCount) TableName() string {
	return "read_count"
}

//阅读记录列表（非表）
type RecordList struct {
	DocId    int
	Title    string
	Identify string
	CreateAt int
}

//阅读进度(非表)
type ReadProgress struct {
	Cnt          int    `json:"cnt"`     //已阅读过的文档
	Total        int    `json:"total"`   //总文档
	Percent      string `json:"percent"` //占的百分比
	BookIdentify string `json:"book_identify"`
}

// 阅读计时规则
type ReadingRule struct {
	Min       int
	Max       int
	MaxReward int
	Invalid   int
}

func NewReadRecord() *ReadRecord {
	return &ReadRecord{}
}

//添加阅读记录
func (this *ReadRecord) Add(docId, uid int) (err error) {
	// 1、根据文档id查询书籍id
	// 2、写入或者更新阅读记录
	// 3、更新书籍被阅读的文档统计
	// 4、更新用户阅读时长
	//var (
	//	doc      Document
	//	r        ReadRecord
	//	o        = orm.NewOrm()
	//	tableDoc = NewDocument()
	//	member   = NewMember()
	//	now      = time.Now()
	//	rt       = NewReadingTime()
	//)
	//
	//err = o.QueryTable(tableDoc).Filter("document_id", docId).One(&doc, "book_id")
	//if err != nil {
	//	beego.Error(err)
	//	return
	//}
	//
	//if doc.BookId <= 0 {
	//	return
	//}
	//
	//record := ReadRecord{
	//	BookId:   doc.BookId,
	//	DocId:    docId,
	//	Uid:      uid,
	//	CreateAt: int(now.Unix()),
	//}
	//
	//// 更新书架中的书籍最后的阅读时间
	//go new(Star).SetLastReadTime(uid, doc.BookId)
	//
	//// 计算奖励的阅读时长
	//readingTime := this.calcReadingTime(uid, docId, now)
	//
	//o.Begin()
	//defer func() {
	//	if err != nil {
	//		o.Rollback()
	//		beego.Error(err)
	//	} else {
	//		o.Commit()
	//	}
	//}()
	//
	//o.QueryTable(tableReadRecord).Filter("doc_id", docId).Filter("uid", uid).One(&r, "id")
	//
	//readCnt := 1
	//if r.Id > 0 { // 先删再增，以便根据主键id索引的倒序查询列表
	//	o.QueryTable(tableReadRecord).Filter("id", r.Id).Delete()
	//	readCnt = 0 // 如果是更新，则阅读次数
	//}
	//
	//// 更新阅读记录
	//_, err = o.Insert(&record)
	//if err != nil {
	//	return
	//}
	//
	//if readCnt == 1 {
	//	rc := &ReadCount{}
	//	o.QueryTable(tableReadCount).Filter("uid", uid).Filter("book_id", doc.BookId).One(rc)
	//	if rc.Id > 0 { // 更新已存在的阅读进度统计记录
	//		rc.Cnt += 1
	//		_, err = o.Update(rc)
	//	} else { // 增加阅读进度统计记录
	//		rc = &ReadCount{BookId: doc.BookId, Uid: uid, Cnt: 1}
	//		_, err = o.Insert(rc)
	//	}
	//}
	//if err != nil {
	//	return
	//}
	//if readingTime <= 0 {
	//	return
	//}
	//
	//o.QueryTable(member).Filter("member_id", uid).One(member, "member_id", "total_reading_time")
	//if member.MemberId > 0 {
	//	_, err = o.QueryTable(member).Filter("member_id", uid).Update(orm.Params{"total_reading_time": member.TotalReadingTime + readingTime})
	//	if err != nil {
	//		return
	//	}
	//	o.QueryTable(rt).Filter("uid", uid).Filter("day", now.Format(signDayLayout)).One(rt)
	//	if rt.Id > 0 {
	//		rt.Duration += readingTime
	//		_, err = o.Update(rt)
	//	} else {
	//		rt.Day, _ = strconv.Atoi(now.Format(signDayLayout))
	//		rt.Uid = uid
	//		rt.Duration = readingTime
	//		_, err = o.Insert(rt)
	//	}
	//}
	return
}

//查询阅读记录
func (this *ReadRecord) List(uid, bookId int) (lists []RecordList, cnt int64, err error) {
	//if uid*bookId == 0 {
	//	err = errors.New("用户id和项目id不能为空")
	//	return
	//}
	//fields := "r.doc_id,r.create_at,d.document_name title,d.identify"
	//sql := "select %v from %v r left join md_documents d on r.doc_id=d.document_id where r.book_id=? and r.uid=? order by r.id desc limit 5000"
	//sql = fmt.Sprintf(sql, fields, tableReadRecord)
	//cnt, err = orm.NewOrm().Raw(sql, bookId, uid).QueryRows(&lists)
	return
}
