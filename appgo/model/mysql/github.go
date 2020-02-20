package mysql

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/i2eco/ecology/appgo/pkg/oauth"
)

var ModelGithub = new(Github)

type Github struct {
	oauth.GithubUser
}

//GitHub 用户数据
//用户使用GitHub登录的时候，直接根据GitHub的id获取数据
type GithubUser struct {
	Id        int       `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"` //用户id
	MemberId  int       `gorm:"not null;"json:"memberId"`                     //绑定的用户id
	UpdatedAt time.Time `gorm:"not null;"json:"updatedAt"`                    //用户资料更新时间
	AvatarURL string    `gorm:"not null;"json:"avatarUrl"`                    //用户头像链接
	Email     string    `gorm:"not null;"json:"email"`                        //电子邮箱
	Login     string    `gorm:"not null;"json:"login"`                        //用户名
	Name      string    `gorm:"not null;"json:"name"`                         //昵称
	HtmlURL   string    `gorm:"not null;"json:"htmlUrl"`                      //github主页
}

// TableName 获取对应数据库表名.
func (m *GithubUser) TableName() string {
	return "github_user"
}

//gitee用户的登录流程是这样的
//1、获取gitee的用户信息，用gitee的用户id查询member_id是否大于0，大于0则表示已绑定了用户信息，直接登录
//2、未绑定用户，先把gitee数据入库，然后再跳转绑定页面

//根据giteeid获取用户的gitee数据。这里可以查询用户是否绑定了或者数据是否在库中存在
func (this *Github) GetUserByGithubId(id int, cols ...string) (user Github, err error) {
	qs := orm.NewOrm().QueryTable("md_github").Filter("id", id)
	if len(cols) > 0 {
		err = qs.One(&user, cols...)
	} else {
		err = qs.One(&user)
	}
	return
}

//绑定用户
func (this *Github) Bind(githubId, memberId interface{}) (err error) {
	_, err = orm.NewOrm().QueryTable("md_github").Filter("id", githubId).Update(orm.Params{"member_id": memberId})
	return
}
