package conf

import (
	"github.com/spf13/viper"
)

var (
	Conf config // holds the global app config.
)

type config struct {
	// 应用配置
	App   app
	Oss   oss
	Image image
	Info  info
}

type info struct {
	RegexpAccount     string
	MemberSuperRole   int // 超级管理员.
	MemberAdminRole   int //普通管理员.
	MemberGeneralRole int //普通用户.
	RegexpEmail       string
	PageSize          int
	RollPage          int
	DefaultAvatar     string //获取默认头像
	DefaultCover      string
	BookFounder       int
	BookAdmin         int
	BookEditor        int
	BookObserver      int
	WorkingDirectory  string
	TokenSize         int // 获取阅读令牌长度
}

type image struct {
	Domain string
	Path   string
}

type app struct {
	SitemapHost string
	Name        string `toml:"name"`
	Wechat      wechat `toml:"wechat"`
	WechatPay   wechatPay
	WechatOpen  wechatOpen
	CdnName     string
	File        string `toml:"file"`
	DbPrefix    string `toml:"dbPrefix"`

	AppKey             string `toml:"appKey"`
	Baidumapkey        string
	ExportHeader       string
	ExportFooter       string
	ExportFontSize     string
	ExportPagerSize    string
	ExportCreator      string
	ExportMarginLeft   string
	ExportMarginRight  string
	ExportMarginTop    string
	ExportMarginBottom string
}

type wechat struct {
	CodeToSessURL string
	AppID         string
	AppSecret     string
}

type wechatPay struct {
	AppID       string
	MchID       string
	Key         string
	CallbackApi string
}

type wechatOpen struct {
	AppID       string
	AppSecret   string
	RedirectURI string
	Scope       string
}

type oss struct {
	Domain string
}

func Init() error {
	// Set defaults.
	Conf = config{
		Info: info{
			RegexpAccount:     `^[a-zA-Z][a-zA-z0-9\.]{2,50}$`,
			MemberSuperRole:   0,
			MemberAdminRole:   1,
			MemberGeneralRole: 2,
			RegexpEmail:       `^(\w)+(\.\w+)*@(\w)+((\.\w+)+)$`,
			PageSize:          20,
			RollPage:          4,
			DefaultAvatar:     "/static/images/headimgurl.jpg",
			DefaultCover:      "/static/images/book.jpg",
			// 创始人.
			BookFounder: 0,
			//管理者
			BookAdmin: 1,
			//编辑者.
			BookEditor: 2,
			//观察者
			BookObserver:     3,
			WorkingDirectory: "./",
			TokenSize:        12,
		},
	}
	err := viper.Unmarshal(&Conf)
	return err
}
