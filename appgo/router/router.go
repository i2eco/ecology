package router

import (
	"github.com/gin-gonic/gin"
	"github.com/i2eco/ecology/appgo/command"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/router/admin/adminuser"
	"github.com/i2eco/ecology/appgo/router/admin/auth"
	"github.com/i2eco/ecology/appgo/router/core"
	"github.com/i2eco/ecology/appgo/router/web/account"
	"github.com/i2eco/ecology/appgo/router/web/awesome"
	"github.com/i2eco/ecology/appgo/router/web/book"
	"github.com/i2eco/ecology/appgo/router/web/bookMember"
	"github.com/i2eco/ecology/appgo/router/web/bookmark"
	"github.com/i2eco/ecology/appgo/router/web/captcha"
	"github.com/i2eco/ecology/appgo/router/web/document"
	"github.com/i2eco/ecology/appgo/router/web/home"
	"github.com/i2eco/ecology/appgo/router/web/label"
	"github.com/i2eco/ecology/appgo/router/web/localhost"
	"github.com/i2eco/ecology/appgo/router/web/manager"
	"github.com/i2eco/ecology/appgo/router/web/rank"
	"github.com/i2eco/ecology/appgo/router/web/record"
	"github.com/i2eco/ecology/appgo/router/web/setting"
	"github.com/i2eco/ecology/appgo/router/web/tutorial"
	"github.com/i2eco/ecology/appgo/router/web/user"
	"github.com/spf13/viper"
)

func InitRouter() *gin.Engine {
	r := mus.Gin

	if command.Mode == "all" || command.Mode == "web" {
		webGrp(r) // 小程序api路由组
	}
	if command.Mode == "all" || command.Mode == "admin" {
		adminGrp(r)
	}
	r.Static("/"+viper.GetString("app.osspic"), viper.GetString("app.osspic"))

	return r
}

func webGrp(r *gin.Engine) {
	r.Static("/static", "static")

	r.GET("/captcha/:captchaId", captcha.CaptchaPng)

	r.Use(mus.Session)
	tplGrp := r.Group("", core.FrontLoginRequired(), core.FrontTplRequired())
	{
		//tplGrp.GET("/", core.Handle(cate.index))
		tplGrp.GET("/", core.Handle(home.Home))
		tplGrp.GET("/ecology", core.Handle(home.Ecology))
		tplGrp.GET("/opensource", core.Handle(home.Opensource))
		tplGrp.GET("/original", core.Handle(home.Original))
		tplGrp.GET("/login", core.Handle(account.LoginHtml))
		tplGrp.GET("/logout", core.Handle(account.Logout))
		tplGrp.GET("/note", core.Handle(account.Note))
		tplGrp.GET("/login/:oauth", core.Handle(account.OauthHtml))
		tplGrp.GET("/account/find_password", core.Handle(account.FindPasswordHtml))

		tplGrp.GET("/book", core.Handle(book.Index))
		tplGrp.GET("/book/:key/setting", core.Handle(book.Setting))
		tplGrp.GET("/book/:key/dashboard", core.Handle(book.Dashboard))
		tplGrp.GET("/book/:key/users", core.Handle(book.Users))

		//tplGrp.GET("/books/:key", core.Handle(document.index))
		tplGrp.GET("/books/:key/:id", core.Handle(document.ReadHtml))
		tplGrp.GET("/books/:key", core.Handle(document.ReadHtml))
		tplGrp.GET("/document/content/:key", core.Handle(document.Edit))
		tplGrp.GET("/document/content/:key/:id", core.Handle(document.Edit))
		tplGrp.GET("/rank", core.Handle(rank.Index))
		tplGrp.GET("/tags", core.Handle(label.List))

		// 个人设置
		{
			tplGrp.GET("/setting", core.Handle(setting.Index))
			tplGrp.GET("/setting/password", core.Handle(setting.Password))
			tplGrp.GET("/setting/star", core.Handle(setting.Star))

		}

		tplGrp.GET("/manager", core.Handle(manager.Index))
		tplGrp.GET("/manager/books", core.Handle(manager.Books))
		tplGrp.GET("/manager/comments", core.Handle(manager.Comments))
		tplGrp.GET("/manager/users", core.Handle(manager.Users))
		tplGrp.GET("/manager/edit/:key", core.Handle(manager.EditBookHtml))
		tplGrp.GET("/manager/banner", core.Handle(manager.Banner))
		//tplGrp.GET("/manager/submit-book", core.Handle(manager.sub))
		tplGrp.GET("/manager/setting", core.Handle(manager.SettingHtml))
		tplGrp.GET("/manager/seo", core.Handle(manager.SeoHtml))
		tplGrp.GET("/manager/category", core.Handle(manager.Category))
		tplGrp.GET("/manager/ads", core.Handle(manager.Ads))
		tplGrp.GET("/manager/tags", core.Handle(manager.Tags))
		tplGrp.GET("/manager/friendlink", core.Handle(manager.FriendLink))

		{
			tplGrp.GET("/awesome", core.Handle(awesome.Index))
		}

		{
			tplGrp.GET("/tutorial", core.Handle(tutorial.Index))
			tplGrp.GET("/tutorial/:key", core.Handle(tutorial.Index))
		}
		tplGrp.GET("/document/history", core.Handle(document.History))

		//用户中心 【start】
		tplGrp.GET("/u/:account", core.Handle(user.Index))
		tplGrp.GET("/u/:account/collection", core.Handle(user.Collection))
		tplGrp.GET("/u/:account/follow", core.Handle(user.Follow))
		tplGrp.GET("/u/:account/fans", core.Handle(user.Fans))
		//用户中心 【end】

		tplGrp.GET("/local-render", core.Handle(localhost.RenderMarkdownHtml))
		tplGrp.POST("/local-render", core.Handle(localhost.RenderMarkdownApi))
		tplGrp.GET("/local-render-cover", core.Handle(localhost.RenderCover))
	}

	apiGrp := r.Group("/api/web", core.FrontLoginRequired())
	{

		apiGrp.GET("/awesome/info", core.Handle(awesome.Info))
		apiGrp.GET("/awesome/gen", core.Handle(awesome.Gen))
		apiGrp.GET("/awesome/all", core.Handle(awesome.All))
		apiGrp.GET("/awesome/click", core.Handle(awesome.Click))

		apiGrp.POST("/login", core.Handle(account.LoginApi))
		apiGrp.POST("/account/bind", core.Handle(account.BindApi))
		apiGrp.POST("/find_password", core.Handle(account.FindPasswordApi))
		apiGrp.GET("/valid_email", core.Handle(account.ValidEmail))

		apiGrp.GET("/book/star/:id", core.Handle(book.Star))
		apiGrp.GET("/book/score/:id", core.Handle(book.Score))
		apiGrp.POST("/book/comment/:id", core.Handle(book.Comment))

		apiGrp.POST("/book/replace/:key", core.Handle(book.Replace))
		apiGrp.POST("/book/dashboard/:key", core.Handle(book.Dashboard))
		apiGrp.POST("/book/release/:key", core.Handle(book.Release))
		apiGrp.POST("/book/sort/:key", core.Handle(book.SaveSort))

		apiGrp.POST("/book/uploadProject", core.Handle(book.UploadProject))
		apiGrp.POST("/book/downloadProject", core.Handle(book.DownloadProject))
		apiGrp.POST("/book/git-pull", core.Handle(book.GitPull))
		apiGrp.POST("/book/create", core.Handle(book.Create))
		apiGrp.POST("/book/setting/save", core.Handle(book.SaveBook))
		apiGrp.POST("/book/setting/open", core.Handle(book.PrivatelyOwned))
		apiGrp.POST("/book/setting/transfer", core.Handle(book.Transfer))
		apiGrp.POST("/book/setting/uploadCover", core.Handle(book.UploadCover))
		apiGrp.POST("/book/setting/token", core.Handle(book.CreateToken))
		apiGrp.POST("/book/setting/delete", core.Handle(book.Delete))

		apiGrp.POST("/book/users/create", core.Handle(bookMember.AddMember))
		apiGrp.POST("/book/users/change", core.Handle(bookMember.ChangeRole))
		apiGrp.POST("/book/users/delete", core.Handle(bookMember.RemoveMember))

		apiGrp.GET("/books/:key/:id", core.Handle(document.ReadApi))

		apiGrp.GET("/document/content/:key/:id", core.Handle(document.ContentGet))
		apiGrp.POST("/document/content/:key/:id", core.Handle(document.ContentPost))
		apiGrp.POST("/document/create/:key", core.Handle(document.CreateApi))
		apiGrp.POST("/document/update/:key", core.Handle(document.UpdateApi))
		apiGrp.POST("/document/upload/:key", core.Handle(document.Upload))
		apiGrp.POST("/document/delete/:key", core.Handle(document.Delete))

		apiGrp.GET("/u/follow/:uid", core.Handle(user.SetFollow))
		apiGrp.GET("/u/sign", core.Handle(user.SignToday))

		apiGrp.POST("/manager/category", core.Handle(manager.CategoryApi))
		apiGrp.POST("/manager/member/delete", core.Handle(manager.DeleteMember))
		apiGrp.GET("/manager/updateCate", core.Handle(manager.UpdateCate))

		apiGrp.POST("/setting/password", core.Handle(setting.PasswordUpdate))
		apiGrp.POST("/setting/upload", core.Handle(setting.Upload))
		apiGrp.POST("/setting/update", core.Handle(setting.Update))

		{
			apiGrp.POST("/record/list/:bookId", core.Handle(record.List))
			apiGrp.POST("/record/clear/:bookId", core.Handle(record.Clear))
			apiGrp.POST("/record/delete/:docId", core.Handle(record.Delete))

		}
		apiGrp.POST("/bookmark/info/:id", core.Handle(bookmark.Bookmark))
		apiGrp.POST("/bookmark/list/:book_id", core.Handle(bookmark.List))
		apiGrp.POST("/account/findPassword", core.Handle(account.FindPasswordApi))
	}

}

func adminGrp(r *gin.Engine) {
	adGrp := r.Group("/admin", mus.Session)
	adGrp.POST("/auth/login", mus.Session, core.Handle(auth.Login))

	// TODO 路由规范
	// 权限模块
	authGrp := adGrp.Group("/auth")
	authGrp.Use(core.AdminLoginRequired())
	{
		authGrp.GET("/self", core.Handle(auth.Self))
		authGrp.GET("/logout", core.Handle(auth.Logout))
	}

	// 参考 https://github.com/gin-gonic/gin#parameters-in-path
	// TODO 路由规范

	// 用户模块
	usersGrp := adGrp.Group("/users")
	usersGrp.Use(core.AdminLoginRequired())
	{
		usersGrp.GET("/list", core.Handle(adminuser.List)) // 获取所有用户
	}
}
