package dynamic

import (
	"strconv"

	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

func (i *Item) FromNewactHeaderModule(mou *natpagegrpc.NativeModule, items []*Item) {
	i.Goto = GotoNewactHeaderModule
	if mou == nil || len(items) == 0 {
		return
	}
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
	i.Item = items
}

func (i *Item) FromNewactAwardModule(mou *natpagegrpc.NativeModule, items []*Item) {
	i.Goto = GotoNewactAwardModule
	if mou == nil || len(items) == 0 {
		return
	}
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
	i.Item = items
}

func (i *Item) FromNewactStatementModule(mou *natpagegrpc.NativeModule, items []*Item) {
	i.Goto = GotoNewactStatementModule
	if mou == nil || len(items) == 0 {
		return
	}
	i.ItemID = mou.ID
	i.Param = strconv.FormatInt(mou.ID, 10)
	i.Ukey = mou.Ukey
	i.Bar = mou.Bar
	i.Item = items
}
