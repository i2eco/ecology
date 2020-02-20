package setting

type ReqPasswordUpdate struct {
	Password1 string `form:"password1"`
	Password2 string `form:"password2"`
	Password3 string `form:"password3"`
}

type ReqUpload struct {
	X      float64 `form:"x"`
	Y      float64 `form:"y"`
	Width  float64 `form:"width"`
	Height float64 `form:"height"`
}

type ReqUpdate struct {
	Email       string `form:"password1"`
	Phone       string `form:"password2"`
	Description string `form:"password3"`
}
