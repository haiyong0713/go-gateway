package threePointMeta

import (
	"go-gateway/app/app-svr/app-card/interface/model"
)

const (
	_typeNotInterested = 1
	_typeWatchLater    = 2
	_typeCollect       = 3
	_typeSpeedPlay     = 4
	_typeAutoPlay      = 5
	_typeSwitchColumn  = 6

	_notInterestedIcon       = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/lJVNJwZCfW.png"
	_watchLaterIcon          = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/NyPAqcn0QF.png"
	_collectIcon             = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/1gIJ91DqKx.png"
	_collectedIcon           = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/MqCrPs0cNW.png"
	_speedPlay50percentIcon  = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/TVoxpmCjvd.png"
	_speedPlay75percentIcon  = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/gRhcAhjuAN.png"
	_speedPlay100percentIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/3cjWGnLzWt.png"
	_speedPlay125percentIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/qlwZFJfB9L.png"
	_speedPlay150percentIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/G9vdWrFHmt.png"
	_speedPlay200percentIcon = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/shm60E8ATG.png"
	_autoPlayIcon            = "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/fPIoe4K0dA.png"
	_switchColumnSingleIcon  = "https://i0.hdslb.com/bfs/activity-plat/static/20210527/0977767b2e79d8ad0a36a731068a83d7/HyW0YhNAE8.png"
	_switchColumnDoubleIcon  = "https://i0.hdslb.com/bfs/activity-plat/static/20210526/0977767b2e79d8ad0a36a731068a83d7/ebAygiMdLC.png"
)

type PanelMeta struct {
	PanelType         int8                `json:"panel_type,omitempty"`
	ShareOrigin       string              `json:"share_origin,omitempty"`
	ShareId           string              `json:"share_id,omitempty"`
	FunctionalButtons []*FunctionalButton `json:"functional_buttons,omitempty"`
}

type FunctionalButton struct {
	Type        int32                   `json:"type,omitempty"`
	ButtonMetas []*FunctionalButtonMeta `json:"button_metas,omitempty"`
}

type FunctionalButtonMeta struct {
	Icon         string `json:"icon,omitempty"`
	Text         string `json:"text,omitempty"`
	ButtonStatus string `json:"button_status,omitempty"`
	Toast        string `json:"toast,omitempty"`
}

func ConstructFunctionalButton(isSimplePanel bool, needSwitchColumnThreePoint bool, column model.ColumnStatus, replaceDislikeTitle bool) []*FunctionalButton {
	out := make([]*FunctionalButton, 0)
	out = append(out, &FunctionalButton{
		Type:        _typeNotInterested,
		ButtonMetas: []*FunctionalButtonMeta{{Icon: _notInterestedIcon, Text: dislikeTitle(replaceDislikeTitle)}},
	})
	if !isSimplePanel {
		out = append(out, &FunctionalButton{
			Type: _typeWatchLater,
			ButtonMetas: []*FunctionalButtonMeta{
				{
					Icon: _watchLaterIcon,
					Text: "稍后再看",
				},
			},
		}, &FunctionalButton{
			Type: _typeCollect,
			ButtonMetas: []*FunctionalButtonMeta{
				{
					Icon:         _collectIcon,
					Text:         "收藏",
					ButtonStatus: "collect",
				},
				{
					Icon:         _collectedIcon,
					Text:         "收藏",
					ButtonStatus: "collected",
				},
			},
		}, &FunctionalButton{
			Type: _typeSpeedPlay,
			ButtonMetas: []*FunctionalButtonMeta{
				{
					Icon:         _speedPlay50percentIcon,
					Text:         "倍速播放",
					ButtonStatus: "0.5x",
				},
				{
					Icon:         _speedPlay75percentIcon,
					Text:         "倍速播放",
					ButtonStatus: "0.75x",
				},
				{
					Icon:         _speedPlay100percentIcon,
					Text:         "倍速播放",
					ButtonStatus: "1.0x",
				},
				{
					Icon:         _speedPlay125percentIcon,
					Text:         "倍速播放",
					ButtonStatus: "1.25x",
				},
				{
					Icon:         _speedPlay150percentIcon,
					Text:         "倍速播放",
					ButtonStatus: "1.5x",
				},
				{
					Icon:         _speedPlay200percentIcon,
					Text:         "倍速播放",
					ButtonStatus: "2.0x",
				},
			},
		})
	}
	if needSwitchColumnThreePoint {
		out = append(out, &FunctionalButton{
			Type:        _typeSwitchColumn,
			ButtonMetas: constructColumnButtonMetas(column),
		})
	}
	out = append(out, &FunctionalButton{
		Type: _typeAutoPlay,
		ButtonMetas: []*FunctionalButtonMeta{
			{
				Icon: _autoPlayIcon,
				Text: "自动播放",
			},
		},
	})
	return out
}

func dislikeTitle(replaceDislikeTitle bool) string {
	if replaceDislikeTitle {
		return "我不想看"
	}
	return "不感兴趣"
}

func constructColumnButtonMetas(column model.ColumnStatus) []*FunctionalButtonMeta {
	text := "切换至单列"
	toast := "已成功切换至单列模式"
	buttonStatus := "single"
	icon := _switchColumnSingleIcon
	if column == model.ColumnSvrSingle || column == model.ColumnUserSingle {
		text = "切换至双列"
		buttonStatus = "double"
		toast = "已成功切换至双列模式"
		icon = _switchColumnDoubleIcon
	}
	return []*FunctionalButtonMeta{
		{
			Icon:         icon,
			Text:         text,
			ButtonStatus: buttonStatus,
			Toast:        toast,
		},
	}
}

func VerticalUGCThreePoint(text *ThreePointMetaText) *PanelMeta {
	const (
		_shareOrigin = "ugc"
		_shareID     = "main.composite-tab.0.0.pv"
	)
	return &PanelMeta{
		PanelType:   1,
		ShareOrigin: _shareOrigin,
		ShareId:     _shareID,
		FunctionalButtons: []*FunctionalButton{
			{
				Type: _typeWatchLater,
				ButtonMetas: []*FunctionalButtonMeta{
					{
						Icon: _watchLaterIcon,
						Text: text.WatchLater,
					},
				},
			},
		},
	}
}

func VerticalOGVThreePoint(_ *ThreePointMetaText) *PanelMeta {
	const (
		_shareOrigin = "ogv"
		_shareID     = "main.composite-tab.0.0.pv"
	)
	return &PanelMeta{
		PanelType:   1,
		ShareOrigin: _shareOrigin,
		ShareId:     _shareID,
		//FunctionalButtons: []*FunctionalButton{
		//	{
		//		Type: _typeWatchLater,
		//		ButtonMetas: []*FunctionalButtonMeta{
		//			{
		//				Icon: _watchLaterIcon,
		//				Text: text.WatchLater,
		//			},
		//		},
		//	},
		//},
	}
}

type ThreePointMetaText struct {
	WatchLater string
}
