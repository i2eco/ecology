package setting

type ReqPasswordUpdate struct {
	Password1 string `form:"password1"`
	Password2 string `form:"password2"`
	Password3 string `form:"password3"`
}
