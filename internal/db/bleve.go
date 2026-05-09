package db

import (
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
)

func BleveInit(path string) (bleve.Index, error) {
	datapath := filepath.Join(path, "index", "notes")

	// 确保目录存在
	if err := os.MkdirAll(datapath, 0755); err != nil {
		return nil, err
	}

	// 尝试打开已有索引
	index, err := bleve.Open(datapath)
	if err == nil {
		return index, nil
	}

	// 打开失败则创建新索引，配置文档映射以正确存储字段并支持中文
	mapping := bleve.NewIndexMapping()

	// 创建文档类型映射
	docMapping := bleve.NewDocumentMapping()

	// ID 字段 - 存储但不分词
	idMapping := bleve.NewKeywordFieldMapping()
	idMapping.Name = "id"
	idMapping.Store = true
	docMapping.AddFieldMapping(idMapping)

	// Title 字段 - 标准分析器
	titleMapping := bleve.NewTextFieldMapping()
	titleMapping.Name = "title"
	titleMapping.Store = true
	docMapping.AddFieldMapping(titleMapping)

	// Content 字段 - 标准分析器
	contentMapping := bleve.NewTextFieldMapping()
	contentMapping.Name = "content"
	contentMapping.Store = true
	docMapping.AddFieldMapping(contentMapping)

	mapping.AddDocumentMapping("note", docMapping)

	newIndex, newErr := bleve.New(datapath, mapping)
	if newErr != nil {
		// 如果创建也失败，尝试删除目录后重试
		os.RemoveAll(datapath)
		os.MkdirAll(datapath, 0755)
		newIndex, newErr = bleve.New(datapath, mapping)
		if newErr != nil {
			return nil, newErr
		}
	}
	return newIndex, nil
}