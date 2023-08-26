package show

import xtime "go-common/library/time"

const (
	MenuLimitType = 0
	SideType      = 1 //固定导航
	MenuType      = 0 //运营导航模块
	//attribute bit
	AttrYes = int64(1)
	// attribute bit
	AttrBitImage          = uint(0)
	AttrBitColor          = uint(1)
	AttrBitBgImage        = uint(2)
	AttrBitFollowBusiness = uint(3)
)

// MenuTabExt .
type MenuExt struct {
	*TabExt
	Limit []*TabLimit
}

// MenuTabExt .
type TabExt struct {
	ID             int64      `json:"id"`
	Type           int64      `json:"type"`
	TabID          int64      `json:"tab_id"`
	Attribute      int64      `json:"attribute"`
	InactiveIcon   string     `json:"inactive_icon"`
	Inactive       int64      `json:"inactive"`
	InactiveType   int64      `json:"inactive_type"`
	ActiveIcon     string     `json:"active_icon"`
	Active         int64      `json:"active"`
	ActiveType     int64      `json:"active_type"`
	TabTopColor    string     `json:"tab_top_color"`
	TabMiddleColor string     `json:"tab_middle_color"`
	TabBottomColor string     `json:"tab_bottom_color"`
	BgImage1       string     `json:"bg_image1"`
	BgImage2       string     `json:"bg_image2"`
	FontColor      string     `json:"font_color"`
	BarColor       int64      `json:"bar_color"`
	State          int64      `json:"state"`
	Stime          xtime.Time `json:"stime"`
	Etime          xtime.Time `json:"etime"`
	Ver            string     `json:"ver"`
}

// AttrVal get attr val by bit.
func (a *TabExt) AttrVal(bit uint) int64 {
	return (a.Attribute >> bit) & int64(1)
}

// MenuTabLimit .
type TabLimit struct {
	ID         int64  `json:"id"`
	Type       int64  `json:"type"`
	TID        int64  `json:"t_id"`
	Plat       int64  `json:"plat"`
	Build      int64  `json:"build"`
	Conditions string `json:"conditions"`
	State      int64  `json:"state"`
}
