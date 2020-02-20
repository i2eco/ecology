package manager

import (
	"encoding/json"
	"fmt"
	"html/template"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/ecology/appgo/router/core"
)

func Index(c *core.Context) {
	c.Tpl().Data["Model"] = mysql.NewDashboard().Query()
	c.GetSeoByPage("manage_dashboard", map[string]string{
		"title":       "仪表盘 - " + dao.Global.GetSiteName(),
		"keywords":    "仪表盘",
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Tpl().Data["IsDashboard"] = true
	c.Html("manager/index")
}

// 用户列表.
func Users(c *core.Context) {
	c.Tpl().Data["IsUsers"] = true
	wd, _ := c.GetQuery("wd")
	roleStr, _ := c.GetQuery("role")
	pageIndexStr, _ := c.GetQuery("page")

	role, _ := strconv.Atoi(roleStr)
	if role == 0 {
		role = -1
	}
	pageIndex, _ := strconv.Atoi(pageIndexStr)
	if pageIndex == 0 {
		pageIndex = 1
	}

	c.GetSeoByPage("manage_users", map[string]string{
		"title":       "用户管理 - " + dao.Global.GetSiteName(),
		"keywords":    "用户管理",
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})

	members, totalCount, err := dao.Member.FindToPager(pageIndex, conf.PageSize, wd, role)

	if err != nil {
		c.Tpl().Data["ErrorMessage"] = err.Error()
		return
	}

	if totalCount > 0 {
		c.Tpl().Data["PageHtml"] = utils.NewPaginations(conf.RollPage, int(totalCount), conf.PageSize, pageIndex, "/manager/users", "")
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}

	b, err := json.Marshal(members)

	if err != nil {
		c.Tpl().Data["Result"] = template.JS("[]")
	} else {
		c.Tpl().Data["Result"] = template.JS(string(b))
	}
	c.Tpl().Data["Role"] = role
	c.Tpl().Data["Wd"] = wd
	c.Html("manager/users")
}

// 标签管理.
func Tags(c *core.Context) {
	c.Tpl().Data["IsTag"] = true
	size := 150
	wd := c.GetString("wd")
	pageIndex := c.GetInt("page")
	tags, totalCount, err := dao.Label.FindToPager(pageIndex, size, wd)
	if err != nil {
		c.Tpl().Data["ErrorMessage"] = err.Error()
		c.Html("manager/tags")
		return
	}
	if totalCount > 0 {
		c.Tpl().Data["PageHtml"] = utils.NewPaginations(conf.RollPage, int(totalCount), size, pageIndex, beego.URLFor("ManagerController.Tags"), "")
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}
	c.Tpl().Data["Total"] = totalCount
	c.Tpl().Data["Tags"] = tags
	c.Tpl().Data["Wd"] = wd
	c.Html("manager/tags")
}

func AddTags(c *core.Context) {
	tags := c.GetString("tags")
	if tags != "" {
		tags = strings.Join(strings.Split(tags, "\n"), ",")
		dao.Label.InsertOrUpdateMulti(tags)
	}
	c.JSONOK()
}

func DelTags(c *core.Context) {
	id := c.GetInt("id")
	if id > 0 {
		mus.Db.Where("label_id = ?", id).Delete(&mysql.Label{})
	}
	c.JSONOK()
}

//编辑用户信息.
func EditMemberHtml(c *core.Context) {
	memberId := c.GetInt(":id")
	if memberId <= 0 {
		c.Html404()
		return
	}

	member, err := dao.Member.Find(memberId)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.Html404()
		return
	}

	c.GetSeoByPage("manage_users_edit", map[string]string{
		"title":       "用户编辑 - " + dao.Global.GetSiteName(),
		"keywords":    "用户标记",
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Tpl().Data["IsUsers"] = true
	c.Tpl().Data["Model"] = member
	c.Html("manager/edit_users")
}

//项目列表.
func Books(c *core.Context) {
	pageIndexStrc, _ := c.GetQuery("page")
	privateStr, _ := c.GetQuery("private")
	pageIndex, _ := strconv.Atoi(pageIndexStrc)
	private, _ := strconv.Atoi(privateStr)
	if pageIndex == 0 {
		pageIndex = 1
	}

	books, totalCount, err := dao.Book.ResultFindToPager(pageIndex, conf.PageSize, c.Member().MemberId, private)
	if err != nil {
		c.Html404()
		return
	}

	if totalCount > 0 {
		c.Tpl().Data["PageHtml"] = utils.NewPaginations(conf.RollPage, totalCount, conf.PageSize, pageIndex, beego.URLFor("ManagerController.Books"), fmt.Sprintf("&private=%v", private))
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}

	c.Tpl().Data["Lists"] = books
	c.Tpl().Data["IsBooks"] = true
	c.GetSeoByPage("manage_project_list", map[string]string{
		"title":       "项目管理 - " + dao.Global.GetSiteName(),
		"keywords":    "项目管理",
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Tpl().Data["Private"] = private
	c.Html("manager/books")
}

//编辑项目.
func EditBookHtml(c *core.Context) {
	identify := c.Param("key")
	if identify == "" {
		c.Html404()
		return
	}

	book, err := dao.Book.FindByFieldFirst("identify", identify)
	if err != nil {
		c.Html404()
		return
	}

	if book.PrivateToken != "" {
		book.PrivateToken = c.BaseUrl() + beego.URLFor("DocumentController.Index", ":key", book.Identify, "token", book.PrivateToken)
	}
	c.Tpl().Data["Model"] = book

	c.GetSeoByPage("manage_project_edit", map[string]string{
		"title":       "项目设置 - " + dao.Global.GetSiteName(),
		"keywords":    "项目设置",
		"description": dao.Global.GetSiteName() + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	c.Html("manager/edit_book")
}

// 删除项目.
func DeleteBook(c *core.Context) {
	bookId := c.GetInt("book_id")
	if bookId <= 0 {
		c.JSONErrStr(6001, "参数错误")
		return
	}

	//用户密码
	pwd := c.GetString("password")
	if m, err := dao.Member.Login(c.Member().Account, pwd); err != nil || m.MemberId == 0 {
		c.JSONErrStr(1, "项目删除失败，您的登录密码不正确")
		return
	}

	b, _ := dao.Book.Find(bookId)
	if b.Identify != c.GetString("identify") {
		c.JSONErrStr(1, "项目删除失败，您输入的文档标识不正确")
		return
	}
	err := dao.Book.ThoroughDeleteBook(bookId)

	if err != nil {
		c.JSONErrStr(6003, "删除失败")
		return
	}

	go func() {
		client := dao.NewElasticSearchClient()
		if errDel := client.DeleteIndex(bookId, true); errDel != nil && client.On {
			mus.Logger.Error(errDel.Error())
		}
	}()

	c.JSONOK()
}

// CreateToken 创建访问来令牌.
func CreateToken(c *core.Context) {

	if c.ForbidGeneralRole() {
		c.JSONErrStr(6001, "您的角色非作者和管理员，无法创建访问令牌")
		return
	}

	action := c.GetString("action")
	identify := c.GetString("identify")

	book, err := dao.Book.FindByFieldFirst("identify", identify)
	if err != nil {
		c.JSONErrStr(6001, "项目不存在")
		return
	}

	if action == "create" {
		if book.PrivatelyOwned == 0 {
			c.JSONErrStr(6001, "公开项目不能创建阅读令牌")
		}

		book.PrivateToken = string(utils.Krand(conf.GetTokenSize(), utils.KC_RAND_KIND_ALL))
		if err := mus.Db.UpdateColumns(book); err != nil {
			logs.Error("生成阅读令牌失败 => ", err)
			c.JSONErrStr(6003, "生成阅读令牌失败")
			return
		}
		c.JSONOK()
		c.JSONOK(c.BaseUrl() + beego.URLFor("DocumentController.Index", ":key", book.Identify, "token", book.PrivateToken))
	}

	book.PrivateToken = ""
	if err := mus.Db.UpdateColumns(book); err != nil {
		c.JSONErrStr(6004, "删除令牌失败")
		return
	}
	c.JSONOK()
}

func SettingHtml(c *core.Context) {
	options := dao.Global.All()

	for _, item := range options {
		if item.OptionName == "APP_PAGE" {
			c.Tpl().Data["APP_PAGE"] = item.OptionValue
			c.Tpl().Data["M_APP_PAGE"] = item
		} else {
			c.Tpl().Data[item.OptionName] = item
		}
	}
	c.Tpl().Data["SITE_TITLE"] = dao.Global.GetSiteName()
	c.Tpl().Data["IsSetting"] = true
	c.Tpl().Data["SeoTitle"] = "配置管理"
	c.Html("manager/setting")
}

// Transfer 转让项目.
func Transfer(c *core.Context) {
	account := c.GetString("account")
	if account == "" {
		c.JSONErrStr(6004, "接受者账号不能为空")
		return
	}

	member, err := mysql.NewMember().FindByAccount(account)
	if err != nil {
		c.JSONErrStr(6005, "接受用户不存在")
		return
	}

	if member.Status != 0 {
		c.JSONErrStr(6006, "接受用户已被禁用")
		return
	}

	if !c.Member().IsAdministrator() {
		c.Html404()
		return
	}

	identify := c.GetString("identify")

	book, err := dao.Book.FindByFieldFirst("identify", identify)
	if err != nil {
		c.JSONErrStr(6001, err.Error())
		return
	}

	rel, err := dao.Relationship.FindFounder(book.BookId)
	if err != nil {
		c.JSONErrStr(6009, "查询项目创始人失败")
		return
	}

	if member.MemberId == rel.MemberId {
		c.JSONErrStr(6007, "不能转让给自己")
		return
	}

	err = dao.Relationship.Transfer(book.BookId, rel.MemberId, member.MemberId)
	if err != nil {
		c.JSONErrStr(6008, err.Error())
		return
	}
	c.JSONOK()
}

func Comments(c *core.Context) {
	status := c.GetString("status")
	statusNum, _ := strconv.Atoi(status)
	p := c.GetInt("page")
	size := c.GetInt("size")
	if status == "" {
		c.Tpl().Data["Comments"], _ = dao.Comments.Comments(p, size, 0)
	} else {
		c.Tpl().Data["Comments"], _ = dao.Comments.Comments(p, size, 0, statusNum)
	}
	c.Tpl().Data["IsComments"] = true
	c.Tpl().Data["Status"] = status
	count, _ := dao.Comments.Count(0, statusNum)
	c.Tpl().Data["Count"] = count
	if count > 0 {
		html := utils.GetPagerHtml(c.Context.Request.RequestURI, p, size, int(count))
		c.Tpl().Data["PageHtml"] = html
	}
	c.Html("manager/comments")
}

func ClearComments(c *core.Context) {
	uid := c.GetInt("uid")
	if uid > 0 {
		dao.Comments.ClearComments(uid)
	}
	c.JSONOK()
}

func DeleteComment(c *core.Context) {
	id := c.GetInt("id")
	if id > 0 {
		dao.Comments.DeleteComment(id)
	}
	c.JSONOK()

}

func SetCommentStatus(c *core.Context) {
	id := c.GetInt("id")
	status := c.GetInt("value")
	field := c.GetString("field")
	if id > 0 && field == "status" {
		if err := dao.Comments.SetCommentStatus(id, status); err != nil {
			c.JSONErrStr(1, err.Error())
			return
		}
	}
	c.JSONOK()
	return
}

//设置项目私有状态.
func PrivatelyOwned(c *core.Context) {
	status := c.GetString("status")
	identify := c.GetString("identify")

	if status != "open" && status != "close" {
		c.JSONErrStr(6003, "参数错误")
		return
	}

	state := 0
	if status == "open" {
		state = 0
	} else {
		state = 1
	}

	if !c.Member().IsAdministrator() {
		c.Html404()
		return
	}

	book, err := dao.Book.FindByFieldFirst("identify", identify)
	if err != nil {
		c.JSONErrStr(6001, err.Error())
		return
	}

	book.PrivatelyOwned = state

	err = mus.Db.UpdateColumns(book).Error
	if err != nil {
		c.JSONErrStr(6004, "保存失败")
		return
	}

	go func() {
		mysql.CountCategory()
		public := true
		if state == 1 {
			public = false
		}
		client := dao.NewElasticSearchClient()
		if errSet := client.SetBookPublic(book.BookId, public); errSet != nil && client.On {
			mus.Logger.Error(errSet.Error())
		}
	}()
	c.JSONOK()
}

//附件列表.
func AttachList(c *core.Context) {

	pageIndex := c.GetInt("page")

	attachList, totalCount, err := dao.Attachment.FindToPager(pageIndex, conf.PageSize)
	if err != nil {
		c.Html404()
	}

	if totalCount > 0 {
		html := utils.GetPagerHtml(c.Context.Request.RequestURI, pageIndex, conf.PageSize, int(totalCount))
		c.Tpl().Data["PageHtml"] = html
	} else {
		c.Tpl().Data["PageHtml"] = ""
	}

	for _, item := range attachList {
		p := filepath.Join("./", item.FilePath)
		item.IsExist = utils.FileExists(p)
	}

	c.Tpl().Data["Lists"] = attachList
	c.Tpl().Data["IsAttach"] = true
	c.Html("manager/attach_list")
}

//附件详情.
func AttachDetailed(c *core.Context) {
	attachId, _ := strconv.Atoi(c.Param("id"))
	if attachId <= 0 {
		c.Html404()
		return
	}

	attach, err := dao.Attachment.ResultFind(attachId)
	if err != nil {
		c.Html404()
		return
	}

	attach.FilePath = filepath.Join("./", attach.FilePath)
	attach.HttpPath = c.BaseUrl() + attach.HttpPath
	attach.IsExist = utils.FileExists(attach.FilePath)
	c.Tpl().Data["Model"] = attach
	c.Html("manager/attach_detailed")
}

//删除附件.
func AttachDelete(c *core.Context) {
	attachId := c.GetInt("attach_id")
	if attachId <= 0 {
		c.Html404()
		return
	}

	attach, err := dao.Attachment.Find(attachId)
	if err != nil {
		c.JSONErrStr(6001, err.Error())
		return
	}

	if err := dao.Attachment.DeleteFilePath(attach); err != nil {
		c.JSONErrStr(6002, err.Error())
		return
	}
	c.JSONOK()
}

//
////SEO管理
func SeoHtml(c *core.Context) {
	//SEO展示
	var seos []mysql.Seo
	mus.Db.Find(&seos)
	c.Tpl().Data["Lists"] = seos
	c.Tpl().Data["IsManagerSeo"] = true
	c.Html("manager/seo")
}

//
//func  UpdateAds(c *core.Context)  {
//	id, _ := this.GetInt("id")
//	field := this.GetString("field")
//	value := this.GetString("value")
//	if field == "" {
//		c.JSONErrStr(1, "字段不能为空")
//	}
//	_, err := orm.NewOrm().QueryTable(mysql.NewAdsCont()).Filter("id", id).Update(orm.Params{field: value})
//	if err != nil {
//		c.JSONErrStr(1, err.Error())
//	}
//	go mysql.UpdateAdsCache()
//	c.JSONErrStr(0, "操作成功")
//}
//
//func  DelAds(c *core.Context)  {
//	id, _ := this.GetInt("id")
//	_, err := orm.NewOrm().QueryTable(mysql.NewAdsCont()).Filter("id", id).Delete()
//	if err != nil {
//		c.JSONErrStr(1, err.Error())
//	}
//	go mysql.UpdateAdsCache()
//	c.JSONErrStr(0, "删除成功")
//}
//
//广告管理
func Ads(c *core.Context) {
	mobile, _ := c.GetQuery("mobile")
	if mobile == "" {
		mobile = "0"
	}

	layout := "2006-01-02"
	c.Tpl().Data["Mobile"] = mobile
	c.Tpl().Data["Positions"] = mysql.NewAdsCont().GetPositions()
	c.Tpl().Data["Lists"] = mysql.NewAdsCont().Lists(mobile == "1")
	c.Tpl().Data["IsAds"] = true
	c.Tpl().Data["Now"] = time.Now().Format(layout)
	c.Tpl().Data["Next"] = time.Now().Add(time.Hour * 24 * 730).Format(layout)
	c.Html("manager/ads")
}

//
////更行书籍项目的排序
//func  UpdateBookSort(c *core.Context)  {
//	bookId, _ := this.GetInt("book_id")
//	orderIndex, _ := this.GetInt("value")
//	if bookId > 0 {
//		if _, err := orm.NewOrm().QueryTable("md_books").Filter("book_id", bookId).Update(orm.Params{
//			"order_index": orderIndex,
//		}); err != nil {
//			c.JSONErrStr(1, err.Error())
//		}
//	}
//	c.JSONErrStr(0, "排序更新成功")
//}
//
//func  Sitemap(c *core.Context)  {
//	baseUrl := c.Context.Input.Scheme() + "://" + c.Context.Request.Host
//	if host := beego.AppConfig.String("sitemap_host"); len(host) > 0 {
//		baseUrl = c.Context.Input.Scheme() + "://" + host
//	}
//	go mysql.SitemapUpdate(baseUrl)
//	c.JSONErrStr(0, "站点地图更新提交成功，已交由后台执行更新，请耐心等待。")
//}

//分类管理
func Category(c *core.Context) {
	//查询所有分类
	cates, err := dao.Category.GetCates(c.Context, -1, -1)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.Html404()
		return
	}

	var parents []mysql.Category
	for idx, item := range cates {
		if strings.TrimSpace(item.Icon) == "" { //赋值为默认图片
			item.Icon = "/static/images/icon.png"
		} else {
			item.Icon = utils.ShowImg(item.Icon)
		}
		if item.Pid == 0 {
			parents = append(parents, item)
		}
		cates[idx] = item
	}

	c.Tpl().Data["Parents"] = parents
	c.Tpl().Data["Cates"] = cates
	c.Tpl().Data["IsCategory"] = true
	c.Html("manager/category")
}

//更新分类字段内容
func UpdateCate(c *core.Context) {
	field := c.Query("field")
	val := c.Query("value")
	id, _ := strconv.Atoi(c.Query("id"))
	if err := dao.Category.UpdateByField(id, field, val); err != nil {
		c.JSONErrStr(1, "更新失败："+err.Error())
		return
	}
	c.JSONOK()
}

//删除分类
func DelCate(c *core.Context) {
	//var err error
	if id := c.GetInt("id"); id > 0 {
		// todo fix
		//err = new(mysql.Category).Del(id)
	}
	//if err != nil {
	//	c.JSONErrStr(1, err.Error())
	//}
	c.JSONOK()
}

//
////更新分类的图标
//func  UpdateCateIcon(c *core.Context)  {
//	var err error
//	id, _ := this.GetInt("id")
//	if id == 0 {
//		c.JSONErrStr(1, "参数不正确")
//	}
//	model := new(mysql.Category)
//	if cate := model.Find(id); cate.Id > 0 {
//		cate.Icon = strings.TrimLeft(cate.Icon, "/")
//		f, h, err1 := this.GetFile("icon")
//		if err1 != nil {
//			err = err1
//		}
//		defer f.Close()
//
//		tmpFile := fmt.Sprintf("uploads/icons/%v%v"+filepath.Ext(h.Filename), id, time.Now().Unix())
//		os.MkdirAll(filepath.Dir(tmpFile), os.ModePerm)
//		if err = this.SaveToFile("icon", tmpFile); err == nil {
//			switch utils.StoreType {
//			case utils.StoreOss:
//				store.ModelStoreOss.MoveToOss(tmpFile, tmpFile, true, false)
//				store.ModelStoreOss.DelFromOss(cate.Icon)
//			case utils.StoreLocal:
//				store.ModelStoreLocal.DelFiles(cate.Icon)
//			}
//			err = model.UpdateByField(cate.Id, "icon", "/"+tmpFile)
//		}
//	}
//
//	if err != nil {
//		c.JSONErrStr(1, err.Error())
//	}
//	c.JSONErrStr(0, "更新成功")
//}
//
//友情链接
func FriendLink(c *core.Context) {
	c.Tpl().Data["SeoTitle"] = "友链管理"
	c.Tpl().Data["Links"] = new(mysql.FriendLink).GetList(true)
	c.Tpl().Data["IsFriendlink"] = true
	c.Html("manager/friendlink")
}

//
////添加友链
//func  AddFriendlink(c *core.Context)  {
//	if err := new(mysql.FriendLink).Add(this.GetString("title"), this.GetString("link")); err != nil {
//		c.JSONErrStr(1, "新增友链失败:"+err.Error())
//	}
//	c.JSONErrStr(0, "新增友链成功")
//}
//
////更新友链
//func  UpdateFriendlink(c *core.Context)  {
//	id, _ := this.GetInt("id")
//	if err := new(mysql.FriendLink).Update(id, this.GetString("field"), this.GetString("value")); err != nil {
//		c.JSONErrStr(1, "操作失败："+err.Error())
//	}
//	c.JSONErrStr(0, "操作成功")
//}
//
////删除友链
//func  DelFriendlink(c *core.Context)  {
//	id, _ := this.GetInt("id")
//	if err := new(mysql.FriendLink).Del(id); err != nil {
//		c.JSONErrStr(1, "删除失败："+err.Error())
//	}
//	c.JSONErrStr(0, "删除成功")
//}
//
//// 重建全量索引
//func  RebuildAllIndex(c *core.Context)  {
//	go mysql.NewElasticSearchClient().RebuildAllIndex()
//	c.JSONErrStr(0, "提交成功，请耐心等待")
//}
//
func Banner(c *core.Context) {
	c.Tpl().Data["SeoTitle"] = "横幅管理"
	c.Tpl().Data["Banners"], _ = dao.Banner.All()
	c.Tpl().Data["IsBanner"] = true
	c.Html("manager/banners")
}

//
//func  DeleteBanner(c *core.Context)  {
//	id, _ := this.GetInt("id")
//	if id > 0 {
//		err := mysql.NewBanner().Delete(id)
//		if err != nil {
//			c.JSONErrStr(1, err.Error())
//		}
//	}
//	c.JSONErrStr(0, "删除成功")
//}
//
//func  UpdateBanner(c *core.Context)  {
//	id, _ := this.GetInt("id")
//	field := this.GetString("field")
//	value := this.GetString("value")
//	if id > 0 {
//		err := mysql.NewBanner().Update(id, field, value)
//		if err != nil {
//			c.JSONErrStr(1, err.Error())
//		}
//	}
//	c.JSONErrStr(0, "更新成功")
//}
//
//func  UploadBanner(c *core.Context)  {
//	f, h, err := this.GetFile("image")
//	if err != nil {
//		c.JSONErrStr(1, err.Error())
//	}
//	ext := strings.ToLower(filepath.Ext(strings.ToLower(h.Filename)))
//	tmpFile := fmt.Sprintf("uploads/tmp/banner-%v-%v%v", c.Member().MemberId, time.Now().Unix(), ext)
//	destFile := fmt.Sprintf("uploads/banners/%v.%v%v", c.Member().MemberId, time.Now().Unix(), ext)
//	defer func(c *core.Context)  {
//		f.Close()
//		os.Remove(tmpFile)
//		if err != nil {
//			utils.DeleteFile(destFile)
//		}
//	}()
//
//	os.MkdirAll(filepath.Dir(tmpFile), os.ModePerm)
//	err = this.SaveToFile("image", tmpFile)
//	if err != nil {
//		c.JSONErrStr(1, err.Error())
//	}
//	err = utils.UploadFile(tmpFile, destFile)
//	if err != nil {
//		c.JSONErrStr(1, err.Error())
//	}
//	banner := &mysql.Banner{
//		Image:     "/" + destFile,
//		Type:      this.GetString("type"),
//		Title:     this.GetString("title"),
//		Link:      this.GetString("link"),
//		Status:    true,
//		CreatedAt: time.Now(),
//	}
//	banner.Sort, _ = this.GetInt("sort")
//	_, err = orm.NewOrm().Insert(banner)
//	if err != nil {
//		c.JSONErrStr(1, err.Error())
//	}
//	c.JSONErrStr(0, "横幅上传成功")
//}

//func  SubmitBook(c *core.Context)  {
//	m := mysql.NewSubmitBooks()
//	page, _ := this.GetInt("page", 1)
//	size, _ := this.GetInt("size", 100)
//	books, total, _ := m.Lists(page, size)
//	if total > 0 {
//		c.Tpl().Data["PageHtml"] = utils.NewPaginations(conf.RollPage, int(total), size, page, beego.URLFor("ManagerController.SubmitBook"), "")
//	} else {
//		c.Tpl().Data["PageHtml"] = ""
//	}
//	c.Tpl().Data["Books"] = books
//	c.Tpl().Data["IsSubmitBook"] = true
//	c.Html("manager/submit_book")
//}
//
//func  DeleteSubmitBook(c *core.Context)  {
//	id, _ := this.GetInt("id")
//	orm.NewOrm().QueryTable(mysql.NewSubmitBooks()).Filter("id", id).Delete()
//	c.JSONErrStr(0, "删除成功")
//}
//
//func  UpdateSubmitBook(c *core.Context)  {
//	field := this.GetString("field")
//	value := this.GetString("value")
//	id, _ := this.GetInt("id")
//	if id > 0 {
//		_, err := orm.NewOrm().QueryTable(mysql.NewSubmitBooks()).Filter("id", id).Update(orm.Params{field: value})
//		if err != nil {
//			c.JSONErrStr(1, err.Error())
//		}
//	}
//	c.JSONErrStr(0, "更新成功")
//}
//
