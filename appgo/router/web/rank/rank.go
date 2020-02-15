package rank

import (
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/router/core"
	"strconv"
)

func Index(c *core.Context) {
	limitStr, _ := c.GetQuery("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit == 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	tab, flag2 := c.GetQuery("tab")
	if !flag2 {
		tab = "all"
	}

	switch tab {
	case "reading":
		rt := mysql.NewReadingTime()
		c.Tpl().Data["SeoTitle"] = "阅读时长榜"
		c.Tpl().Data["TodayReading"] = rt.Sort(mysql.PeriodDay, limit, true)
		c.Tpl().Data["WeekReading"] = rt.Sort(mysql.PeriodWeek, limit, true)
		c.Tpl().Data["MonthReading"] = rt.Sort(mysql.PeriodMonth, limit, true)
		c.Tpl().Data["LastWeekReading"] = rt.Sort(mysql.PeriodLastWeek, limit, true)
		c.Tpl().Data["LastMonthReading"] = rt.Sort(mysql.PeriodLastMoth, limit, true)
		c.Tpl().Data["AllReading"] = rt.Sort(mysql.PeriodAll, limit, true)
	case "sign":
		c.Tpl().Data["SeoTitle"] = "用户签到榜"
		sign := mysql.NewSign()
		c.Tpl().Data["ContinuousSignUsers"] = sign.Sorted(limit, "total_continuous_sign", true)
		c.Tpl().Data["TotalSignUsers"] = sign.Sorted(limit, "total_sign", true)
		c.Tpl().Data["HistoryContinuousSignUsers"] = sign.Sorted(limit, "history_total_continuous_sign", true)
	case "popular":
		c.Tpl().Data["SeoTitle"] = "文档人气榜"
		bookCounter := mysql.NewBookCounter()
		//c.Tpl().Data["Today"] = bookCounter.PageViewSort(mysql.PeriodDay, limit, true)
		c.Tpl().Data["Week"] = bookCounter.PageViewSort(mysql.PeriodWeek, limit, true)
		c.Tpl().Data["Month"] = bookCounter.PageViewSort(mysql.PeriodMonth, limit, true)
		//c.Tpl().Data["LastWeek"] = bookCounter.PageViewSort(mysql.PeriodLastWeek, limit, true)
		//c.Tpl().Data["LastMonth"] = bookCounter.PageViewSort(mysql.PeriodLastMoth, limit, true)
		c.Tpl().Data["All"] = bookCounter.PageViewSort(mysql.PeriodAll, limit, true)
	case "star":
		c.Tpl().Data["SeoTitle"] = "热门收藏榜"
		bookCounter := mysql.NewBookCounter()
		//c.Tpl().Data["Today"] = bookCounter.StarSort(mysql.PeriodDay, limit, true)
		c.Tpl().Data["Week"] = bookCounter.StarSort(mysql.PeriodWeek, limit, true)
		c.Tpl().Data["Month"] = bookCounter.StarSort(mysql.PeriodMonth, limit, true)
		//c.Tpl().Data["LastWeek"] = bookCounter.StarSort(mysql.PeriodLastWeek, limit, true)
		//c.Tpl().Data["LastMonth"] = bookCounter.StarSort(mysql.PeriodLastMoth, limit, true)
		c.Tpl().Data["All"] = bookCounter.StarSort(mysql.PeriodAll, limit, true)
	default:
		tab = "all"
		c.Tpl().Data["SeoTitle"] = "总榜"
		limit = 10
		sign := mysql.NewSign()
		book := dao.Book
		c.Tpl().Data["ContinuousSignUsers"] = sign.Sorted(limit, "total_continuous_sign", true)
		c.Tpl().Data["TotalSignUsers"] = sign.Sorted(limit, "total_sign", true)
		c.Tpl().Data["TotalReadingUsers"] = sign.Sorted(limit, "total_reading_time", true)
		c.Tpl().Data["StarBooks"] = book.Sorted(limit, "star")
		c.Tpl().Data["VcntBooks"] = book.Sorted(limit, "vcnt")
		c.Tpl().Data["CommentBooks"] = book.Sorted(limit, "cnt_comment")
	}
	c.Tpl().Data["Tab"] = tab
	c.Tpl().Data["IsRank"] = true
	c.Html("rank/index")
}
