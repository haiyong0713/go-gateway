package threePointMeta

import (
	protoV2 "go-gateway/app/app-svr/app-card/interface/model/card/proto"
)

func ConstructPanelMeta(shareID, shareOrigin string) *protoV2.PanelMeta {
	return &protoV2.PanelMeta{
		PanelType:         1,
		ShareOrigin:       shareOrigin,
		ShareId:           shareID,
		FunctionalButtons: constructFunctionalButtons(),
	}
}

func constructFunctionalButtons() []*protoV2.FunctionalButton {
	out := make([]*protoV2.FunctionalButton, 0)
	out = append(out, &protoV2.FunctionalButton{
		Type: _typeWatchLater,
		ButtonMetas: []*protoV2.FunctionalButtonMeta{
			{
				Icon: _watchLaterIcon,
				Text: "稍后再看",
			},
		},
	}, &protoV2.FunctionalButton{
		Type: _typeCollect,
		ButtonMetas: []*protoV2.FunctionalButtonMeta{
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
	}, &protoV2.FunctionalButton{
		Type: _typeSpeedPlay,
		ButtonMetas: []*protoV2.FunctionalButtonMeta{
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
	out = append(out, &protoV2.FunctionalButton{
		Type: _typeAutoPlay,
		ButtonMetas: []*protoV2.FunctionalButtonMeta{
			{
				Icon: _autoPlayIcon,
				Text: "自动播放",
			},
		},
	})
	return out
}
