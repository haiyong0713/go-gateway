package act

import (
	"context"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	egv2 "go-common/library/sync/errgroup.v2"

	"go-gateway/app/web-svr/native-page/interface/api"
)

// NatConfig .
func (d *Dao) NatConfig(c context.Context, arg *api.NatConfigReq) (*api.NatConfigReply, error) {
	return d.natRPC.NatConfig(c, arg)
}

// ModuleMixExt .
// nolint:gomnd
func (d *Dao) ModuleMixExt(c context.Context, arg *api.ModuleMixExtReq) (reply *api.ModuleMixExtReply, err error) {
	if arg.Ps > 100 {
		arg.Ps = 100
	}
	return d.natRPC.ModuleMixExt(c, arg)
}

// ModuleConfig .
func (d *Dao) ModuleConfig(c context.Context, arg *api.ModuleConfigReq) (req *api.ModuleConfigReply, err error) {
	return d.natRPC.ModuleConfig(c, arg)
}

// BaseConfig .
func (d *Dao) BaseConfig(c context.Context, arg *api.BaseConfigReq) (*api.BaseConfigReply, error) {
	return d.natRPC.BaseConfig(c, arg)
}

// ModuleMixExt .
func (d *Dao) ModuleMixExts(c context.Context, arg *api.ModuleMixExtsReq) (reply *api.ModuleMixExtsReply, err error) {
	return d.natRPC.ModuleMixExts(c, arg)
}

// NativePages .
func (d *Dao) NativePages(c context.Context, pids []int64) (map[int64]*api.NativePage, error) {
	var maxLimit = 100
	g := egv2.WithContext(c)
	mu := sync.Mutex{}
	rly := make(map[int64]*api.NativePage)
	for i := 0; i < len(pids); i += maxLimit {
		var pidGroups []int64
		if i+maxLimit <= len(pids) {
			pidGroups = pids[i : i+maxLimit]
		} else {
			pidGroups = pids[i:]
		}
		g.Go(func(ctx context.Context) error {
			tmpRes, err := d.natRPC.NativePages(ctx, &api.NativePagesReq{Pids: pidGroups})
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			if tmpRes == nil {
				return nil
			}
			mu.Lock()
			for key, value := range tmpRes.List {
				rly[key] = value
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return rly, nil
}

// NativePage .
func (d *Dao) NativePage(c context.Context, pid int64) (*api.NativePage, error) {
	rly, e := d.natRPC.NativePage(c, &api.NativePageReq{Pid: pid})
	if e != nil {
		return nil, e
	}
	if rly == nil || rly.Item == nil {
		return nil, ecode.NothingFound
	}
	return rly.Item, nil
}

func (d *Dao) NatProgressParams(c context.Context, pageID int64) ([]*api.ProgressParam, error) {
	reply, err := d.natRPC.GetNatProgressParams(c, &api.GetNatProgressParamsReq{PageID: pageID})
	if err != nil || reply == nil {
		log.Error("Fail to get natProgressParams, pageID=%+v error=%+v", pageID, err)
		return nil, err
	}
	return reply.List, nil
}

func (d *Dao) NativePageCards(c context.Context, req *api.NativePageCardsReq) (map[int64]*api.NativePageCard, error) {
	if len(req.Pids) == 0 {
		return map[int64]*api.NativePageCard{}, nil
	}
	var maxLimit = 100
	g := egv2.WithContext(c)
	mu := sync.Mutex{}
	rly := make(map[int64]*api.NativePageCard)
	pids := req.Pids
	for i := 0; i < len(pids); i += maxLimit {
		var pidGroups []int64
		if i+maxLimit <= len(pids) {
			pidGroups = pids[i : i+maxLimit]
		} else {
			pidGroups = pids[i:]
		}
		tmpReq := &api.NativePageCardsReq{
			Pids:     pidGroups,
			Device:   req.Device,
			MobiApp:  req.MobiApp,
			Build:    req.Build,
			Buvid:    req.Buvid,
			Platform: req.Platform,
		}
		g.Go(func(ctx context.Context) error {
			tmpRes, err := d.natRPC.NativePageCards(ctx, tmpReq)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			if tmpRes == nil {
				return nil
			}
			mu.Lock()
			for key, value := range tmpRes.List {
				rly[key] = value
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) NativeAllPages(c context.Context, pids []int64) (map[int64]*api.NativePage, error) {
	if len(pids) == 0 {
		return map[int64]*api.NativePage{}, nil
	}
	var maxLimit = 100
	g := egv2.WithContext(c)
	mu := sync.Mutex{}
	rly := make(map[int64]*api.NativePage)
	for i := 0; i < len(pids); i += maxLimit {
		var pidGroups []int64
		if i+maxLimit <= len(pids) {
			pidGroups = pids[i : i+maxLimit]
		} else {
			pidGroups = pids[i:]
		}
		g.Go(func(ctx context.Context) error {
			tmpRes, err := d.natRPC.NativeAllPages(ctx, &api.NativeAllPagesReq{Pids: pidGroups})
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			if tmpRes == nil {
				return nil
			}
			mu.Lock()
			for key, value := range tmpRes.List {
				rly[key] = value
			}
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return rly, nil
}

// NatTabModules .
func (d *Dao) NatTabModules(c context.Context, tabID int64) (*api.NatTabModulesReply, error) {
	return d.natRPC.NatTabModules(c, &api.NatTabModulesReq{TabID: tabID})
}

// NativePagesTab .
func (d *Dao) NativePagesTab(c context.Context, pids []int64, category int32) (map[int64]*api.PagesTab, error) {
	rly, e := d.natRPC.NativePagesTab(c, &api.NativePagesTabReq{Pids: pids, Category: category})
	if e != nil {
		log.Error("d.natRPC.NativePagesTab(%v,%d) error(%v)", pids, category, e)
		return nil, e
	}
	if rly != nil {
		return rly.List, nil
	}
	return make(map[int64]*api.PagesTab), nil
}
