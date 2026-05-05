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

type CreateNoteReq struct {
	Name       string  `json:"name"`
	File       string  `json:"file"`
	ParentID   string  `json:"parentId"`
	ParentPath string  `json:"parentPath"`
	Others     JSONMap `json:"others"`
}

type CreateTagReq struct {
	Name     string  `json:"name"`
	ParentID string  `json:"parentId"`
	Others   JSONMap `json:"others"`
}

type CreateLinkReq struct {
	FromID string `json:"fromId"`
	ToID   string `json:"toId"`
}

type ChangeHonorReq struct {
	Honor string `json:"honor"`
}
