package mysql

// Option struct .
type Option struct {
	OptionId    int    `gorm:"not null;primary_key;AUTO_INCREMENT"json:"optionId"`
	OptionTitle string `gorm:"not null;"json:"optionTitle"`
	OptionName  string `gorm:"not null;"json:"optionName"`
	OptionValue string `gorm:"not null;"json:"optionValue"`
	Remark      string `gorm:"not null;"json:"remark"`
}

// TableName 获取对应数据库表名.
func (m Option) TableName() string {
	return "option"
}
