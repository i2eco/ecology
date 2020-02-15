package mysql

import (
	"time"
)

//全文搜索客户端
type ElasticSearchClient struct {
	Host           string        //host
	Index          string        //索引
	Type           string        //type
	On             bool          //是否启用全文搜索
	Timeout        time.Duration //超时时间
	IsRelateSearch bool
}

//全文搜索
type ElasticSearchData struct {
	Id       int    `json:"id"`       //文档或书籍id
	BookId   int    `json:"book_id"`  //书籍id。这里的book_id起到的作用是IsBooK的布尔，以及搜索书籍文档时候的过滤
	Title    string `json:"title"`    //文档标题或书籍名称
	Keywords string `json:"keywords"` //文档或书籍关键字
	Content  string `json:"content"`  //文档摘要或书籍文本内容
	Vcnt     int    `json:"vcnt"`     //浏览量
	Private  int    `json:"private"`  //书籍或者文档是否是公开的
}

//统计信息结构
type ElasticSearchCount struct {
	Shards struct {
		Failed     int `json:"failed"`
		Skipped    int `json:"skipped"`
		Successful int `json:"successful"`
		Total      int `json:"total"`
	} `json:"_shards"`
	Count int `json:"count"`
}

// 分词
type Token struct {
	EndOffset   int    `json:"end_offset"`
	Position    int    `json:"position"`
	StartOffset int    `json:"start_offset"`
	Token       string `json:"token"`
	Type        string `json:"type"`
}
type Tokens struct {
	Tokens []Token `json:"tokens"`
}

//搜索结果结构
type ElasticSearchResult struct {
	Shards struct {
		Failed     int `json:"failed"`
		Skipped    int `json:"skipped"`
		Successful int `json:"successful"`
		Total      int `json:"total"`
	} `json:"_shards"`
	Hits struct {
		Hits []struct {
			ID     string      `json:"_id"`
			Index  string      `json:"_index"`
			Score  interface{} `json:"_score"`
			Source struct {
				Id       int    `json:"id"`
				BookId   int    `json:"book_id"`
				Title    string `json:"title"`
				Keywords string `json:"keywords"`
				Content  string `json:"content"`
				Vcnt     int    `json:"vcnt"`
				Private  int    `json:"private"`
			} `json:"_source"`
			Type string `json:"_type"`
			Sort []int  `json:"sort"`
		} `json:"hits"`
		MaxScore interface{} `json:"max_score"`
		Total    int         `json:"total"`
	} `json:"hits"`
	TimedOut bool `json:"timed_out"`
	Took     int  `json:"took"`
}
