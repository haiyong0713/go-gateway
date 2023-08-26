package config

import (
	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type BaseCfgManager interface {
	ModuleBase() *ModuleBase
	AddMaterialParam(matType model.MaterialType, data ...interface{}) (kernel.RequestID, error)
	MaterialParams() *kernel.MaterialLoader
}

type ModuleBase struct {
	ModuleID int64
	Category int64
	Rank     int64
	Ukey     string
	Bar      string //导航栏
}

type BaseCfg struct {
	moduleBase     *ModuleBase
	materialLoader *kernel.MaterialLoader
}

func NewBaseCfg(module *natpagegrpc.NativeModule) *BaseCfg {
	return &BaseCfg{
		moduleBase: &ModuleBase{
			ModuleID: module.ID,
			Category: module.Category,
			Rank:     module.Rank,
			Ukey:     module.Ukey,
			Bar:      module.Bar,
		},
		materialLoader: &kernel.MaterialLoader{},
	}
}

func (cc *BaseCfg) ModuleBase() *ModuleBase {
	return cc.moduleBase
}

func (cc *BaseCfg) AddMaterialParam(matType model.MaterialType, data ...interface{}) (kernel.RequestID, error) {
	reqID, err := cc.materialLoader.AddItem(matType, data...)
	if err != nil {
		log.Error("Fail to add MaterialParam of %s, error=%+v", matType, err)
		return "", err
	}
	return reqID, nil
}

func (cc *BaseCfg) MaterialParams() *kernel.MaterialLoader {
	return cc.materialLoader
}

type Area struct {
	Height int64  //区域高
	Width  int64  //区域宽
	X      int64  //区域偏移x
	Y      int64  //区域偏移y
	Ukey   string //区域唯一标识
}

func (a *Area) ToGrpcArea() *api.Area {
	if a == nil {
		return nil
	}
	return &api.Area{Height: a.Height, Width: a.Width, X: a.X, Y: a.Y, Ukey: a.Ukey}
}

type SizeImage = model.MixExtImage

type SortListItem struct {
	SortType int64
	SortName string
}
