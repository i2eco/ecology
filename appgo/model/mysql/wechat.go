package mysql

type Wechat struct {
	Id        int    `gorm:"not null;primary_key;AUTO_INCREMENT"json:"id"`
	MemberId  int    `gorm:"not null;"json:"memberId"` //绑定的用户id
	Openid    string `gorm:"not null;"json:"openid"`
	Unionid   string `gorm:"not null;"json:"unionid"`
	AvatarURL string `gorm:"not null;"json:"avatarUrl"`
	Nickname  string `gorm:"not null;"json:"nickname"`
	SessKey   string `gorm:"not null;"json:"sessKey"`
}

func (Wechat) TableName() string {
	return "wechat"
}
