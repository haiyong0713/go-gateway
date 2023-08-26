package service

import (
	"context"
	"fmt"
	midrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/space/admin/dao"
	"go-gateway/app/web-svr/space/admin/model"
	util "go-gateway/app/web-svr/space/admin/util"
	"time"
)

func (s *Service) UpdateTabState() (err error) {
	c := cron.New()
	// 每10秒更新下tag状态
	err = c.AddFunc("*/10 * * * *", func() {
		var tabs []*model.UserTabReq
		if tabs, err = s.dao.FindUserTabByTime(dao.OFFLINE); err != nil {
			log.Error("service.UpdateTabState FindUserTabByTime arg(%v) error(%v)", dao.ONLINE, err)
			return
		}
		for _, item := range tabs {
			if err = s.OnlineUserTab(item.ID, dao.OFFLINE); err != nil {
				log.Error("service.UpdateTabState SpaceUserTabOnline arg(%v) error(%v)", item, err)
			}
			if err = util.AddLogs(model.LogExamine, "system", 0, item.ID, "offline", item); err != nil {
				log.Error("onlineUserTab AddLog arg(%v) error(%v)", item, err)
				return
			}
		}
		if tabs, err = s.dao.FindUserTabByTime(dao.ONLINE); err != nil {
			log.Error("service.UpdateTabState FindUserTabByTime arg(%v) error(%v)", dao.ONLINE, err)
			return
		}
		for _, item := range tabs {
			if err = s.OnlineUserTab(item.ID, dao.ONLINE); err != nil {
				log.Error("service.UpdateTabState SpaceUserTabOnline arg(%v) error(%v)", item, err)
			}
			if err = util.AddLogs(model.LogExamine, "system", 0, item.ID, "online", item); err != nil {
				log.Error("onlineUserTab AddLog arg(%v) error(%v)", item, err)
				return
			}
		}
	})
	if err != nil {
		log.Error("usertab.service UpadateTabState error: %s", err)
	}
	c.Start()
	return
}

// AddSpaceUserTab .
func (s *Service) AddSpaceUserTab(req *model.UserTabReq) (err error) {
	var (
		tmp *model.UserTabReq
		ok  bool
	)
	// 检查mid
	if tmp, err = s.dao.ValidUserTabFindByMid(req.Mid); err != nil {
		if err == gorm.ErrRecordNotFound {
			if ok, err = s.dao.EtimeUserTabFindByMid(req); err != nil {
				log.Error("serivce.AddSpaceUserTab Add arg(%v) error(%v)", req, err)
				return
			}
			if !ok {
				err = fmt.Errorf("当前时间已有配置")
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
	err = fmt.Errorf("MID已存在")
	return
}

// ModifySpaceUserTab
func (s *Service) ModifySpaceUserTab(req *model.UserTabReq) (err error) {
	var (
		tmp *model.UserTabReq
		ok  bool
	)
	if tmp, err = s.dao.ValidUserTabFindByMid(req.Mid); err != nil {
		if err == gorm.ErrRecordNotFound {
			if ok, err = s.dao.EtimeUserTabFindByMid(req); err != nil {
				log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
				return
			}
			if !ok {
				err = fmt.Errorf("当前时间已有配置")
				log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
				return err
			}
			if err = s.dao.SpaceUserTabModify(req); err != nil {
				log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
				return
			}
			return nil
		}
		log.Error("service.ModifySpaceUserTab error(%v)", err)
		return
	}
	if tmp.ID != req.ID && req.Online == 1 {
		err = fmt.Errorf("当前MID已存在")
		log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
		return
	}
	if err = s.dao.SpaceUserTabModify(req); err != nil {
		log.Error("serivce.ModifySpaceUserTab modify arg(%v) error(%v)", req, err)
		return
	}
	return
}

func (s *Service) OnlineUserTab(id int64, state int) (err error) {
	var (
		tmpid  *model.UserTabReq
		tmpmid *model.UserTabReq
		ok     bool
	)
	if tmpid, err = s.dao.SpaceUserTabFindById(id); err != nil {
		log.Error("service.OnlineUserTab error(%v)", err)
		return
	}
	now := xtime.Time(time.Now().Unix())
	if state == dao.ONLINE {
		tmpid.Online = dao.ONLINE
		tmpid.Stime = now
		if tmpid.Etime.Time().Unix() < time.Now().Unix() {
			tmpid.Etime = dao.LastTime
		}
	} else if state == dao.OFFLINE {
		tmpid.Online = dao.OFFLINE
		tmpid.Etime = now
	}
	if tmpmid, err = s.dao.ValidUserTabFindByMid(tmpid.Mid); err != nil {
		if err == gorm.ErrRecordNotFound {
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
			if tmpid.TabType == 2 && tmpid.Online == dao.OFFLINE && tmpid.TabCont != 0 {
				// 空间native页下线，通知网关
				if err = s.dao.NoticeUserTab(tmpid.Mid, tmpid.TabCont, dao.NativeType); err != nil {
					log.Error("service.OnlineUserTab NoticeUserTab mid(%+v) tab_cont(%+v) error(%+v)", tmpid.Mid, tmpid.TabCont, err)
					return
				}
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
		log.Error("service.OnlineUserTab error(%v)", err)
		return
	}
	if tmpid.TabType == 2 && tmpid.Online == dao.OFFLINE && tmpid.TabCont != 0 {
		// 空间native页下线，通知网关
		if err = s.dao.NoticeUserTab(tmpid.Mid, tmpid.TabCont, dao.NativeType); err != nil {
			log.Error("service.OnlineUserTab NoticeUserTab mid(%+v) tab_cont(%+v) error(%+v)", tmpid.Mid, tmpid.TabCont, err)
			return
		}
	}
	return
}

func (s *Service) UserTabList(arg *model.UserTabListReq) (pager *model.UserTabList, err error) {
	var (
		list  []*model.UserTabListReply
		count int
	)
	pager = &model.UserTabList{
		Page: model.Page{
			Num:  arg.Pn,
			Size: arg.Ps,
		},
	}
	if list, count, err = s.dao.SpaceUserTabList(arg); err != nil {
		log.Error("service.UserTabList error: %v", err)
		return
	}
	pager.Page.Total = count
	pager.List = list
	return
}

func (s *Service) MidInfo(c context.Context, mid int64) (info *model.MidInfoReply, err error) {
	var midinfo *midrpc.CardReply
	if midinfo, err = s.dao.MidInfoReply(c, mid); err != nil {
		log.Error("service.MidInfo mid(%v) error(%v)", mid, err)
		return
	}
	info = &model.MidInfoReply{
		Mid:      mid,
		MidName:  midinfo.Card.Name,
		Official: midinfo.Card.Official.Role,
	}
	return
}

func (s *Service) DeleteSpaceUserTab(id int64, t int) (err error) {
	if err = s.dao.SpaceUserTabDelete(id, t); err != nil {
		log.Error("serviece.SpaceUserTabDelte id(%v) error(%v)", id, err)
		return
	}
	return
}
