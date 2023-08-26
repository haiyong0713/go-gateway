package builder

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

const (
	_partiLightImage   = "https://i0.hdslb.com/bfs/activity-plat/static/20210331/467746a96c68611c46194c29089d62f5/eD8wsjzfZc.png"
	_partiNightImage   = "https://i0.hdslb.com/bfs/activity-plat/static/20210331/467746a96c68611c46194c29089d62f5/frJ4deI8im.png"
	_partiDynamicImage = "https://i0.hdslb.com/bfs/activity-plat/static/4f3662116d8ab4ee084213142492fc16/0-uIlgov_w156_h156.png"
	_partiVideoImage   = "https://i0.hdslb.com/bfs/activity-plat/static/4f3662116d8ab4ee084213142492fc16/E-vXzW-~_w156_h156.png"
	_partiArticleImage = "https://i0.hdslb.com/bfs/activity-plat/static/4f3662116d8ab4ee084213142492fc16/50d91hpX_w156_h156.png"
)

type Participation struct{}

func (bu Participation) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	partCfg, ok := cfg.(*config.Participation)
	if !ok {
		logCfgAssertionError(config.Participation{})
		return nil
	}
	if len(partCfg.Items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeParticipation.String(),
		ModuleId:    cfg.ModuleBase().ModuleID,
		ModuleItems: bu.buildModuleItems(partCfg, material),
		ModuleUkey:  cfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Participation) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Participation) buildModuleItems(cfg *config.Participation, material *kernel.Material) []*api.ModuleItem {
	cd := &api.ParticipationCard{
		Image:         _partiLightImage,
		SelectedImage: _partiNightImage,
	}
	for _, v := range cfg.Items {
		switch v.Type {
		case natpagegrpc.PartDynamic:
			cd.Items = append(cd.Items, bu.buildDynamicItem(v))
		case natpagegrpc.PartVideo:
			cd.Items = append(cd.Items, bu.buildVideoItem(v, material))
		case natpagegrpc.PartArticle:
			cd.Items = append(cd.Items, bu.buildArticleItem(v))
		default:
			log.Warn("unknown participation_type=%+v", v.Type)
			continue
		}
	}
	item := &api.ModuleItem{
		CardType:   model.CardTypeParticipation.String(),
		CardId:     strconv.FormatInt(cfg.ModuleBase().ModuleID, 10),
		CardDetail: &api.ModuleItem_ParticipationCard{ParticipationCard: cd},
	}
	return []*api.ModuleItem{item}
}

func (bu Participation) buildDynamicItem(data *config.ParticipationItem) *api.ParticipationCardItem {
	var newTid string
	if data.NewTid > 0 {
		newTid = strconv.FormatInt(data.NewTid, 10)
	}
	return &api.ParticipationCardItem{
		Image: _partiDynamicImage,
		Uri:   fmt.Sprintf("bilibili://following/publish?topicV2ID=%s", newTid),
		Title: data.ButtonContent,
		Type:  model.ParticipationDynamic,
	}
}

func (bu Participation) buildVideoItem(data *config.ParticipationItem, material *kernel.Material) *api.ParticipationCardItem {
	from := "1"
	typ := model.ParticipationVideoShoot
	if data.UploadType == 0 {
		from = "0"
		typ = model.ParticipationVideoChoose
	}
	urlValues := url.Values{}
	urlValues.Set("copyright", "1")
	urlValues.Set("from", from)
	urlValues.Set("relation_from", "NAactivityb")
	urlValues.Set("is_new_ui", "1")
	if data.NewTid > 0 {
		urlValues.Set("topic_id", strconv.FormatInt(data.NewTid, 10))
	}
	if data.Sid > 0 {
		urlValues.Set("mission_id", strconv.FormatInt(data.Sid, 10))
		if sub, ok := material.ActSubProtos[data.Sid]; ok && sub != nil && sub.Subject != nil && sub.Protocol != nil {
			urlValues.Set("mission_name", sub.Protocol.Tags)
		}
	}
	uri := fmt.Sprintf("bilibili://uper/user_center/add_archive/?%s", urlValues.Encode())
	return &api.ParticipationCardItem{
		Image: _partiVideoImage,
		Uri:   uri,
		Title: data.ButtonContent,
		Type:  typ,
	}
}

func (bu Participation) buildArticleItem(data *config.ParticipationItem) *api.ParticipationCardItem {
	return &api.ParticipationCardItem{
		Image: _partiArticleImage,
		Uri:   "https://member.bilibili.com/article-text/mobile",
		Title: data.ButtonContent,
		Type:  model.ParticipationArticle,
	}
}
