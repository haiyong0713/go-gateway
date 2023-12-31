package operate

import (
	"strconv"

	"go-gateway/app/app-svr/app-card/interface/model"
)

type Special struct {
	ID          int64  `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Desc        string `json:"desc,omitempty"`
	Cover       string `json:"cover,omitempty"`
	SingleCover string `json:"single_cover,omitempty"`
	GifCover    string `json:"gif_cover,omitempty"`
	ReType      int    `json:"re_type,omitempty"`
	ReValue     string `json:"re_value,omitempty"`
	Badge       string `json:"badge,omitempty"`
	Size        string `json:"size,omitempty"`
	// extra
	Ratio int      `json:"ratio,omitempty"`
	Goto  model.Gt `json:"goto,omitempty"`
	Param string   `json:"param,omitempty"`
	Pid   int64    `json:"pid,omitempty"`
	// channel
	BgCover        string  `json:"bg_cover,omitempty"`
	Reason         string  `json:"reason,omitempty"`
	TabURI         string  `json:"tab_uri,omitempty"`
	PowerPicSun    string  `json:"power_pic_sun,omitempty"`
	PowerPicNight  string  `json:"power_pic_night,omitempty"`
	PowerPicWidth  float64 `json:"power_pic_width,omitempty"`
	PowerPicHeight float64 `json:"power_pic_height,omitempty"`
}

func (c *Special) Change() {
	if c.SingleCover == "" {
		c.SingleCover = c.Cover
	}
	if c.Size == "1020x300" {
		c.Ratio = 34
	} else if c.Size == "1020x378" {
		c.Ratio = 24
	}
	c.Goto = model.OperateType[c.ReType]
	c.Param = c.ReValue
	c.Pid, _ = strconv.ParseInt(c.Param, 10, 64)
}
