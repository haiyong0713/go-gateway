package model

import "fmt"

type TopArc struct {
	Mid             int64  `json:"mid" gorm:"column:mid"`
	Aid             int64  `json:"aid" gorm:"column:aid"`
	RecommendReason string `json:"recommend_reason" gorm:"column:recommend_reason"`
}

// TableName top arc
func (c *TopArc) TableName() string {
	return fmt.Sprintf("member_top%d", c.Mid%10)
}

type Masterpiece struct {
	ID              int64  `json:"id" gorm:"column:id"`
	Mid             int64  `json:"mid" gorm:"column:mid"`
	Aid             int64  `json:"aid" gorm:"column:aid"`
	RecommendReason string `json:"recommend_reason" gorm:"column:recommend_reason"`
}

// TableName masterpiece
func (c *Masterpiece) TableName() string {
	return fmt.Sprintf("member_masterpiece%d", c.Mid%10)
}
