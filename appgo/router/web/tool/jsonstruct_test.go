package tool

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/i2eco/ecology/appgo/router/core"
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// go test -v
func TestJsonToStruct(t *testing.T) {
	// 初始化请求地址和请求参数
	uri := "/jsontostruct"

	param := url.Values{
		"json": {`{"test":"info"}`},
	}

	router := gin.Default()
	router.POST("/jsontostruct", core.Handle(JsonToStruct))

	// 发起post请求，以表单形式传递参数
	body := PostForm(uri, param, router)

	fmt.Println(string(body))

	// 解析响应，判断响应是否与预期一致
	//response := &Response{}
	//if err := json.Unmarshal(body, response); err != nil {
	//	t.Errorf("解析响应出错，err:%v\n",err)
	//}
	//t.Log(response.Data)
	//if response.Data.Strname != "test" {
	//	t.Errorf("响应数据不符，errmsg:%v, data:%v\n",response.Errmsg,response.Data.Strname)
	//}
	//convey.Convey("测试POST接口", t, func() {
	//	convey.So(response.Data.Strname, convey.ShouldEqual, "test")
	//})

}

// Get 根据特定请求uri，发起get请求返回响应
func Get(uri string, router *gin.Engine) []byte {
	// 构造get请求
	req := httptest.NewRequest("GET", uri, nil)
	// 初始化响应
	w := httptest.NewRecorder()

	// 调用相应的handler接口
	router.ServeHTTP(w, req)

	// 提取响应
	result := w.Result()
	defer result.Body.Close()

	// 读取响应body
	body, _ := ioutil.ReadAll(result.Body)
	return body
}

// PostForm 根据特定请求uri和参数param，以表单形式传递参数，发起post请求返回响应
func PostForm(uri string, param url.Values, router *gin.Engine) []byte {

	// 构造post请求
	req := httptest.NewRequest("POST", uri, strings.NewReader(param.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 初始化响应
	w := httptest.NewRecorder()

	// 调用相应handler接口
	router.ServeHTTP(w, req)

	// 提取响应
	result := w.Result()
	defer result.Body.Close()

	// 读取响应body
	body, _ := ioutil.ReadAll(result.Body)
	return body
}
