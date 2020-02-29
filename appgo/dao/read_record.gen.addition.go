package dao

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
)

var (
	signDayLayout = "20060102"
	_readingRule  = &ReadingRule{}
)

// 阅读计时规则
type ReadingRule struct {
	Min       int
	Max       int
	MaxReward int
	Invalid   int
}

//添加阅读记录
func (this *readRecord) Add(docId, uid int) (err error) {
	// 1、根据文档id查询书籍id
	// 2、写入或者更新阅读记录
	// 3、更新书籍被阅读的文档统计
	// 4、更新用户阅读时长
	var (
		doc    mysql.Document
		r      mysql.ReadRecord
		member = mysql.Member{}
		now    = time.Now()
	)

	err = mus.Db.Where("document_id = ?", docId).Find(&doc).Error
	if err != nil {
		return
	}

	if doc.BookId <= 0 {
		return
	}

	record := mysql.ReadRecord{
		BookId:   doc.BookId,
		DocId:    docId,
		Uid:      uid,
		CreateAt: int(now.Unix()),
	}

	// 更新书架中的书籍最后的阅读时间
	Star.SetLastReadTime(uid, doc.BookId)

	// 计算奖励的阅读时长
	readingTime := this.calcReadingTime(uid, docId, now)

	db := mus.Db.Begin()
	defer func() {
		if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	mus.Db.Where("uid = ? and doc_id = ?", uid, docId).Find(&r)
	readCnt := 1
	if r.Id > 0 { // 先删再增，以便根据主键id索引的倒序查询列表
		mus.Db.Where("id = ?", r.Id).Delete(&r)
		readCnt = 0 // 如果是更新，则阅读次数
	}

	// 更新阅读记录
	err = db.Create(&record).Error
	if err != nil {
		return
	}

	if readCnt == 1 {
		rc := &mysql.ReadCount{}
		db.Where("uid = ? and book_id = ?", uid, doc.BookId).Find(rc)
		if rc.Id > 0 { // 更新已存在的阅读进度统计记录
			rc.Cnt += 1
			err = db.Where("id = ?", rc.Id).UpdateColumns(rc).Error
		} else { // 增加阅读进度统计记录
			rc = &mysql.ReadCount{BookId: doc.BookId, Uid: uid, Cnt: 1}
			err = db.Create(rc).Error
		}
	}
	if err != nil {
		return
	}
	if readingTime <= 0 {
		return
	}

	db.Where("member_id = ?", uid).Find(&member)
	if member.MemberId > 0 {
		err = db.Where("member_id = ?", uid).Updates(mysql.Ups{"total_reading_time": member.TotalReadingTime + readingTime}).Error
		if err != nil {
			return
		}
		var rt mysql.ReadingTime
		db.Model(mysql.ReadingTime{}).Where("day = ?", now.Format(signDayLayout)).Find(&rt)
		if rt.Id > 0 {
			rt.Duration += readingTime
			db.Model(mysql.ReadingTime{}).Where("day = ?", now.Format(signDayLayout)).Updates(mysql.Ups{
				"duration": rt.Duration,
			})
		} else {
			rt.Day, _ = strconv.Atoi(now.Format(signDayLayout))
			rt.Uid = uid
			rt.Duration = readingTime
			err = db.Create(&rt).Error
		}
	}
	return
}

// 在 5 - 600 秒之间的阅读计时，正常计时
// 在 600 - 1800 秒(半个小时)之间的计时，按最大计时来计算时长
// 超过半个小时之后才有阅读记录，则在此期间的阅读时长为0
func (r *readRecord) calcReadingTime(uid, docId int, t time.Time) int {
	rr := r.LastReading(uid, "uid", "doc_id", "created_at")
	if rr.DocId == docId {
		return 0
	}

	rule := r.GetReadingRule()
	diff := int(t.Unix()) - rr.CreateAt
	if diff <= 0 || diff < rule.Min || diff >= rule.Invalid {
		return 0
	}

	if diff > rule.MaxReward {
		return rule.Max
	}
	return diff
}

// 查询用户最后的一条阅读记录
func (this *readRecord) LastReading(uid int, cols ...string) (r mysql.ReadRecord) {
	mus.Db.Where("uid = ?", uid).Order("id desc").Find(&r)
	return
}

// 获取阅读计时规则
func (*readRecord) GetReadingRule() (r *ReadingRule) {
	return _readingRule
}

// 更新签到奖励规则
func (*readRecord) UpdateReadingRule() {
	ops := []string{"READING_MIN_INTERVAL", "READING_MAX_INTERVAL", "READING_INTERVAL_MAX_REWARD", "READING_INVALID_INTERVAL"}
	for _, op := range ops {
		num, _ := strconv.Atoi(Global.GetOptionValue(op, ""))
		switch op {
		case "READING_MIN_INTERVAL":
			_readingRule.Min = num
		case "READING_MAX_INTERVAL":
			_readingRule.Max = num
		case "READING_INTERVAL_MAX_REWARD":
			_readingRule.MaxReward = num
		case "READING_INVALID_INTERVAL":
			_readingRule.Invalid = num
		}
	}
}

//查询阅读记录
func (*readRecord) ListXX(uid, bookId int) (lists []mysql.RecordList, cnt int, err error) {
	if uid*bookId == 0 {
		err = errors.New("用户id和项目id不能为空")
		return
	}
	fields := "r.doc_id,r.create_at,d.document_name title,d.identify"
	sql := "select %v from %v r left join " + mysql.Document{}.TableName() + " d on r.doc_id=d.document_id where r.book_id=? and r.uid=? order by r.id desc limit 5000"
	sql = fmt.Sprintf(sql, fields, mysql.ReadRecord{}.TableName())
	err = mus.Db.Raw(sql, bookId, uid).Scan(&lists).Error
	cnt = len(lists)
	return
}

//查询阅读进度
func (this *readRecord) Progress(uid, bookId int) (rp mysql.ReadProgress, err error) {
	if uid*bookId == 0 {
		err = errors.New("用户id和书籍id均不能为空")
		return
	}
	var (
		rc   mysql.ReadCount
		book = new(mysql.Book)
	)

	err = mus.Db.Where("uid = ? and book_id = ?", uid, bookId).Find(&rc).Error
	if err == nil {
		err = mus.Db.Where("book_id=?", bookId).Find(book).Error
		if err == nil {
			rp.Total = book.DocCount
		}
	}

	rp.Cnt = rc.Cnt
	rp.BookIdentify = book.Identify
	if rp.Total == 0 {
		rp.Percent = "0.00%"
	} else {
		if rp.Cnt > rp.Total {
			rp.Cnt = rp.Total
		}
		f := float32(rp.Cnt) / float32(rp.Total)
		rp.Percent = fmt.Sprintf("%.2f", f*100) + "%"
	}
	return
}

//清空阅读记录
//当删除文档项目时，直接删除该文档项目的所有记录
func (this *readRecord) Clear(uid, bookId int) (err error) {
	if bookId > 0 && uid > 0 {
		err = mus.Db.Where("uid = ? and book_id = ?").Delete(&mysql.ReadCount{}).Error
		if err != nil {
			return
		}
		err = mus.Db.Where("uid = ? and book_id = ?").Delete(&mysql.ReadRecord{}).Error
		if err != nil {
			return
		}
	} else if uid == 0 && bookId > 0 {
		err = mus.Db.Where("book_id = ?").Delete(&mysql.ReadCount{}).Error
		if err != nil {
			return
		}
		err = mus.Db.Where("book_id = ?").Delete(&mysql.ReadRecord{}).Error
		if err != nil {
			return
		}
	}
	return
}

//删除单条阅读记录
func (this *readRecord) DeleteXX(uid, docId int) (err error) {
	if uid*docId == 0 {
		err = errors.New("用户id和文档id不能为空")
		return
	}

	var record mysql.ReadRecord

	mus.Db.Where("uid = ? and doc_id = ?", uid, docId).Find(&record)
	if record.BookId > 0 { //存在，则删除该阅读记录
		err = mus.Db.Where("id = ?", record.Id).Delete(&record).Error
		if err != nil {
			return
		}
		err = SetIncreAndDecre(mysql.ReadCount{}.TableName(), "cnt", "book_id="+strconv.Itoa(record.BookId)+" and uid="+strconv.Itoa(uid), false, 1)
	}
	return
}
