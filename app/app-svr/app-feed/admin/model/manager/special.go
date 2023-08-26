package manager

// SpecialCard .
type SpecialCard struct {
	ID       int64  `gorm:"column:id" json:"id"`
	Title    string `gorm:"column:title" json:"title"`
	Desc     string `gorm:"column:desc" json:"desc"`
	Cover    string `gorm:"column:cover" json:"cover"`
	ReType   int64  `gorm:"column:re_type" json:"re_type"`
	ReValue  string `gorm:"column:re_value" json:"re_value"`
	Person   string `gorm:"column:person" json:"person"`
	Corner   string `gorm:"column:corner" json:"corner"`
	Card     int64  `gorm:"column:card" json:"card"`
	Info     string `gorm:"column:info" json:"info"`
	Scover   string `gorm:"column:scover" json:"scover"`
	Size     string `gorm:"column:size" json:"size"`
	Gifcover string `gorm:"column:gifcover" json:"gifcover"`
}

// TableName .
func (a SpecialCard) TableName() string {
	return "special_card"
}
