package search

import (
	managersearch "git.bilibili.co/bapis/bapis-go/ai/search/mgr/interface"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	esportsservice "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
)

type SportsMaterial struct {
	SportsEventMatches       map[int64]*esportsservice.SportsEventMatchItem
	Configs                  map[int64]*managersearch.EsportConfigInfo
	InlineFns                map[int64]func(i *Item)
	MatchVersusLiveEntryRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
}
