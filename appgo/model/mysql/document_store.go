package mysql

// Document Store，文档存储，将大内容分发到专门的数据表里面
type DocumentStore struct {
	DocumentId int    `gorm:"not null;primary_key"json:"documentId"` //文档id，对应Document中的document_id
	Markdown   string `gorm:"not null;type:longtext"json:"markdown"` //markdown内容
	Content    string `gorm:"not null;type:longtext"json:"content"`  //文本内容
}

func (DocumentStore) TableName() string {
	return "document_store"
}
