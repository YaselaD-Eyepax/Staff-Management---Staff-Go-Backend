package models

type GlobalTag struct {
	ID  int    `gorm:"primaryKey"`
	Tag string `json:"tag"`
}

func (GlobalTag) TableName() string {
	return "global_tags"
}
