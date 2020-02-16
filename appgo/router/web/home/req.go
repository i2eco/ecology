package home

type ReqHome struct {
	Tab  string `form:"tab"`
	Cid  int    `form:"cid"`
	Page int    `form:"page"`
}
