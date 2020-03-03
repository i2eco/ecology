package mus

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/i2eco/muses/pkg/cache/mixcache"
	mmysql "github.com/i2eco/muses/pkg/database/mysql"
	"github.com/i2eco/muses/pkg/logger"
	"github.com/i2eco/muses/pkg/open/github"
	"github.com/i2eco/muses/pkg/oss"
	musgin "github.com/i2eco/muses/pkg/server/gin"
	"github.com/i2eco/muses/pkg/session/ginsession"
	"github.com/jinzhu/gorm"
)

var (
	Cfg             musgin.Cfg
	Logger          *logger.Client
	Gin             *gin.Engine
	Db              *gorm.DB
	Session         gin.HandlerFunc
	Oss             *oss.Client
	Mixcache        *mixcache.Client
	GithubClient    *github.Client
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

	FormRestyClient = resty.New().SetDebug(true).SetTimeout(3*time.Second).SetHeader("Content-Type", "multipart/form-data")
	JsonRestyClient = resty.New().SetDebug(true).SetTimeout(10*time.Second).SetHeader("Content-Type", "application/json;charset=utf-8")
	GithubClient = github.Caller()

	return nil

}
