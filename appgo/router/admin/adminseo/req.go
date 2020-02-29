package adminseo

import "github.com/i2eco/ecology/appgo/model/trans"

type ReqList struct {
	Name string `form:"name"`
	trans.ReqPage
}

type ReqInfo struct {
	Id int `form:"id"`
}

type ReqCreate struct {
	Page        string `json:"page"`        //页面
	Statement   string `json:"statement"`   //页面说明
	Title       string `json:"title"`       //SEO标题
	Keywords    string `json:"keywords"`    //SEO关键字
	Description string `json:"description"` //SEO摘要
}

type ReqUpdate struct {
	Id          int    `json:"id"`
	Page        string `json:"page"`        //页面
	Statement   string `json:"statement"`   //页面说明
	Title       string `json:"title"`       //SEO标题
	Keywords    string `json:"keywords"`    //SEO关键字
	Description string `json:"description"` //SEO摘要
}
