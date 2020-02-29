package install

import (
	"fmt"
	"github.com/i2eco/ecology/appgo/pkg/util"
	"github.com/jinzhu/gorm"
	"time"

	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	mmysql "github.com/i2eco/muses/pkg/database/mysql"
)

var Models []interface{}

func init() {
	Models = []interface{}{
		&mysql.Attachment{},
		&mysql.Book{},
		&mysql.BookCategory{},
		&mysql.Member{},
		&mysql.Option{},
		&mysql.GithubUser{},
		&mysql.MemberToken{},
		&mysql.Category{},
		&mysql.Star{},
		&mysql.Score{},
		&mysql.Comments{},
		&mysql.Relationship{},
		&mysql.Banner{},
		&mysql.Logs{},
		&mysql.Label{},
		&mysql.Document{},
		&mysql.DocumentStore{},
		&mysql.DocumentHistory{},
		&mysql.Wechat{},
		&mysql.ReadRecord{},
		&mysql.ReadCount{},
		&mysql.Bookmark{},
		&mysql.BookCounter{},
		&mysql.ReadingTime{},
		&mysql.Sign{},
		&mysql.Seo{},
		&mysql.AdsCont{},
		&mysql.FriendLink{},
		&mysql.Fans{},
		&mysql.Awesome{},
		&mysql.AwesomeCate{},
		&mysql.User{},
	}
}
func Create(isClear bool) error {
	db := mmysql.Caller("ecology")
	if isClear {
		db.DropTable(Models...)
	}

	db.SingularTable(true)
	if db.Error != nil {
		return db.Error
	}

	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(Models...)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func Mock() error {
	db := mmysql.Caller("ecology")
	options := []mysql.Option{
		{
			OptionValue: "true",
			OptionName:  "ENABLED_REGISTER",
			OptionTitle: "是否启用注册",
		},
		{
			OptionValue: "100",
			OptionName:  "ENABLE_DOCUMENT_HISTORY",
			OptionTitle: "版本控制",
		}, {
			OptionValue: "true",
			OptionName:  "ENABLED_CAPTCHA",
			OptionTitle: "是否启用验证码",
		}, {
			OptionValue: "true",
			OptionName:  "ENABLE_ANONYMOUS",
			OptionTitle: "启用匿名访问",
		}, {
			OptionValue: "Ecology",
			OptionName:  "SITE_NAME",
			OptionTitle: "站点名称",
		}, {
			OptionValue: "",
			OptionName:  "ICP",
			OptionTitle: "网站备案",
		}, {
			OptionValue: "",
			OptionName:  "TONGJI",
			OptionTitle: "站点统计",
		}, {
			OptionValue: "true",
			OptionName:  "SPIDER",
			OptionTitle: "采集器，是否只对管理员开放",
		}, {
			OptionValue: "false",
			OptionName:  "ELASTICSEARCH_ON",
			OptionTitle: "是否开启全文搜索",
		}, {
			OptionValue: "http://localhost:9200/",
			OptionName:  "ELASTICSEARCH_HOST",
			OptionTitle: "ElasticSearch Host",
		}, {
			OptionValue: "book",
			OptionName:  "DEFAULT_SEARCH",
			OptionTitle: "默认搜索",
		}, {
			OptionValue: "50",
			OptionName:  "SEARCH_ACCURACY",
			OptionTitle: "搜索精度",
		}, {
			OptionValue: "true",
			OptionName:  "LOGIN_QQ",
			OptionTitle: "是否允许使用QQ登录",
		}, {
			OptionValue: "true",
			OptionName:  "LOGIN_GITHUB",
			OptionTitle: "是否允许使用Github登录",
		}, {
			OptionValue: "true",
			OptionName:  "LOGIN_GITEE",
			OptionTitle: "是否允许使用码云登录",
		}, {
			OptionValue: "0",
			OptionName:  "RELATE_BOOK",
			OptionTitle: "是否开始关联书籍",
		}, {
			OptionValue: "true",
			OptionName:  "ALL_CAN_WRITE_BOOK",
			OptionTitle: "是否都可以创建项目",
		}, {
			OptionValue: "false",
			OptionName:  "CLOSE_SUBMIT_ENTER",
			OptionTitle: "是否关闭收录入口",
		}, {
			OptionValue: "true",
			OptionName:  "CLOSE_OPEN_SOURCE_LINK",
			OptionTitle: "是否关闭开源项目入口",
		}, {
			OptionValue: "0",
			OptionName:  "HOUR_REG_NUM",
			OptionTitle: "同一IP每小时允许注册人数",
		}, {
			OptionValue: "0",
			OptionName:  "DAILY_REG_NUM",
			OptionTitle: "同一IP每天允许注册人数",
		}, {
			OptionValue: "X-Real-Ip",
			OptionName:  "REAL_IP_FIELD",
			OptionTitle: "request中获取访客真实IP的header",
		}, {
			OptionValue: "",
			OptionName:  "APP_PAGE",
			OptionTitle: "手机APP下载单页",
		}, {
			OptionValue: "false",
			OptionName:  "HIDE_TAG",
			OptionTitle: "是否隐藏标签在导航栏显示",
		}, {
			OptionValue: "",
			OptionName:  "DOWNLOAD_LIMIT",
			OptionTitle: "是否需要登录才能下载电子书",
		}, {
			OptionValue: "",
			OptionName:  "MOBILE_BANNER_SIZE",
			OptionTitle: "手机端横幅宽高比",
		}, {
			OptionValue: "false",
			OptionName:  "AUTO_HTTPS",
			OptionTitle: "图片链接HTTP转HTTPS",
		}, {
			OptionValue: "0",
			OptionName:  "APP_VERSION",
			OptionTitle: "Android APP版本号（数字）",
		}, {
			OptionValue: "",
			OptionName:  "APP_QRCODE",
			OptionTitle: "是否在用户下载电子书的时候显示APP下载二维码",
		},
		{
			OptionValue: "5",
			OptionName:  "SIGN_BASIC_REWARD",
			OptionTitle: "用户每次签到基础奖励阅读时长(秒)",
		},
		{
			OptionValue: "10",
			OptionName:  "SIGN_APP_REWARD",
			OptionTitle: "使用APP签到额外奖励阅读时长(秒)",
		},
		{
			OptionValue: "0",
			OptionName:  "SIGN_CONTINUOUS_REWARD", //
			OptionTitle: "用户连续签到奖励阅读时长(秒)",
		}, {
			OptionValue: "0",
			OptionName:  "SIGN_CONTINUOUS_MAX_REWARD",
			OptionTitle: "连续签到奖励阅读时长上限(秒)",
		},
		{
			OptionValue: "0",
			OptionName:  "READING_MIN_INTERVAL",
			OptionTitle: "内容最小阅读计时间隔(秒)",
		},
		{
			OptionValue: "600",
			OptionName:  "READING_MAX_INTERVAL",
			OptionTitle: "内容最大阅读计时间隔(秒)",
		},
		{
			OptionValue: "1200",
			OptionName:  "READING_INVALID_INTERVAL",
			OptionTitle: "内容阅读无效计时间隔(秒)",
		},
		{
			OptionValue: "600",
			OptionName:  "READING_INTERVAL_MAX_REWARD",
			OptionTitle: "内容阅读计时间隔最大奖励(秒)",
		},
	}

	for _, op := range options {
		db.Create(&op)
	}
	hash, err := utils.PasswordHash("123456")
	if err != nil {
		return err
	}
	db.Create(&mysql.Member{
		Account:                    "ecology",
		Nickname:                   "ecology",
		Password:                   hash,
		AuthMethod:                 "local",
		Description:                "",
		Email:                      "ecology@163.com",
		Phone:                      "",
		Avatar:                     "",
		Role:                       1,
		RoleName:                   "管理员",
		Status:                     0,
		CreateTime:                 time.Now(),
		CreateAt:                   0,
		LastLoginTime:              time.Now(),
		Wxpay:                      "",
		Alipay:                     "",
		TotalReadingTime:           0,
		TotalSign:                  0,
		TotalContinuousSign:        0,
		HistoryTotalContinuousSign: 0,
	})
	createAdminUser(db)
	return nil
}

func createAdminUser(db *gorm.DB) {
	pwdHash, err := util.Hash("123456")
	if err != nil {
		fmt.Println("err", err)
		return
	}
	user := mysql.User{
		Name:          "ecologyadmin",
		Password:      pwdHash,
		Status:        1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LastLoginIP:   "127.0.0.1",
		LastLoginTime: time.Now(),
	}
	if err = db.Create(&user).Error; err != nil {
		fmt.Println("err", err)
		return
	}
}
