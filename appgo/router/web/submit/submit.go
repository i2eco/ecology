package submit

import (
	"strings"

	"github.com/goecology/webhook/app/pkg/mus"
)

func (this *SubmitController) Index() {
	c.Tpl().Data["SeoTitle"] = "开源书籍和文档收录"
	c.Tpl().Data["IsSubmit"] = true
	this.TplName = "submit/index.html"
}

func (this *SubmitController) Post() {
	uid := c.Member().MemberId
	if uid <= 0 {
		c.JSONErrStr(1, "请先登录")
	}

	form := &mysql.SubmitBooks{}
	err := this.ParseForm(form)
	if err != nil {
		mus.Logger.Error(err.Error())
		c.JSONErrStr(1, "数据解析失败")
	}

	lowerURL := strings.ToLower(form.Url)
	if !(strings.HasPrefix(lowerURL, "https://") || strings.HasPrefix(lowerURL, "http://")) {
		c.JSONErrStr(1, "URL链接地址格式不正确")
	}

	if form.Url == "" || form.Title == "" {
		c.JSONErrStr(1, "请填写必填项")
	}
	form.Uid = uid
	if err = form.Add(); err != nil {
		c.JSONErrStr(1, err.Error())
	}
	c.JSONErrStr(0, "提交成功，感谢您的分享。")
}
