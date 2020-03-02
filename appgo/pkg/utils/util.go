package utils

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/i2eco/ecology/appgo/pkg/sego"
	"github.com/mssola/user_agent"
	html1 "html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

//存储类型

//更多存储类型有待扩展
const (
	StoreLocal = "local"
	StoreOss   = "oss"
)

//分词器
var (
	Segmenter   sego.Segmenter
	BasePath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	StoreType   = "local" //存储类型
	langs       sync.Map
)

func init() {
	//加载分词字典

	langs.Store("zh", "中文")
	langs.Store("en", "英文")
	langs.Store("other", "其他")
}

func GetLang(lang string) string {
	if val, ok := langs.Load(lang); ok {
		return val.(string)
	}
	return "中文"
}

//分词
//@param            str         需要分词的文字
func SegWord(str interface{}) (wds string) {
	//如果已经成功加载字典
	if Segmenter.Dictionary() != nil {
		wds = sego.SegmentsToString(Segmenter.Segment([]byte(fmt.Sprintf("%v", str))), true)
		var wdslice []string
		slice := strings.Split(wds, " ")
		for _, wd := range slice {
			w := strings.Split(wd, "/")[0]
			if (strings.Count(w, "") - 1) >= 2 {
				if i, _ := strconv.Atoi(w); i == 0 { //如果为0，则表示非数字
					wdslice = append(wdslice, w)
				}
			}
		}
		wds = strings.Join(wdslice, ",")
	}
	return
}

//评分处理
func ScoreFloat(score int) string {
	return fmt.Sprintf("%1.1f", float32(score)/10.0)
}

//操作图片显示
//如果用的是oss存储，这style是avatar、cover可选项
func ShowImg(img string, style ...string) (url string) {
	if strings.HasPrefix(img, "https://") || strings.HasPrefix(img, "http://") {
		return img
	}
	img = "/" + strings.TrimLeft(img, "./")
	switch StoreType {
	case StoreOss:
		//s := ""
		//if len(style) > 0 && strings.TrimSpace(style[0]) != "" {
		//	s = "/" + style[0]
		//}
		//url = strings.TrimRight(beego.AppConfig.String("oss::Domain"), "/ ") + img + s
	case StoreLocal:
		url = img
	}
	return
}

// Substr returns the substr from start to length.
func Substr(s string, length int) string {
	bt := []rune(s)
	start := 0
	dot := false

	if start > len(bt) {
		start = start % len(bt)
	}
	var end int
	if (start + length) > (len(bt) - 1) {
		end = len(bt)
	} else {
		end = start + length
		dot = true
	}

	str := string(bt[start:end])
	if dot {
		str = str + "..."
	}
	return str
}

//分页函数（这个分页函数不具有通用性）
//rollPage:展示分页的个数
//totalRows：总记录
//currentPage:每页显示记录数
//urlPrefix:url链接前缀
//urlParams:url键值对参数
func NewPaginations(rollPage, totalRows, listRows, currentPage int, urlPrefix string, urlSuffix string, urlParams ...interface{}) html1.HTML {
	var (
		htmlPage, path string
		pages          []int
		params         []string
	)
	//总页数
	totalPage := totalRows / listRows
	if totalRows%listRows > 0 {
		totalPage += 1
	}
	//只有1页的时候，不分页
	if totalPage < 2 {
		return ""
	}
	paramsLen := len(urlParams)
	if paramsLen > 0 {
		if paramsLen%2 > 0 {
			paramsLen = paramsLen - 1
		}
		for i := 0; i < paramsLen; {
			key := strings.TrimSpace(fmt.Sprintf("%v", urlParams[i]))
			val := strings.TrimSpace(fmt.Sprintf("%v", urlParams[i+1]))
			//键存在，同时值不为0也不为空
			if len(key) > 0 && len(val) > 0 && val != "0" {
				params = append(params, key, val)
			}
			i = i + 2
		}
	}

	path = strings.Trim(urlPrefix, "/")
	if len(params) > 0 {
		path = path + "/" + strings.Trim(strings.Join(params, "/"), "/")
	}
	//最后再处理一次“/”，是为了防止urlPrifix参数为空时，出现多余的“/”
	path = "/" + strings.Trim(path, "/")

	if currentPage > totalPage {
		currentPage = totalPage
	}
	if currentPage < 1 {
		currentPage = 1
	}
	index := 0
	rp := rollPage * 2
	for i := rp; i > 0; i-- {
		p := currentPage + rollPage - i
		if p > 0 && p <= totalPage {

			pages = append(pages, p)
		}
	}
	for k, v := range pages {
		if v == currentPage {
			index = k
		}
	}
	pages_len := len(pages)
	if currentPage > 1 {
		htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`?page=1%v">1..</a></li><li><a class="num" href="`+path+`?page=%d%v">«</a></li>`, urlSuffix, currentPage-1, urlSuffix)
	}
	if pages_len <= rollPage {
		for _, v := range pages {
			if v == currentPage {
				htmlPage += fmt.Sprintf(`<li class="active"><a href="javascript:void(0);">%d</a></li>`, v)
			} else {
				htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`?page=%d%v">%d</a></li>`, v, urlSuffix, v)
			}
		}

	} else {
		var pageSlice []int
		indexMin := index - rollPage/2
		indexMax := index + rollPage/2
		if indexMin > 0 && indexMax < pages_len { //切片索引未越界
			pageSlice = pages[indexMin:indexMax]
		} else {
			if indexMin < 0 {
				pageSlice = pages[0:rollPage]
			} else if indexMax > pages_len {
				pageSlice = pages[(pages_len - rollPage):pages_len]
			} else {
				pageSlice = pages[indexMin:indexMax]
			}

		}

		for _, v := range pageSlice {
			if v == currentPage {
				htmlPage += fmt.Sprintf(`<li class="active"><a href="javascript:void(0);">%d</a></li>`, v)
			} else {
				htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`?page=%d%v">%d</a></li>`, v, urlSuffix, v)
			}
		}

	}
	if currentPage < totalPage {
		htmlPage += fmt.Sprintf(`<li><a class="num" href="`+path+`?page=%v%v">»</a></li><li><a class="num" href="`+path+`?page=%v%v">..%d</a></li>`, currentPage+1, urlSuffix, totalPage, urlSuffix, totalPage)
	}

	return html1.HTML(`<ul class="pagination">` + htmlPage + `</ul>`)
}

func NewPaginations2(rollPage, totalRows, listRows, currentPage int, urlPrefix string, urlSuffix string, urlParams ...interface{}) html1.HTML {
	var (
		htmlPage, path string
		pages          []int
		params         []string
	)
	//总页数
	totalPage := totalRows / listRows
	if totalRows%listRows > 0 {
		totalPage += 1
	}
	//只有1页的时候，不分页
	if totalPage < 2 {
		return ""
	}
	paramsLen := len(urlParams)
	if paramsLen > 0 {
		if paramsLen%2 > 0 {
			paramsLen = paramsLen - 1
		}
		for i := 0; i < paramsLen; {
			key := strings.TrimSpace(fmt.Sprintf("%v", urlParams[i]))
			val := strings.TrimSpace(fmt.Sprintf("%v", urlParams[i+1]))
			//键存在，同时值不为0也不为空
			if len(key) > 0 && len(val) > 0 && val != "0" {
				params = append(params, key, val)
			}
			i = i + 2
		}
	}

	path = strings.Trim(urlPrefix, "/")
	if len(params) > 0 {
		path = path + "/" + strings.Trim(strings.Join(params, "/"), "/")
	}
	//最后再处理一次“/”，是为了防止urlPrifix参数为空时，出现多余的“/”
	path = "/" + strings.Trim(path, "/")

	if currentPage > totalPage {
		currentPage = totalPage
	}
	if currentPage < 1 {
		currentPage = 1
	}
	index := 0
	rp := rollPage * 2
	for i := rp; i > 0; i-- {
		p := currentPage + rollPage - i
		if p > 0 && p <= totalPage {

			pages = append(pages, p)
		}
	}
	for k, v := range pages {
		if v == currentPage {
			index = k
		}
	}
	pages_len := len(pages)
	if currentPage > 1 {
		htmlPage += fmt.Sprintf(`<li><a class="show-loader" href="`+path+`?page=1%v">1..</a></li><li><a class="show-loader" href="`+path+`?page=%d%v">«</a></li>`, urlSuffix, currentPage-1, urlSuffix)
	}
	if pages_len <= rollPage {
		for _, v := range pages {
			if v == currentPage {
				htmlPage += fmt.Sprintf(`<li class="active"><a href="javascript:void(0);">%d</a></li>`, v)
			} else {
				htmlPage += fmt.Sprintf(`<li><a class="show-loader" href="`+path+`?page=%d%v">%d</a></li>`, v, urlSuffix, v)
			}
		}

	} else {
		var pageSlice []int
		indexMin := index - rollPage/2
		indexMax := index + rollPage/2
		if indexMin > 0 && indexMax < pages_len { //切片索引未越界
			pageSlice = pages[indexMin:indexMax]
		} else {
			if indexMin < 0 {
				pageSlice = pages[0:rollPage]
			} else if indexMax > pages_len {
				pageSlice = pages[(pages_len - rollPage):pages_len]
			} else {
				pageSlice = pages[indexMin:indexMax]
			}

		}

		for _, v := range pageSlice {
			if v == currentPage {
				htmlPage += fmt.Sprintf(`<li class="active"><a href="javascript:void(0);">%d</a></li>`, v)
			} else {
				htmlPage += fmt.Sprintf(`<li><a class="show-loader" href="`+path+`?page=%d%v">%d</a></li>`, v, urlSuffix, v)
			}
		}

	}
	if currentPage < totalPage {
		htmlPage += fmt.Sprintf(`<li><a class="show-loader" href="`+path+`?page=%v%v">»</a></li><li><a class="num" href="`+path+`?page=%v%v">..%d</a></li>`, currentPage+1, urlSuffix, totalPage, urlSuffix, totalPage)
	}

	return html1.HTML(`<ul class="pagination text-center">` + htmlPage + `</ul>`)
}

//判断数据是否在map中
func InMap(maps map[int]bool, key int) (ret bool) {
	if _, ok := maps[key]; ok {
		return true
	}
	return
}

// 从HTML中获取text
func GetTextFromHtml(htmlStr string) (txt string) {
	if doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr)); err == nil {
		txt = strings.TrimSpace(doc.Find("body").Text())
	}
	return
}

// Git Clone
func GitClone(url, folder string) error {
	os.RemoveAll(folder)
	args := []string{"clone", url, folder}
	return exec.Command("git", args...).Run()
}

type SplitMD struct {
	Identify string
	Cont     string
}

// 处理http响应成败
func HandleResponse(resp *http.Response, err error) error {
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			b, _ := ioutil.ReadAll(resp.Body)
			err = errors.New(resp.Status + "；" + string(b))
		}
	}
	return err
}

type ReflectVal struct {
	T reflect.Type
	V reflect.Value
}

func IsMobile(userAgent string) bool {
	return user_agent.New(userAgent).Mobile()
}

func FormatReadingTime(seconds int, withoutTag ...bool) string {
	strFmt := "<span>%v</span> <small>小时</small> <span>%v</span> <small>分钟</small>"
	if len(withoutTag) > 0 && withoutTag[0] {
		strFmt = "%v 小时 %v 分钟"
	}
	hour := int(seconds / 3600)
	second := int(seconds % 3600 / 60)
	secondStr := strconv.Itoa(second)
	if second < 10 {
		secondStr = "0" + secondStr
	}
	return fmt.Sprintf(strFmt, hour, secondStr)
}
