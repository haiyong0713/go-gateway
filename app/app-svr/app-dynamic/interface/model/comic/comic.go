package comic

const (
	_inSerial = 0
	_end      = 1
)

// Comic get from comit.
type Comic struct {
	ID            int64    `json:"id"`
	Title         string   `json:"title"`
	Author        []string `json:"author"`
	Evaluate      string   `json:"evaluate"`       //漫画简介
	VerticalCover string   `json:"vertical_cover"` // 竖版封面
	IsFinish      int8     `json:"is_finish"`      // 完结状态 1:完结 0:连载 -1:未开刊
	Styles        []*struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"styles"` // 风格标签
	Total          int    `json:"total"`            // 总话数（全x话）
	LastShortTitle string `json:"last_short_title"` // 最新话短标题
	LastUpdateTime string `json:"last_update_time"` // 最新话更新时间: 秒级时间戳, 当更新时间不存在时, 最新话更新时间为0
	URL            string `json:"url"`              // h5 跳转url
	PcURL          string `json:"pc_url"`           // pc 跳转链接
	FavStatus      int    `json:"fav_status"`       // 用户是否追漫，0 未追；1 已追
}

// info http://comic.bilibili.co/api-doc/sniper/dynamic/comic/v0/dynamic.html#comicv0dynamicbatchgetinfo
type Batch struct {
	ID           int64  `json:"id"`
	Face         string `json:"face"`
	Cover        string `json:"cover"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	Area         string `json:"area"`
	Style        string `json:"style"`
	IsFinish     int    `json:"is_finish"`
	UpdateFreq   string `json:"update_freq"`
	PayMode      int    `json:"pay_mode"`
	JumpURL      string `json:"jump_url"`
	InteractArea string `json:"interact_area"`
}

func (b *Batch) FromBatchFinish() string {
	switch b.IsFinish {
	case _inSerial:
		return "连载中"
	case _end:
		return "完结"
	}
	return "未开刊"
}

func (b *Batch) FromBatchPay() string {
	const (
		pay       = 1
		payVolume = 2
	)
	switch b.PayMode {
	case pay:
		return "付费漫画"
	case payVolume:
		return "按卷付费"
	}
	return "漫画"
}
