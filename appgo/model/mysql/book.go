package mysql

import (
	"strings"
	"time"

	"github.com/goecology/ecology/appgo/pkg/mus"

	"github.com/goecology/ecology/appgo/pkg/utils"
)

//定义书籍排序类型
type BookOrder string

const (
	OrderRecommend       BookOrder = "recommend"
	OrderPopular         BookOrder = "popular"          //热门
	OrderLatest          BookOrder = "latest"           //最新
	OrderNew             BookOrder = "new"              //最新
	OrderScore           BookOrder = "score"            //评分排序
	OrderComment         BookOrder = "comment"          //评论排序
	OrderStar            BookOrder = "star"             //收藏排序
	OrderView            BookOrder = "vcnt"             //浏览排序
	OrderLatestRecommend BookOrder = "latest-recommend" //最新推荐
)

// Book struct .
type Book struct {
	BookId            int       `gorm:"not null;primary_key;AUTO_INCREMENT"json:"bookId"`
	BookName          string    `gorm:"not null;"json:"bookName"` // BookName 项目名称.
	Identify          string    `gorm:"not null;"json:"identify"` // Identify 项目唯一标识.
	OrderIndex        int       `gorm:"not null;"json:"orderIndex"`
	Pin               int       `gorm:"not null;"json:"pin"`         // pin值，用于首页固定显示
	Description       string    `gorm:"not null;"json:"description"` // Description 项目描述.
	Label             string    `gorm:"not null;"json:"label"`
	PrivatelyOwned    int       `gorm:"not null;"json:"privatelyOwned"` // PrivatelyOwned 项目私有： 0 公开/ 1 私有
	PrivateToken      string    `gorm:"not null;"json:"privateToken"`   // 当项目是私有时的访问Token.
	Status            int       `gorm:"not null;"json:"status"`         //状态：0 正常/1 已删除
	Editor            string    `gorm:"not null;"json:"editor"`         //默认的编辑器.
	DocCount          int       `gorm:"not null;"json:"docCount"`       // DocCount 包含文档数量.
	CommentStatus     string    `gorm:"not null;"json:"commentStatus"`  // CommentStatus 评论设置的状态:open 为允许所有人评论，closed 为不允许评论, group_only 仅允许参与者评论 ,registered_only 仅允许注册者评论.
	CommentCount      int       `gorm:"not null;"json:"commentCount"`
	Cover             string    `gorm:"not null;"json:"cover"` //封面地址
	Theme             string    `gorm:"not null;"json:"theme"` //主题风格
	CreateTime        time.Time `gorm:""json:"createTime"`     // CreateTime 创建时间 .
	MemberId          int       `gorm:"not null;"json:"memberId"`
	ModifyTime        time.Time `gorm:""json:"modifyTime"`
	ReleaseTime       time.Time `gorm:""json:"releaseTime"`       //项目发布时间，每次发布都更新一次，如果文档更新时间小于发布时间，则文档不再执行发布
	GenerateTime      time.Time `gorm:""json:"generateTime"`      //下载文档生成时间
	LastClickGenerate time.Time `gorm:""json:"lastClickGenerate"` //上次点击上传文档的时间，用于显示频繁点击浪费服务器硬件资源的情况
	Version           int64     `gorm:"not null;"json:"version"`
	Vcnt              int       `gorm:"not null;"json:"vcnt"`       // 文档项目被阅读次数
	Star              int       `gorm:"not null;"json:"star"`       // 文档项目被收藏次数
	Score             int       `gorm:"not null;"json:"score"`      // 文档项目评分，默认40，即4.0星
	CntScore          int       `gorm:"not null;"json:"cntScore"`   // 评分人数
	CntComment        int       `gorm:"not null;"json:"cntComment"` // 评论人数
	Author            string    `gorm:"not null;"json:"author"`     //原作者，即来源
	AuthorURL         string    `gorm:"not null;"json:"authorUrl"`  //原作者链接，即来源链接
	AdTitle           string    `gorm:"not null;"json:"adTitle"`    // 文字广告标题
	AdLink            string    `gorm:"not null;"json:"adLink"`     // 文字广告链接
	Lang              string    `gorm:"not null;"json:"lang"`
	BookType          string    `gorm:"not null;"json:"bookType"` //  original 原创， opensource 开源
	Avatar            string    `gorm:"-"json:"avatar"`
	UserName          string    `gorm:"-"json:"userName"`
}

// TableName 获取对应数据库表名.
func (m Book) TableName() string {
	return "book"
}

func NewBook() *Book {
	return &Book{}
}

func (b *Book) DealCover() {
	b.Cover = mus.Oss.ShowImg(b.Cover)
}

func (b *Book) DealAll() {
	b.Cover = mus.Oss.ShowImg(b.Cover)
	var member Member
	mus.Db.Where("member_id = ?", b.MemberId).Find(&member)
	b.Avatar = member.Avatar
	b.UserName = member.Account
}

func (book *Book) ToBookResult() (m *BookResult) {
	m = &BookResult{}
	m.BookId = book.BookId
	m.BookName = book.BookName
	m.Identify = book.Identify
	m.OrderIndex = book.OrderIndex
	m.Description = strings.Replace(book.Description, "\r\n", "<br/>", -1)
	m.PrivatelyOwned = book.PrivatelyOwned
	m.PrivateToken = book.PrivateToken
	m.DocCount = book.DocCount
	m.CommentStatus = book.CommentStatus
	m.CommentCount = book.CommentCount
	m.CreateTime = book.CreateTime
	m.ModifyTime = book.ModifyTime
	m.Cover = book.Cover
	m.MemberId = book.MemberId
	m.Label = book.Label
	m.Status = book.Status
	m.Editor = book.Editor
	m.Theme = book.Theme
	m.Vcnt = book.Vcnt
	m.Star = book.Star
	m.Score = book.Score
	m.ScoreFloat = utils.ScoreFloat(book.Score)
	m.CntScore = book.CntScore
	m.CntComment = book.CntComment
	m.Author = book.Author
	m.AuthorURL = book.AuthorURL
	m.AdTitle = book.AdTitle
	m.AdLink = book.AdLink
	m.Lang = book.Lang

	if book.Theme == "" {
		m.Theme = "default"
	}

	if book.Editor == "" {
		m.Editor = "markdown"
	}
	return m
}
