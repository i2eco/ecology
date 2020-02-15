package oauth

import (
	"encoding/json"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/goecology/ecology/appgo/pkg/mus"
	"github.com/spf13/viper"
)

//gitee accesstoken 数据
type GiteeAccessToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int    `json:"created_at"`
}

//gitee 用户数据
//用户使用gitee登录的时候，直接根据gitee的id获取数据
type GiteeUser struct {
	Id        int       `json:"id"`                                  //用户id
	MemberId  int       `json:"member_id"`                           //绑定的用户id
	UpdatedAt time.Time `json:"updated_at"`                          //用户资料更新时间
	AvatarURL string    `json:"avatar_url" orm:"column(avatar_url)"` //用户头像链接
	Email     string    `json:"email" orm:"size(50)"`                //电子邮箱
	Login     string    `json:"login" orm:"size(50)"`                //用户名
	Name      string    `json:"name" orm:"size(50)"`                 //昵称
	HtmlURL   string    `json:"html_url" orm:"column(html_url)"`     //gitee主页
}

//获取accessToken
func GetGiteeAccessToken(code string) (token GiteeAccessToken, err error) {
	Api := viper.GetString("oauth.giteeAccesstoken")
	ClientId := viper.GetString("oauth.giteeClientId")
	ClientSecret := viper.GetString("oauth.giteeClientSecret")
	Callback := viper.GetString("oauth.giteeCallback")
	param := map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"client_id":     ClientId,
		"redirect_uri":  Callback,
		"client_secret": ClientSecret,
	}
	var output *resty.Response
	output, err = mus.FormRestyClient.R().SetFormData(param).Post(Api)
	if err != nil {
		return
	}

	err = json.Unmarshal(output.Body(), &token)
	if err != nil {
		return
	}

	//if strings.HasPrefix(Api, "https") {
	//	req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	//}
	return
}

//获取用户信息
func GetGiteeUserInfo(accessToken string) (info GiteeUser, err error) {
	Api := viper.GetString("oauth.giteeUserInfo") + "?access_token=" + accessToken
	var output *resty.Response

	output, err = mus.FormRestyClient.R().Get(Api)
	if err != nil {
		return
	}
	//if strings.HasPrefix(Api, "https") {
	//	req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	//}
	err = json.Unmarshal(output.Body(), &info)
	if err != nil {
		return
	}
	return
}
