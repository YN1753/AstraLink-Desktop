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
