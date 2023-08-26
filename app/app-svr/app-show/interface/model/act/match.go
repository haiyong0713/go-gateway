package act

import (
	"strconv"

	"go-gateway/app/app-svr/app-show/interface/conf"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

func (i *Item) FromMatchMedalModule(mou *natpagegrpc.NativeModule, items []*Item) {
	i.Goto = GotoMatchMedalModule
	if mou == nil || len(items) == 0 {
		return
	}
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
	i.Item = items
	i.Color = &Color{BgColor: mou.BgColor}
}

func (i *Item) FromMatchEventModule(mou *natpagegrpc.NativeModule, items []*Item, cfg *conf.WinterOlyEvent) {
	i.Goto = GotoMatchEventModule
	if mou == nil || len(items) == 0 {
		return
	}
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
	i.Item = []*Item{
		{
			Goto:   GotoMatchEvent,
			ItemID: mou.Fid,
			Item:   items,
		},
	}
	i.Color = &Color{
		BgColor:      mou.BgColor,
		TitleColor:   cfg.TitleColor,
		TitleBgColor: cfg.TitleBgColor,
	}
}
