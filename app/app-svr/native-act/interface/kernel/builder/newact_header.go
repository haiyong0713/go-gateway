package builder

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type NewactHeader struct{}

func (bu NewactHeader) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	nhCfg, ok := cfg.(*config.NewactHeader)
	if !ok {
		logCfgAssertionError(config.NewactHeader{})
		return nil
	}
	items := bu.buildModuleItems(nhCfg, material)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeNewactHeader.String(),
		ModuleId:    nhCfg.ModuleBase().ModuleID,
		ModuleItems: items,
		ModuleUkey:  nhCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu NewactHeader) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu NewactHeader) buildModuleItems(cfg *config.NewactHeader, material *kernel.Material) []*api.ModuleItem {
	st, ok := material.ActSubjects[cfg.ReqID][cfg.Sid]
	if !ok {
		return nil
	}
	cd := &api.NewactHeader{
		Title:        st.Name,
		Time:         fmt.Sprintf("任务时间：%s-%s", st.Stime.Time().Format("2006.01.02"), st.Etime.Time().Format("2006.01.02")),
		Image:        st.ActivityImage,
		SponsorTitle: "发起",
	}
	if user, ok := material.AccountCards[st.ActivityInitiator]; ok {
		cd.Mid = user.Mid
		cd.UserName = user.Name
		cd.UserFace = user.Face
		cd.UserUrl = fmt.Sprintf("bilibili://space/%d?defaultTab=dynamic", user.Mid)
	}
	if time.Now().After(st.Etime.Time()) {
		cd.Features = append(cd.Features, &api.NewactFeature{Name: "已结束", BorderColor: "#9499A0"})
	}
	for _, ft := range st.ActFeature {
		if ft == nil {
			continue
		}
		cd.Features = append(cd.Features, &api.NewactFeature{Name: ft.Title, BorderColor: "#FF7F24"})
	}
	moduleItem := &api.ModuleItem{
		CardType:   model.CardTypeNewactHeader.String(),
		CardId:     strconv.FormatInt(cfg.Sid, 10),
		CardDetail: &api.ModuleItem_NewactHeaderCard{NewactHeaderCard: cd},
	}
	return []*api.ModuleItem{moduleItem}
}
