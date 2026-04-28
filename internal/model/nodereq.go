package model

type MergeNodeReq struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	File       string  `json:"file"`
	ParentPath string  `json:"parentPath"`
	Others     JSONMap `json:"others"`
}
