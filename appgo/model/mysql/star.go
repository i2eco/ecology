package mysql

type DataCount struct {
	Cnt int64
}

type Star struct {
	Id       int `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"`
	Uid      int `gorm:"not null;index"json:"uid"` //用户id,user id
	Bid      int `gorm:"not null;"json:"bid"`      //书籍id,book id
	LastRead int `gorm:"not null;"json:"lastRead"` //最后阅读书剑
}

func (Star) TableName() string {
	return "star"
}

// 多字段唯一键
func (this *Star) TableUnique() [][]string {
	return [][]string{
		[]string{"Uid", "Bid"},
	}
}

type StarResult struct {
	BookId      int    `json:"book_id"`
	BookName    string `json:"book_name"`
	Identify    string `json:"identify"`
	Description string `json:"description"`
	DocCount    int    `json:"doc_count"`
	Cover       string `json:"cover"`
	MemberId    int    `json:"member_id"`
	Nickname    string `json:"user_name"`
	Vcnt        int    `json:"vcnt"`
	Star        int    `json:"star"`
	Score       int    `json:"score"`
	CntComment  int    `json:"cnt_comment"`
	CntScore    int    `json:"cnt_score"`
	ScoreFloat  string `json:"score_float"`
	OrderIndex  int    `json:"order_index"`
}
