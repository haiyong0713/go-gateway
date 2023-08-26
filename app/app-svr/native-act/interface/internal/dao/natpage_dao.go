package dao

import (
	"context"

	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type natpageDao struct {
	client natpagegrpc.NaPageClient
}

func (d *natpageDao) ModuleMixExts(c context.Context, req *natpagegrpc.ModuleMixExtsReq) (*natpagegrpc.ModuleMixExtsReply, error) {
	rly, err := d.client.ModuleMixExts(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *natpageDao) NatConfig(c context.Context, pageID, ps, offset int64, pType int32) (*natpagegrpc.NatConfigReply, error) {
	req := &natpagegrpc.NatConfigReq{
		Pid:    pageID,
		Offset: offset,
		Ps:     ps,
		PType:  pType,
	}
	rly, err := d.client.NatConfig(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *natpageDao) BaseConfig(c context.Context, pageID, ps, offset int64, pType int32) (*natpagegrpc.BaseConfigReply, error) {
	req := &natpagegrpc.BaseConfigReq{
		Pid:    pageID,
		Offset: offset,
		Ps:     ps,
		PType:  pType,
	}
	rly, err := d.client.BaseConfig(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *natpageDao) NativePage(c context.Context, pageID int64) (*natpagegrpc.NativePage, error) {
	rly, err := d.client.NativePage(c, &natpagegrpc.NativePageReq{Pid: pageID})
	if err != nil {
		return nil, err
	}
	return rly.Item, nil
}

func (d *natpageDao) ModuleConfig(c context.Context, moduleID, primaryID int64) (*natpagegrpc.ModuleConfigReply, error) {
	rly, err := d.client.ModuleConfig(c, &natpagegrpc.ModuleConfigReq{ModuleID: moduleID, PrimaryPageID: primaryID})
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *natpageDao) ModuleMixExt(c context.Context, req *natpagegrpc.ModuleMixExtReq) (*natpagegrpc.ModuleMixExtReply, error) {
	rly, err := d.client.ModuleMixExt(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *natpageDao) NativePageCards(c context.Context, req *natpagegrpc.NativePageCardsReq) (map[int64]*natpagegrpc.NativePageCard, error) {
	rly, err := d.client.NativePageCards(c, req)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return map[int64]*natpagegrpc.NativePageCard{}, nil
	}
	return rly.List, nil
}

func (d *natpageDao) NativeAllPages(c context.Context, req *natpagegrpc.NativeAllPagesReq) (map[int64]*natpagegrpc.NativePage, error) {
	rly, err := d.client.NativeAllPages(c, req)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return map[int64]*natpagegrpc.NativePage{}, nil
	}
	return rly.List, nil
}

func (d *natpageDao) NativePages(c context.Context, req *natpagegrpc.NativePagesReq) (*natpagegrpc.NativePagesReply, error) {
	rly, err := d.client.NativePages(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *natpageDao) GetNatProgressParams(c context.Context, pageID int64) ([]*natpagegrpc.ProgressParam, error) {
	req := &natpagegrpc.GetNatProgressParamsReq{PageID: pageID}
	rly, err := d.client.GetNatProgressParams(c, req)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return []*natpagegrpc.ProgressParam{}, nil
	}
	return rly.List, nil
}

func (d *natpageDao) NatTabModules(c context.Context, tabID int64) (*natpagegrpc.NatTabModulesReply, error) {
	req := &natpagegrpc.NatTabModulesReq{TabID: tabID}
	rly, err := d.client.NatTabModules(c, req)
	if err != nil {
		return nil, err
	}
	return rly, nil
}
