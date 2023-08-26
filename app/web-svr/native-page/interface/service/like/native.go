package like

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/web-svr/native-page/ecode"
	v1 "go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/app/web-svr/native-page/interface/dao/lottery"
	actmdl "go-gateway/app/web-svr/native-page/interface/model/act"
	dynmdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
	lmdl "go-gateway/app/web-svr/native-page/interface/model/like"
	lott "go-gateway/app/web-svr/native-page/interface/model/lottery"
)

const (
	_rcmdOffset = 40
	_maxPartNum = 10
)

// AutoDispense 仅仅更新缓存，等interface服务下线后开启tag等通知服务
// nolint:gocognit
func (s *Service) AutoDispense(c context.Context, msg string) (err error) {
	var m *lmdl.Message
	if err = json.Unmarshal([]byte(msg), &m); err != nil {
		log.Error("NatPageCache json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	log.Info("AutoDispense json.Unmarshal msg(%s)", msg)
	switch m.Table {
	case _nativeTsModuleTable:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert || m.Action == lmdl.ActDelete {
			return s.ntTsModuleUpdate(c, m.New)
		}
	case _nativeTsPageTable:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert || m.Action == lmdl.ActDelete {
			return s.ntTsPageUpdate(c, m.New)
		}
	case _nativePageTable:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert {
			return s.PageUpdate(c, m.New, m.Old, m.Action)
		} else if m.Action == lmdl.ActDelete {
			return s.PageDel(c, m.New)
		}
	case _nativeModuleTable:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert {
			return s.ModuleUpdate(c, m.New, m.Old)
		} else if m.Action == lmdl.ActDelete {
			return s.ModuleDel(c, m.New)
		}
	case _nativeClickTable:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert {
			return s.NatClickUpdate(c, m.New)
		} else if m.Action == lmdl.ActDelete {
			return s.NatClickDel(c, m.New)
		}
	case _nativeActTable:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert {
			return s.NatActUpdate(c, m.New)
		} else if m.Action == lmdl.ActDelete {
			return s.NatActDel(c, m.New)
		}
	case _nativeDnamicTable:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert {
			return s.NatDynamicUpdate(c, m.New)
		} else if m.Action == lmdl.ActDelete {
			return s.NatDynamicDel(c, m.New)
		}
	case _nativeVideoTable:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert {
			return s.NatVideoUpdate(c, m.New)
		} else if m.Action == lmdl.ActDelete {
			return s.NatVideoDel(c, m.New)
		}
	case _nativeMixtureExt:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert || m.Action == lmdl.ActDelete {
			return s.NatMixtureUpdate(c, m.New)
		}
	case _nativeParticipationExt:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert || m.Action == lmdl.ActDelete {
			return s.NatParticipationDel(c, m.New)
		}
	case _nativeActTab:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert || m.Action == lmdl.ActDelete {
			return s.natTabUpdate(c, m.New)
		}
	case _nativeTabModule:
		if m.Action == lmdl.ActUpdate || m.Action == lmdl.ActInsert || m.Action == lmdl.ActDelete {
			return s.natTabModuleUpdate(c, m.New)
		}
	}
	return
}

func (s *Service) natTabModuleUpdate(c context.Context, msg json.RawMessage) (err error) {
	var (
		m struct {
			ID       int64 `json:"id"`
			TabID    int64 `json:"tab_id"`
			Category int32 `json:"category"`
			State    int32 `json:"state"`
			Pid      int64 `json:"pid"`
		}
	)
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Error("natTabModuleUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroup.WithCancel(c)
	eg.Go(func(ctx context.Context) error {
		if e := s.natDao.DelCacheNativeTabModule(c, m.ID); e != nil {
			log.Error("s.natDao.DelCacheNativeTabModule(%d) error(%v)", m.ID, e)
			return e
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if e := s.natDao.DelNativeTabSort(c, m.TabID); e != nil {
			log.Error("s.natDao.DelNativeTabSort(%d) error(%v)", m.TabID, e)
			return e
		}
		return nil
	})
	if m.Category == v1.TabPageCategory {
		eg.Go(func(ctx context.Context) error {
			if e := s.natDao.DelCacheNativeTabBind(c, m.Pid, m.Category); e != nil {
				log.Error("s.natDao.DelCacheNativeTabBind(%d) error(%v)", m.TabID, e)
				return e
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("natTabModuleUpdate success %d", m.ID)
	return
}

// natTabUpdate .
func (s *Service) natTabUpdate(c context.Context, msg json.RawMessage) (err error) {
	var (
		m struct {
			ID int64 `json:"id"`
		}
	)
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Error("natTabUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if err = s.natDao.DelCacheNativeTab(c, m.ID); err != nil {
		log.Error("s.natDao.DelCacheNativeTab(%d) error(%v)", m.ID, err)
		return
	}
	log.Info("natTabUpdate success %d", m.ID)
	return
}

// NatParticipationDel .
func (s *Service) NatParticipationDel(c context.Context, msg json.RawMessage) (err error) {
	var (
		m struct {
			ID       int64 `json:"id"`
			ModuleID int64 `json:"module_id"`
		}
	)
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Error("NatParticipationDel json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelParticipationCache(ctx, m.ModuleID); e != nil {
			log.Error("s.natDao.DelParticipationCache(%d) error(%v)", m.ID, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNativePart(ctx, m.ID); e != nil {
			log.Error("s.natDao.DelCacheNativePart(%d) error(%v)", m.ID, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("NatParticipationDel success %d", m.ID)
	return

}

// NatMixtureUpdate .
func (s *Service) NatMixtureUpdate(c context.Context, msg json.RawMessage) (err error) {
	var (
		mobj struct {
			ID       int64 `json:"id"`
			ModuleID int64 `json:"module_id"`
			MType    int32 `json:"m_type"`
		}
	)
	if err = json.Unmarshal(msg, &mobj); err != nil {
		log.Error("NatMixtureUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNatMixIDs(ctx, mobj.ModuleID, mobj.MType); e != nil {
			log.Error("s.natDao.DelMixCache(%d,%d) error(%v)", mobj.MType, mobj.ID, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNatAllMixIDs(ctx, mobj.ModuleID); e != nil {
			log.Error("s.natDao.DelCacheNatAllMixIDs(%d) error(%v)", mobj.ID, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNativeMixtures(ctx, []int64{mobj.ID}); e != nil {
			log.Error("s.natDao.DelCacheNativeMixtures(%d) error(%v)", mobj.ID, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("NatMixtureUpdate success %d", mobj.ID)
	return
}

// ntTsPageUpdate .
func (s *Service) ntTsPageUpdate(c context.Context, msg json.RawMessage) (err error) {
	var (
		mobj struct {
			ID  int64 `json:"id"`
			PID int64 `json:"pid"`
		}
	)
	if err = json.Unmarshal(msg, &mobj); err != nil {
		log.Error("ntTsPageUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		if e := s.natDao.DelCacheNtPidToTsID(ctx, mobj.PID); e != nil {
			log.Error("s.natDao.DelCacheNtPidToTsID(%d) error(%v)", mobj.PID, e)
			return e
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if e := s.natDao.DelCacheNtTsPages(ctx, []int64{mobj.ID}); e != nil {
			log.Error("s.natDao.DelCacheNtTsPages(%d) error(%v)", mobj.ID, e)
			return e
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("ntTsPageUpdate success %d", mobj.ID)
	return
}

// ntTsModuleUpdate .
func (s *Service) ntTsModuleUpdate(c context.Context, msg json.RawMessage) (err error) {
	var (
		mobj struct {
			ID   int64 `json:"id"`
			TsID int64 `json:"ts_id"`
		}
	)
	if err = json.Unmarshal(msg, &mobj); err != nil {
		log.Error("ntTsPageUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		if e := s.natDao.DelCacheNtTsModulesExt(ctx, []int64{mobj.ID}); e != nil {
			log.Error("s.natDao.DelCacheNtTsPages(%d) error(%v)", mobj.ID, e)
			return e
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if e := s.natDao.DelCacheNtTsModuleIDs(ctx, mobj.TsID); e != nil {
			log.Error("s.natDao.DelCacheNtTsPages(%d) error(%v)", mobj.ID, e)
			return e
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("ntTsModuleUpdate success %d", mobj.ID)
	return
}

// PageCommon .
// nolint:gocognit
func (s *Service) PageUpdate(c context.Context, msg json.RawMessage, oldMsg json.RawMessage, mAction string) (err error) {
	var (
		m, mold dynmdl.PageMsg
		list    map[int64]*v1.NativePage
	)
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Error("PageCommon json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if list, err = s.natDao.RawNativePages(c, []int64{m.ID}); err != nil {
		log.Error("s.natDao.RawNativePages(%d) error(%v)", m.ID, err)
		return
	}
	if v, ok := list[m.ID]; ok {
		eg := errgroup.WithContext(c)
		if v.IsTopicAct() && v.ForeignID > 0 {
			isOnLine := false
			if v.IsOnline() { //话题活动上线状态
				isOnLine = true
				eg.Go(func(ctx context.Context) (e error) {
					if e = s.natDao.AddCacheNativeForeign(ctx, v.ForeignID, v.ID, v.Type); e != nil {
						log.Error("s.natDao.AddCacheNativeForeign(%d) error(%v)", m.ID, e)
					}
					return
				})
			} else { //其余状态，下线，待审核，草稿箱，打回
				eg.Go(func(ctx context.Context) (e error) {
					if e = s.natDao.DelCacheNativeForeign(ctx, v.ForeignID, v.Type); e != nil {
						log.Error("s.natDao.DelCachenativeForeign(%d) error(%v)", m.ID, e)
					}
					return
				})
			}
			//通知动态话题活动绑定消息
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.SendMsg(ctx, v, isOnLine); e != nil {
					log.Error("s.natDao.SendMsg(%v) online error(%v)", v, e)
				}
				return
			})
			// 通知tag话题活动绑定消息
			func() {
				if mAction == lmdl.ActInsert {
					if !isOnLine {
						return
					}
					eg.Go(func(ctx context.Context) (e error) {
						if e = s.tagSendMsg(ctx, v.ForeignID, isOnLine); e != nil {
							log.Error("s.tagSendMsg(%d) %v error(%v)", v.ForeignID, isOnLine, e)
						}
						return
					})
					return
				}
				if err = json.Unmarshal(oldMsg, &mold); err != nil {
					log.Error("PageCommon json.Unmarshal msg(%s) error(%v)", msg, err)
					return
				}
				var tagIsOnline bool
				if v.IsOnline() && mold.State != v1.OnlineState {
					tagIsOnline = true
				} else if !v.IsOnline() && mold.State == v1.OnlineState {
					tagIsOnline = false
				} else {
					return
				}
				eg.Go(func(ctx context.Context) (e error) {
					if e = s.tagSendMsg(ctx, v.ForeignID, tagIsOnline); e != nil {
						log.Error("s.tagSendMsg(%d) %v error(%v)", v.ForeignID, tagIsOnline, e)
					}
					return
				})
			}()
			//tag 绑定page更新
			switch {
			case v.IsOnline(), v.IsWaitForCheck(), v.IsWaitOnline():
				if !v1.IsFromTopicUpg(v.FromType) {
					eg.Go(func(ctx context.Context) (e error) {
						if e = s.natDao.AddCacheNatTagIDExist(ctx, v.ForeignID, v.ID); e != nil {
							log.Error("s.natDao.AddCacheNatTagIDExist(%d,%d) error(%v)", v.ForeignID, v.ID, e)
						}
						return
					})
				}
			default:
				eg.Go(func(ctx context.Context) (e error) {
					if e = s.natDao.DelCacheNatTagIDExist(ctx, v.ForeignID); e != nil {
						log.Error("s.natDao.DelCacheNatTagIDExist(%d) error(%v)", v.ForeignID, e)
					}
					return
				})
			}
			//tag 绑定page更新
		}
		eg.Go(func(ctx context.Context) (e error) {
			save := map[int64]*v1.NativePage{m.ID: v}
			if e = s.natDao.AddCacheNativePages(ctx, save); e != nil {
				log.Error("s.natDao.AddCacheNativePages(%d) error(%v)", m.ID, e)
			}
			return
		})
		// 如果是up主发起，并且是下线，则下线对应数据源
		if isUpAct(&m) && v.IsOffline() {
			eg.Go(func(ctx context.Context) error {
				var err error
				for i := 0; i < 3; i++ {
					if err = s.offlineActSubject(ctx, v.ID); err == nil {
						return nil
					}
					time.Sleep(10 * time.Millisecond)
				}
				log.Error("日志告警 通知活动数据源下线失败, page={%+v} error=%+v", v, err)
				return nil
			})
		}
		if isUpAct(&m) {
			eg.Go(func(ctx context.Context) error {
				_ = s.natDao.AddCacheSponsoredUp(ctx, m.RelatedUid)
				return nil
			})
		}
		if mAction == lmdl.ActUpdate {
			if err = json.Unmarshal(oldMsg, &mold); err != nil {
				log.Error("PageCommon json.Unmarshal msg(%s) error(%v)", msg, err)
				return
			}
			//通知评论
			if m.Type == v1.TopicActType && m.State != mold.State {
				eg.Go(func(ctx context.Context) error {
					_ = s.replySendMsg(ctx, m.ID, int64(m.State))
					return nil
				})
			}
			// 处理up主发起活动有效列表和在线列表数据
			// up发起活动私信通知
			s.pageTsUpdate(c, eg, m, mold)
			if m.ActType != mold.ActType || m.State != mold.State {
				s.delCacheNatIDsByActType(eg, m.ActType)
				if m.ActType != mold.ActType {
					s.delCacheNatIDsByActType(eg, mold.ActType)
				}
			}
			// 如果UP主发起活动是首次审核通过则自动发布动态
			if canPublishDynamic(&m, &mold) {
				eg.Go(func(ctx context.Context) error {
					_ = s.publishDynamic(ctx, m.RelatedUid, m.ID)
					return nil
				})
			}
		} else if mAction == lmdl.ActInsert {
			//通知评论
			if m.Type == v1.TopicActType {
				eg.Go(func(ctx context.Context) error {
					_ = s.replySendMsg(ctx, m.ID, int64(m.State))
					return nil
				})
			}
			if m.Type == v1.TopicActType && m.FromType == v1.PageFromUid && m.RelatedUid > 0 {
				// 处理up主发起活动有效列表和在线列表数据
				s.pageTsAdd(eg, m)
			}
			if m.State == v1.OnlineState {
				s.delCacheNatIDsByActType(eg, m.ActType)
			}
		}
		if err = eg.Wait(); err != nil {
			return
		}
		log.Info("PageCommon success %d", m.ID)
	}
	return
}

func (s *Service) pageTsAdd(eg *errgroup.Group, m dynmdl.PageMsg) {
	var tMime int64
	if tt, er := time.Parse("2006-01-02 15:04:05", m.Mtime); er == nil {
		tMime = tt.Unix()
	} else {
		tMime = time.Now().Unix()
	}
	switch m.State {
	case v1.WaitForCheck, v1.WaitForOnline, v1.CheckOffline, v1.OnlineState, v1.OfflineState: //添加到我的活动列表
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.AddSingleCacheNtTsUIDs(ctx, m.RelatedUid, m.ID, tMime); e != nil {
				log.Error("s.natDao.AddSingleCacheNtTsUIDs(%d,%d,%s) error(%v)", m.RelatedUid, m.ID, m.Mtime, e)
			}
			return
		})
		if m.State == v1.OnlineState { //添加到有效活动列表
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.AddSingleCacheNtTsOnlineIDs(ctx, m.RelatedUid, m.ID, tMime); e != nil {
					log.Error("s.natDao.AddSingleCacheNtTsOnlineIDs(%d,%d,%s) error(%v)", m.RelatedUid, m.ID, m.Mtime, e)
				}
				return
			})
		}
	default:
	}
}

// pageTsUpdate .
// nolint:gocognit
func (s *Service) pageTsUpdate(c context.Context, eg *errgroup.Group, m, mld dynmdl.PageMsg) {
	var newData bool
	// 活动类型发生变更 一般不允许修改
	if m.FromType != mld.FromType || m.RelatedUid != mld.RelatedUid || m.Type != mld.Type {
		// 移除变更前数据
		if mld.FromType == v1.PageFromUid && mld.Type == v1.TopicActType && mld.RelatedUid > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.ZremSingleCacheNtTsUIDs(ctx, mld.RelatedUid, mld.ID); e != nil {
					log.Error("s.natDao.DelCacheNtTsUIDs(%d) error(%v)", mld.RelatedUid, e)
				}
				return
			})
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.ZremSingleCacheNtTsOnlineIDs(ctx, mld.RelatedUid, mld.ID); e != nil {
					log.Error("s.natDao.DelCacheNtTsOnlineIDs(%d) error(%v)", mld.RelatedUid, e)
				}
				return
			})
		}
		//变更后数据是up主发起活动类型
		if m.FromType == v1.PageFromUid && m.Type == v1.TopicActType && m.RelatedUid > 0 {
			newData = true
		}
	} else if (m.State != mld.State || m.Mtime != mld.Mtime) && m.FromType == v1.PageFromUid && m.Type == v1.TopicActType && m.RelatedUid > 0 { //类型数据都没有发生变更 状态发生变化的up主发起活动
		//up主发起话题活动，且状态发生变更或者时间发生变化，更新score
		newData = true
	}
	if newData {
		var tMime int64
		if tt, er := time.Parse("2006-01-02 15:04:05", m.Mtime); er == nil {
			tMime = tt.Unix()
		} else {
			tMime = time.Now().Unix()
		}
		switch m.State {
		case v1.WaitForCheck, v1.WaitForOnline, v1.CheckOffline, v1.OnlineState, v1.OfflineState: //添加到我的活动列表
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.AddSingleCacheNtTsUIDs(ctx, m.RelatedUid, m.ID, tMime); e != nil {
					log.Error("s.natDao.AddSingleCacheNtTsUIDs(%d,%d,%s) error(%v)", m.RelatedUid, m.ID, m.Mtime, e)
				}
				return
			})
		default: //从我的活动列表移除
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.ZremSingleCacheNtTsUIDs(ctx, m.RelatedUid, m.ID); e != nil {
					log.Error("s.natDao.ZremSingleCacheNtTsUIDs(%d,%d) error(%v)", m.RelatedUid, m.ID, e)
				}
				return
			})
		}
		switch m.State {
		case v1.OnlineState: //添加到有效活动列表
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.AddSingleCacheNtTsOnlineIDs(ctx, m.RelatedUid, m.ID, tMime); e != nil {
					log.Error("s.natDao.AddSingleCacheNtTsOnlineIDs(%d,%d,%s) error(%v)", m.RelatedUid, m.ID, m.Mtime, e)
				}
				return
			})
		default: //从有效活动列表移除
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.ZremSingleCacheNtTsOnlineIDs(ctx, m.RelatedUid, m.ID); e != nil {
					log.Error("s.natDao.ZremSingleCacheNtTsOnlineIDs(%d,%d) error(%v)", m.RelatedUid, m.ID, e)
				}
				return
			})
		}
	}
	// up主发起话题活动，从上线状态变为下线状态
	if m.FromType == v1.PageFromUid && m.Type == v1.TopicActType && m.RelatedUid > 0 && m.State == v1.OfflineState && mld.State == v1.OnlineState {
		//私信通知
		eg.Go(func(ctx context.Context) (e error) {
			lotReq := &lott.LetterParam{RecverIDs: []uint64{uint64(m.RelatedUid)}, SenderUID: s.c.Rule.UpSenderUid, MsgType: 10}
			lotReq.NotifyCode = s.c.Rule.NotifyCodeOffline
			lotReq.Params = lottery.BuildNotifyParams([]string{m.Title, offReason(m.OffReason)})
			if _, e = s.lottDao.SendLetter(c, lotReq); e != nil {
				log.Error("s.lottDao.SendLetter %d,%s error(%v)", m.RelatedUid, m.Title, e)
				e = nil
			}
			return
		})
	}
}

// PageDel .
func (s *Service) PageDel(c context.Context, msg json.RawMessage) (err error) {
	var (
		m struct {
			ID           int64  `json:"id"`
			Title        string `json:"title"`
			Type         int64  `json:"type"`
			ForeignID    int64  `json:"foreign_id"`
			Creator      string `json:"creator"`
			Operator     string `json:"operator"`
			ShareTitle   string `json:"share_title"`
			ShareImage   string `json:"share_image"`
			ShareURL     string `json:"share_url"`
			State        int64  `json:"state"`
			SkipURL      string `json:"skip_url"`
			PcURL        string `json:"pc_url"`
			AnotherTitle string `json:"another_title"`
			FromType     int32  `json:"from_type"`
			RelatedUid   int64  `json:"related_uid"`
			ActType      int32  `json:"act_type"`
			OffReason    string `json:"off_reason"`
		}
	)
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Error("PageAdd json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	vpage := &v1.NativePage{
		ID:           m.ID,
		Title:        m.Title,
		Type:         m.Type,
		ForeignID:    m.ForeignID,
		Creator:      m.Creator,
		Operator:     m.Operator,
		ShareTitle:   m.ShareTitle,
		ShareImage:   m.ShareImage,
		ShareURL:     m.SkipURL,
		State:        m.State,
		SkipURL:      m.SkipURL,
		PcURL:        m.PcURL,
		AnotherTitle: m.AnotherTitle,
		FromType:     m.FromType,
		ActType:      m.ActType,
		OffReason:    m.OffReason,
	}
	eg := errgroup.WithContext(c)
	if vpage.IsTopicAct() && vpage.ForeignID > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.SendMsg(ctx, vpage, false); e != nil {
				log.Error("s.natDao.SendMsg(%v) offline error(%v)", vpage, e)
			}
			return
		})
		// 通知tag话题活动绑定消息
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.tagSendMsg(ctx, vpage.ForeignID, false); e != nil {
				log.Error("s.tagSendMsg(%d) offline error(%v)", vpage.ForeignID, e)
			}
			return
		})
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.DelCacheNativeForeign(ctx, vpage.ForeignID, vpage.Type); e != nil {
				log.Error("s.natDao.DelCachenativeForeign(%d) error(%v)", vpage.ID, e)
			}
			return
		})
	}
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNativePages(ctx, []int64{vpage.ID}); e != nil {
			log.Error("s.natDao.DelCacheNativePages(%d) error(%v)", vpage.ID, e)
		}
		return
	})
	// 移除变更前数据 up主发起的话题活动
	if vpage.FromType == v1.PageFromUid && vpage.Type == v1.TopicActType && vpage.RelatedUid > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.ZremSingleCacheNtTsUIDs(ctx, vpage.RelatedUid, vpage.ID); e != nil {
				log.Error("s.natDao.DelCacheNtTsUIDs(%d) error(%v)", vpage.RelatedUid, e)
			}
			return
		})
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.ZremSingleCacheNtTsOnlineIDs(ctx, vpage.RelatedUid, vpage.ID); e != nil {
				log.Error("s.natDao.DelCacheNtTsOnlineIDs(%d) error(%v)", vpage.RelatedUid, e)
			}
			return
		})
		// up主发起话题活动，从上线状态变为下线状态
		//私信通知
		eg.Go(func(ctx context.Context) (e error) {
			lotReq := &lott.LetterParam{RecverIDs: []uint64{uint64(vpage.RelatedUid)}, SenderUID: s.c.Rule.UpSenderUid, MsgType: 10}
			lotReq.NotifyCode = s.c.Rule.NotifyCodeOffline
			lotReq.Params = lottery.BuildNotifyParams([]string{vpage.Title, offReason(vpage.OffReason)})
			if _, e = s.lottDao.SendLetter(c, lotReq); e != nil {
				log.Error("s.lottDao.SendLetter %d,%s error(%v)", vpage.RelatedUid, vpage.Title, e)
				e = nil
			}
			return
		})
	}
	s.delCacheNatIDsByActType(eg, int64(vpage.ActType))
	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("PageDel success %d", vpage.ID)
	return
}

// tagSendMsg .
func (s *Service) tagSendMsg(c context.Context, tid int64, isOnline bool) (e error) {
	stat := int32(0)
	if isOnline {
		stat = 1
	}
	if e = s.tagDao.UpdateExtraAttr(c, tid, stat); e == nil {
		return
	}
	//通知失败，异步重复通知
	if err := s.cache.Do(c, func(c context.Context) {
		for i := 0; i < 2; i++ {
			time.Sleep(10 * time.Millisecond)
			if e = s.tagDao.UpdateExtraAttr(c, tid, stat); e == nil {
				return
			}
			log.Error("Fail to send TagMsg, tid=%+v state=%+v error=%+v", tid, stat, e)
		}
		log.Error("日志告警:通知tag绑定关系失败s.tagDao.UpdateExtraAttr(%d) ofline error(%v)", tid, e)
	}); err != nil {
		log.Errorc(c, "tagSendMsg fanout.Do() failed, tid=%+v error=%+v", tid, err)
	}
	return
}

// replySendMsg .
func (s *Service) replySendMsg(c context.Context, pid, stat int64) (e error) {
	if e = s.replyDao.UpdateActivityState(c, pid, stat); e == nil {
		log.Info("replySendMsg send success (%d,%d)", pid, stat)
		return
	}
	//通知失败，异步重复通知
	if err := s.cache.Do(c, func(c context.Context) {
		for i := 0; i < 2; i++ {
			time.Sleep(10 * time.Millisecond)
			if e = s.replyDao.UpdateActivityState(c, pid, stat); e == nil {
				return
			}
			log.Error("Fail to send ReplyMsg, pid=%+v stat=%+v error=%+v", pid, stat, e)
		}
		log.Error("日志告警:通知评论绑定关系失败replySendMsg s.replyDao.UpdateActivityState(%d,%d) ofline error(%v)", pid, stat, e)
	}); err != nil {
		log.Errorc(c, "replySendMsg fanout.Do() failed, pid=%+v error=%+v", pid, err)
	}
	return
}

// ModuleUpdate .
func (s *Service) ModuleUpdate(c context.Context, msg json.RawMessage, old json.RawMessage) (err error) {
	var (
		mobj struct {
			ID int64 `json:"id"`
		}
		oldObj struct {
			NativeID int64  `json:"native_id"`
			Ukey     string `json:"ukey"`
		}
		list map[int64]*v1.NativeModule
	)
	if err = json.Unmarshal(msg, &mobj); err != nil {
		log.Error("ModuleUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if list, err = s.natDao.RawNativeModules(c, []int64{mobj.ID}); err != nil {
		log.Error("s.natDao.Modules(%d) error(%v)", mobj.ID, err)
		return
	}
	if v, ok := list[mobj.ID]; ok {
		eg := errgroup.WithContext(c)
		if v.IsOnline() {
			eg.Go(func(ctx context.Context) (e error) {
				save := map[int64]*v1.NativeModule{v.ID: v}
				if e = s.natDao.AddCacheNativeModules(ctx, save); e != nil {
					log.Error("s.natDao.AddCacheNativeModules(%d) error(%v)", v.ID, e)
				}
				return
			})
		} else {
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.DelCacheNativeModules(ctx, []int64{v.ID}); e != nil {
					log.Error("s.natDao.DelCacheNativeModules(%d) error(%v)", v.ID, e)
				}
				return
			})
		}
		// 删除native_page对应的moduleids
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.DeleteModuleCache(ctx, v.NativeID, v.PType); e != nil {
				log.Error("s.natDao.DeleteModuleCache(%d,%d) error(%v)", v.NativeID, v.PType, e)
			}
			return
		})
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.DelCacheNativeUkey(ctx, v.NativeID, v.Ukey); e != nil {
				log.Error("s.natDao.DelCacheNativeUkey(%d,%s) error(%v)", v.NativeID, v.Ukey, e)
			}
			return
		})
		// 删除old对应的key
		if len(old) > 0 {
			eg.Go(func(ctx context.Context) (e error) {
				if e = json.Unmarshal(old, &oldObj); e != nil {
					log.Error("ModuleUpdate json.Unmarshal old(%s) error(%v)", old, e)
					e = nil
					return
				}
				if oldObj.NativeID > 0 {
					if e = s.natDao.DelCacheNativeUkey(ctx, oldObj.NativeID, oldObj.Ukey); e != nil {
						log.Error("s.natDao.DelCacheNativeUkey(%d,%s) error(%v)", oldObj.NativeID, oldObj.Ukey, e)
						e = nil
					}
				}
				return
			})
		}
		if err = eg.Wait(); err != nil {
			return
		}
		log.Info("ModuleUpdate success %d", v.ID)
	}
	return
}

// ModuleDel .
func (s *Service) ModuleDel(c context.Context, msg json.RawMessage) (err error) {
	var (
		m struct {
			ID       int64  `json:"id"`
			NativeID int64  `json:"native_id"`
			Ukey     string `json:"ukey"`
			PType    int32  `json:"p_type"`
		}
	)
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Error("ModuleUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DeleteModuleCache(ctx, m.NativeID, m.PType); e != nil {
			log.Error("s.natDao.DeleteModuleCache(%d) error(%v)", m.ID, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNativeModules(ctx, []int64{m.ID}); e != nil {
			log.Error("s.natDao.DelCacheNativeModules(%d) error(%v)", m.ID, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNativeUkey(ctx, m.NativeID, m.Ukey); e != nil {
			log.Error("s.natDao.DelCacheNativeUkey(%d,%s) error(%v)", m.NativeID, m.Ukey, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("ModuleDel success %d", m.ID)
	return
}

// NatClickUpdate .
func (s *Service) NatClickUpdate(c context.Context, msg json.RawMessage) (err error) {
	var (
		mobj struct {
			ID int64 `json:"id"`
		}
		list map[int64]*v1.NativeClick
	)
	if err = json.Unmarshal(msg, &mobj); err != nil {
		log.Error("NatClickUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if list, err = s.natDao.Clicks(c, []int64{mobj.ID}); err != nil {
		log.Error("s.natDao.Clicks(%d) error(%v)", mobj.ID, err)
		return
	}
	if v, ok := list[mobj.ID]; ok {
		eg := errgroup.WithContext(c)
		if v.IsOnline() {
			eg.Go(func(ctx context.Context) (e error) {
				save := map[int64]*v1.NativeClick{v.ID: v}
				if e = s.natDao.AddCacheNativeClicks(ctx, save); e != nil {
					log.Error("s.natDao.AddCacheNativeClicks(%d) error(%v)", v.ID, e)
				}
				return
			})
		} else {
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.DelCacheNativeClicks(ctx, []int64{v.ID}); e != nil {
					log.Error("s.natDao.DelCacheNativeClicks(%d) error(%v)", v.ID, e)
				}
				return
			})
		}
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.DelNativeClickIDs(ctx, v.ModuleID); e != nil {
				log.Error("s.natDao.AddClickCache(%d) error(%v)", v.ID, e)
			}
			return
		})
		if err = eg.Wait(); err != nil {
			return
		}
		log.Info("NatClickUpdate success %d", v.ID)
	}
	return
}

// NatClicksDel .
func (s *Service) NatClickDel(c context.Context, msg json.RawMessage) (err error) {
	var (
		m struct {
			ID       int64 `json:"id"`
			ModuleID int64 `json:"module_id"`
		}
	)
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Error("NatClickDel json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelNativeClickIDs(ctx, m.ModuleID); e != nil {
			log.Error("s.natDao.DelNativeClickIDs(%d) error(%v)", m.ID, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNativeClicks(ctx, []int64{m.ID}); e != nil {
			log.Error("s.natDao.DelCacheNativeClicks(%d) error(%v)", m.ID, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("NatClickDel success %d", m.ID)
	return
}

// NatActUpdate .
func (s *Service) NatActUpdate(c context.Context, msg json.RawMessage) (err error) {
	var (
		mobj struct {
			ModuleID int64 `json:"module_id"`
		}
	)
	if err = json.Unmarshal(msg, &mobj); err != nil {
		log.Error("NatActUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if err = s.natDao.DelNativeActIDs(c, mobj.ModuleID); err != nil {
		log.Error("s.natDao.DelNativeActIDs(%d) error(%v)", mobj.ModuleID, err)
		return
	}
	log.Info("NatActUpdate success %d", mobj.ModuleID)
	return
}

// DelNatAct .
func (s *Service) NatActDel(c context.Context, msg json.RawMessage) (err error) {
	var (
		v struct {
			ModuleID int64 `json:"module_id"`
		}
	)
	if err = json.Unmarshal(msg, &v); err != nil {
		log.Error("NatActDel json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if err = s.natDao.DelNativeActIDs(c, v.ModuleID); err != nil {
		log.Error("s.natDao.DelNativeActIDs(%d) error(%v)", v.ModuleID, err)
		return
	}
	log.Info("NatActDel success %d", v.ModuleID)
	return
}

// NatDynamicUpdate .
func (s *Service) NatDynamicUpdate(c context.Context, msg json.RawMessage) (err error) {
	var (
		mobj struct {
			ID int64 `json:"id"`
		}
		list map[int64]*v1.NativeDynamicExt
	)
	if err = json.Unmarshal(msg, &mobj); err != nil {
		log.Error("NatDynamicUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if list, err = s.natDao.Dynamics(c, []int64{mobj.ID}); err != nil {
		log.Error("s.natDao.Dynamics(%d) error(%v)", mobj.ID, err)
		return
	}
	if v, ok := list[mobj.ID]; ok {
		eg := errgroup.WithContext(c)
		if v.IsOnline() {
			eg.Go(func(ctx context.Context) (e error) {
				save := map[int64]*v1.NativeDynamicExt{v.ID: v}
				if e = s.natDao.AddCacheNativeDynamics(ctx, save); e != nil {
					log.Error("s.natDao.AddCacheNativeDynamics(%d) error(%v)", v.ID, e)
				}
				return
			})
		} else {
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.DelCacheNativeDynamics(ctx, []int64{v.ID}); e != nil {
					log.Error("s.natDao.DelCacheNativeDynamics(%d) error(%v)", v.ID, e)
				}
				return
			})
		}
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.DelNativeDynamicIDs(ctx, v.ModuleID); e != nil {
				log.Error("s.natDao.DelNativeDynamicIDs(%d) error(%v)", v.ID, e)
			}
			return
		})
		if err = eg.Wait(); err != nil {
			return
		}
		log.Info("NatDynamicUpdate success %d", v.ID)
	}
	return
}

// NatDynamicDel .
func (s *Service) NatDynamicDel(c context.Context, msg json.RawMessage) (err error) {
	var (
		m struct {
			ID       int64 `json:"id"`
			ModuleID int64 `json:"module_id"`
		}
	)
	if err = json.Unmarshal(msg, &m); err != nil {
		log.Error("NatDynamicDel json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelNativeDynamicIDs(ctx, m.ModuleID); e != nil {
			log.Error("s.natDao.DelNativeDynamicIDs(%d) error(%v)", m.ID, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNativeDynamics(ctx, []int64{m.ID}); e != nil {
			log.Error("s.natDao.DelCacheNativeDynamics(%d) error(%v)", m.ID, e)
		}
		return
	})
	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("NatDynamicDel success %d", m.ID)
	return
}

// NatVideoUpdate .
func (s *Service) NatVideoUpdate(c context.Context, msg json.RawMessage) (err error) {
	var (
		mobj struct {
			ID int64 `json:"id"`
		}
		list map[int64]*v1.NativeVideoExt
	)
	if err = json.Unmarshal(msg, &mobj); err != nil {
		log.Error("NatVideoUpdate json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	if list, err = s.natDao.NatVideos(c, []int64{mobj.ID}); err != nil {
		log.Error("s.natDao.NatVideos(%d) error(%v)", mobj.ID, err)
		return
	}
	if v, ok := list[mobj.ID]; ok {
		eg := errgroup.WithContext(c)
		if v.IsOnline() {
			eg.Go(func(ctx context.Context) (e error) {
				save := map[int64]*v1.NativeVideoExt{v.ID: v}
				if e = s.natDao.AddCacheNativeVideos(ctx, save); e != nil {
					log.Error("s.natDao.AddCacheNativeVideos(%d) error(%v)", v.ID, e)
				}
				return
			})
		} else {
			eg.Go(func(ctx context.Context) (e error) {
				if e = s.natDao.DelCacheNativeVideos(ctx, []int64{v.ID}); e != nil {
					log.Error("s.natDao.DelCacheNativeVideos(%d) error(%v)", v.ID, e)
				}
				return
			})
		}
		eg.Go(func(ctx context.Context) (e error) {
			if e = s.natDao.DelNativeVideoIDs(ctx, v.ModuleID); e != nil {
				log.Error("s.natDao.DelNativeVideoIDs(%d) error(%v)", v.ID, e)
			}
			return
		})
		if err = eg.Wait(); err != nil {
			return
		}
		log.Info("NatVideoUpdate success %d", v.ID)
	}
	return
}

// DelNatAct .
func (s *Service) NatVideoDel(c context.Context, msg json.RawMessage) (err error) {
	var (
		v struct {
			ID       int64 `json:"id"`
			ModuleID int64 `json:"module_id"`
		}
	)
	if err = json.Unmarshal(msg, &v); err != nil {
		log.Error("NatVideoDel json.Unmarshal msg(%s) error(%v)", msg, err)
		return
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelNativeVideoIDs(ctx, v.ModuleID); e != nil {
			log.Error("s.natDao.DelNativeVideoIDs(%d) error(%v)", v.ID, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if e = s.natDao.DelCacheNativeVideos(ctx, []int64{v.ID}); e != nil {
			log.Error("s.natDao.DelCacheNativeVideos(%d) error(%v)", v.ID, e)
		}
		return
	})

	if err = eg.Wait(); err != nil {
		return
	}
	log.Info("NatVideoDel success %d", v.ID)
	return
}

// ModuleConfig .
func (s *Service) ModuleConfig(c context.Context, arg *v1.ModuleConfigReq) (res *v1.ModuleConfigReply, err error) {
	var (
		vMls map[int64]*v1.NativeModule
	)
	// 返回module信息，包括已经删除，优化已删除组件在前台后,用户无法点击查看更多后去内容
	if vMls, err = s.natDao.NativeModules(c, []int64{arg.ModuleID}); err != nil {
		log.Info(" s.natDao.NativeModules(%d) error(%v)", arg.ModuleID, err)
		return
	}
	res = &v1.ModuleConfigReply{}
	if _, ok := vMls[arg.ModuleID]; !ok {
		return
	}
	v := vMls[arg.ModuleID]
	if v == nil || v.ID <= 0 {
		return
	}
	res.Module = &v1.Module{NativeModule: v}
	eg := errgroup.WithContext(c)
	switch {
	case v.IsClick():
		eg.Go(func(ctx context.Context) (e error) {
			if res.Module.Click, e = s.clickData(ctx, arg.ModuleID); e != nil {
				log.Error("s.clickData(%d) error(%v)", arg.ModuleID, e)
				e = nil
			}
			return
		})
	case v.IsVideo(), v.IsVideoAct(), v.IsResourceAct(), v.IsNewVideoAct():
		eg.Go(func(ctx context.Context) (e error) {
			if res.Module.VideoAct, e = s.videoData(ctx, arg.ModuleID); e != nil {
				log.Error("s.videoData(%d) error(%v)", arg.ModuleID, e)
				e = nil
			}
			return
		})
	case v.IsResourceOrigin():
		if v.ConfUnmarshal().RdbType == v1.RDBLive {
			eg.Go(func(ctx context.Context) (e error) {
				if res.Module.VideoAct, e = s.videoData(ctx, arg.ModuleID); e != nil {
					log.Error("s.videoData(%d) error(%v)", arg.ModuleID, e)
					e = nil
				}
				return
			})
		}
	case v.IsAct():
		eg.Go(func(ctx context.Context) (e error) {
			if res.Module.Act, e = s.actData(ctx, arg.ModuleID); e != nil {
				log.Error("s.actData(%d) error(%v)", arg.ModuleID, e)
				e = nil
			}
			return
		})
	case v.IsDynamic(), v.IsResourceDyn():
		eg.Go(func(ctx context.Context) (e error) {
			if res.Module.Dynamic, e = s.dynamicData(ctx, arg.ModuleID); e != nil {
				log.Error("s.dynamicData(%d) error(%v)", arg.ModuleID, e)
				e = nil
			}
			return
		})
	}
	eg.Go(func(ctx context.Context) error {
		pageIDs := []int64{v.NativeID}
		if arg.PrimaryPageID != 0 {
			pageIDs = append(pageIDs, arg.PrimaryPageID)
		}
		pgs, e := s.natDao.NativePages(ctx, pageIDs)
		if e != nil {
			log.Error("s.NativePages(%d) error(%v)", v.NativeID, e)
			return nil
		}
		if val, ok := pgs[v.NativeID]; ok && val.IsOnline() {
			res.NativePage = val
		}
		if res.NativePage != nil && res.NativePage.IsInlineAct() {
			if primaryPage, ok := pgs[arg.PrimaryPageID]; ok && primaryPage.IsOnline() {
				res.PrimaryPage = primaryPage
			}
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "ModuleConfig eg.Wait() failed, req=%+v error=%+v", arg, err)
	}
	return
}

// NatConfig .
// nolint:gocognit
func (s *Service) NatConfig(c context.Context, arg *v1.NatConfigReq) (res *v1.NatConfigReply, err error) {
	var (
		natPages              map[int64]*v1.NativePage
		modules               map[int64]*v1.NativeModule
		sortReply             *lmdl.SortModule
		baseMIDs, pageBaseIDs []int64
		baseMul               []*v1.Module
	)
	res = &v1.NatConfigReply{}
	if natPages, err = s.natDao.NativePages(c, []int64{arg.Pid}); err != nil {
		log.Error("s.natDao.NativePages(%d) error(%v)", arg.Pid, err)
		return
	}
	if _, ok := natPages[arg.Pid]; !ok || !natPages[arg.Pid].IsOnline() {
		err = ecode.NativePageOffline
		return
	}
	res.NativePage = natPages[arg.Pid]
	//兼容老的话题数据，有跳转地址，直接返回
	if natPages[arg.Pid].IsTopicAct() && arg.PType == v1.CommonPage && res.NativePage.SkipURL != "" {
		return
	}
	//页面至少含有一个组件
	eeg := errgroup.WithContext(c)
	//需要获取最上层页面cache start
	var firstPage *v1.FirstPage
	if natPages[arg.Pid].IsInlineAct() {
		if natPages[arg.Pid].FirstPid > 0 { //内嵌页存在父id的数据
			eeg.Go(func(ctx context.Context) (e error) {
				firstPage, e = s.inlineFirstPage(ctx, natPages[arg.Pid].FirstPid)
				if e != nil {
					log.Error("s.inlineFirstPage(%d) error(%v)", arg.Pid, e)
				}
				return
			})
		}
	} else {
		eeg.Go(func(ctx context.Context) error {
			firstPage = &v1.FirstPage{Item: natPages[arg.Pid]}
			if natPages[arg.Pid].IsAttrWhiteSwitch() == v1.AttrModuleYes { //开启了白名单逻辑
				extPage, e := s.natDao.NativeExtend(ctx, arg.Pid)
				if e != nil { //父页面获取失败
					log.Error("s.natDao.NativeExtend(%d) error(%v)", arg.Pid, e)
					return e
				}
				firstPage.Ext = extPage
			}
			return nil
		})
	}
	//需要获取最上层页面cache end
	eeg.Go(func(ctx context.Context) (e error) {
		if sortReply, e = s.BaseModules(ctx, arg.Pid, arg.PType, arg.Offset, arg.Ps); e != nil {
			log.Error("s.natDao.SortModules(%d) error(%v)", arg.Pid, e)
		}
		return
	})
	// 页面公共基础组件
	if arg.PType == v1.CommonPage || arg.PType == v1.FeedPage {
		eeg.Go(func(ctx context.Context) error {
			baseMIDsRly, e := s.BaseModules(ctx, arg.Pid, v1.BasePage, 0, -1)
			if e != nil {
				log.Error("s.BaseModules(%d) error(%v)", arg.Pid, e)
				return nil
			}
			if baseMIDsRly != nil {
				baseMIDs = baseMIDsRly.IDs
			}
			return nil
		})
		baseH := v1.FeedBaseModule
		if arg.PType == v1.CommonPage {
			baseH = v1.CommonBaseModule
		}
		//页面基础组件
		eeg.Go(func(ctx context.Context) error {
			pageBaseIDsRly, e := s.BaseModules(ctx, arg.Pid, int32(baseH), 0, -1)
			if e != nil {
				log.Error("s.BaseModules(%d) error(%v)", arg.Pid, e)
				return nil
			}
			if pageBaseIDsRly != nil {
				pageBaseIDs = pageBaseIDsRly.IDs
			}
			return nil
		})
	}
	if err = eeg.Wait(); err != nil {
		return
	}
	var modulesIDs []int64
	if sortReply != nil {
		res.Page = &v1.Page{HasMore: sortReply.HasMore, Offset: sortReply.Offset}
		modulesIDs = sortReply.IDs
	}
	if len(pageBaseIDs) > 0 {
		baseMIDs = append(baseMIDs, pageBaseIDs...)
	}
	if len(baseMIDs) > 0 {
		modulesIDs = append(modulesIDs, baseMIDs...)
	}
	if len(modulesIDs) == 0 {
		return
	}
	if modules, err = s.natDao.OnlineNativeModules(c, modulesIDs); err != nil {
		log.Error("s.natDao.OnlineNativeModules(%v) error(%v)", modulesIDs, err)
		err = nil
		return
	}
	if len(modules) == 0 {
		return
	}
	temp := s.ModuleDetail(c, modules)
	mul := make([]*v1.Module, 0, len(sortReply.IDs))
	for _, v := range sortReply.IDs {
		if _, k := temp[v]; !k {
			continue
		}
		if temp[v].NativeModule.IsOnline() {
			mul = append(mul, temp[v])
		}
	}
	res.Modules = mul
	for _, v := range baseMIDs {
		if _, k := temp[v]; !k {
			continue
		}
		if temp[v].NativeModule.IsOnline() {
			baseMul = append(baseMul, temp[v])
		}
	}
	res.Bases = baseMul
	res.FirstPage = firstPage
	return
}

// clickData 组装自定义点击区域配置.
func (s *Service) clickData(c context.Context, pid int64) (click *v1.Click, err error) {
	var (
		ids  []int64
		list map[int64]*v1.NativeClick
		area []*v1.NativeClick
	)
	if ids, err = s.natDao.NativeClickIDs(c, pid); err != nil {
		log.Error("s.natDao.NativeClickIDs(%d) error(%v)", pid, err)
		return
	}
	if len(ids) == 0 {
		return
	}
	if list, err = s.natDao.NativeClicks(c, ids); err != nil {
		return
	}
	if len(list) == 0 {
		return
	}
	area = make([]*v1.NativeClick, 0, len(list))
	for _, v := range list {
		if v.IsOnline() {
			area = append(area, v)
		}
	}
	click = &v1.Click{Areas: area}
	return
}

// dynamicData 动态列表配置信息 .
func (s *Service) dynamicData(c context.Context, pid int64) (res *v1.Dynamic, err error) {
	var (
		ids        []int64
		list       map[int64]*v1.NativeDynamicExt
		selectList []*v1.NativeDynamicExt
	)
	if ids, err = s.natDao.NativeDynamicIDs(c, pid); err != nil {
		log.Error("s.natDao.NativeDynamicIDs(%d) error(%v)", pid, err)
		return
	}
	if len(ids) == 0 {
		return
	}
	if list, err = s.natDao.NativeDynamics(c, ids); err != nil {
		return
	}
	if len(list) == 0 {
		return
	}
	selectList = make([]*v1.NativeDynamicExt, 0, len(list))
	for _, v := range list {
		if v.IsOnline() {
			selectList = append(selectList, v)
		}
	}
	sort.Slice(selectList, func(i, j int) bool { return selectList[i].ID < selectList[j].ID })
	res = &v1.Dynamic{SelectList: selectList}
	return
}

// videoData 活动视频信息.
func (s *Service) videoData(c context.Context, pid int64) (res *v1.VideoAct, err error) {
	var (
		ids      []int64
		list     map[int64]*v1.NativeVideoExt
		SortList []*v1.NativeVideoExt
	)
	if ids, err = s.natDao.NativeVideoIDs(c, pid); err != nil {
		log.Error("s.natDao.NativeVideoIDs(%d) error(%v)", pid, err)
		return
	}
	if len(ids) == 0 {
		return
	}
	if list, err = s.natDao.NativeVideos(c, ids); err != nil {
		return
	}
	if len(list) == 0 {
		return
	}
	SortList = make([]*v1.NativeVideoExt, 0, len(list))
	for _, id := range ids {
		if _, ok := list[id]; !ok {
			continue
		}
		if list[id].IsOnline() {
			SortList = append(SortList, list[id])
		}
	}
	res = &v1.VideoAct{SortList: SortList}
	return
}

// actData .
func (s *Service) actData(c context.Context, pid int64) (res *v1.Act, err error) {
	var (
		pids []int64
		Acts map[int64]*v1.NativePage
		List []*v1.NativePage
	)
	if pids, err = s.natDao.NativeActIDs(c, pid); err != nil {
		log.Error("s.natDao.NativeActIDs(%d) error(%v)", pid, err)
		return
	}
	if len(pids) == 0 {
		return
	}
	if Acts, err = s.natDao.NativePages(c, pids); err != nil {
		return
	}
	if len(Acts) == 0 {
		return
	}
	List = make([]*v1.NativePage, 0, len(Acts))
	for _, id := range pids {
		if aVal, ok := Acts[id]; !ok || aVal == nil {
			continue
		}
		List = append(List, Acts[id])
	}
	res = &v1.Act{List: List}
	return
}

func (s *Service) actCapsuleData(c context.Context, moduleID int64) (*v1.ActPage, error) {
	pageIDs, err := s.natDao.NativeActIDs(c, moduleID)
	if err != nil {
		log.Errorc(c, "Fail to get nativeActIDs, moduleID=%+v error=%+v", moduleID, err)
		return nil, err
	}
	list := make([]*v1.ActPageItem, 0, len(pageIDs))
	for _, pid := range pageIDs {
		list = append(list, &v1.ActPageItem{PageID: pid})
	}
	return &v1.ActPage{List: list}, nil
}

func (s *Service) partData(c context.Context, pid int64) (res *v1.Participation, err error) {
	var (
		pids  []int64
		Parts map[int64]*v1.NativeParticipationExt
		List  []*v1.NativeParticipationExt
		pt    *dynmdl.ModuleIDsReply
	)
	if pt, err = s.natDao.PartPids(c, pid, 0, _maxPartNum); err != nil {
		log.Error("s.natDao.PartPids(%d) error(%v)", pid, err)
		return
	}
	if pt == nil {
		return
	}
	for _, v := range pt.IDs {
		if v <= 0 {
			continue
		}
		pids = append(pids, v)
	}
	if len(pids) == 0 {
		return
	}
	if Parts, err = s.natDao.NativePart(c, pids); err != nil {
		log.Error("s.natDao.NativePart(pids:%+v) error(%v)", pids, err)
		return
	}
	if len(Parts) == 0 {
		return
	}
	List = make([]*v1.NativeParticipationExt, 0, len(Parts))
	for _, id := range pids {
		if _, ok := Parts[id]; !ok {
			continue
		}
		List = append(List, Parts[id])
	}
	res = &v1.Participation{List: List}
	return
}

// RecommendData return recommend list
func (s *Service) RecommendData(c context.Context, moduleID int64, mixType int32) (res *v1.Recommend, err error) {
	var (
		list   map[int64]*v1.NativeMixtureExt
		mixIds *dynmdl.ModuleIDsReply
	)
	if mixIds, err = s.natDao.NatMixIDs(c, moduleID, mixType, 0, _rcmdOffset); err != nil {
		log.Error("s.natDao.NatMixIDs(%d) error(%v)", moduleID, err)
		return
	}
	if mixIds == nil {
		return
	}
	var ids []int64
	for _, v := range mixIds.IDs {
		if v <= 0 {
			continue
		}
		ids = append(ids, v)
	}
	res = &v1.Recommend{}
	idLen := len(ids)
	if idLen == 0 {
		return
	}
	if list, err = s.natDao.NativeMixtures(c, ids); err != nil {
		return
	}
	if len(list) == 0 {
		return
	}
	res.List = make([]*v1.NativeMixtureExt, 0, len(list))
	for _, id := range ids {
		if _, ok := list[id]; !ok {
			continue
		}
		if list[id].IsOnline() && list[id].MType == mixType {
			res.List = append(res.List, list[id])
		}
	}
	return
}

func (s *Service) CarouselData(c context.Context, natModule *v1.NativeModule) (*v1.Carousel, error) {
	mixIDs, err := s.NatAllMixIDs(c, natModule.ID, 0, _rcmdOffset)
	if err != nil {
		log.Error("Fail to get natAllMixIDs, moduleID=%d error=%+v", natModule.ID, err)
		return nil, err
	}
	if mixIDs == nil || len(mixIDs.IDs) == 0 {
		return &v1.Carousel{}, nil
	}
	mixtures, err := s.natDao.NativeMixtures(c, mixIDs.IDs)
	if err != nil {
		log.Error("Fail to get nativeMixtures, ids=%+v error=%+v", mixIDs.IDs, err)
		return nil, err
	}
	if len(mixtures) == 0 {
		return &v1.Carousel{}, nil
	}
	carousel := &v1.Carousel{
		List: make([]*v1.NativeMixtureExt, 0, len(mixtures)),
	}
	for _, id := range mixIDs.IDs {
		mixture, ok := mixtures[id]
		if !ok || mixture == nil || !mixture.IsOnline() {
			continue
		}
		if (natModule.IsCarouselImg() && mixtures[id].MType == v1.MixCarouselImg) ||
			(natModule.IsCarouselWord() && mixtures[id].MType == v1.MixCarouselWord) {
			carousel.List = append(carousel.List, mixture)
		}
	}
	return carousel, nil
}

func (s *Service) IconData(c context.Context, natModule *v1.NativeModule) (*v1.Icon, error) {
	mixIDs, err := s.NatAllMixIDs(c, natModule.ID, 0, _rcmdOffset)
	if err != nil {
		log.Error("Fail to get natAllMixIDs, moduleID=%d error=%+v", natModule.ID, err)
		return nil, err
	}
	if len(mixIDs.IDs) == 0 {
		return &v1.Icon{}, nil
	}
	mixtures, err := s.natDao.NativeMixtures(c, mixIDs.IDs)
	if err != nil {
		log.Error("Fail to get nativeMixtures, ids=%+v error=%+v", mixIDs.IDs, err)
		return nil, err
	}
	if len(mixtures) == 0 {
		return &v1.Icon{}, nil
	}
	icon := &v1.Icon{
		List: make([]*v1.NativeMixtureExt, 0, len(mixtures)),
	}
	for _, id := range mixIDs.IDs {
		mixture, ok := mixtures[id]
		if !ok || mixture == nil || !mixture.IsOnline() {
			continue
		}
		icon.List = append(icon.List, mixture)
	}
	return icon, nil
}

// NatInfoFromForeign .
func (s *Service) NatInfoFromForeign(c context.Context, fids []int64, pageType int64, content map[string]string) (rs map[int64]*v1.NativePage, err error) {
	var (
		fidMap map[int64]int64
		ids    []int64
		pages  map[int64]*v1.NativePage
	)
	if fidMap, err = s.natDao.NativeForeigns(c, fids, pageType); err != nil {
		log.Error("s.natDao.NativeForeigns(%v %d) error(%v)", fids, pageType, err)
		return
	}
	for _, v := range fidMap {
		if v > 0 {
			ids = append(ids, v)
		}
	}
	if len(ids) == 0 {
		return
	}
	if pages, err = s.natDao.NativePages(c, ids); err != nil {
		log.Error("s.natDao.NativePages(%v) error(%v)", ids, err)
		return
	}
	rs = make(map[int64]*v1.NativePage)
	for k, v := range fidMap {
		//开启白名单逻辑的页面也不下发
		if _, ok := pages[v]; ok && pages[v].IsOnline() && pages[v].IsAttrWhiteSwitch() != v1.AttrModuleYes {
			rs[k] = pages[v]
		}
	}
	for _, v := range rs {
		if v.SkipURL == "" {
			continue
		}
		for k, cv := range content {
			v.SkipURL = strings.ReplaceAll(v.SkipURL, fmt.Sprintf("__%s__", k), cv)
		}
	}
	return
}

// NatAllMixIDs .
func (s *Service) NatAllMixIDs(c context.Context, moduleID, offset, ps int64) (*dynmdl.ModuleIDsReply, error) {
	var (
		end int64
	)
	if ps <= 0 {
		end = -1
	} else {
		end = offset + ps - 1
	}
	rly, err := s.natDao.NatAllMixIDs(c, moduleID, offset, end)
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return &dynmdl.ModuleIDsReply{}, nil
	}
	moduleReply := &dynmdl.ModuleIDsReply{Count: rly.Count}
	moduleReply.Count = rly.Count
	for _, v := range rly.IDs {
		if v <= 0 {
			continue
		}
		moduleReply.IDs = append(moduleReply.IDs, v)
	}
	return moduleReply, nil
}

// ModuleMixExt .
func (s *Service) ModuleMixExt(c context.Context, moduleID, offset, ps int64, mType int32) (res *v1.ModuleMixExtReply, err error) {
	var (
		list   map[int64]*v1.NativeMixtureExt
		mixIds *dynmdl.ModuleIDsReply
		end    = offset + ps - 1
	)
	if mixIds, err = s.natDao.NatMixIDs(c, moduleID, mType, offset, end); err != nil {
		log.Error("s.natDao.MixCache(%d) error(%v)", moduleID, err)
		return
	}
	if mixIds == nil {
		return
	}
	var xids []int64
	for _, v := range mixIds.IDs {
		if v <= 0 {
			continue
		}
		xids = append(xids, v)
	}
	res = &v1.ModuleMixExtReply{Total: mixIds.Count}
	idLen := len(xids)
	if idLen == 0 {
		return
	}
	if list, err = s.natDao.NativeMixtures(c, xids); err != nil {
		return
	}
	if len(list) == 0 {
		return
	}
	res.List = make([]*v1.NativeMixtureExt, 0, len(list))
	for _, id := range xids {
		if _, ok := list[id]; !ok {
			continue
		}
		if list[id].IsOnline() {
			res.List = append(res.List, list[id])
		}
	}
	res.Offset = offset + int64(idLen)
	if res.Offset < mixIds.Count {
		res.HasMore = 1
	}
	return
}

// ModuleMixExts .
func (s *Service) ModuleMixExts(c context.Context, moduleID, offset, ps int64) (res *v1.ModuleMixExtsReply, err error) {
	var (
		list   map[int64]*v1.NativeMixtureExt
		mixIds *dynmdl.ModuleIDsReply
	)
	if mixIds, err = s.NatAllMixIDs(c, moduleID, offset, ps); err != nil {
		log.Error("s.natDao.NatAllMixIDs(%d) error(%v)", moduleID, err)
		return
	}
	if mixIds == nil {
		res = &v1.ModuleMixExtsReply{}
		return
	}
	res = &v1.ModuleMixExtsReply{Total: mixIds.Count}
	idLen := len(mixIds.IDs)
	if idLen == 0 {
		return
	}
	if list, err = s.natDao.NativeMixtures(c, mixIds.IDs); err != nil {
		return
	}
	if len(list) == 0 {
		return
	}
	res.List = make([]*v1.NativeMixtureExt, 0, len(list))
	for _, id := range mixIds.IDs {
		if _, ok := list[id]; !ok {
			continue
		}
		if list[id].IsOnline() {
			res.List = append(res.List, list[id])
		}
	}
	res.Offset = offset + int64(idLen)
	if res.Offset < mixIds.Count {
		res.HasMore = 1
	}
	return
}

func (s *Service) inlineFirstPage(c context.Context, pid int64) (*v1.FirstPage, error) {
	parentPages, e := s.natDao.NativePages(c, []int64{pid})
	if e != nil { //父页面获取失败
		log.Error("s.natDao.NativePages(%d) error(%v)", pid, e)
		return nil, e
	}
	if _, ok := parentPages[pid]; !ok {
		return nil, ecode.NativePageOffline
	}
	rly := &v1.FirstPage{Item: parentPages[pid]}
	if parentPages[pid].IsAttrWhiteSwitch() == v1.AttrModuleYes { //开启白名单逻辑
		if rly.Ext, e = s.natDao.NativeExtend(c, pid); e != nil { //父页面获取失败
			log.Error("s.natDao.NativeExtend(%d) error(%v)", pid, e)
			return nil, e
		}
	}
	return rly, nil
}

// BaseModules .
func (s *Service) BaseModules(c context.Context, nid int64, pType int32, offset, ps int64) (res *lmdl.SortModule, err error) {
	var (
		moduleReply *dynmdl.ModuleIDsReply
		end         int64
	)
	res = &lmdl.SortModule{HasMore: 0}
	if nid == 0 {
		return
	}
	if ps <= 0 {
		end = -1
	} else {
		end = offset + ps - 1
	}
	if moduleReply, err = s.natDao.ModuleIDs(c, nid, pType, offset, end); err != nil {
		log.Error("s.natDao.ModuleCache(%d) error(%v)", nid, err)
		return
	}
	if moduleReply == nil {
		return
	}
	idLen := len(moduleReply.IDs)
	for _, v := range moduleReply.IDs {
		if v > 0 { //过滤掉default值
			res.IDs = append(res.IDs, v)
		}
	}
	res.Offset = offset + int64(idLen)
	if res.Offset < moduleReply.Count {
		res.HasMore = 1
	}
	return
}

func (s *Service) BaseConfig(c context.Context, arg *v1.BaseConfigReq) (res *v1.BaseConfigReply, err error) {
	var (
		natPages    map[int64]*v1.NativePage
		modules     map[int64]*v1.NativeModule
		baseMIDsRly *lmdl.SortModule
	)
	if natPages, err = s.natDao.NativePages(c, []int64{arg.Pid}); err != nil {
		log.Error("s.natDao.NativePages(%d) error(%v)", arg.Pid, err)
		return
	}
	// inlineact,menu,botton 类型下线后也支持获取页面信息，容错逻辑
	if _, ok := natPages[arg.Pid]; !ok || (!natPages[arg.Pid].IsOnline() && !natPages[arg.Pid].IsTabAct()) {
		err = ecode.NativePageOffline
		return
	}
	res = &v1.BaseConfigReply{NativePage: natPages[arg.Pid], Offset: arg.Offset}
	eg := errgroup.WithContext(c)
	//需要获取最上层页面cache start
	var firstPage *v1.FirstPage
	if natPages[arg.Pid].IsInlineAct() {
		if natPages[arg.Pid].FirstPid > 0 { //内嵌页存在父id的数据
			eg.Go(func(ctx context.Context) (e error) {
				firstPage, e = s.inlineFirstPage(ctx, natPages[arg.Pid].FirstPid)
				if e != nil {
					log.Error("s.inlineFirstPage(%d) error(%v)", arg.Pid, e)
				}
				return
			})
		}
	} else {
		eg.Go(func(ctx context.Context) error {
			firstPage = &v1.FirstPage{Item: natPages[arg.Pid]}
			if natPages[arg.Pid].IsAttrWhiteSwitch() == v1.AttrModuleYes { //开启了白名单逻辑
				extPage, e := s.natDao.NativeExtend(ctx, arg.Pid)
				if e != nil { //父页面获取失败
					log.Error("s.natDao.NativeExtend(%d) error(%v)", arg.Pid, e)
					return e
				}
				firstPage.Ext = extPage
			}
			return nil
		})
	}
	//需要获取最上层页面cache end
	eg.Go(func(ctx context.Context) error {
		if rly, err := s.BaseModules(ctx, arg.Pid, arg.PType, arg.Offset, arg.Ps); err == nil {
			baseMIDsRly = rly
		}
		return nil
	})
	var baseModuleIDs []int64
	eg.Go(func(ctx context.Context) error {
		if rly, err := s.BaseModules(ctx, arg.Pid, v1.CommonBaseModule, 0, -1); err == nil && rly != nil {
			baseModuleIDs = rly.IDs
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return res, nil
	}
	res.FirstPage = firstPage
	if baseMIDsRly == nil {
		return
	}
	moduleIDs := append(baseMIDsRly.IDs, baseModuleIDs...)
	res.Offset = baseMIDsRly.Offset
	res.HasMore = baseMIDsRly.HasMore
	if modules, err = s.natDao.OnlineNativeModules(c, moduleIDs); err != nil {
		log.Error("s.natDao.OnlineNativeModules(%v) error(%v)", moduleIDs, err)
		err = nil
		return
	}
	if len(modules) == 0 {
		return
	}
	temp := s.ModuleDetail(c, modules)
	for _, v := range baseMIDsRly.IDs {
		if _, k := temp[v]; !k {
			continue
		}
		if temp[v].NativeModule.IsOnline() {
			res.Bases = append(res.Bases, temp[v])
		}
	}
	for _, id := range baseModuleIDs {
		if module, ok := temp[id]; ok && module.NativeModule.IsOnline() {
			res.BaseModules = append(res.BaseModules, temp[id])
		}
	}
	return
}

// ModuleDetail 组件信息组装.
// nolint:gocognit
func (s *Service) ModuleDetail(c context.Context, modules map[int64]*v1.NativeModule) (temp map[int64]*v1.Module) {
	temp = make(map[int64]*v1.Module)
	mutex := sync.Mutex{}
	eg := errgroup.WithContext(c)
	for _, v := range modules {
		if !v.IsOnline() {
			continue
		}
		natModule := v
		switch {
		case v.IsClick(), v.IsVote(), v.IsBaseBottomButton():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				if a.Click, e = s.clickData(ctx, natModule.ID); e != nil {
					log.Error("s.clickData(%d) error(%v)", natModule.ID, e)
					e = nil
					return
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsDynamic(), v.IsResourceDyn():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				if a.Dynamic, e = s.dynamicData(ctx, natModule.ID); e != nil {
					log.Error("s.dynamicData(%d) error(%v)", natModule.ID, e)
					e = nil
					return
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsVideo(), v.IsVideoAct(), v.IsResourceAct(), v.IsNewVideoAct():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				if a.VideoAct, e = s.videoData(ctx, natModule.ID); e != nil {
					log.Error("s.videoData(%d) error(%v)", natModule.ID, e)
					e = nil
					return
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsResourceOrigin():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				if natModule.ConfUnmarshal().RdbType == v1.RDBLive {
					if a.VideoAct, e = s.videoData(ctx, natModule.ID); e != nil {
						log.Error("s.videoData(%d) error(%v)", natModule.ID, e)
						e = nil
						return
					}
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsAct():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				if a.Act, e = s.actData(ctx, natModule.ID); e != nil {
					log.Error("s.actData(%d) error(%v)", natModule.ID, e)
					e = nil
					return
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsActCapsule():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				if a.ActPage, e = s.actCapsuleData(ctx, natModule.ID); e != nil {
					log.Error("Fail to get act_capsule data, moduleID=%+v error=%+v", natModule.ID, e)
					e = nil
					return
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsVideoDyn(), v.IsVideoAvid(), v.IsBanner(), v.IsSingleDyn(), v.IsStatement(), v.IsNavigation(), v.IsBaseHead(),
			v.IsResourceID(), v.IsLive(), v.IsNewVideoDyn(), v.IsNewVideoID(), v.IsEditor(), v.IsEditorOrigin(), v.IsResourceRole(), v.IsTimelineIDs(), v.IsTimelineSource(), v.IsOgvSeasonID(), v.IsOgvSeasonSource(), v.IsReply(),
			v.IsCarouselSource(), v.IsRcmdSource(), v.IsRcmdVerticalSource(), v.IsProgress(), v.IsBaseHoverButton(),
			v.IsNewactHeaderModule(), v.IsNewactAwardModule(), v.IsNewactStatementModule(), v.IsMatchMedal():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsPart():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				if a.Participation, e = s.partData(ctx, natModule.ID); e != nil {
					log.Error("s.partData(%d) error(%v)", natModule.ID, e)
					e = nil
					return
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsInlineTab():
			eg.Go(func(ctx context.Context) error {
				tabRly, e := s.ModuleMixExts(ctx, natModule.ID, 0, -1)
				if e != nil {
					log.Error("s.partData(%d) error(%v)", natModule.ID, e)
					return nil
				}
				if tabRly == nil || len(tabRly.List) == 0 {
					return nil
				}
				a := &v1.Module{NativeModule: natModule, InlineTab: &v1.InlineTab{List: tabRly.List}}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return nil
			})
		case v.IsGame():
			eg.Go(func(ctx context.Context) error {
				tabRly, e := s.ModuleMixExts(ctx, natModule.ID, 0, -1)
				if e != nil {
					log.Error("s.partData(%d) error(%v)", natModule.ID, e)
					return nil
				}
				if tabRly == nil || len(tabRly.List) == 0 {
					return nil
				}
				a := &v1.Module{NativeModule: natModule, Game: &v1.Game{List: tabRly.List}}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return nil
			})
		case v.IsReserve():
			eg.Go(func(ctx context.Context) error {
				tabRly, e := s.ModuleMixExts(ctx, natModule.ID, 0, -1)
				if e != nil {
					log.Error("s.partData(%d) error(%v)", natModule.ID, e)
					return nil
				}
				if tabRly == nil || len(tabRly.List) == 0 {
					return nil
				}
				a := &v1.Module{NativeModule: natModule, Reserve: &v1.Reserve{List: tabRly.List}}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return nil
			})
		case v.IsRecommend(), v.IsRcmdVertical():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				var mixType int32 = v1.MixTypeRcmd
				if natModule.IsRcmdVertical() {
					mixType = v1.MixRcmdVertical
				}
				if a.Recommend, e = s.RecommendData(ctx, natModule.ID, mixType); e != nil {
					log.Error("s.RcmdData(%d) error(%v)", natModule.ID, e)
					e = nil
					return
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsCarouselImg(), v.IsCarouselWord():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				if a.Carousel, e = s.CarouselData(ctx, natModule); e != nil {
					log.Error("Fail to get carouselData, natModule=%+v error=%+v", natModule, e)
					e = nil
					return
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsIcon():
			eg.Go(func(ctx context.Context) (e error) {
				a := &v1.Module{NativeModule: natModule}
				if a.Icon, e = s.IconData(ctx, natModule); e != nil {
					log.Error("Fail to get iconData, natModule=%+v error=%+v", natModule, e)
					e = nil
					return
				}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return
			})
		case v.IsSelect():
			eg.Go(func(ctx context.Context) error {
				tabRly, e := s.ModuleMixExts(ctx, natModule.ID, 0, -1)
				if e != nil {
					log.Error("s.partData(%d) error(%v)", natModule.ID, e)
					return nil
				}
				if tabRly == nil || len(tabRly.List) == 0 {
					return nil
				}
				a := &v1.Module{NativeModule: natModule, Select: &v1.Select{List: tabRly.List}}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return nil
			})
		case v.IsMatchEvent():
			eg.Go(func(ctx context.Context) error {
				events, e := s.ModuleMixExts(ctx, natModule.ID, 0, -1)
				if e != nil {
					log.Error("s.partData(%d) error(%v)", natModule.ID, e)
					return nil
				}
				if events == nil || len(events.List) == 0 {
					return nil
				}
				a := &v1.Module{NativeModule: natModule, MatchEvent: &v1.MatchEvent{List: events.List}}
				mutex.Lock()
				temp[natModule.ID] = a
				mutex.Unlock()
				return nil
			})
		}
	}
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "ModuleDetail eg.Wait() failed, error=%+v", err)
	}
	return
}

// NativePages .
func (s *Service) NativePages(c context.Context, ids []int64) (map[int64]*v1.NativePage, error) {
	pages, err := s.natDao.NativePages(c, ids)
	if err != nil {
		log.Error("s.natDao.NativePages ids(%v) error(%v)", ids, err)
		return nil, err
	}
	rly := make(map[int64]*v1.NativePage)
	for k, v := range pages {
		if v != nil && v.ID > 0 && v.IsOnline() {
			rly[k] = v
		}
	}
	return rly, nil
}

func (s *Service) NativeAllPages(c context.Context, ids []int64) (map[int64]*v1.NativePage, error) {
	pages, err := s.natDao.NativePages(c, ids)
	if err != nil {
		log.Error("Fail to get nativePages, pageIDs=%+v error=%+v", ids, err)
		return nil, err
	}
	return pages, nil
}

func (s *Service) NativePageCards(c context.Context, ids []int64, build int32, mobiApp, platform string) (map[int64]*v1.NativePageCard, error) {
	uniqueIDs := make(map[int64]struct{})
	var lastIDs []int64
	for _, v := range ids {
		if _, ok := uniqueIDs[v]; ok {
			continue
		}
		uniqueIDs[v] = struct{}{}
		lastIDs = append(lastIDs, v)
	}
	pages, err := s.NativePages(c, lastIDs)
	if err != nil {
		log.Error("s.NativePages ids(%v) error(%v)", ids, err)
		return nil, err
	}
	var tabIDs []int64
	for _, val := range pages {
		if val == nil {
			continue
		}
		//请求来自于iphone和android粉版，iphone build号大于10020， android build好大于等于6020000，并且该活动配置的h5跳链为空
		if val.SkipURL == "" && ((mobiApp == "iphone" && build > 10020) || (mobiApp == "android" && build > 6020000)) {
			tabIDs = append(tabIDs, val.ID)
		}
	}
	//没有需要处理的跳转地址
	var tabRly map[int64]*v1.PagesTab
	if len(tabIDs) > 0 {
		if tabRly, err = s.nativeTab(c, tabIDs, v1.TopicActType); err != nil { //降级处理,错误忽略
			log.Error("s.nativeTab(%v) error(%v)", tabIDs, err)
		}
	}
	cardRly := make(map[int64]*v1.NativePageCard)
	for _, val := range pages {
		if val == nil {
			continue
		}
		// 分享title为空则取话题名
		if val.ShareCaption == "" {
			val.ShareCaption = val.Title
		}
		if val.SkipURL == "" {
			if tv, ok := tabRly[val.ID]; ok && tv != nil {
				val.SkipURL = tv.Url
			}
		}
		//默认跳转地址为na页
		if val.SkipURL == "" {
			val.SkipURL = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d", val.ID, time.Now().Unix())
		}
		if val.ShareURL == "" { //分享使用h5跳转地址
			val.ShareURL = val.SkipURL
		}
		if val.PcURL == "" {
			val.PcURL = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d", val.ID, time.Now().Unix())
		}
		cardRly[val.ID] = &v1.NativePageCard{
			Id:           val.ID,
			Title:        val.Title,
			Type:         val.Type,
			ForeignID:    val.ForeignID,
			ShareCaption: val.ShareCaption,
			ShareImage:   val.ShareImage,
			ShareTitle:   val.ShareTitle,
			SkipURL:      val.SkipURL,
			PcURL:        val.PcURL,
			ShareURL:     val.ShareURL,
			RelatedUid:   val.RelatedUid,
			State:        val.State,
		}
	}
	return cardRly, nil
}

func (s *Service) NativeAllPageCards(c context.Context, ids []int64) (map[int64]*v1.NativePageCard, error) {
	pageIDs := UniqueArray(ids)
	pages, err := s.natDao.NativePages(c, pageIDs)
	if err != nil {
		return nil, err
	}
	tabs := s.nativeTabsFromPages(c, pages)
	cardRly := make(map[int64]*v1.NativePageCard, len(pageIDs))
	for _, val := range pages {
		if val == nil {
			continue
		}
		// 分享title为空则取话题名
		if val.ShareCaption == "" {
			val.ShareCaption = val.Title
		}
		if val.IsOnline() && val.SkipURL == "" {
			if tv, ok := tabs[val.ID]; ok && tv != nil {
				val.SkipURL = tv.Url
			}
		}
		//默认跳转地址为na页
		if val.SkipURL == "" {
			val.SkipURL = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d", val.ID, time.Now().Unix())
		}
		if val.ShareURL == "" { //分享使用h5跳转地址
			val.ShareURL = val.SkipURL
		}
		if val.PcURL == "" {
			val.PcURL = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d?ts=%d", val.ID, time.Now().Unix())
		}
		cardRly[val.ID] = &v1.NativePageCard{
			Id:           val.ID,
			Title:        val.Title,
			Type:         val.Type,
			ForeignID:    val.ForeignID,
			ShareCaption: val.ShareCaption,
			ShareImage:   val.ShareImage,
			ShareTitle:   val.ShareTitle,
			SkipURL:      val.SkipURL,
			PcURL:        val.PcURL,
			ShareURL:     val.ShareURL,
			RelatedUid:   val.RelatedUid,
			State:        val.State,
		}
	}
	return cardRly, nil
}

func (s *Service) nativeTabsFromPages(c context.Context, pages map[int64]*v1.NativePage) map[int64]*v1.PagesTab {
	var tabIDs []int64
	for _, val := range pages {
		if val == nil {
			continue
		}
		if val.SkipURL == "" && val.IsOnline() {
			tabIDs = append(tabIDs, val.ID)
		}
	}
	if len(tabIDs) == 0 {
		return map[int64]*v1.PagesTab{}
	}
	tabRly, _ := s.nativeTab(c, tabIDs, v1.TopicActType)
	return tabRly
}

// NativePagesExt .
func (s *Service) NativePagesExt(c context.Context, ids []int64) (map[int64]*v1.NativePageExt, error) {
	eg := errgroup.WithContext(c)
	// 获取page
	var pages map[int64]*v1.NativePage
	eg.Go(func(ctx context.Context) error {
		var e error
		if pages, e = s.natDao.NativePages(ctx, ids); e != nil {
			log.Error("s.natDao.NativePages ids(%v) error(%v)", ids, e)
		}
		return e
	})
	////获取ext
	var exts map[int64]*v1.NativePageDyn
	eg.Go(func(ctx context.Context) error {
		var e error
		if exts, e = s.natDao.NativePagesExt(ctx, ids); e != nil {
			log.Error("s.natDao.NativePagesExt ids(%v) error(%v)", ids, e)
			return nil
		}
		//降级处理
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	rly := make(map[int64]*v1.NativePageExt)
	for _, v := range pages {
		if v == nil || v.ID == 0 || !v.IsOnline() || v.IsAttrWhiteSwitch() == v1.AttrModuleYes { //开启白名单逻辑
			continue
		}
		tmp := &v1.NativePageExt{Item: v}
		if ev, ok := exts[v.ID]; ok && ev != nil {
			tmp.DynExt = ev
		}
		rly[v.ID] = tmp
	}
	return rly, nil
}

// NativeValidPagesExt .
func (s *Service) NativeValidPagesExt(c context.Context, req *v1.NativeValidPagesExtReq) (map[int64]*v1.NativePageExt, error) {
	// 获取page_ids
	pageIDs, err := s.natDao.NatIDsByActType(c, req.ActType)
	if err != nil {
		log.Error("Fail to get NatIDsByActType, actType=%d err=%+v", req.ActType, err)
		return nil, err
	}
	maxPageIDsLen := 200
	if len(pageIDs) > maxPageIDsLen {
		pageIDs = pageIDs[:maxPageIDsLen]
	}
	// 获取page_ext
	exts, err := s.natDao.NativePagesExt(c, pageIDs)
	if err != nil {
		log.Error("Fail to get NativePagesExt, pageIDs=%+v err=%+v", pageIDs, err)
		return nil, err
	}
	// 获取上榜时间内的pageIDs
	validPageIDs := getValidPageIDs(exts)
	validLen := 50
	if len(validPageIDs) > validLen {
		validPageIDs = validPageIDs[:validLen]
	}
	// 获取page详情
	pages, err := s.natDao.NativePages(c, validPageIDs)
	if err != nil {
		log.Error("Fail to get NativePages, pageIDs=%+v err=%+v", validPageIDs, pages)
		return nil, err
	}
	rly := make(map[int64]*v1.NativePageExt)
	for _, v := range pages {
		if v == nil || v.ID == 0 || !v.IsOnline() || v.IsAttrWhiteSwitch() == v1.AttrModuleYes { //开启白名单逻辑
			continue
		}
		tmp := &v1.NativePageExt{Item: v}
		if ev, ok := exts[v.ID]; ok && ev != nil {
			tmp.DynExt = ev
		}
		rly[v.ID] = tmp
	}
	return rly, nil
}

// NativePages .
func (s *Service) NativePage(c context.Context, id int64) (*v1.NativePage, error) {
	pages, err := s.natDao.NativePages(c, []int64{id})
	if err != nil {
		log.Error("s.natDao.NativePages id(%d) error(%v)", id, err)
		return nil, err
	}
	if pv, ok := pages[id]; ok {
		return pv, nil
	}
	return nil, nil
}

// NatTabModules .
func (s *Service) NatTabModules(c context.Context, req *v1.NatTabModulesReq) (*v1.NatTabModulesReply, error) {
	eg := errgroup.WithCancel(c)
	var tabRly *v1.NativeActTab
	eg.Go(func(ctx context.Context) error {
		tabs, e := s.natDao.NativeTabs(c, []int64{req.TabID})
		if e != nil {
			log.Error("s.natDao.NativeTa  %d error(%v)", req.TabID, e)
			return e
		}
		if v, ok := tabs[req.TabID]; ok {
			tabRly = v
		}
		return nil
	})
	var mIDs []int64
	eg.Go(func(ctx context.Context) (e error) {
		if mIDs, e = s.natDao.NativeTabSort(c, req.TabID); e != nil {
			log.Error("s.natDao.NativeTabSort(%d) error(%v)", req.TabID, e)
		}
		return
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if tabRly == nil {
		return &v1.NatTabModulesReply{}, nil
	}
	modus, err := s.natDao.NativeTabModules(c, mIDs)
	if err != nil {
		log.Error("s.natDao.NativeTabModules(%v) error(%v)", mIDs, err)
		return nil, err
	}
	rly := &v1.NatTabModulesReply{Tab: tabRly}
	for _, v := range mIDs {
		if _, ok := modus[v]; !ok {
			continue
		}
		rly.List = append(rly.List, modus[v])
	}
	return rly, nil
}

func (s *Service) nativeTab(c context.Context, ids []int64, category int32) (map[int64]*v1.PagesTab, error) {
	// 获取绑定信息
	bind, e := s.natDao.NativeTabBind(c, ids, category)
	if e != nil {
		log.Error("s.natDao.NativeTabBind ids(%v) error(%v)", ids, e)
		return nil, e
	}
	var (
		moduleIDs []int64
		pageIDs   = make(map[int64]int64)
	)
	for _, v := range bind {
		if _, ok := pageIDs[v]; !ok { //去重逻辑
			moduleIDs = append(moduleIDs, v)
		}
		pageIDs[v] = v
	}
	if len(moduleIDs) == 0 {
		return make(map[int64]*v1.PagesTab), nil
	}
	// 获取module信息
	moduleRly, e := s.natDao.NativeTabModules(c, moduleIDs)
	if e != nil {
		log.Error("s.natDao.NativeTabModules ids(%v) error(%v)", ids, e)
		return nil, e
	}
	moduleToTab := make(map[int64]int64)
	tabIDs := make([]int64, 0)
	for _, v := range moduleRly {
		if _, ok := moduleToTab[v.TabID]; !ok { //去重逻辑
			tabIDs = append(tabIDs, v.TabID)
		}
		moduleToTab[v.TabID] = v.TabID
	}
	//获取tab信息
	tabRly, e := s.natDao.NativeTabs(c, tabIDs)
	if e != nil {
		log.Error("s.natDao.NativeTabs ids(%v) error(%v)", ids, e)
		return nil, e
	}
	// 过滤出有效的tab
	onlineTab := make(map[int64]*v1.NativeActTab)
	nowTime := time.Now().Unix()
	for k, v := range tabRly {
		if v != nil && v.IsOnline() && int64(v.Stime) > 0 && int64(v.Stime) <= nowTime && (int64(v.Etime) >= nowTime || int64(v.Etime) <= 0) {
			onlineTab[k] = v
		}
	}
	res := make(map[int64]*v1.PagesTab)
	for _, v := range ids {
		if _, ok := bind[v]; !ok {
			continue
		}
		//module 是否有效
		if mVal, ok := moduleRly[bind[v]]; !ok || mVal == nil || mVal.Pid != v || mVal.Category != category {
			continue
		}
		// tab 是否有效
		if tVal, ok := onlineTab[moduleRly[bind[v]].TabID]; ok && tVal != nil {
			url := fmt.Sprintf("https://www.bilibili.com/blackboard/group/%d?tab_id=%d&tab_module_id=%d&ts=%d", v, tVal.ID, bind[v], time.Now().Unix())
			res[v] = &v1.PagesTab{TabID: tVal.ID, TabModuleID: bind[v], PageID: v, Url: url}
		}
	}
	return res, nil
}

// NativePagesTab .
func (s *Service) NativePagesTab(c context.Context, ids []int64, category int32) (*v1.NativePagesTabReply, error) {
	// 获取绑定信息
	list, e := s.nativeTab(c, ids, category)
	if e != nil {
		log.Error("s.nativeTab(%v) error(%v)", ids, e)
		return nil, e
	}
	return &v1.NativePagesTabReply{List: list}, nil
}

func (s *Service) delCacheNatIDsByActType(eg *errgroup.Group, actType int64) {
	// 更新上榜有效期内的缓存
	eg.Go(func(ctx context.Context) error {
		err := s.natDao.DelCacheNatIDsByActType(ctx, actType)
		if err != nil {
			log.Error("Fail to delete natIDsByActType cache, actType=%d err=%+v", actType, err)
			return err
		}
		return nil
	})
}

func (s *Service) GetNatProgressParams(c context.Context, req *v1.GetNatProgressParamsReq) (*v1.GetNatProgressParamsReply, error) {
	reply, err := s.natDao.CachePageProgressParams(c, req.PageID)
	if err != nil {
		log.Error("Fail to get pageProgressParams cache, pageID=%+v error=%+v", req.PageID, err)
		return nil, err
	}
	return &v1.GetNatProgressParamsReply{List: reply}, nil
}

func getValidPageIDs(exts map[int64]*v1.NativePageDyn) []int64 {
	var validPageIDs []int64
	now := time.Now()
	for _, v := range exts {
		if v.Stime <= 0 || v.Stime.Time().AddDate(0, 0, int(v.Validity)).Before(now) || v.Stime.Time().After(now) {
			continue
		}
		validPageIDs = append(validPageIDs, v.Pid)
	}
	return validPageIDs
}

func (s *Service) Progress(c context.Context, req *dynmdl.ProgressReq, mid int64) (*dynmdl.ProgressRly, error) {
	page, err := s.NatConfig(c, &v1.NatConfigReq{Pid: req.PageID, Ps: 41})
	if err != nil {
		return nil, err
	}
	for _, v := range page.Modules {
		if v.NativeModule == nil {
			continue
		}
		switch {
		case req.From == dynmdl.ProgressFromProg && v.NativeModule.IsProgress() && v.NativeModule.Ukey == req.WebKey:
			progress, err := s.progressOfModule(c, v.NativeModule, mid)
			if err != nil {
				return nil, err
			}
			return &dynmdl.ProgressRly{Num: progress}, nil
		case req.From == dynmdl.ProgressFromClick && v.NativeModule.IsClick():
			if v.Click == nil {
				continue
			}
			for _, a := range v.Click.Areas {
				if a.UnfinishedImage != req.WebKey {
					continue
				}
				progress, err := s.progressOfClick(c, a, mid)
				if err != nil {
					return nil, err
				}
				return &dynmdl.ProgressRly{Num: progress}, nil
			}
		}
	}
	return &dynmdl.ProgressRly{}, nil
}

func (s *Service) progressOfModule(c context.Context, module *v1.NativeModule, mid int64) (int64, error) {
	groupID := module.Width
	progRly, err := s.actDao.ActivityProgress(c, module.Fid, 2, mid, []int64{groupID})
	if err != nil {
		return 0, err
	}
	prog, ok := progRly.Groups[groupID]
	if !ok {
		return 0, nil
	}
	return prog.Total, nil
}

func (s *Service) progressOfClick(c context.Context, click *v1.NativeClick, mid int64) (int64, error) {
	if click.Tip == "" {
		return 0, nil
	}
	tip := &v1.ClickTip{}
	if err := json.Unmarshal([]byte(click.Tip), tip); err != nil {
		log.Error("Fail to unmarshal clickTip, clickTip=%+v error=%+v", click.Tip, err)
		return 0, err
	}
	progRly, err := s.actDao.ActivityProgress(c, click.ForeignID, 2, mid, []int64{tip.GroupId})
	if err != nil {
		return 0, err
	}
	prog, ok := progRly.Groups[tip.GroupId]
	if !ok {
		return 0, nil
	}
	return prog.Total, nil
}

func (s *Service) offlineActSubject(c context.Context, pageID int64) error {
	sources, err := s.natDao.PageSourcesByPid(c, pageID)
	if err != nil {
		return err
	}
	source, ok := sources[actmdl.ActTypeCollect]
	if !ok || source.Sid == 0 {
		return nil
	}
	return s.actDao.OfflineActSubject(c, source.Sid)
}

func (s *Service) publishDynamic(c context.Context, mid, pageID int64) error {
	pubFunc := func() error {
		list, err := s.natDao.RawNativePagesExt(c, []int64{pageID})
		if err != nil {
			return err
		}
		pageDyn, ok := list[pageID]
		if !ok {
			return nil
		}
		dynID, err := s.dynamicDao.CreateDynamic(c, pageDyn.Dynamic, mid, pageID)
		if err != nil {
			return err
		}
		_ = s.cache.Do(c, func(ctx context.Context) {
			_ = s.natDao.UpdateNatDynDynID(ctx, pageDyn.Id, dynID)
		})
		return nil
	}
	if err := pubFunc(); err != nil {
		log.Error("日志告警 UP主发起活动自动发布动态失败，pageID=%+v error=%+v", pageID, err)
		return err
	}
	return nil
}

func extractDimension(click *v1.NativeClick) (actGRPC.GetReserveProgressDimension, error) {
	tmpDimension, err := strconv.ParseInt(click.FinishedImage, 10, 64)
	if err != nil {
		log.Error("Fail to parse dimension, dimension=%+v err=%+v", click.FinishedImage, err)
		return 0, err
	}
	return actGRPC.GetReserveProgressDimension(tmpDimension), nil
}

func extractProgressParamFromClick(click *v1.NativeClick) (sid, groupID int64) {
	sid = click.ForeignID
	if click.Tip == "" {
		return
	}
	tip := &v1.ClickTip{}
	if err := json.Unmarshal([]byte(click.Tip), tip); err != nil {
		log.Error("Fail to unmarshal clickTip, clickTip=%+v error=%+v", click.Tip, err)
		return
	}
	groupID = tip.GroupId
	return
}

func calculateProgress(progress, interveNum int64) int64 {
	total := progress + interveNum
	if total < 0 {
		total = 0
	}
	return total
}

func offReason(reason string) string {
	if reason == "" {
		return "系统下架"
	}
	return reason
}

func extractProgNum(group *actGRPC.ActivityProgressGroup, nodeID int64) (total, targetNum int64) {
	total = group.Total
	for _, node := range group.Nodes {
		if node.Nid == nodeID {
			targetNum = node.Val
			return
		}
	}
	return
}

func isUpAct(m *dynmdl.PageMsg) bool {
	return m.Type == v1.TopicActType && m.FromType == v1.PageFromUid && m.RelatedUid > 0
}

func UniqueArray(arr []int64) []int64 {
	if len(arr) == 0 {
		return []int64{}
	}
	m := make(map[int64]struct{}, len(arr))
	uniq := make([]int64, 0, len(arr))
	for _, v := range arr {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		uniq = append(uniq, v)
	}
	return uniq
}

func canPublishDynamic(new, old *dynmdl.PageMsg) bool {
	if !isUpAct(new) {
		return false
	}
	return new.State == v1.OnlineState && (old.State == v1.CheckOffline || old.State == v1.WaitForCheck)
}
