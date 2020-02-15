package user

import (
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/router/core"
	"github.com/kataras/iris/core/errors"
	"strconv"
	"time"

	"github.com/TruthHun/BookStack/conf"
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/pkg/utils"
)

func prepare(c *core.Context) (UcenterMember mysql.Member, err error) {
	account := c.Param("account")
	UcenterMember, _ = dao.Member.GetByUsername(account)
	if UcenterMember.MemberId == 0 {
		err = errors.New("not exist")
		return
	}
	rt := mysql.NewReadingTime()
	if c.Member() != nil {
		c.Tpl().Data["IsSelf"] = UcenterMember.MemberId == c.Member().MemberId
	}
	c.Tpl().Data["User"] = UcenterMember
	c.Tpl().Data["JoinedDays"] = int(time.Now().Sub(UcenterMember.CreateTime).Seconds()/(24*3600)) + 1
	c.Tpl().Data["TotalReading"] = utils.FormatReadingTime(UcenterMember.TotalReadingTime)
	c.Tpl().Data["MonthReading"] = utils.FormatReadingTime(rt.GetReadingTime(UcenterMember.MemberId, mysql.PeriodMonth))
	c.Tpl().Data["WeekReading"] = utils.FormatReadingTime(rt.GetReadingTime(UcenterMember.MemberId, mysql.PeriodWeek))
	c.Tpl().Data["TodayReading"] = utils.FormatReadingTime(rt.GetReadingTime(UcenterMember.MemberId, mysql.PeriodDay))
	c.Tpl().Data["Tab"] = "share"
	c.Tpl().Data["IsSign"] = false
	if c.Member() != nil && c.Member().MemberId > 0 {
		c.Tpl().Data["LatestSign"] = mysql.NewSign().LatestOne(c.Member().MemberId)
	}
	return
}

//首页
func Index(c *core.Context) {
	pageStr, _ := c.GetQuery("page")
	page, _ := strconv.Atoi(pageStr)
	pageSize := 10
	if page < 1 {
		page = 1
	}
	ucenterMember, err := prepare(c)
	if err != nil {
		c.Html404()
		return
	}
	books, totalCount, _ := dao.Book.FindToPager(page, pageSize, ucenterMember.MemberId, 0)
	c.Tpl().Data["Books"] = books

	if totalCount > 0 {
		html := utils.NewPaginations(conf.RollPage, totalCount, pageSize, page, "/u/"+ucenterMember.Account, "")
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}
	c.Tpl().Data["Total"] = totalCount
	c.GetSeoByPage("ucenter-share", map[string]string{
		"title":       "分享 - " + ucenterMember.Nickname,
		"keywords":    "用户主页," + ucenterMember.Nickname,
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})

	c.Html("user/index")
}

//收藏
func Collection(c *core.Context) {
	pageStr, _ := c.GetQuery("page")
	page, _ := strconv.Atoi(pageStr)
	pageSize := 10
	if page < 1 {
		page = 1
	}

	ucenterMember, err := prepare(c)
	if err != nil {
		c.Html404()
		return
	}

	totalCount, books, _ := dao.Star.ListXX(ucenterMember.MemberId, page, pageSize)
	c.Tpl().Data["Books"] = books

	if totalCount > 0 {
		html := utils.NewPaginations(conf.RollPage, int(totalCount), pageSize, page, "/u/"+ucenterMember.Account+"/collection", "")
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}
	c.GetSeoByPage("ucenter-collection", map[string]string{
		"title":       "收藏 - " + ucenterMember.Nickname,
		"keywords":    "用户收藏," + ucenterMember.Nickname,
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Tpl().Data["Total"] = totalCount
	c.Tpl().Data["Tab"] = "collection"
	c.Html("user/collection")
}

//关注
func Follow(c *core.Context) {
	pageStr, _ := c.GetQuery("page")
	page, _ := strconv.Atoi(pageStr)
	pageSize := 10
	if page < 1 {
		page = 1
	}

	ucenterMember, err := prepare(c)
	if err != nil {
		c.Html404()
		return
	}

	fans, totalCount, _ := new(mysql.Fans).GetFollowList(ucenterMember.MemberId, page, pageSize)
	if totalCount > 0 {
		html := utils.NewPaginations(conf.RollPage, int(totalCount), pageSize, page, "/u/"+ucenterMember.Account+"/follow", "")
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}
	c.GetSeoByPage("ucenter-follow", map[string]string{
		"title":       "关注 - " + ucenterMember.Nickname,
		"keywords":    "用户关注," + ucenterMember.Nickname,
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Tpl().Data["Fans"] = fans
	c.Tpl().Data["Tab"] = "follow"
	c.Html("user/fans")
}

//粉丝和关注
func Fans(c *core.Context) {
	pageStr, _ := c.GetQuery("page")
	page, _ := strconv.Atoi(pageStr)
	pageSize := 10
	if page < 1 {
		page = 1
	}

	ucenterMember, err := prepare(c)
	if err != nil {
		c.Html404()
		return
	}

	fans, totalCount, _ := new(mysql.Fans).GetFansList(ucenterMember.MemberId, page, pageSize)
	if totalCount > 0 {
		html := utils.NewPaginations(conf.RollPage, int(totalCount), pageSize, page, "/u/"+ucenterMember.Account+"/fans", "")
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}
	c.GetSeoByPage("ucenter-fans", map[string]string{
		"title":       "粉丝 - " + ucenterMember.Nickname,
		"keywords":    "用户粉丝," + ucenterMember.Nickname,
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Tpl().Data["Fans"] = fans
	c.Tpl().Data["Tab"] = "fans"
	c.Html("user/fans")
}
