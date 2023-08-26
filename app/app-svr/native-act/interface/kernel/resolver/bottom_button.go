package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type BottomButton struct{}

func (r BottomButton) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	if natPage.IsNewact() && ss.TabFrom == model.TabFromTopicLayer {
		return nil
	}
	return Click{}.Resolve(c, ss, natPage, module)
}
