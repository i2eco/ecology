package mysql

import (
	"time"

	"github.com/astaxie/beego"
	"github.com/i2eco/ecology/appgo/pkg/mus"
)

type period string

const (
	PeriodDay      period = "day"
	PeriodWeek     period = "week"
	PeriodLastWeek period = "last-week"
	PeriodMonth    period = "month"
	PeriodLastMoth period = "last-month"
	PeriodAll      period = "all"
	PeriodYear     period = "year"
)

const dateFormat = "20060102"

var cacheTime = beego.AppConfig.DefaultFloat("CacheTime", 60) // 1 分钟

func Init() {
	initAPI()
	initAdsCache()
	//NewSign().UpdateSignRule() // 更新签到规则的全局变量
	//NewReadRecord().UpdateReading  Rule() // 更新阅读计时规则的全局变量
}

type SitemapDocs struct {
	DocumentId   int
	DocumentName string
	Identify     string
	BookId       int
}

// 统计书籍分类
var counting = false

type Count struct {
	Cnt        int
	CategoryId int
}

func CountCategory() {
	if counting {
		return
	}
	counting = true
	defer func() {
		counting = false
	}()

	var count []Count

	sql := "select count(bc.id) cnt, bc.category_id from " + BookCategory{}.TableName() + " bc left join " + Book{}.TableName() + " b on b.book_id=bc.book_id where b.privately_owned=0 group by bc.category_id"
	mus.Db.Raw(sql).Scan(&count)
	if len(count) == 0 {
		return
	}

	var cates []Category

	mus.Db.Select("id,pid,cnt").Find(&cates)
	if len(cates) == 0 {
		return
	}

	var err error

	db := mus.Db.Begin()
	defer func() {
		if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	db.Model(Category{}).Updates(Ups{"cnt": 0})

	cateChild := make(map[int]int)
	for _, item := range count {
		if item.Cnt > 0 {
			cateChild[item.CategoryId] = item.Cnt
			err = db.Model(Category{}).Where("id=?", item.CategoryId).Updates(Ups{"cnt": item.Cnt}).Error
			if err != nil {
				return
			}
		}
	}
}

func getTimeRange(t time.Time, prd period) (start, end string) {
	switch prd {
	case PeriodWeek:
		start, end = getWeek(t)
	case PeriodLastWeek:
		start, end = getWeek(t.AddDate(0, 0, -7))
	case PeriodMonth:
		start, end = getMonth(t)
	case PeriodLastMoth:
		start, end = getMonth(t.AddDate(0, -1, 0))
	case PeriodAll:
		start = "20060102"
		end = "20401231"
	case PeriodDay:
		start = t.Format(dateFormat)
		end = start
	case PeriodYear:
		start, end = getYear(t.AddDate(-1, 0, 0))
	default:
		start = t.Format(dateFormat)
		end = start
	}
	return
}

func getWeek(t time.Time) (start, end string) {
	if t.Weekday() == 0 {
		start = t.Add(-7 * 24 * time.Hour).Format(dateFormat)
		end = t.Format(dateFormat)
	} else {
		s := t.Add(-time.Duration(t.Weekday()-1) * 24 * time.Hour)
		start = s.Format(dateFormat)
		end = s.Add(6 * 24 * time.Hour).Format(dateFormat)
	}
	return
}

func getYear(t time.Time) (start, end string) {
	month := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.Local)
	start = month.Format(dateFormat)
	end = month.AddDate(0, 12, 0).Add(-24 * time.Hour).Format(dateFormat)
	return
}

func getMonth(t time.Time) (start, end string) {
	month := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
	start = month.Format(dateFormat)
	end = month.AddDate(0, 1, 0).Add(-24 * time.Hour).Format(dateFormat)
	return
}
