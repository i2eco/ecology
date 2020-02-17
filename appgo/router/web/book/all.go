package book

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/code"
	"github.com/goecology/ecology/appgo/pkg/conf"
	"github.com/goecology/ecology/appgo/pkg/graphics"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/goecology/ecology/appgo/router/core"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

//保存项目信息
func SaveBook(c *core.Context) {
	bookResult, err := isFormPermission(c)
	if err != nil {
		c.JSONErrStr(6001, err.Error())
		return
	}

	book, err := dao.Book.Find(bookResult.BookId)
	if err != nil {
		logs.Error("SaveBook => ", err)
		c.JSONErrStr(6002, err.Error())
		return
	}

	bookName := strings.TrimSpace(c.GetPostFormString("book_name"))
	description := strings.TrimSpace(c.GetPostFormString("description"))
	commentStatus := c.GetPostFormString("comment_status")
	tag := strings.TrimSpace(c.GetPostFormString("label"))
	editor := strings.TrimSpace(c.GetPostFormString("editor"))

	if strings.Count(description, "") > 500 {
		c.JSONErrStr(6004, "项目描述不能大于500字")
		return

	}
	if commentStatus != "open" && commentStatus != "closed" && commentStatus != "group_only" && commentStatus != "registered_only" {
		commentStatus = "closed"
	}
	if tag != "" {
		tags := strings.Split(tag, ",")
		if len(tags) > 10 {
			c.JSONErrStr(6005, "最多允许添加10个标签")
			return

		}
	}
	if editor != "markdown" && editor != "html" {
		editor = "markdown"
	}

	book.BookName = bookName
	book.Description = description
	book.CommentStatus = commentStatus
	book.Label = tag
	book.Editor = editor
	book.Author = c.GetPostFormString("author")
	book.AuthorURL = c.GetPostFormString("author_url")
	book.Lang = c.GetPostFormString("lang")
	book.AdTitle = c.GetPostFormString("ad_title")
	book.AdLink = c.GetPostFormString("ad_link")

	if err := dao.Book.UpdateXX(book); err != nil {
		c.JSONErrStr(6006, "保存失败")
		return
	}
	bookResult.BookName = bookName
	bookResult.Description = description
	bookResult.CommentStatus = commentStatus
	bookResult.Label = tag

	//更新书籍分类
	if cids, ok := c.GetPostFormArray("cid"); ok {
		dao.BookCategory.SetBookCates(book.BookId, cids)
	}

	go func() {
		es := mysql.ElasticSearchData{
			Id:       book.BookId,
			BookId:   0,
			Title:    book.BookName,
			Keywords: book.Label,
			Content:  book.Description,
			Vcnt:     book.Vcnt,
			Private:  book.PrivatelyOwned,
		}
		client := dao.NewElasticSearchClient()
		if errSearch := client.BuildIndex(es); errSearch != nil && client.On {
			mus.Logger.Error(errSearch.Error())
		}
	}()

	c.JSONOK(bookResult)
}

//设置项目私有状态.
func PrivatelyOwned(c *core.Context) {
	status := c.GetPostFormString("status")
	if c.ForbidGeneralRole() && status == "open" {
		c.JSONErrStr(6001, "您的角色非作者和管理员，无法将项目设置为公开")
		return
	}
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

	bookResult, err := isPermission(c)
	if err != nil {
		c.JSONErrStr(6001, err.Error())
		return
	}

	//只有创始人才能变更私有状态
	if bookResult.RoleId != conf.Conf.Info.BookFounder {
		c.JSONErrStr(6002, "权限不足")
		return
	}

	if err = mus.Db.Model(mysql.Book{}).Where("book_id = ?", bookResult.BookId).Updates(mysql.Ups{
		"privately_owned": state,
	}).Error; err != nil {
		c.JSONErrStr(6004, "保存失败")
		return
	}
	go func() {
		// todo
		//mysql.CountCategory()

		public := true
		if state == 1 {
			public = false
		}
		client := dao.NewElasticSearchClient()
		if errSet := client.SetBookPublic(bookResult.BookId, public); errSet != nil && client.On {
			mus.Logger.Error(errSet.Error())
		}
	}()
	c.JSONOK("ok")
}

// Transfer 转让项目.
func Transfer(c *core.Context) {
	userMember := c.Member()
	account := c.GetPostFormString("account")
	if account == "" {
		c.JSONErrStr(6004, "接受者账号不能为空")
		return
	}

	member, err := dao.Member.FindByAccount(account)
	if err != nil {
		logs.Error("FindByAccount => ", err)
		c.JSONErrStr(6005, "接受用户不存在")
		return

	}

	if member.Status != 0 {
		c.JSONErrStr(6006, "接受用户已被禁用")
		return

	}

	if member.MemberId == userMember.MemberId {
		c.JSONErrStr(6007, "不能转让给自己")
		return

	}

	bookResult, err := isPermission(c)
	if err != nil {
		c.JSONErrStr(6001, err.Error())
		return

	}

	err = dao.Relationship.Transfer(bookResult.BookId, userMember.MemberId, member.MemberId)
	if err != nil {
		logs.Error("Transfer => ", err)
		c.JSONErrStr(6008, err.Error())
		return

	}

	c.JSONOK("ok")
}

//上传项目封面.
func UploadCover(c *core.Context) {
	req := ReqUploadCover{}
	err := c.Bind(&req)
	if err != nil {
		c.JSONErr(code.UploadCoverErr0, err)
		return
	}

	bookResult, err := isFormPermission(c)
	if err != nil {
		c.JSONErr(code.UploadCoverErr1, err)
		return
	}

	book, err := dao.Book.Find(bookResult.BookId)
	if err != nil {
		c.JSONErr(code.UploadCoverErr2, err)
		return
	}

	fileHeader, err := c.FormFile("image-file")
	if err != nil {
		c.JSONErr(code.UploadCoverErr3, err)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSONErr(code.UploadCoverErr4, err)
		return
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)

	if !strings.EqualFold(ext, ".png") && !strings.EqualFold(ext, ".jpg") && !strings.EqualFold(ext, ".gif") && !strings.EqualFold(ext, ".jpeg") {
		c.JSONErr(code.UploadCoverErr5, err)
		return
	}

	fileName := strconv.FormatInt(time.Now().UnixNano(), 16)

	filePath := filepath.Join("uploads", time.Now().Format("200601"), fileName+ext)

	path := filepath.Dir(filePath)

	os.MkdirAll(path, os.ModePerm)

	err = c.SaveToFile("image-file", filePath)

	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	if utils.StoreType != utils.StoreLocal {
		defer func(filePath string) {
			os.Remove(filePath)
		}(filePath)
	}

	//剪切图片
	subImg, err := graphics.ImageCopyFromFile(filePath, int(req.X), int(req.Y), int(req.Width), int(req.Height))
	if err != nil {
		c.JSONErr(code.UploadCoverErr6, err)
		return
	}

	filePath = filepath.Join(conf.Conf.Info.WorkingDirectory, "uploads", time.Now().Format("200601"), fileName+ext)

	//生成缩略图并保存到磁盘
	err = graphics.ImageResizeSaveFile(subImg, 175, 230, filePath)
	if err != nil {
		c.JSONErr(code.UploadCoverErr7, err)
		return
	}

	dstPath := mus.Oss.GenerateKey("eco-book")

	oldCover := book.Cover
	book.Cover = dstPath
	if utils.StoreType == utils.StoreLocal {
		book.Cover = dstPath
	}

	if err := mus.Db.Model(book).Where("book_id = ?", book.BookId).Update(map[string]interface{}{
		"cover": book.Cover,
	}).Error; err != nil {
		c.JSONErr(code.UploadCoverErr8, err)
		return
	}

	//如果原封面不是默认封面则删除
	if oldCover != conf.Conf.Info.DefaultCover {
		err = mus.Oss.DeleteObject(oldCover)
		if err != nil {
			mus.Logger.Warn("remove error", zap.Error(err))
		}
	}

	err = mus.Oss.PutObjectFromFile(dstPath, filePath)
	if err != nil {
		c.JSONErr(code.UploadCoverErr10, err)
		return
	}
	c.JSONOK(mus.Oss.ShowImg(dstPath))
}

// Create 创建项目.
func Create(c *core.Context) {
	member := c.Member()

	if opt, err := dao.Global.FindByKey("ALL_CAN_WRITE_BOOK"); err == nil {
		if opt.OptionValue == "false" && member.Role == conf.MemberGeneralRole { // 读者无权限创建项目
			c.JSONErrStr(1, "普通读者无法创建项目，如需创建项目，请向管理员申请成为作者")
			return
		}
	}
	bookName := strings.TrimSpace(c.GetPostFormString("book_name"))
	bookType := strings.TrimSpace(c.GetPostFormString("book_type"))
	identify := strings.TrimSpace(c.GetPostFormString("identify"))
	description := strings.TrimSpace(c.GetPostFormString("description"))
	author := strings.TrimSpace(c.GetPostFormString("author"))
	authorURL := strings.TrimSpace(c.GetPostFormString("author_url"))
	privatelyOwned, _ := strconv.Atoi(c.GetPostFormString("privately_owned"))
	commentStatus := c.GetPostFormString("comment_status")

	if bookName == "" {
		c.JSONErrStr(6001, "项目名称不能为空")
		return
	}

	if identify == "" {
		c.JSONErrStr(6002, "项目标识不能为空")
		return

	}

	ok, err1 := regexp.MatchString(`^[a-zA-Z0-9_\-\.]*$`, identify)
	if !ok || err1 != nil {
		c.JSONErrStr(6003, "项目标识只能包含字母、数字，以及“-”、“.”和“_”符号，且不能是纯数字")
		return

	}

	if num, _ := strconv.Atoi(identify); strconv.Itoa(num) == identify {
		c.JSONErrStr(6003, "项目标识不能是纯数字")
		return

	}

	if strings.Count(identify, "") > 50 {
		c.JSONErrStr(6004, "项目标识不能超过50字")
		return

	}

	if strings.Count(description, "") > 500 {
		c.JSONErrStr(6004, "项目描述不能大于500字")
		return
	}

	if privatelyOwned != 0 && privatelyOwned != 1 {
		privatelyOwned = 1
	}
	if commentStatus != "open" && commentStatus != "closed" && commentStatus != "group_only" && commentStatus != "registered_only" {
		commentStatus = "closed"
	}

	var book mysql.Book

	if books, _ := dao.Book.FindByField("identify", identify); len(books) > 0 {
		c.JSONErrStr(6006, "项目标识已存在")
		return
	}

	book.Label = utils.SegWord(bookName)
	book.BookName = bookName
	book.Author = author
	book.AuthorURL = authorURL
	book.Description = description
	book.CommentCount = 0
	book.PrivatelyOwned = privatelyOwned
	book.CommentStatus = commentStatus
	book.Identify = identify
	book.DocCount = 0
	book.MemberId = member.MemberId
	book.CommentCount = 0
	book.Version = time.Now().Unix()
	book.Cover = conf.GetDefaultCover()
	book.Editor = "markdown"
	book.Theme = "default"
	book.BookType = bookType
	book.Score = 40 //默认评分，40即表示4星

	//设置默认时间，因为beego的orm好像无法设置datetime的默认值
	defaultTime, _ := time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
	book.LastClickGenerate = defaultTime
	book.GenerateTime, _ = time.Parse("2006-01-02 15:04:05", "2000-01-02 15:04:05") //默认生成文档的时间
	book.ReleaseTime = defaultTime
	book.CreateTime = time.Now()
	book.ModifyTime = time.Now()

	if err := dao.Book.Insert(mus.Db, &book); err != nil {
		c.JSONErrStr(6005, "保存项目失败")
		return
	}
	fmt.Println("book------>", book)
	fmt.Println("member------>", member)

	bookResult, err := dao.Book.ResultFindByIdentify(book.Identify, member.MemberId)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	c.JSONOK(bookResult)
}

// CreateToken 创建访问来令牌.
func CreateToken(c *core.Context) {
	if c.ForbidGeneralRole() {
		c.JSONErrStr(6001, "您的角色非作者和管理员，无法创建访问令牌")
		return
	}
	action := c.GetPostFormString("action")

	bookResult, err := isPermission(c)

	if err != nil {
		if err == dao.ErrPermissionDenied {
			c.JSONErrStr(403, "权限不足")
			return

		}

		logs.Error("生成阅读令牌失败 =>", err)
		c.JSONErrStr(6002, err.Error())
		return

	}

	var book mysql.Book
	if _, err := dao.Book.Find(bookResult.BookId); err != nil {
		c.JSONErrStr(6001, "项目不存在")
	}

	if action == "create" {
		if bookResult.PrivatelyOwned == 0 {
			c.JSONErrStr(6001, "公开项目不能创建阅读令牌")
		}

		book.PrivateToken = string(utils.Krand(conf.GetTokenSize(), utils.KC_RAND_KIND_ALL))
		if err := dao.Book.UpdateXX(&book); err != nil {
			logs.Error("生成阅读令牌失败 => ", err)
			c.JSONErrStr(6003, "生成阅读令牌失败")
		}
		//book.PrivateToken = this.BaseUrl() + beego.URLFor("DocumentController.Index", ":key", book.Identify, "token", book.PrivateToken)
		tipsFmt := "访问链接：%v  访问密码：%v"
		privateToken := fmt.Sprintf(tipsFmt, c.BaseUrl()+beego.URLFor("DocumentController.Index", ":key", book.Identify), book.PrivateToken)
		c.JSONOK(privateToken)
		return
	}

	book.PrivateToken = ""
	if err := dao.Book.UpdateXX(&book); err != nil {
		logs.Error("CreateToken => ", err)
		c.JSONErrStr(6004, "删除令牌失败")
		return
	}
	c.JSONOK()
}

// Delete 删除项目.
func Delete(c *core.Context) {

	bookResult, err := isPermission(c)
	if err != nil {
		c.JSONErrStr(6001, err.Error())
		return
	}

	if bookResult.RoleId != conf.BookFounder {
		c.JSONErrStr(6002, "只有创始人才能删除项目")
		return

	}

	member := c.Member()
	//用户密码
	pwd := c.GetPostFormString("password")
	if m, err := dao.Member.Login(member.Account, pwd); err != nil || m.MemberId == 0 {
		c.JSONErrStr(1, "项目删除失败，您的登录密码不正确")
		return
	}

	err = dao.Book.ThoroughDeleteBook(bookResult.BookId)
	if err == gorm.ErrRecordNotFound {
		c.JSONErrStr(6002, "项目不存在")
		return
	}

	if err != nil {
		logs.Error("删除项目 => ", err)
		c.JSONErrStr(6003, "删除失败")
		return

	}

	go func() {
		client := dao.NewElasticSearchClient()
		if errDel := client.DeleteIndex(bookResult.BookId, true); errDel != nil && client.On {
			mus.Logger.Error(errDel.Error())
		}
	}()

	c.JSONOK("ok")
}
