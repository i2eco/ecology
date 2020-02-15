package manager

type CategoryCreate struct {
	Pid   int    `form:"pid"`
	Cates string `form:"cates"`
}
