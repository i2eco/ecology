package setting

import (
	"fmt"
	"github.com/i2eco/ecology/appgo/dao"
	"github.com/i2eco/ecology/appgo/pkg/conf"
	"github.com/i2eco/ecology/appgo/pkg/constx"
	"github.com/i2eco/ecology/appgo/pkg/graphics"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/i2eco/ecology/appgo/model/mysql"
	"github.com/i2eco/ecology/appgo/model/mysql/store"
	"github.com/i2eco/ecology/appgo/pkg/code"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"github.com/i2eco/ecology/appgo/pkg/utils"
	"github.com/i2eco/ecology/appgo/router/core"
)

//基本信息
func Update(c *core.Context) {
	var req ReqUpdate
	err := c.Bind(&req)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	if req.Email == "" {
		c.JSONErrStr(601, "邮箱不能为空")
		return
	}
	member := c.Member()
	member.Email = req.Email
	member.Phone = req.Phone
	member.Description = req.Description
	if err := member.Update(); err != nil {
		c.JSONErrStr(602, err.Error())
		return
	}
	c.UpdateUser(member)
	c.JSONOK()
}

//修改密码
func PasswordUpdate(c *core.Context) {
	var req ReqPasswordUpdate
	err := c.Bind(&req)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	password1 := req.Password1
	password2 := req.Password2
	password3 := req.Password3

	if password1 == "" {
		c.JSONErrStr(6003, "原密码不能为空")
		return
	}

	if password2 == "" {
		c.JSONErrStr(6004, "新密码不能为空")
		return
	}

	if count := strings.Count(password2, ""); count < 6 || count > 18 {
		c.JSONErrStr(6009, "密码必须在6-18字之间")
		return
	}

	if password2 != password3 {
		c.JSONErrStr(6003, "确认密码不正确")
		return
	}

	if ok, _ := utils.PasswordVerify(c.Member().Password, password1); !ok {
		c.JSONErrStr(6005, "原始密码不正确")
		return
	}

	if password1 == password2 {
		c.JSONErrStr(6006, "新密码不能和原始密码相同")
		return
	}

	pwd, err := utils.PasswordHash(password2)
	if err != nil {
		c.JSONErrStr(6007, "密码加密失败")
		return
	}

	c.Member().Password = pwd
	if err := c.Member().Update(); err != nil {
		c.JSONErrStr(6008, err.Error())
		return
	}
	c.JSONOK()
}

//二维码
func QrcodeUpdate(c *core.Context) {
	header, err := c.FormFile("qrcode")
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	file, err := header.Open()
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	defer file.Close()

	payType, _ := c.GetPostForm("paytype")

	alipay := true
	if payType == "wxpay" {
		alipay = false
	}

	ext := filepath.Ext(header.Filename)

	if !strings.EqualFold(ext, ".png") && !strings.EqualFold(ext, ".jpg") && !strings.EqualFold(ext, ".gif") && !strings.EqualFold(ext, ".jpeg") {
		c.JSONErrStr(500, "不支持的图片格式")
	}

	savePath := fmt.Sprintf("uploads/qrcode/%v/%v%v", c.Member().MemberId, time.Now().Unix(), ext)
	err = os.MkdirAll(filepath.Dir(savePath), 0777)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	if err = c.SaveToFile("qrcode", savePath); err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}
	url := ""
	switch utils.StoreType {
	case utils.StoreOss:
		if err := store.ModelStoreOss.MoveToOss(savePath, savePath, true, false); err != nil {
			mus.Logger.Error(err.Error())
		} else {
			url = strings.TrimRight(beego.AppConfig.String("oss::Domain"), "/ ") + "/" + savePath
		}
	case utils.StoreLocal:
		if err := store.ModelStoreLocal.MoveToStore(savePath, savePath); err != nil {
			mus.Logger.Error(err.Error())
		} else {
			url = "/" + savePath
		}
	}

	var member mysql.Member
	mus.Db.Where("member_id = ?", c.Member().MemberId).Find(&member)

	if member.MemberId > 0 {
		dels := []string{}

		if alipay {
			dels = append(dels, member.Alipay)
			member.Alipay = savePath
		} else {
			dels = append(dels, member.Wxpay)
			member.Wxpay = savePath
		}

		err = mus.Db.Where("member_id = ?", member.MemberId).Updates(mysql.Ups{
			"wxpay":  member.Wxpay,
			"alipay": member.Alipay,
		}).Error

		if err == nil {
			switch utils.StoreType {
			case utils.StoreOss:
				go store.ModelStoreOss.DelFromOss(dels...)
			case utils.StoreLocal:
				go store.ModelStoreLocal.DelFiles(dels...)
			}
		}
	}
	//删除旧的二维码，并更新新的二维码
	data := map[string]interface{}{
		"url":    url,
		"alipay": alipay,
	}
	c.JSONOK(data)

}

// Upload 上传图片
func Upload(c *core.Context) {
	req := ReqUpload{}
	err := c.Bind(&req)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	var (
		filePath string
		fileName string
	)

	filePath, fileName, err = c.SaveToFileImg("image-file")

	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	//剪切图片
	subImg, err := graphics.ImageCopyFromFile(filePath, int(req.X), int(req.Y), int(req.Width), int(req.Height))

	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	defer func(filePath string) {
		os.Remove(filePath)
	}(filePath)

	filePath = filepath.Join(conf.Conf.Info.WorkingDirectory, "uploads", time.Now().Format("200601"), fileName)

	err = graphics.ImageResizeSaveFile(subImg, 120, 120, filePath)
	if err != nil {
		c.JSONErr(code.MsgErr, err)
		return
	}

	dstPath := mus.Oss.GenerateKey(constx.OssUser)

	if member, err := dao.Member.Find(c.Member().MemberId); err == nil {
		oldAvatar := member.Avatar

		err = mus.Oss.PutObjectFromFile(dstPath, filePath)
		if err != nil {
			c.JSONErr(code.UploadCoverErr10, err)
			return
		}
		member.Avatar = dstPath
		err = dao.Member.UpdateX(c.Context, mus.Db, mysql.Conds{"member_id": c.Member().MemberId}, mysql.Ups{"avatar": dstPath})
		if err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
		err = mus.Oss.DeleteObject(oldAvatar)
		if err != nil {
			mus.Logger.Warn("remove error", zap.Error(err))
		}
		err = c.UpdateUser(member)
		if err != nil {
			c.JSONErr(code.MsgErr, err)
			return
		}
	}

	c.JSONOK(mus.Oss.ShowImg(dstPath))
}
