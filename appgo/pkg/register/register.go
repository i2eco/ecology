package register

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	"github.com/goecology/ecology/appgo/dao"
	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/captcha"
	"github.com/goecology/ecology/appgo/pkg/utils"
	"github.com/goecology/muses/pkg/tpl/tplbeego"
)

var cpt *captcha.Captcha

func Init() (err error) {

	fc := &cache.FileCache{CachePath: "./cache/captcha"}
	cpt = captcha.NewWithFilter("/captcha/", fc)
	fmt.Println("Init------>", Init)
	err = tplbeego.AddFuncMap("config", dao.Global.FindByKey)
	if err != nil {
		panic(err)
	}
	err = tplbeego.AddFuncMap("cdn", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdn", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})
	if err != nil {
		panic(err)
	}

	err = tplbeego.AddFuncMap("cdnjs", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdnjs", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})
	err = tplbeego.AddFuncMap("cdncss", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdncss", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})
	err = tplbeego.AddFuncMap("cdnimg", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdnimg", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})
	err = tplbeego.AddFuncMap("getUsernameByUid", func(id interface{}) string {
		return dao.Member.GetUsernameByUid(id)
	})
	err = tplbeego.AddFuncMap("getNicknameByUid", func(id interface{}) string {
		return dao.Member.GetNicknameByUid(id)
	})
	err = tplbeego.AddFuncMap("inMap", utils.InMap)
	//将标签转成a链接
	err = tplbeego.AddFuncMap("tagsToLink", func(tags string) (links string) {
		var linkArr []string
		if slice := strings.Split(tags, ","); len(slice) > 0 {
			for _, tag := range slice {
				if tag = strings.TrimSpace(tag); len(tag) > 0 {
					linkArr = append(linkArr, fmt.Sprintf(`<a target="_blank" title="%v" href="%v">%v</a>`, tag, "/tag/"+tag, tag))
				}
			}
		}
		return strings.Join(linkArr, "")
	})

	//用户是否收藏了文档
	err = tplbeego.AddFuncMap("doesStar", dao.Star.DoesStar)
	err = tplbeego.AddFuncMap("scoreFloat", utils.ScoreFloat)
	err = tplbeego.AddFuncMap("showImg", utils.ShowImg)
	err = tplbeego.AddFuncMap("IsFollow", new(mysql.Fans).Relation)
	err = tplbeego.AddFuncMap("isubstr", utils.Substr)
	err = tplbeego.AddFuncMap("ads", mysql.GetAdsCode)
	err = tplbeego.AddFuncMap("formatReadingTime", utils.FormatReadingTime)
	err = tplbeego.AddFuncMap("add", func(a, b int) int { return a + b })
	return
}
