package model

type Relation struct {
	FromID    string `gorm:"primaryKey;index" json:"fromId"`
	ToID      string `gorm:"primaryKey;index" json:"toId"`
	Type      string `gorm:"primaryKey;index" json:"type"` // contains, link, attach
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
}

func (Relation) TableName() string {
	return "relations"
}
