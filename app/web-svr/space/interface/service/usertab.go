package service

import (
	"context"
	"fmt"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"
	"time"
)

// UserTab
func (s *Service) UserTab(c context.Context, req *pb.UserTabReq) (res *pb.UserTabReply, err error) {
	var tab *model.UserTab
	if tab, err = s.dao.UserTab(c, req); err != nil {
		log.Error("UserTab getCache req(%v) err(%v)", req, err)
		return nil, err
	}
	if tab == nil || tab.TabType == -1 {
		return nil, ecode.NothingFound
	}
	if tab.IsLimitValidated(req.Plat, req.Build) {
		return tab.ConvertToReply(), nil
	}
	return nil, ecode.NothingFound
}

// nolint:gomnd
func (s *Service) UpActivityTab(c context.Context, req *pb.UpActivityTabReq) (ret *pb.UpActivityTabResp, err error) {
	var (
		userTab *model.UserTab
		flag    int // 更新还是新增的标记为，如果是1为新增
	)
	ret = &pb.UpActivityTabResp{}
	arg := &model.UserTab{
		TabType: 3,
		Mid:     req.Mid,
		TabCont: req.TabCont,
		TabName: req.TabName,
		Stime:   xtime.Time(time.Now().Unix()),
		Etime:   model.LastTime,
		Online:  1,
	}
	// 校验MID合法性
	if err = s.MidInfo(c, req.Mid); err != nil {
		err = ecode.Error(-400, "无效MID")
		log.Error("service.UpActivityTab mid(%+v) error(%+v)", req.Mid, err)
		return
	}
	// 校验Native ID合法
	if err = s.dao.CheckNaPage(req.TabCont); err != nil {
		log.Error("service.UpActivityTab tab_cont(%+v) error(%+v)", req.TabCont, err)
		return
	}
	// 查找mid，如果存在，则进行更新操作，否则进行新增操作
	if userTab, err = s.dao.ValidUserTabFindByMid(req.Mid); err != nil {
		if err == sql.ErrNoRows {
			flag = 1
			err = nil
		} else {
			log.Error("service.UpActivityTab error(%+v)", err)
			return
		}
	}
	// 如果是下线，即可直接下线
	if req.State == 0 && flag != 1 {
		if err = s.OnlineUserTab(userTab.ID, 0); err != nil {
			err = ecode.Error(-400, fmt.Sprintf("下线MID失败:%s", err.Error()))
			log.Error("service.UpActivityTab mid(%+v) error(%+v)", req.Mid, err)
			return
		}
		arg.Online = model.OFFLINE
		// 刷新缓存
		if err = s.dao.FlushCache(c, arg); err != nil {
			log.Error("service.FlushCache error(%+v)", err)
		}
		ret.Success = true
		return
	} else if req.State == 0 && flag == 1 {
		err = ecode.Error(-400, fmt.Sprintf("下线失败:Mid(%d)无有效Tab配置", req.Mid))
		log.Error("service.UpActivityTab mid(%+v) error(%+v)", req.Mid, err)
		return
	}

	if flag == 1 {
		if err = s.AddSpaceUserTab(arg); err != nil {
			err = ecode.Error(-400, fmt.Sprintf("绑定活动失败:%s", err.Error()))
			log.Error("service.UpActivityTab mid(%+v) error(%+v)", req.Mid, err)
			return
		}
	} else {
		if userTab.TabType != 3 {
			err = ecode.Error(-400, "请求MID Tab非UP主")
			log.Error("service.UpActivityTab mid(%+v) error(%+v)", req.Mid, err)
			return
		}
		arg.ID = userTab.ID
		if err = s.ModifySpaceUserTab(arg); err != nil {
			err = ecode.Error(-400, fmt.Sprintf("更新活动失败:%s", err.Error()))
			log.Error("service.UpActivityTab mid(%+v) error(%+v)", req.Mid, err)
			return
		}
		// 刷新缓存
		if err = s.dao.FlushCache(c, arg); err != nil {
			log.Error("service.FlushCache error(%+v)", err)
		}
	}
	ret.Success = true
	return
}

// AddSpaceUserTab .
func (s *Service) AddSpaceUserTab(req *model.UserTab) (err error) {
	var (
		tmp *model.UserTab
		ok  bool
	)
	// 检查mid
	if tmp, err = s.dao.ValidUserTabFindByMid(req.Mid); err != nil {
		if err == sql.ErrNoRows {
			if ok, err = s.dao.EtimeUserTabFindByMid(req); err != nil {
				log.Error("serivce.AddSpaceUserTab Add arg(%v) error(%v)", req, err)
				return
			}
			if !ok {
				err = ecode.Error(-400, "当前时间已有配置")
				log.Error("serivce.AddSpaceUserTab Add arg(%v) error(%v)", req, err)
				return err
			}
			if err = s.dao.SpaceUserTabAdd(req); err != nil {
				log.Error("serivce.AddSpaceUserTab Add arg(%v) error(%v)", req, err)
				return
			}
			return nil
		}
		log.Error("service.AddSpaceUserTab Find(%v) Mid error(%v)", req, err)
		return
	}
	if tmp.Etime < req.Stime {
		if err = s.dao.SpaceUserTabAdd(req); err != nil {
			log.Error("serivce.AddSpaceUserTab Add arg(%v) error(%v)", req, err)
			return
		}
		return nil
	}
	err = ecode.Error(-400, "MID已存在")
	return
}

// ModifySpaceUserTab
func (s *Service) ModifySpaceUserTab(req *model.UserTab) (err error) {
	var (
		tmp *model.UserTab
		ok  bool
	)
	if tmp, err = s.dao.ValidUserTabFindByMid(req.Mid); err != nil {
		if err == sql.ErrNoRows {
			if ok, err = s.dao.EtimeUserTabFindByMid(req); err != nil {
				log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
				return
			}
			if !ok {
				err = ecode.Error(-400, "当前时间已有配置")
				log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
				return err
			}
			if err = s.dao.SpaceUserTabModify(req); err != nil {
				log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
				return
			}
			if err := s.dao.DelCacheUserTab(context.Background(), req.Mid); err != nil {
				log.Error("serivce.ModifySpaceUserTab modify DelCacheUserTab arg(%v) error(%v)", req, err)
				return err
			}
			return nil
		}
		log.Error("service.ModifySpaceUserTab error(%v)", err)
		return
	}
	if tmp.ID != req.ID && req.Online == 1 {
		err = ecode.Error(-400, "MID已存在")
		log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
		return
	}
	if err = s.dao.SpaceUserTabModify(req); err != nil {
		log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
		return
	}
	if err := s.dao.DelCacheUserTab(context.Background(), req.Mid); err != nil {
		log.Error("serivce.ModifySpaceUserTab modify DelCacheUserTab arg(%v) error(%v)", req, err)
		return err
	}
	return
}

func (s *Service) OnlineUserTab(id int64, state int) (err error) {
	var (
		tmpid  *model.UserTab
		tmpmid *model.UserTab
		ok     bool
	)
	if tmpid, err = s.dao.SpaceUserTabFindById(id); err != nil {
		log.Error("service.OnlineUserTab error(%v)", err)
		return
	}
	now := xtime.Time(time.Now().Unix())
	if state == model.ONLINE {
		tmpid.Online = model.ONLINE
		tmpid.Stime = now
		if tmpid.Etime.Time().Unix() < time.Now().Unix() {
			tmpid.Etime = model.LastTime
		}
	} else if state == model.OFFLINE {
		tmpid.Online = model.OFFLINE
		tmpid.Etime = now
	}
	if tmpmid, err = s.dao.ValidUserTabFindByMid(tmpid.Mid); err != nil {
		if err == sql.ErrNoRows {
			if ok, err = s.dao.EtimeUserTabFindByMid(tmpid); err != nil {
				log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", tmpid, err)
				return
			}
			if !ok {
				err = fmt.Errorf("当前时间已有配置")
				log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", tmpid, err)
				return err
			}
			if err = s.dao.SpaceUserTabOnline(id, tmpid); err != nil {
				log.Error("service.OnlineUserTab error(%v)", err)
				return
			}
			if err := s.dao.DelCacheUserTab(context.Background(), tmpid.Mid); err != nil {
				log.Error("serivce.ModifySpaceUserTab modify DelCacheUserTab arg(%v) error(%v)", tmpid, err)
				return err
			}
			return nil
		}
		log.Error("service.OnlineSpaceUserTab error(%v)", err)
		return
	}
	if tmpmid.ID != id {
		err = fmt.Errorf("当前MID已在线")
		log.Error("serivce.OnlineSpaceUserTab online arg(%v) error(%v)", id, err)
		return
	}
	if err = s.dao.SpaceUserTabOnline(id, tmpid); err != nil {
		log.Error("servie.OnlineUserTab error(%v)", err)
		return
	}
	if err := s.dao.DelCacheUserTab(context.Background(), tmpid.Mid); err != nil {
		log.Error("serivce.OnlineUserTab modify DelCacheUserTab arg(%v) error(%v)", tmpid, err)
		return err
	}
	return
}

func (s *Service) MidInfo(c context.Context, mid int64) (err error) {
	if _, err = s.dao.MidInfoReply(c, mid); err != nil {
		log.Error("service.MidInfo mid(%v) error(%v)", mid, err)
		return
	}
	return
}
