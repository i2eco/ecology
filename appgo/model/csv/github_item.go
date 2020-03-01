package csv

type GithubItem struct {
	Name string `csv:"name"`
	Desc string `csv:"desc"`
}
