package mysql

import (
	"github.com/goecology/ecology/appgo/pkg/mus"
)

type Dashboard struct {
	BookNumber       int `json:"book_number"`
	DocumentNumber   int `json:"document_number"`
	MemberNumber     int `json:"member_number"`
	CommentNumber    int `json:"comment_number"`
	AttachmentNumber int `json:"attachment_number"`
}

func NewDashboard() *Dashboard {
	return &Dashboard{}
}

func (m *Dashboard) Query() *Dashboard {
	var number int
	mus.Db.Model(Book{}).Count(&number)
	m.BookNumber = number

	mus.Db.Model(Document{}).Count(&number)
	m.DocumentNumber = number

	mus.Db.Model(Member{}).Count(&number)
	m.MemberNumber = number

	m.CommentNumber = 0

	mus.Db.Model(Attachment{}).Count(&number)

	m.AttachmentNumber = number

	return m
}
