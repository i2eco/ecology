package trans

type ReqPage struct {
	Current  int    `json:"currentPage" form:"currentPage"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Sort     string `json:"sort" form:"sort"`
}

type RespOauthLogin struct {
	CurrentAuthority string `json:"currentAuthority"`
}

type ReqOauthLogin struct {
	Name string `json:"userName" binding:"required"`
	Pwd  string `json:"password" binding:"required"`
	Type string `json:"type"`
}
