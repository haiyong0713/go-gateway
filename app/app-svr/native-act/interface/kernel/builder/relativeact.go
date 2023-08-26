package builder

import (
	"context"
	"fmt"
	"strconv"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type Relativeact struct{}

func (bu Relativeact) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	raCfg, ok := cfg.(*config.Relativeact)
	if !ok {
		logCfgAssertionError(config.Relativeact{})
		return nil
	}
	items := bu.buildModuleItems(raCfg, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeRelact.String(),
		ModuleId:    raCfg.ModuleBase().ModuleID,
		ModuleColor: &api.Color{BgColor: raCfg.BgColor, CardTitleFontColor: raCfg.CardTitleFontColor},
		ModuleItems: items,
		ModuleUkey:  raCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu Relativeact) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Relativeact) buildModuleItems(cfg *config.Relativeact, ss *kernel.Session) []*api.ModuleItem {
	items := make([]*api.ModuleItem, 0, len(cfg.Acts))
	for _, act := range cfg.Acts {
		if act == nil {
			continue
		}
		cd := &api.RelativeactCard{
			Image: act.ShareImage,
			Title: act.Title,
			Desc:  act.ShareTitle,
		}
		cd.Uri = func() string {
			if !act.IsOnline() {
				return fmt.Sprintf("bilibili://pegasus/channel/%d?type=topic", act.ForeignID)
			}
			if act.SkipURL == "" {
				return fmt.Sprintf("bilibili://following/activity_landing/%d", act.ID)
			}
			return act.SkipURL
		}()
		items = append(items, &api.ModuleItem{
			CardType:   model.CardTypeRelact.String(),
			CardId:     strconv.FormatInt(act.ID, 10),
			CardDetail: &api.ModuleItem_RelativeactCard{RelativeactCard: cd},
		})
	}
	return unshiftTitleCard(items, cfg.ImageTitle, "", ss.ReqFrom)
}
