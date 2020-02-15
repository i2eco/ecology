package static

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/model/mysql/store"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/goecology/ecology/appgo/pkg/utils"
)

func (this *StaticController) APP() {
	link := strings.TrimSpace(mysql.GetOptionValue("APP_PAGE", ""))
	if link != "" {
		this.Redirect(link, 302)
	}
	c.Html404()
}

//静态文件，这个加在路由的最后
func (this *StaticController) StaticFile() {
	file := this.GetString(":splat")
	if strings.HasPrefix(file, ".well-known") || file == "sitemap.xml" {
		http.ServeFile(c.Context.ResponseWriter, c.Context.Request, file)
		return
	}
	file = strings.TrimLeft(file, "./")
	path := filepath.Join(utils.VirtualRoot, file)
	http.ServeFile(c.Context.ResponseWriter, c.Context.Request, path)
}

// 项目静态文件
func (this *StaticController) ProjectsFile() {
	prefix := "projects/"
	object := prefix + strings.TrimLeft(this.GetString(":splat"), "./")

	//这里的时间只是起到缓存的作用
	t, _ := time.Parse("2006-01-02 15:04:05", "2006-01-02 15:04:05")
	date := t.Format(http.TimeFormat)
	since := c.Context.Request.Header.Get("If-Modified-Since")
	if since == date {
		c.Context.ResponseWriter.WriteHeader(http.StatusNotModified)
		return
	}

	if utils.StoreType == utils.StoreOss { //oss
		reader, err := store.NewOss().GetFileReader(object)
		if err != nil {
			mus.Logger.Error(err.Error())
			c.Html404()
		}
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			mus.Logger.Error(err.Error())
			c.Html404()
		}
		c.Context.ResponseWriter.Header().Set("Last-Modified", date)
		if strings.HasSuffix(object, ".svg") {
			c.Context.ResponseWriter.Header().Set("Content-Type", "image/svg+xml")
		}
		c.Context.ResponseWriter.Write(b)
	} else { //local
		c.Html404()
	}
}
