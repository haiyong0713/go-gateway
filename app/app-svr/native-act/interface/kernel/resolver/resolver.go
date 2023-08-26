package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Resolver interface {
	Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager
}

func actSortType(module *natpagegrpc.Module, ss *kernel.Session) int64 {
	if model.IsFromIndex(ss.ReqFrom) {
		if module.VideoAct == nil || len(module.VideoAct.SortList) == 0 {
			return 0
		}
		return module.VideoAct.SortList[0].SortType
	}
	return ss.SortType
}

func actSortList(module *natpagegrpc.Module) []*config.SortListItem {
	if module.VideoAct == nil || len(module.VideoAct.SortList) == 0 {
		return nil
	}
	list := make([]*config.SortListItem, 0, len(module.VideoAct.SortList))
	for _, ext := range module.VideoAct.SortList {
		if ext == nil {
			continue
		}
		var name string
		if ext.Category == 1 {
			name = ext.SortName
		} else {
			switch ext.SortType {
			case model.ActOrderTime:
				name = "时间"
			case model.ActOrderRandom:
				name = "随机"
			case model.ActOrderHot:
				name = "热度"
			default:
				name = "分数"
			}
		}
		list = append(list, &config.SortListItem{
			SortType: ext.SortType,
			SortName: name,
		})
	}
	return list
}
