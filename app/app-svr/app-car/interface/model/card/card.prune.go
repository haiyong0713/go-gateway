package card

import (
	"encoding/json"
	"net/url"

	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	"go-gateway/app/app-svr/app-car/interface/model/dynamic"
	"go-gateway/app/app-svr/app-car/interface/model/popular"

	hisApi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	cardappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
)

func PGCFollowPrune(c *cardappgrpc.CardSeasonProto) *Prune {
	return &Prune{
		SeasonID: int64(c.SeasonId),
	}
}

func PGCModulePrune(c *bangumi.Module) *Prune {
	return &Prune{
		SeasonID: int64(c.SeasonID),
	}
}

func CursorPrune(c *hisApi.ModelResource) *Prune {
	return &Prune{
		Business: c.Business,
		Oid:      c.Oid,
		Epid:     c.Epid,
		Cid:      c.Cid,
	}
}

func PopularPrune(c *popular.PopularCard) *Prune {
	return &Prune{
		Goto: c.Type,
		ID:   c.Value,
	}
}

func GtPrune(gt string, value int64) *Prune {
	return &Prune{
		Goto: gt,
		ID:   value,
	}
}

func GtWebPrune(gt string, id, cid int64) *Prune {
	return &Prune{
		Goto:    gt,
		ID:      id,
		ChildID: cid,
	}
}

func DynamicPrune(c *dynamic.Dynamic) *Prune {
	return &Prune{
		DynamicID: c.DynamicID,
		Dtype:     c.Type,
		Drid:      c.Rid,
	}
}

func FromPGCFollow(value string) (*cardappgrpc.CardSeasonProto, bool) {
	res, ok := valueToPrune(value)
	if !ok {
		return nil, false
	}
	return &cardappgrpc.CardSeasonProto{SeasonId: int32(res.SeasonID)}, true
}

func FromPGCModule(value string) (*bangumi.Module, bool) {
	res, ok := valueToPrune(value)
	if !ok {
		return nil, false
	}
	return &bangumi.Module{SeasonID: int32(res.SeasonID)}, true
}

func FromCursor(value string) (*hisApi.ModelResource, bool) {
	res, ok := valueToPrune(value)
	if !ok {
		return nil, false
	}
	return &hisApi.ModelResource{
		Business: res.Business,
		Oid:      res.Oid,
		Epid:     res.Epid,
		Cid:      res.Cid,
	}, true
}

func FromPopular(value string) (*popular.PopularCard, bool) {
	res, ok := valueToPrune(value)
	if !ok {
		return nil, false
	}
	return &popular.PopularCard{Type: res.Goto, Value: res.ID}, true
}

func FromDynamicPrune(value string) (*dynamic.Dynamic, bool) {
	res, ok := valueToPrune(value)
	if !ok {
		return nil, false
	}
	return &dynamic.Dynamic{DynamicID: res.DynamicID, Type: res.Dtype, Rid: res.Drid}, true
}

func FromGtPrune(value string) (gt string, id, childID int64, isok bool) {
	res, ok := valueToPrune(value)
	if !ok {
		return "", 0, 0, false
	}
	return res.Goto, res.ID, res.ChildID, true
}

func FromGtPrunes(value string) ([]*Prune, bool) {
	res, ok := valueToPrunes(value)
	if !ok {
		return nil, false
	}
	return res, true
}

func FromGtPruneItem(value string) (*Prune, bool) {
	res, ok := valueToPrune(value)
	if !ok {
		return nil, false
	}
	return res, true
}

func valueToPrune(value string) (*Prune, bool) {
	var res *Prune
	unescape, err := url.QueryUnescape(value)
	if err != nil {
		return nil, false
	}
	if err := json.Unmarshal([]byte(unescape), &res); err != nil {
		return nil, false
	}
	if res == nil {
		return nil, false
	}
	return res, true
}

func valueToPrunes(value string) ([]*Prune, bool) {
	var res []*Prune
	unescape, err := url.QueryUnescape(value)
	if err != nil {
		return nil, false
	}
	if err := json.Unmarshal([]byte(unescape), &res); err != nil {
		return nil, false
	}
	if res == nil {
		return nil, false
	}
	return res, true
}

func (item *Prune) FromAiItemToDyn() *dynamic.Dynamic {
	d := &dynamic.Dynamic{
		Rid: item.ID,
	}
	switch item.Goto {
	case model.GotoAv:
		d.Type = dynamic.DynTypeVideo
	case model.GotoPGC:
		d.Type = dynamic.DynTypeBangumi
	}
	return d
}
