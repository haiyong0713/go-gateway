package frontpage

import xtime "go-common/library/time"

var CategoriesMap = map[string]string{
	"fp-global":      "全局",
	"fp-index":       "首页",
	"fp-animation":   "动画",
	"fp-drama":       "番剧",
	"fp-music":       "音乐",
	"fp-dance":       "舞蹈",
	"fp-game":        "游戏",
	"fp-tech":        "知识",
	"fp-life":        "生活",
	"fp-kichiku":     "鬼畜",
	"fp-fashion":     "时尚",
	"fp-play":        "娱乐",
	"fp-film":        "放映厅",
	"fp-comic":       "国创",
	"fp-newtv":       "新影视",
	"fp-digital":     "科技",
	"fp-food":        "美食",
	"fp-animal":      "动物圈",
	"fp-motor":       "汽车",
	"fp-sport":       "运动",
	"fp-information": "资讯",
}

// Menu resource table
type Menu struct {
	ID       int64      `json:"id" gorm:"column:id"`
	Platform int        `json:"platform" gorm:"column:platform"`
	Name     string     `json:"name" gorm:"column:name"`
	Parent   int64      `json:"parent" gorm:"column:parent"`
	State    int        `json:"-" gorm:"column:state"`
	Counter  int        `json:"counter" gorm:"column:counter"`
	Position int        `json:"position" gorm:"column:position"`
	Rule     string     `json:"rule" gorm:"column:rule"`
	Size     string     `json:"size" gorm:"column:size"`
	Preview  string     `json:"preview" gorm:"column:preview"`
	Desc     string     `json:"description" gorm:"column:description"`
	Mark     string     `json:"mark" gorm:"column:mark"`
	CTime    xtime.Time `json:"ctime" gorm:"column:ctime"`
	MTime    xtime.Time `json:"mtime" gorm:"column:mtime"`
	Level    int64      `json:"level" gorm:"column:level"`
	Type     int        `json:"type" gorm:"column:type"`
	IsAd     int        `json:"is_ad" gorm:"column:is_ad"`
}

func (t *Menu) TableName() string {
	return "resource"
}
