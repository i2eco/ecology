package dao

import (
	"fmt"
	"github.com/TruthHun/BookStack/utils"
	"github.com/i2eco/ecology/appgo/pkg/constx"
	"github.com/i2eco/ecology/appgo/pkg/mus"
	"os"
	"strings"
	"time"
)

// 版本控制，文件存储于获取
type versionControl struct {
	DocId    int    //文档id
	Version  int64  //版本(时间戳)
	HtmlFile string //HTML文件名
	MdFile   string //md文件名
}

func NewVersionControl(docId int, version int64) *versionControl {
	t := time.Unix(version, 0).Format("2006/01/02/%v/15/04/05")
	folder := "./version_control/" + fmt.Sprintf(t, docId)
	if utils.StoreType == utils.StoreLocal {
		os.MkdirAll(folder, os.ModePerm)
	}
	return &versionControl{
		DocId:    docId,
		Version:  version,
		HtmlFile: folder + "master.html",
		MdFile:   folder + "master.md",
	}
}

// 保存版本数据
func (v *versionControl) SaveVersion(htmlContent, mdContent string) (err error) {
	err = mus.Oss.PutObject(constx.OssVersion+"/"+strings.TrimLeft(v.HtmlFile, "./"), strings.NewReader(htmlContent))
	if err != nil {
		return
	}
	err = mus.Oss.PutObject(constx.OssVersion+"/"+strings.TrimLeft(v.MdFile, "./"), strings.NewReader(mdContent))
	if err != nil {
		return
	}
	return
}

// 获取版本数据
func (v *versionControl) GetVersionContent(isHtml bool) (content string, err error) {
	file := v.MdFile
	if isHtml {
		file = v.HtmlFile
	}
	var contentByte []byte
	contentByte, err = mus.Oss.GetObject(file)
	if err != nil {
		return
	}

	content = string(contentByte)
	return
}

// 删除版本文件
func (v *versionControl) DeleteVersion() (err error) {
	_, err = mus.Oss.DeleteObjects([]string{v.HtmlFile, v.MdFile})
	return
}
