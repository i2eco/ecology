package mysql

type AttachmentResult struct {
	Attachment
	IsExist       bool
	BookName      string
	DocumentName  string
	FileShortSize string
	Account       string
	LocalHttpPath string
}

func NewAttachmentResult() *AttachmentResult {
	return &AttachmentResult{IsExist: false}
}
