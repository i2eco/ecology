package mysql

type DocumentTree struct {
	DocumentId   int               `json:"id"`
	DocumentName string            `json:"text"`
	ParentId     interface{}       `json:"parent"`
	Identify     string            `json:"identify"`
	BookIdentify string            `json:"-"`
	Version      int64             `json:"version"`
	State        *DocumentSelected `json:"state,omitempty"`
}
type DocumentSelected struct {
	Selected bool `json:"selected"`
	Opened   bool `json:"opened"`
}
