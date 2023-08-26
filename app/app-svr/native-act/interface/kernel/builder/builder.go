package builder

import (
	"context"
	"reflect"

	"go-common/library/log"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type AfterContextData struct {
	NaviItems []*api.NavigationItem
}

type Builder interface {
	Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module
	After(data *AfterContextData, current *api.Module) bool
}

func unshiftTitleCard(items []*api.ModuleItem, imageTitle, textTitle, reqFrom string) []*api.ModuleItem {
	if !model.IsFromIndex(reqFrom) {
		return items
	}
	if imageTitle != "" {
		items = append([]*api.ModuleItem{card.NewImageTitle(imageTitle).Build()}, items...)
	} else if textTitle != "" {
		items = append([]*api.ModuleItem{card.NewTextTitle(textTitle).Build()}, items...)
	}
	return items
}

func StatString(number int64) string {
	if number == 0 {
		return "0"
	}
	return appcardmdl.Stat64String(number, "")
}

func logCfgAssertionError(cfg interface{}) {
	log.Error("Fail to build, config type is not %+v", reflect.TypeOf(cfg).Name())
}

const SubpageCurrSortKey = -1

func buildSubpageData(title string, sortList []*config.SortListItem, newParams func(sort int64) string) *api.SubpageData {
	if newParams == nil {
		return nil
	}
	data := &api.SubpageData{
		Title:  title,
		Params: newParams(SubpageCurrSortKey),
		Tabs:   make([]*api.SubpageTab, 0, len(sortList)),
	}
	for _, sort := range sortList {
		data.Tabs = append(data.Tabs, &api.SubpageTab{
			Name:   sort.SortName,
			Params: newParams(sort.SortType),
			Sort:   sortChange(sort.SortType),
		})
	}
	return data
}

func sortChange(st int64) api.SortCategory {
	switch st {
	case model.ActOrderRandom:
		return api.SortCategory_StRandom
	default:
		return api.SortCategory_StTypeDefault
	}
}
