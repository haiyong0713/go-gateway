package resolver

import (
	"context"
	"sort"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Vote struct{}

func (r Vote) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	confSort := natModule.ConfUnmarshal()
	cfg := &config.Vote{
		BaseCfgManager:  config.NewBaseCfg(natModule),
		BgImage:         config.SizeImage{Image: natModule.Meta, Height: natModule.Length, Width: natModule.Width},
		DisplayNum:      natModule.IsAttrDisplayNum() == natpagegrpc.AttrModuleYes,
		DoneButtonImage: confSort.Image,
		Sid:             natModule.Fid,
		Gid:             confSort.Sid,
		SourceType:      confSort.SourceType,
	}
	r.setVote(cfg, module.Click)
	r.setBaseCfg(cfg, ss)
	return cfg
}

func (r Vote) setVote(cfg *config.Vote, click *natpagegrpc.Click) {
	if click == nil {
		return
	}
	sort.Slice(click.Areas, func(i, j int) bool {
		return click.Areas[i].Leftx < click.Areas[j].Leftx
	})
	for _, area := range click.Areas {
		if area == nil {
			continue
		}
		areaExt := area.ExtUnmarshal()
		areaCfgBuilder := func() config.Area {
			return config.Area{Height: area.Length, Width: area.Width, X: area.Leftx, Y: area.Lefty, Ukey: areaExt.Ukey}
		}
		switch area.Type {
		case model.ClickTypeVoteButton:
			cfg.VoteButtons = append(cfg.VoteButtons, &config.VoteButton{
				Area:        areaCfgBuilder(),
				UndoneImage: area.UnfinishedImage,
			})
		case model.ClickTypeVoteProcess:
			cfg.VoteProgress = &config.VoteProgress{
				Area:  areaCfgBuilder(),
				Style: areaExt.Style,
			}
			for _, v := range areaExt.Items {
				cfg.VoteProgress.OptionColors = append(cfg.VoteProgress.OptionColors, v.BgColor)
			}
		case model.ClickTypeVoteUser:
			cfg.VoteLeftNum = &config.VoteNum{
				Area: areaCfgBuilder(),
			}
		default:
			log.Warn("unknown area.Type of Vote, type=%d", area.Type)
		}
	}
}

func (r Vote) setBaseCfg(cfg *config.Vote, ss *kernel.Session) {
	switch cfg.SourceType {
	case model.SourceTypeVoteAct, "":
		if cfg.Sid <= 0 || cfg.Gid <= 0 {
			return
		}
		cfg.VoteRankReqID, _ = cfg.AddMaterialParam(model.MaterialVoteRankRly, &activitygrpc.GetVoteActivityRankReq{
			ActivityId:    cfg.Sid,
			SourceGroupId: cfg.Gid,
			Pn:            1,
			Ps:            model.VoteOptionNum,
			Sort:          3,
			Mid:           ss.Mid(),
		})
	case model.SourceTypeVoteUp:
		if cfg.Sid <= 0 {
			return
		}
		_, _ = cfg.AddMaterialParam(model.MaterialDynVoteInfo, []int64{cfg.Sid})
	}
}
