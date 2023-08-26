package model

import "fmt"

type Channel struct {
	ID    int64  `json:"id" gorm:"column:id"`
	Mid   int64  `json:"mid" gorm:"column:mid"`
	Name  string `json:"name" gorm:"column:name"`
	Intro string `json:"intro" gorm:"column:intro"`
}

// TableName channel
func (c *Channel) TableName() string {
	return fmt.Sprintf("member_channel%d", c.Mid%10)
}
