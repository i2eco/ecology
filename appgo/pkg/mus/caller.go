package mus

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/goecology/muses/pkg/cache/mixcache"
	mmysql "github.com/goecology/muses/pkg/database/mysql"
	"github.com/goecology/muses/pkg/logger"
	"github.com/goecology/muses/pkg/oss"
	musgin "github.com/goecology/muses/pkg/server/gin"
	"github.com/goecology/muses/pkg/session/ginsession"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

var (
	Cfg             musgin.Cfg
	Logger          *logger.Client
	Gin             *gin.Engine
	Db              *gorm.DB
	Session         gin.HandlerFunc
	Oss             *oss.Client
	Mixcache        *mixcache.Client
	JsonRestyClient *resty.Client
	FormRestyClient *resty.Client
)

// Init 初始化muses相关容器
func Init() error {
	Cfg = musgin.Config()
	Db = mmysql.Caller("ecology")
	Logger = logger.Caller("system")
	Gin = musgin.Caller()
	Oss = oss.Caller("ecology")
	Mixcache = mixcache.Caller("ecology")
	Session = ginsession.Caller()

	dataSource := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", viper.GetString("muses.mysql.egoshop.username"), viper.GetString("muses.mysql.egoshop.password"), viper.GetString("muses.mysql.egoshop.addr"), viper.GetString("muses.mysql.egoshop.db"))

	FormRestyClient = resty.New().SetDebug(true).SetTimeout(3*time.Second).SetHeader("Content-Type", "multipart/form-data")
	JsonRestyClient = resty.New().SetDebug(true).SetTimeout(3*time.Second).SetHeader("Content-Type", "application/json;charset=utf-8")

	orm.RegisterDataBase("default", "mysql", dataSource)
	return nil

}
