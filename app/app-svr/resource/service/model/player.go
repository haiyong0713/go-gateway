package model

type CustomizedPanel struct {
	ID           int32  `json:"id" gorm:"column:id"`
	Tids         string `json:"tids" gorm:"column:tids"`
	BtnImg       string `json:"btn_img" gorm:"column:btn_img"`
	BtnText      string `json:"btn_text" gorm:"column:btn_text"`
	TextColor    string `json:"text_color" gorm:"column:text_color"`
	Link         string `json:"link" gorm:"column:link"`
	Label        string `json:"label" gorm:"column:label"`
	DisplayStage string `json:"display_stage" gorm:"column:display_stage"`
	Operator     string `json:"operator" gorm:"column:operator"`
	Priority     int32  `json:"priority" gorm:"column:priority"`
}
