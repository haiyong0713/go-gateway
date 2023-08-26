package dynamicV2

import api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"

type FoldList struct {
	List []*FoldItem
}

type FoldItem struct {
	Item *api.DynamicItem
}

type CtrlSort []*Ctrl

func (t CtrlSort) Len() int {
	return len(t)
}

func (t CtrlSort) Less(i, j int) bool {
	return t[i].Location < t[j].Location
}

func (t CtrlSort) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type FoldResItem struct {
	DynID     int64
	FoldType  api.FoldType
	Group     int
	Statement string
}
