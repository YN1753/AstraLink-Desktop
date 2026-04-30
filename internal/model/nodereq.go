package model

type MergeNodeReq struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	File       string  `json:"file"`
	ParentPath string  `json:"parentPath"`
	ParentID   string  `json:"parentId"` // 父节点ID，用于创建relation
	Others     JSONMap `json:"others"`
}
