package cheese

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"

	cheeseEp "git.bilibili.co/bapis/bapis-go/cheese/service/season/episode"
	cheeseSeason "git.bilibili.co/bapis/bapis-go/cheese/service/season/season"

	"github.com/pkg/errors"
)

// Dao is cheese dao.
type Dao struct {
	c *conf.Config
	//grpc
	csClient cheeseSeason.SeasonClient
	ceClient cheeseEp.EpisodeClient
}

// New new a cheese dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	//grpc
	var err error
	if d.csClient, err = cheeseSeason.NewClient(c.CheeseGRPC); err != nil {
		panic(fmt.Sprintf("cheeseSeason.NewClientt error (%+v)", err))
	}
	if d.ceClient, err = cheeseEp.NewClient(c.CheeseGRPC); err != nil {
		panic(fmt.Sprintf("cheeseEp.NewClientt error (%+v)", err))
	}
	return
}

func (d *Dao) UserSeason(c context.Context, vmid int64, pn, ps int) (res []*cheeseSeason.SeasonCard, total int64, err error) {
	var (
		// NeedAll 0=下架、修复待审、修复打回，已发布 1=拉取全部 2=修复待审、修复打回，已发布
		req = &cheeseSeason.UserSeasonReq{Mid: vmid, Pn: int32(pn), Ps: int32(ps), NeedAll: 2}
		rep *cheeseSeason.UserSeasonReply
	)
	if rep, err = d.csClient.UserSeason(c, req); err != nil {
		err = errors.Wrapf(err, "%v", req)
		return
	}
	if rep != nil {
		res = rep.SeasonList
		total = rep.Total
	}
	return
}

func (d *Dao) EpCards(c context.Context, epids []int32) (res map[int32]*cheeseEp.EpisodeCard, err error) {
	var (
		req = &cheeseEp.EpisodeCardsReq{Ids: epids}
		rep *cheeseEp.EpisodeCardsReply
	)
	if rep, err = d.ceClient.Cards(c, req); err != nil {
		err = errors.Wrapf(err, "%v", req)
		return
	}
	if rep != nil {
		res = rep.Cards
	}
	return
}

func (d *Dao) HasCheese(plat int8, build int, all bool) (res bool) {
	// 历史记录课堂类型在phone和pad是不同的
	res = (plat == model.PlatAndroid && build > d.c.BuildLimit.AndroidCheese) ||
		(plat == model.PlatIPhone && build > d.c.BuildLimit.IOSCheese) ||
		(plat == model.PlatAndroidB && build >= d.c.BuildLimit.AndroidBCheese) ||
		(plat == model.PlatIPhoneB && build >= d.c.BuildLimit.IPhoneBCheese)
	if all {
		res = res || (plat == model.PlatIPad && build > d.c.BuildLimit.IPadCheese) ||
			(plat == model.PlatIpadHD && build > d.c.BuildLimit.IPadCheese) ||
			(plat == model.PlatAndroidHD && build > d.c.BuildLimit.AndroidHDCheese)
	}
	return res
}
