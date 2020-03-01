package mysql

import "time"

type Awesome struct {
	Id             int        `gorm:"not null;"json:"id"`
	Name           string     `gorm:"not null;UNIQUE_INDEX"json:"name"` // 唯一标识
	GitName        string     `gorm:"not null;"json:"gitName"`
	OwnerAvatarUrl string     `gorm:"not null;"json:"ownerAvatarUrl"`
	HtmlUrl        string     `gorm:"not null;"json:"htmlUrl"`
	GitDescription string     `gorm:"not null;"json:"gitDescription"`
	GitCreatedAt   time.Time  `gorm:""json:"gitCreatedAt"`
	GitUpdatedAt   time.Time  `gorm:""json:"gitUpdatedAt"`
	GitUrl         string     `gorm:"not null;"json:"gitUrl"`
	SshUrl         string     `gorm:"not null;"json:"sshUrl"`
	CloneUrl       string     `gorm:"not null;"json:"cloneUrl"`
	HomePage       string     `gorm:"not null;"json:"homePage"`
	StarCount      int        `gorm:"not null;"json:"starCount"`
	WatcherCount   int        `gorm:"not null;"json:"watcherCount"`
	Language       string     `gorm:"not null;"json:"language"`
	ForkCount      int        `gorm:"not null;"json:"forkCount"`
	LicenseKey     string     `gorm:"not null;"json:"licenseKey"`
	LicenseName    string     `gorm:"not null;"json:"licenseName"`
	LicenseUrl     string     `gorm:"not null;"json:"licenseUrl"`
	CreatedAt      time.Time  `gorm:""json:"createdAt"`
	UpdatedAt      time.Time  `gorm:""json:"updatedAt"`
	DeletedAt      *time.Time `gorm:""json:"deletedAt"`
	Desc           string     `gorm:"not null;"json:"desc"`
	LongDesc       string     `gorm:"not null;"json:"longDesc"`
	Version        int        `gorm:"not null;"json:"version"`
	ReadCount      int        `gorm:"not null;"json:"readCount"`
}

func (Awesome) TableName() string {
	return "awesome"
}
