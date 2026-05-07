package model

type Relation struct {
	FromID    string `gorm:"uniqueIndex:idx_relation_identity" json:"fromId"`
	ToID      string `gorm:"uniqueIndex:idx_relation_identity" json:"toId"`
	Type      string `gorm:"uniqueIndex:idx_relation_identity" json:"type"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
}

func (Relation) TableName() string {
	return "relations"
}

type DeleteRelationReq struct {
	FromId string `json:"fromId"`
	ToId   string `json:"toId"`
	Type   string `json:"type"`
}
