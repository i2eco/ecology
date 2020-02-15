package dao

import (
	"bytes"
	"html/template"
	"strconv"

	"github.com/goecology/ecology/appgo/model/mysql"
	"github.com/goecology/ecology/appgo/pkg/mus"
)

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

//使用递归的方式获取指定ID的顶级ID
func getSelectedNode(array []*DocumentTree, parentId int) int {
	for _, item := range array {
		if _, ok := item.ParentId.(string); ok && item.DocumentId == parentId {
			return item.DocumentId
		} else if pid, ok := item.ParentId.(int); ok && item.DocumentId == parentId {
			if pid == parentId {
				return 0
			}
			return getSelectedNode(array, pid)
		}
	}
	return 0
}

func getDocumentTree(array []*DocumentTree, parentId int, selectedId int, selectedParentId int, buf *bytes.Buffer) {
	buf.WriteString("<ul>")

	for _, item := range array {
		pid := 0

		if p, ok := item.ParentId.(int); ok {
			pid = p
		}
		if pid == parentId {

			selected := ""
			if item.DocumentId == selectedId {
				selected = ` class="jstree-clicked"`
			}
			selectedLi := ""
			if item.DocumentId == selectedParentId {
				selectedLi = ` class="jstree-open"`
			}
			buf.WriteString("<li id=\"")
			buf.WriteString(strconv.Itoa(item.DocumentId))
			buf.WriteString("\"")
			buf.WriteString(selectedLi)
			buf.WriteString("><a href=\"")
			if item.Identify != "" {
				uri := "/api/web/read/" + item.BookIdentify + "/" + item.Identify
				buf.WriteString(uri)
			} else {
				uri := "/api/web/read/" + item.BookIdentify + "/" + strconv.Itoa(item.DocumentId)
				buf.WriteString(uri)
			}
			buf.WriteString("\" title=\"")
			buf.WriteString(template.HTMLEscapeString(item.DocumentName) + "\"")
			buf.WriteString(selected + ">")
			buf.WriteString(template.HTMLEscapeString(item.DocumentName) + "</a>")

			for _, sub := range array {
				if p, ok := sub.ParentId.(int); ok && p == item.DocumentId {
					getDocumentTree(array, p, selectedId, selectedParentId, buf)
					break
				}
			}
			buf.WriteString("</li>")
		}
	}
	buf.WriteString("</ul>")
}

func (m *document) CreateDocumentTreeForHtml(bookId, selectedId int) (string, error) {
	trees, err := m.FindDocumentTree(bookId, selectedId)
	if err != nil {
		return "", err
	}
	parentId := getSelectedNode(trees, selectedId)

	buf := bytes.NewBufferString("")

	getDocumentTree(trees, 0, selectedId, parentId, buf)

	return buf.String(), nil

}

//获取项目的文档树状结构
func (m *document) FindDocumentTree(bookId int, selectedId int, isEdit ...bool) ([]*DocumentTree, error) {
	trees := make([]*DocumentTree, 0)

	var docs []*mysql.Document

	err := mus.Db.Select("document_id, version, document_name, parent_id, identify").Where("book_id = ?", bookId).Order("order_sort asc,identify asc").Limit(5000).Find(&docs).Error

	if err != nil {
		return trees, err
	}
	book, _ := Book.Find(bookId)

	trees = make([]*DocumentTree, len(docs))

	for index, item := range docs {
		tree := &DocumentTree{}
		if selectedId > 0 {
			if selectedId == item.DocumentId {
				tree.State = &DocumentSelected{Selected: true, Opened: true}
			}
		} else {
			if index == 0 {
				tree.State = &DocumentSelected{Selected: true, Opened: true}
			}
		}

		tree.DocumentId = item.DocumentId
		tree.Identify = item.Identify
		tree.Version = item.Version
		tree.BookIdentify = book.Identify
		if item.ParentId > 0 {
			tree.ParentId = item.ParentId
		} else {
			tree.ParentId = "#"
		}
		idf := item.Identify
		if idf == "" {
			idf = strconv.Itoa(item.DocumentId)
		}
		if len(isEdit) > 0 && isEdit[0] == true {
			tree.DocumentName = item.DocumentName + "<small class='text-danger'>(" + idf + ")</small>"
		} else {
			tree.DocumentName = item.DocumentName
		}

		trees[index] = tree
	}

	return trees, nil
}

//使用递归
