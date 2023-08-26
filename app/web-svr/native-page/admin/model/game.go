package model

import (
	"fmt"
	"strings"
)

type GameItem struct {
	GameBaseId   int64    `json:"game_base_id"`  // 游戏唯一标识
	GameName     string   `json:"game_name"`     // 游戏名称
	GameIcon     string   `json:"game_icon"`     // 游戏icon
	GameTags     []string `json:"game_tags"`     // 游戏标签
	GameSubtitle string   `json:"game_subtitle"` // 小标题
	GameLink     string   `json:"game_link"`     // 游戏跳转链接
	GameButton   string   `json:"game_button"`   // 按钮文案 2种：预约，进入
	GameStatus   int      `json:"game_status"`   // 游戏状态：0 下载，1 预约（跳过详情），2 预约，3 测试，4 测试+预约，5 跳过详情页，6 仅展示， 7 社区，只有动态小卡行动点游戏信息接口会返回
}

type GameList struct {
	GameBaseId    int64  `json:"game_base_id"`    // 游戏唯一标识
	GameName      string `json:"game_name"`       // 游戏名称
	IsShowAndroid int    `json:"is_show_android"` //平台类型：安卓，0=下架，1=上架
	IsShowIos     int    `json:"is_show_ios"`     //平台类型：IOS，0=下架, 1=上架
	IsShowPc      int    `json:"is_show_pc"`      //平台类型：PC，0=下架，1=上架
}

type FormatItem struct {
	Param    string `json:"param"`
	Image    string `json:"image"`
	Title    string `json:"title"`
	URI      string `json:"uri"`
	Subtitle string `json:"subtitle"`
	Content  string `json:"content"`
}

func (i *FormatItem) FromGameExt(act *GameItem) {
	i.Param = fmt.Sprintf("%d", act.GameBaseId)
	i.Image = act.GameIcon
	i.Title = act.GameName
	i.URI = act.GameLink
	i.Subtitle = act.GameSubtitle
	i.Content = strings.Join(act.GameTags, "/")
}
