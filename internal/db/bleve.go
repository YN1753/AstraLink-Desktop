package db

import (
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve/v2"
)

func BleveInit(path string) (bleve.Index, error) {
	datapath := filepath.Join(path, "data", "index", "notes")

	// 确保目录存在
	if err := os.MkdirAll(datapath, 0755); err != nil {
		return nil, err
	}

	// 尝试打开已有索引
	index, err := bleve.Open(datapath)
	if err == nil {
		return index, nil
	}

	// 打开失败则创建新索引
	newIndex, newErr := bleve.New(datapath, nil)
	if newErr != nil {
		return nil, newErr
	}
	return newIndex, nil
}
