package service

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/web/interface/model"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

func (s *Service) RegionIndex(ctx context.Context, plat int, lang, ip string, tinyMode, teenageMode int) ([]*model.Region, error) {
	if lang != model.LangHant {
		lang = model.LangHans
	}
	key := fmt.Sprintf("%v_%v", plat, lang)
	data, ok := s.regionList[key]
	if !ok {
		log.Error("日志告警 分区列表数据空,key:%v", key)
		return nil, ecode.NothingFound
	}
	var pids []int64
	exists := map[int64]struct{}{}
	for _, val := range data {
		pid, _ := strconv.ParseInt(val.Area, 10, 64)
		if pid == 0 {
			continue
		}
		if _, ok := exists[pid]; ok {
			continue
		}
		pids = append(pids, pid)
		exists[pid] = struct{}{}
	}
	auths, err := s.authPIDs(ctx, pids, ip)
	if err != nil {
		log.Error("%+v", err)
	}

	var res []*model.Region
	for _, val := range data {
		pid, _ := strconv.ParseInt(val.Area, 10, 64)
		if auth, ok := auths[pid]; ok && auth.GetPlay() == int64(locgrpc.Status_Forbidden) {
			log.Warn("分区地区屏蔽 region:%+v,ip:%v", val, ip)
			continue
		}
		r := &model.Region{}
		*r = *val
		r.Area = ""
		res = append(res, r)
	}
	res = filterTinyRegions(res, s.c.Rule.TinyPackageRegion, s.c.Rule.TeenageModeRegion, tinyMode, teenageMode)
	return res, nil
}

func filterTinyRegions(original []*model.Region, tinyRids, teenageRids []int64, tinyMode, teenageMode int) []*model.Region {
	var res []*model.Region
	_flagYes := 1
	if tinyMode == _flagYes {
		tag := "tiny"
		rids := tinyRids
		// 过滤极小包
		if teenageMode == _flagYes {
			// 青少年
			rids = teenageRids
			tag = "teenage"
		}
		ridMap := make(map[int64]int)
		for _, id := range rids {
			ridMap[id]++
		}
		for _, v := range original {
			if ridMap[v.Rid] > 0 {
				res = append(res, v)
			}
		}
		log.Info("分区%s屏蔽, res:%+v", tag, res)
		return res
	}
	return original
}

func (s *Service) loadRegionList() error {
	data, err := s.dao.RegionList(context.Background())
	if err != nil {
		return err
	}
	s.regionList = data
	return nil
}

func (s *Service) authPIDs(ctx context.Context, pids []int64, ip string) (map[int64]*locgrpc.Auth, error) {
	if len(pids) == 0 {
		return nil, nil
	}
	req := &locgrpc.AuthPIDsReq{
		Pids:   xstr.JoinInts(pids),
		IpAddr: ip,
	}
	res, err := s.locGRPC.AuthPIDs(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.GetAuths(), nil
}
