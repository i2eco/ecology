package book

type ReqIndex struct {
	Private int `json:"private"form:"private"`
	Page    int `json:"page"form:"page"`
}

type ReqUploadCover struct {
	X      float64 `form:"x"`
	Y      float64 `form:"y"`
	Width  float64 `form:"width"`
	Height float64 `form:"height"`
}
