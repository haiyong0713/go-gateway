package icon

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/dao/show"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/model/icon"
	"go-gateway/app/app-svr/app-feed/admin/model/menu"
	smdl "go-gateway/app/app-svr/app-feed/admin/model/show"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

// Service is menu service
type Service struct {
	showDao *show.Dao
}

// New new a search service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		showDao: show.New(c),
	}
	return
}

// IconList is
func (s *Service) IconList(c context.Context, arg *icon.ListParam) (*icon.ListReply, error) {
	ics, total, err := s.showDao.Icons(c, arg.Pn, arg.Ps)
	if err != nil {
		log.Error("s.showDao.Icons arg(%+v) err(%+v)", arg, err)
		return nil, err
	}
	if total == 0 || len(ics) == 0 {
		return nil, nil
	}
	var (
		oids          []int64
		sideBars      map[int64]*menu.Sidebar
		sideBarLimits map[int64][]*smdl.SidebarLimit
		icsInfo       []*icon.IconInfo
	)
	for _, v := range ics {
		var m []*icon.Module
		var mi []*icon.ModuleInfo
		if err := json.Unmarshal([]byte(v.Module), &m); err != nil {
			continue
		}
		for _, v := range m {
			oids = append(oids, v.Oid)
			mi = append(mi, &icon.ModuleInfo{Plat: v.Plat, Oid: v.Oid})
		}
		icsInfo = append(icsInfo, &icon.IconInfo{Icon: v, ModuleInfo: mi})
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		sideBars, err = s.showDao.SideBars(c, oids)
		return err
	})
	eg.Go(func(ctx context.Context) (e error) {
		sideBarLimits, err = s.showDao.SideBarLimits(c, oids)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Error("IconList eg.Wait() err(%+v) oids(%+v)", err, oids)
		return nil, err
	}
	res := &icon.ListReply{Page: &icon.Page{Num: arg.Pn, Size: arg.Ps, Total: total}}
	//拼接信息
	for _, v := range icsInfo {
		for _, m := range v.ModuleInfo {
			if s, ok := sideBars[m.Oid]; ok {
				m.Name = s.Name
			}
			if sl, ok := sideBarLimits[m.Oid]; ok {
				m.Limit = sl
			}
		}
	}
	res.List = icsInfo
	return res, nil
}

// IconDetail is
func (s *Service) IconDetail(c context.Context, id int64) (*icon.IconInfo, error) {
	ic, err := s.showDao.Icon(c, id)
	if err != nil || ic == nil {
		log.Error("s.showDao.Icon id(%d) err(%+v) or ic=nil", id, err)
		return nil, err
	}
	var (
		sideBars      map[int64]*menu.Sidebar
		sideBarLimits map[int64][]*smdl.SidebarLimit
		mi            []*icon.Module
		oids          []int64
	)
	if err := json.Unmarshal([]byte(ic.Module), &mi); err != nil {
		log.Error("IconDetail  Unmarshal err(%+v) mi(%s)", err, ic.Module)
		return nil, err
	}
	for _, m := range mi {
		oids = append(oids, m.Oid)
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		sideBars, err = s.showDao.SideBars(c, oids)
		return err
	})
	eg.Go(func(ctx context.Context) (e error) {
		sideBarLimits, err = s.showDao.SideBarLimits(c, oids)
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Error("IconDetail eg.Wait() err(%+v) oids(%+v)", err, oids)
		return nil, err
	}
	res := &icon.IconInfo{Icon: ic}
	for _, m := range mi {
		side, ok := sideBars[m.Oid]
		if !ok {
			continue
		}
		sl, ok := sideBarLimits[m.Oid]
		if !ok {
			continue
		}
		res.ModuleInfo = append(res.ModuleInfo, &icon.ModuleInfo{Oid: m.Oid, Plat: m.Plat, Name: side.Name, Limit: sl})
	}
	return res, nil
}

// SaveLog .
func (s *Service) SaveLog(c context.Context, uid, id int64, opAction, operator string) error {
	ic, err := s.showDao.Icon(c, id)
	if err != nil {
		log.Error("SaveLog s.showDao.Icon (%d) error(%v) or nil", id, err)
		return err
	}
	arg := make(map[string]interface{})
	arg["info"] = ic
	if err = util.AddLog(common.LogMngIcon, operator, uid, id, opAction, arg); err != nil {
		log.Error("SaveLog AddLog error(%v)", err)
		return err
	}
	return nil
}

// IconOpt is
func (s *Service) IconOpt(c context.Context, id, uid int64, state int, operator string) error {
	rows, err := s.showDao.UpdateIconState(c, id, state)
	if err != nil {
		log.Error("IconOpt s.showDao.UpdateIconState id:%d state:%d error(%v)", id, state, err)
		return err
	}
	if rows == 0 {
		return ecode.NotModified
	}
	// 操作日志
	if err = s.SaveLog(c, uid, id, common.ActionDelete, operator); err != nil {
		log.Error("IconOpt s.SaveLog error(%v)", err)
		return err
	}
	return nil
}

// IconSave .
func (s *Service) IconSave(c context.Context, arg *icon.IconSaveParam, operator string, uid int64) error {
	iconParam := &icon.Icon{
		ID:          arg.ID,
		Module:      arg.Module,
		Icon:        arg.Icon,
		GlobalRed:   arg.GlobalRed,
		EffectGroup: arg.EffectGroup,
		EffectURL:   arg.EffectURL,
		Operator:    operator,
		Stime:       arg.Stime,
		Etime:       arg.Etime,
		State:       icon.StateNormal,
	}
	id, err := s.showDao.IconSave(c, iconParam)
	if err != nil {
		log.Error("IconSave s.showDao.IconSave err(%+v)", err)
		return err
	}
	// 操作日志
	opAction := common.ActionAdd
	if arg.ID > 0 {
		opAction = common.ActionUpdate
	}
	if err = s.SaveLog(c, uid, id, opAction, operator); err != nil {
		log.Error("IconSave s.SaveLog error(%v)", err)
		return err
	}
	return nil
}

// IconModule .
func (s *Service) IconModule(c context.Context, plat int32) ([]*icon.ModuleInfo, error) {
	sideBars, err := s.showDao.SideBarsByPlat(c, plat)
	if err != nil {
		log.Error("IconModule s.showDao.SideBarsByPlat plat(%d) err(%+v)", plat, err)
		return nil, err
	}
	var res []*icon.ModuleInfo
	for _, s := range sideBars {
		res = append(res, &icon.ModuleInfo{Oid: s.SID, Name: fmt.Sprintf("%s：%s", s.ModuleName, s.Name), Plat: s.Plat, Limit: s.Limit})
	}
	return res, nil
}
