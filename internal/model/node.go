package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JSONMap map[string]any

func (m *JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("格式错误")
	}
	return json.Unmarshal(bytes, m)
}

type Node struct {
	// 1. 建议使用 ID (全大写)，并显式标注主键
	ID   string `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Type string `json:"type"`

	// 如果 Address 和 Path 经常变动或可选，保持这样很好
	Address string `json:"address"`
	Path    string `json:"path"`

	Others JSONMap `json:"others" gorm:"serializer:json;type:TEXT"`

	CreateTime int64 `json:"create_time" gorm:"autoCreateTime"`
	UpdateTime int64 `json:"update_time" gorm:"autoUpdateTime"`
}

type UpdateNoteInfoReq struct {
	Id     string  `json:"id" `
	Title  string  `json:"title"`
	Path   string  `json:"path"`
	Others JSONMap `json:"others"`
}

type NoteDoc struct {
	Id      string `json:"id" `
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (Node) TableName() string {
	return "nodes"
}

// TreeNode 用于目录树结构
type TreeNode struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Type     string      `json:"type"` // folder, note, user
	Children []*TreeNode `json:"children,omitempty"`
}
type D3Graph struct {
	Nodes []D3Node `json:"nodes"`
	Links []D3Link `json:"links"`
}
type D3Node struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}
type D3Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

// TagWithCount 标签及其关联的笔记数量
type TagWithCount struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	NoteCount int    `json:"noteCount"`
}

type TagMessage struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type NoteSearchResult struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
