package account

type ReqBind struct {
	Account   string `form:"account"`
	Nickname  string `form:"nickname"`
	Password1 string `form:"password1"`
	Password2 string `form:"password2"`
	Email     string `form:"email"`
	OauthType string `form:"oauth"`
	OauthId   string `form:"id"`
	Avatar    string `form:"avatar"`
	IsBind    int    `form:"isbind"`
}
