package adminawesome

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/model/csv"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/router/core"
	"github.com/i2eco/muses/pkg/system"
	"os"
	"path/filepath"
	"time"
)

func List(c *core.Context) {
	req := ReqList{}
	if err := c.Bind(&req); err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	total, list := dao.Awesome.ListPage(c.Context, mysql.Conds{}, &req.ReqPage)
	c.JSONList(list, req.ReqPage.Current, req.ReqPage.PageSize, total)
}

func Info(c *core.Context) {
	req := ReqInfo{}
	err := c.Bind(&req)
	if err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	resp, err := dao.Awesome.Info(c.Context, req.Id)
	if err != nil {
		c.JSONErrTips("获取信息失败", err)
		return
	}
	c.JSONOK(resp)
}

func Create(c *core.Context) {
	req := ReqCreate{}
	err := c.Bind(&req)
	if err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	createInfo := &mysql.Awesome{
		Name: req.Name,
		Desc: req.Desc,
	}
	err = dao.Awesome.Create(c.Context, mus.Db, createInfo)
	if err != nil {
		c.JSONErrTips("创建失败1", err)
		return
	}
	c.JSONOK()

}

func Update(c *core.Context) {
	req := ReqUpdate{}
	err := c.Bind(&req)
	if err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	err = mus.Db.Model(mysql.Seo{}).Where("id = ?", req.Id).UpdateColumns(&mysql.Awesome{
		Name: req.Name,
		Desc: req.Desc,
	}).Error
	if err != nil {
		c.JSONErrTips("更新失败", err)
		return
	}
	c.JSONOK()
}

func Delete(c *core.Context) {
	req := ReqDelete{}
	err := c.Bind(&req)
	if err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}
	err = mus.Db.Model(mysql.Seo{}).Where("id = ?", req.Id).Delete(&mysql.Awesome{}).Error
	if err != nil {
		c.JSONErrTips("更新失败", err)
		return
	}
	c.JSONOK()
}

func Upload(c *core.Context) {
	file, err := c.FormFile("file-upload")
	if err != nil {
		c.JSONErrTips("参数错误", err)
		return
	}

	fileName := GenerateUniqueMd5()
	filePath := "./cache/" + fileName + filepath.Base(file.Filename)
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		c.JSONErrTips("保存文件失败", err)
		return
	}

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm) // 此处假设当前目录下已存在test目录
	if err != nil {
		c.JSONErrTips("打开文件失败", err)
		return
	}

	defer f.Close()
	var out []*csv.GithubItem
	if err = gocsv.UnmarshalFile(f, &out); err != nil {
		c.JSONErrTips("解析csv失败", err)
		return
	}
	for _, value := range out {
		if value.Name == "" {
			continue
		}
		db := mus.Db
		var info mysql.Awesome
		db.Where("name = ?", value.Name).Find(&info)
		if info.Id > 0 {
			db.Model(mysql.Awesome{}).Where("id=?", info.Id).Updates(mysql.Ups{"desc": value.Desc})
			continue
		}

		db.Create(&mysql.Awesome{
			Name: value.Name,
			Desc: value.Desc,
		})
	}
	c.JSONOK()
}

func GenerateUniqueMd5() string {
	date := time.Now().Format("20060102150405")
	uniqueID := GenerateUniqueID()
	sno := date + system.RunInfo.HostName + string(system.RunInfo.Pid) + uniqueID

	return fmt.Sprintf("%x", md5.Sum([]byte(sno)))
}

func GenerateUniqueID() string {
	b := make([]byte, 16)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}
