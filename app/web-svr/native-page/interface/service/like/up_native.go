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
	"unicode/utf8"

	accountGRPC "git.bilibili.co/bapis/bapis-go/account/service"
	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	channelapi "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	upratingGRPC "git.bilibili.co/bapis/bapis-go/crm/service/uprating"
	topicapi "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	fligrpc "git.bilibili.co/bapis/bapis-go/filter/service"
	upclient "git.bilibili.co/bapis/bapis-go/up-archive/service"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	showmdl "go-gateway/app/app-svr/app-show/interface/model"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	aecode "go-gateway/app/web-svr/native-page/ecode"
	pb "go-gateway/app/web-svr/native-page/interface/api"
	actmdl "go-gateway/app/web-svr/native-page/interface/model/act"
	dynmdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
	spacemdl "go-gateway/app/web-svr/native-page/interface/model/space"
	whitemdl "go-gateway/app/web-svr/native-page/interface/model/white_list"
)

const (
	TagTypeVideo = 8
	CompanyRole  = 3
)

func (s *Service) IsUpActUid(c context.Context, mid int64) (bool, error) {
	whitelist, err := s.natDao.WhiteListByMid(c, mid)
	if err != nil {
		return false, nil
	}
	return whitelist != nil, nil
}

// IsUpActUidAuto up发起活动白名单（自动添加）
func (s *Service) IsUpActUidAuto(c context.Context, mid int64) (bool, error) {
	ok, err := s.IsUpActUid(c, mid)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	if !s.passAutoAddWhitelist(c, mid) {
		return false, nil
	}
	// 添加白名单
	_ = s.cache.Do(c, func(ctx context.Context) {
		if _, err := s.tryLockWlist(ctx, fmt.Sprintf("wl_%d", mid)); err != nil {
			return
		}
		saveReq := &whitemdl.WhiteList{
			Mid:      mid,
			Creator:  "system",
			Modifier: "system",
			State:    1,
			FromType: whitemdl.FromWhitelist,
		}
		_, _ = s.natDao.WhiteSave(ctx, saveReq)
	})
	return true, nil
}

// MinePages 我的活动列表页（待审核，已上线，已下线，审核不通过）.
func (s *Service) MinePages(c context.Context, mid, offset, ps int64) (*dynmdl.MinePagesRly, error) {
	end := ps
	if ps != -1 {
		end = offset + ps - 1
	}
	// mid 发起活动列表（审核中,打回,待上线,已上线,已下线）
	pidRly, err := s.natDao.NtTsUIDs(c, mid, offset, end)
	if err != nil {
		log.Error("s.natDao.NtTsUIDs(%d,%d,%d) error(%v)", mid, offset, end, err)
		return nil, err
	}
	rly := &dynmdl.MinePagesRly{Offset: offset}
	if pidRly == nil || len(pidRly.IDs) == 0 {
		return rly, nil
	}
	rly.Offset = offset + int64(len(pidRly.IDs))
	if pidRly.Count > rly.Offset {
		rly.HasMore = 1
	}
	var (
		pages     map[int64]*pb.NativePage
		lastPages map[int64]*pb.NativeTsPage
	)
	var pids []int64
	for _, v := range pidRly.IDs {
		if v == nil {
			continue
		}
		pids = append(pids, v.ID)
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if pages, e = s.natDao.NativePages(ctx, pids); e != nil {
			log.Error("s.natDao.NativePage(%v) error(%v)", pids, e)
			e = nil
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if lastPages, e = s.tsLastIDs(ctx, pids); e != nil {
			log.Error("s.tsLastIDs %v error(%v)", pidRly.IDs, e)
			e = nil
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "MinePages eg.Wait() failed, mid=%+v error=%+v", mid, err)
	}
	for _, v := range pids {
		if pv, ok := pages[v]; !ok || pv == nil {
			continue
		}
		if !pages[v].IsUpTopicAct() || pages[v].RelatedUid != mid {
			continue
		}
		tmp := &dynmdl.MinePage{Title: pages[v].Title, ForeignID: pages[v].ForeignID, Pid: v, State: pages[v].State}
		if tv, ok := lastPages[v]; ok && tv != nil {
			tmp.AuditState = lastPages[v].State
		}
		rly.List = append(rly.List, tmp)
	}
	rly.SpaceButton, _ = s.spaceButton(c, mid)
	return rly, nil
}

func (s *Service) UpActPages(c context.Context, mid, offset, ps int64) (*dynmdl.UpActPagesReply, error) {
	rly, err := s.UpActNativePages(c, mid, offset, ps)
	if err != nil {
		return nil, err
	}
	res := &dynmdl.UpActPagesReply{}
	if rly == nil {
		return res, nil
	}
	res.Offset = rly.Offset
	res.HasMore = rly.HasMore
	for _, v := range rly.List {
		if v == nil || v.Base == nil {
			continue
		}
		res.List = append(res.List, &dynmdl.UpActPages{PID: v.Base.ID, Title: v.Base.Title, ForeignID: v.Base.ForeignID})
	}
	return res, nil
}

// UpActNativePages 有效活动列表（已上线话题）.
func (s *Service) UpActNativePages(c context.Context, mid, offset, ps int64) (*pb.UpActNativePagesReply, error) {
	//check白名单
	ok, err := s.natDao.CacheSponsoredUp(c, mid)
	if err != nil {
		return nil, err
	}
	if !ok { //不是白名单用户返回空数据
		return &pb.UpActNativePagesReply{}, nil
	}
	end := ps
	if ps != -1 {
		end = offset + ps - 1
	}
	// mid 发起活动列表（已上线）
	pidRly, err := s.natDao.NtTsOnlineIDs(c, mid, offset, end)
	if err != nil {
		log.Error("s.natDao.NtTsOnlineIDs(%d,%d,%d) error(%v)", mid, offset, end, err)
		return nil, err
	}
	rly := &pb.UpActNativePagesReply{Offset: offset}
	if pidRly == nil || len(pidRly.IDs) == 0 {
		return rly, nil
	}
	rly.Offset = offset + int64(len(pidRly.IDs))
	if pidRly.Count > rly.Offset {
		rly.HasMore = 1
	}
	var pids []int64
	for _, v := range pidRly.IDs {
		if v == nil {
			continue
		}
		pids = append(pids, v.ID)
	}
	pages, err := s.natDao.NativePages(c, pids)
	if err != nil {
		log.Error("s.natDao.NativePage(%v) error(%v)", pids, err)
		return rly, nil
	}
	for _, v := range pids {
		if pv, ok := pages[v]; !ok || pv == nil {
			continue
		}
		if !pages[v].IsUpTopicAct() || pages[v].RelatedUid != mid {
			continue
		}
		tmp := &pb.UpActNativePages{Base: pages[v]}
		rly.List = append(rly.List, tmp)
	}
	return rly, nil
}

// tsLastIDs 根据pids获取最新提交审核的数据
func (s *Service) tsLastIDs(c context.Context, pids []int64) (map[int64]*pb.NativeTsPage, error) {
	//根据pids获取最新ts_ids
	mapRly, err := s.natDao.NtPidToTsIDs(c, pids)
	if err != nil {
		log.Error("s.natDao.NtPidToTsIDs(%v) error(%v)", pids, err)
		return nil, err
	}
	var tsIDs []int64
	for _, v := range mapRly {
		if v <= 0 {
			continue
		}
		tsIDs = append(tsIDs, v)
	}
	rly := make(map[int64]*pb.NativeTsPage)
	if len(tsIDs) == 0 {
		return rly, nil
	}
	tsInfos, err := s.natDao.NtTsPages(c, tsIDs)
	if err != nil {
		log.Error("s.natDao.NtTsPages %v error(%v)", tsIDs, err)
		return nil, err
	}
	for k, v := range mapRly {
		if v <= 0 {
			continue
		}
		if tv, ok := tsInfos[v]; !ok || tv == nil {
			continue
		}
		rly[k] = tsInfos[v]
	}
	return rly, nil
}

// TsPage pid下最新的提交信息.
func (s *Service) TsPage(c context.Context, mid int64, pid int64) (*dynmdl.TsPageRly, error) {
	var (
		pagesInfo map[int64]*pb.NativePage
		tsID      int64
	)
	eg := errgroup.WithContext(c)
	// 获取pageinfo
	eg.Go(func(ctx context.Context) (e error) {
		if pagesInfo, e = s.natDao.NativePages(ctx, []int64{pid}); e != nil {
			log.Error("s.natDao.NativePages(%d,%d) error(%v)", mid, pid, e)
			return
		}
		// 只能查看本人信息
		if pv, ok := pagesInfo[pid]; !ok || pv == nil || pv.FromType != pb.PageFromUid {
			e = ecode.NothingFound
			return
		}
		//只能查看本人信息
		if pagesInfo[pid].RelatedUid != mid {
			e = aecode.ActivityNtUserLimit
		}
		return
	})
	// 获取pid下提交的tsid
	eg.Go(func(ctx context.Context) (e error) {
		if tsID, e = s.natDao.NtPidToTsID(ctx, pid); e != nil {
			log.Error("s.natDao.NtPidToTsID(%d,%d) error(%v)", mid, pid, e)
			return
		}
		if tsID <= 0 {
			e = ecode.NothingFound
		}
		return
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	page := pagesInfo[pid]
	res := &dynmdl.TsPageRly{Attribute: page.Attribute, TsID: tsID, Pid: pid, Title: page.Title, ForeignID: page.ForeignID, BgColor: page.BgColor, State: page.State, Uid: page.RelatedUid, ShareImage: page.ShareImage}
	if pagesInfo[pid].Ver != "" { //最后编辑来源
		res.IsAdmin = strings.Contains(page.Ver, "admin")
	}
	egTwo := errgroup.WithContext(c)
	//草稿箱
	var tsPageRly map[int64]*pb.NativeTsPage
	egTwo.Go(func(ctx context.Context) (e error) {
		if tsPageRly, e = s.natDao.NtTsPages(ctx, []int64{tsID}); e != nil { //降级处理
			log.Error("s.natDao.NtTsPages(%d) error(%v)", tsID, e)
		}
		return
	})
	egTwo.Go(func(ctx context.Context) (e error) {
		if res.Modules, e = s.tsModulesExt(ctx, tsID); e != nil {
			log.Error("s.tsModule (%d) error(%v)", tsID, e)
		}
		return
	})
	egTwo.Go(func(ctx context.Context) error {
		var err error
		res.PageSources, err = s.natDao.PageSourcesByPid(ctx, pid)
		if err != nil {
			return err
		}
		if source, ok := res.PageSources[actmdl.ActTypeCollect]; ok {
			res.Partitions = source.Partitions
		}
		return nil
	})
	egTwo.Go(func(ctx context.Context) error {
		list, err := s.natDao.RawNativePagesExt(ctx, []int64{pid})
		if err != nil {
			return err
		}
		if pageDyn, ok := list[pid]; ok {
			res.Dynamic = pageDyn.Dynamic
		}
		return nil
	})
	if err := egTwo.Wait(); err != nil {
		return nil, err
	}
	if tv, ok := tsPageRly[tsID]; ok && tv != nil {
		res.VideoDisplay = tv.VideoDisplay
		res.AuditState = tv.State
		res.AuditTime = tv.AuditTime
		res.AuditType = tv.AuditType
		res.UpShareImage = tv.ShareImage
		res.Template = tv.Template
	}
	return res, nil
}

func (s *Service) TsPageResource(c context.Context, mid int64, pid int64) (*dynmdl.TsPageResourceRly, error) {
	tsPage, err := s.TsPage(c, mid, pid)
	if err != nil {
		return nil, err
	}
	rly := tsPage.Trans2TsPageResourceRly()
	// 审核中即展示审核中的内容
	if isAuditing(rly.AuditState) {
		rly.ShareImage = rly.UpShareImage
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		rly.SpaceButton, _ = s.spaceButton(ctx, mid)
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		rly.UserSpace, _ = s.natDao.UserSpaceByMid(ctx, mid)
		return nil
	})
	aids, actPids := extractTsResourceIDs(rly.Modules)
	var arcs map[int64]*arcapi.Arc
	if len(aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			rly, err := s.arcClient.Arcs(ctx, &arcapi.ArcsRequest{Aids: aids})
			if err != nil {
				log.Error("Fail to get arcs, aids=%+v error=%+v", aids, err)
				return err
			}
			arcs = rly.Arcs
			return nil
		})
	}
	var pages map[int64]*pb.NativePage
	if len(actPids) > 0 {
		eg.Go(func(ctx context.Context) error {
			var err error
			pages, err = s.natDao.NativePages(ctx, actPids)
			return err
		})
	}
	if source, ok := tsPage.PageSources[actmdl.ActTypeCollect]; ok && source.Sid > 0 {
		eg.Go(func(ctx context.Context) error {
			actRly, err := s.actDao.ActSubProtocol(ctx, source.Sid)
			if err != nil {
				return nil
			}
			rly.IsPartitionChanged = isPartitionChanged(source.Partitions, actRly)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "Fail to handle TsPageResource errgroup, mid=%d pid=%d error=%+v", mid, pid, err)
		return nil, err
	}
	for _, m := range rly.Modules {
		resources := make([]*dynmdl.ResourceDetail, 0, len(m.Resources))
		for _, r := range m.Resources {
			ok := true
			switch r.ResourceType {
			case pb.MixAvidType:
				ok = setArcOfResDetail(r, arcs)
			case pb.MixActivity:
				ok = setActivityOfResDetail(r, pages)
			}
			if !ok {
				continue
			}
			resources = append(resources, r)
		}
		m.Resources = resources
	}
	return rly, nil
}

// tsModulesExt .
func (s *Service) tsModulesExt(c context.Context, tsID int64) ([]*dynmdl.NativeTsModuleExt, error) {
	//根据tsid获取moduleids
	rly, err := s.natDao.NtTsModuleIDs(c, tsID, 0, -1)
	if err != nil {
		return nil, err
	}
	if rly == nil || len(rly.IDs) == 0 {
		return nil, nil
	}
	//根据moduleids获取详细信息
	mInfos, err := s.natDao.NtTsModulesExt(c, rly.IDs)
	if err != nil {
		return nil, err
	}
	var list []*dynmdl.NativeTsModuleExt
	for _, v := range rly.IDs {
		if mv, ok := mInfos[v]; !ok || mv == nil || mv.State != 1 {
			continue
		}
		list = append(list, mInfos[v])
	}
	return list, nil
}

/**
pageCommonCheck
1.mid白名单校验
2.post 防并发
3.module内remakr&meta校验
*/
// nolint:gocognit
func (s *Service) pageCommonCheck(c context.Context, mid int64, modStr, from string) ([]*dynmdl.NativeTsModuleExt, error) {
	//mid白名单校验
	if from == whitemdl.WhitelistOpAdd {
		ok, err := s.IsUpActUid(c, mid)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, aecode.ActivityNtUserLimit
		}
	}
	// 防并发
	if ok, err := s.natDao.NtTsMidUnique(c, mid); err != nil || !ok {
		return nil, aecode.ActivityFrequence
	}
	if modStr == "" {
		return nil, nil
	}
	var arg []*dynmdl.ParamModule
	if err := json.Unmarshal([]byte(modStr), &arg); err != nil {
		log.Error("pageCommonCheck json.Unmarshal %s error(%v)", modStr, err)
		return nil, ecode.RequestErr
	}
	// module check
	var (
		lastModules []*dynmdl.NativeTsModuleExt
		uniqueMap   = make(map[int64]int64)
		arcNumLeft  = dynmdl.UpArcsMax
	)
	for k, v := range arg {
		tmp := &dynmdl.NativeTsModuleExt{
			NativeTsModule: pb.NativeTsModule{Category: v.Category, Rank: int64(k), PType: pb.CommonPage, State: 1, Ukey: v.Ukey},
		}
		categoryTmp := &pb.NativeModule{Category: int64(v.Category)}
		switch {
		case categoryTmp.IsStatement(): //文本组件
			//只会有一个文本组件
			if _, ok := uniqueMap[v.Category]; ok {
				continue
			}
			// 空组件
			if len(v.Remark) == 0 {
				continue
			}
			if err := s.RemarkCheck(c, v.Remark); err != nil {
				return nil, err
			}
			tmp.Remark = v.Remark
		case categoryTmp.IsClick(): //自定义点击组件
			// 只会有一个自定义点击组件
			if _, ok := uniqueMap[v.Category]; ok {
				continue
			}
			// 空组件
			if v.Meta == "" || v.Width == 0 || v.Length == 0 {
				continue
			}
			//check image width length
			if err := dynmdl.MetaCheck(v.Meta, int32(v.Width), int32(v.Length)); err != nil {
				return nil, err
			}
			tmp.Meta = v.Meta
			tmp.Width = v.Width
			tmp.Length = v.Length
		case categoryTmp.IsResourceID():
			tmp.SetResourceNum(v, &arcNumLeft)
		case categoryTmp.IsNewVideoID():
			tmp.Attribute = 1 << pb.AttrIsAutoPlay
			tmp.SetResourceNum(v, &arcNumLeft)
		case categoryTmp.IsActCapsule():
			//只会有一个相关活动组件
			if _, ok := uniqueMap[v.Category]; ok {
				continue
			}
			tmp.Remark = "相关活动"
			if len(v.Resources) > dynmdl.UpSyncActMax {
				tmp.Resources = v.Resources[:dynmdl.UpSyncActMax]
			} else {
				tmp.Resources = v.Resources
			}
		case categoryTmp.IsCarouselImg():
			if _, ok := uniqueMap[v.Category]; ok {
				continue
			}
			if len(v.Resources) == 0 || v.Resources[0].Ext == "" {
				continue
			}
			extStr, ext, err := buildCarouselExt(v.Resources[0].Ext)
			if err != nil {
				return nil, err
			}
			tmp.Meta = ext.ImgUrl
			tmp.Width = ext.Width
			tmp.Length = ext.Length
			tmp.Resources = append(tmp.Resources, &pb.NativeTsModuleResource{ResourceType: pb.MixCarouselImg, Ext: extStr})
		case categoryTmp.IsRecommend():
			if _, ok := uniqueMap[v.Category]; ok {
				continue
			}
			tmp.Num = 1
			tmp.Resources = append(tmp.Resources, &pb.NativeTsModuleResource{ResourceType: pb.MixTypeRcmd, ResourceID: mid})
		default: //其余组件暂不支持
			continue
		}
		uniqueMap[v.Category] = v.Category
		lastModules = append(lastModules, tmp)
	}
	return lastModules, nil
}

// MinePageSave.
// 草稿箱编辑&上线后编辑
func (s *Service) MinePageSave(c context.Context, mid int64, arg *dynmdl.ParamMinePageSave) (*dynmdl.PageSaveRly, error) {
	//编辑草稿箱OR(已上线&审核通过更新)-根据pid下的state区分
	tsRly, err := s.TsPage(c, mid, arg.PID)
	if err != nil {
		log.Error("s.TsPage(%d,%d) error(%v)", mid, arg.PID, err)
		return nil, err
	}
	if tsRly.State == pb.OfflineState {
		return nil, aecode.NativePageOffline
	}
	if err := s.checkDynamic(c, arg.Dynamic); err != nil {
		return nil, err
	}
	//草稿箱状态，重新保存数据
	if tsRly.State == pb.WaitForCommit {
		return s.pageWaitSave(c, mid, arg, tsRly)
	}
	//(已上线&审核通过更新)
	return s.pageOnlineSave(c, mid, arg, tsRly)
}

// UpActNativePageBind .
func (s *Service) UpActNativePageBind(c context.Context, arg *pb.UpActNativePageBindReq) (*pb.UpActNativePageBindReply, error) {
	//pid 是草稿状态
	ts, err := s.TsPage(c, arg.Mid, arg.PageID)
	if err != nil {
		log.Error("s.TsPage(%d,%d) error(%v)", arg.Mid, arg.PageID, err)
		return nil, err
	}
	//草稿箱状态才支持送审
	if ts == nil {
		log.Error("UpActNativePageBind mid:%d,pageid:%d  ts is nil", arg.Mid, arg.PageID)
		return nil, ecode.RequestErr
	}
	// 兼容重复提交，能返回正确的值
	if ts.State == pb.WaitForCheck {
		return &pb.UpActNativePageBindReply{}, nil
	}
	if ts.State != pb.WaitForCommit {
		return nil, aecode.ActivityNtNoBind
	}
	if ts.Uid != arg.Mid {
		return nil, aecode.ActivityNtUserLimit
	}
	//修改native_page ,修改ts_native_page
	tmpPage := &pb.NativePage{State: pb.WaitForCheck, Type: pb.TopicActType, ID: ts.Pid}
	if err = s.natDao.PageBind(c, tmpPage); err != nil {
		log.Error("s.natDao.PageBind(%d,%d) error(%v)", arg.PageID, arg.Mid, err)
		return nil, err
	}
	if err := s.cache.Do(c, func(c context.Context) {
		var (
			partitions   string
			auditContent dynmdl.AuditContent
		)
		auditContent.SetModule()
		if ts.Template == dynmdl.UpTempCollect {
			if _, ok := ts.PageSources[actmdl.ActTypeCollect]; ok {
				partitions = ts.Partitions
				auditContent.SetCollectTemp()
			}
		}
		// 送审
		tsReq := &dynmdl.TsSendReq{
			TsID:         ts.TsID,
			Title:        ts.Title,
			BgColor:      tmpPage.BgColor,
			State:        tmpPage.State,
			Uid:          ts.Uid,
			Pid:          ts.Pid,
			Modules:      ts.Modules,
			Url:          arg.Url,
			AuditTime:    ts.AuditTime,
			Partitions:   partitions,
			Template:     ts.Template,
			AuditContent: auditContent,
			IsFirstAudit: isFirstAudit(ts.State),
		}
		if e := s.aegisDao.AegisAddByTsReq(c, tsReq); e != nil {
			log.Error("s.natDao.AegisAdd %d error(%v)", ts.TsID, e)
		}
	}); err != nil {
		log.Errorc(c, "UpActNativePageBind fanout.Do() failed, req=%+v error=%+v", arg, err)
	}
	return &pb.UpActNativePageBindReply{}, nil
}

/*
*
tagNameValid
1.tagName 是否是在线话题活动
2.title set 防并发
3.title 转tagid
*/
func (s *Service) tagNameValid(c context.Context, name string, mid int64) (tagName string, tagID int64, err error) {
	tagName = name
	//title不可以重复
	if tagID, err = s.TitleCheck(c, name); err != nil {
		return
	}
	eg := errgroup.WithContext(c)
	//title 防并发
	eg.Go(func(ctx context.Context) (e error) {
		var succ bool
		succ, e = s.natDao.NtTsTitleUnique(ctx, name)
		if e != nil {
			return
		}
		// 并发操作
		if !succ {
			e = ecode.Error(ecode.RequestErr, "活动已存在")
		}
		return
	})
	// 新tag，需要通过tag服务新增
	if tagID == 0 {
		//add tagID
		eg.Go(func(ctx context.Context) error {
			tagRly, e := s.tagDao.AddTag(ctx, name, mid)
			if e != nil {
				return e
			}
			if tagRly == nil || tagRly.Tag == nil || tagRly.Tag.Id == 0 {
				log.Error("s.tagDao.AddTag(%s,%d) 发生未知错误", name, mid)
				return ecode.RequestErr
			}
			tagID = tagRly.Tag.Id
			tagName = tagRly.Tag.Name
			return nil
		})
	}
	err = eg.Wait()
	return
}

// pageOnlineSave
// nolint:gocognit
func (s *Service) pageOnlineSave(c context.Context, mid int64, arg *dynmdl.ParamMinePageSave, tsRly *dynmdl.TsPageRly) (*dynmdl.PageSaveRly, error) {
	lastModules, err := s.pageCommonCheck(c, mid, arg.Modules, whitemdl.WhitelistOpOnlineSave)
	if err != nil {
		log.Error("s.pageCommonCheck(%d) error(%v)", mid, err)
		return nil, err
	}
	expiryArcs := func() bool {
		rly, e := s.tagDao.TagByName(c, arg.Title)
		if e != nil {
			return false
		}
		return s.delInvalidResources(c, rly.GetId(), lastModules)
	}()
	tmpPage := &pb.NativePage{
		ID:      arg.PID,
		BgColor: arg.BgColor,
	}
	svRly := &dynmdl.PageSaveRly{TopicID: tsRly.ForeignID, PID: tsRly.Pid, State: tsRly.State, Title: tsRly.Title, ExpiryArcs: expiryArcs}
	// 新老module 对比
	svRly.AuditState = pb.TsWaitCheck
	auditType := pb.TsAuditManual
	auditContent, tsCompare := tsComparison(tsRly, arg, lastModules)
	switch tsCompare {
	case dynmdl.TsCompareNoChange:
		// 驳回状态禁止修改无需审核部分
		if isNatAuditFail(tsRly.State) && (tsRly.BgColor != arg.BgColor || arg.ShareImage == "") {
			return nil, aecode.NaAuditFail
		}
		svRly.AuditState = tsRly.AuditState
		_ = s.updateSpaceFromTsOp(c, mid, arg.PID, arg.UserSpace, dynmdl.SpaceSaveFromTsSave)
	case dynmdl.TsCompareAuto:
		// 驳回状态禁止修改无需审核部分
		if isNatAuditFail(tsRly.State) {
			return nil, aecode.NaAuditFail
		}
		auditType = pb.TsAuditAuto
	}
	//page的内容可以直接更新,只可编辑 bgcolor
	if tsRly.BgColor != arg.BgColor || arg.ShareImage == "" {
		if arg.BgColor != "" { //夜间模式颜色适配
			tmpPage.Attribute = tsRly.Attribute | pb.AttrIsNotNightNum
		} else {
			tmpPage.Attribute = tsRly.Attribute & (pb.AttrMaxNum - pb.AttrIsNotNightNum)
		}
		tmpPage.ShareImage = tsRly.ShareImage
		if arg.ShareImage == "" {
			tmpPage.ShareImage = "https://i0.hdslb.com/bfs/activity-plat/static/8347b7383c4a730528a82854f98b9b32/sYbPL4QDx9.png"
		}
		if err = s.natDao.PageUpdate(c, tmpPage); err != nil {
			log.Error("s.natDao.PageColorUpdat(%d,%d) error(%v)", mid, arg.PID, err)
			return nil, err
		}
	}
	if tsCompare == dynmdl.TsCompareNoChange {
		return svRly, nil
	}
	if auditType == pb.TsAuditManual && isNatAuditFail(tsRly.State) {
		// 已驳回：更新native_page为待审核
		if err = s.natDao.UpdatePageState(c, tsRly.Pid, pb.WaitForCheck, tsRly.State); err != nil {
			return nil, err
		}
		svRly.State = pb.WaitForCheck
	}
	//module的内容根据是否有变更确定是否要重新送审
	if err := s.cache.Do(c, func(c context.Context) {
		_ = s.updateSpaceFromTsOp(c, mid, arg.PID, arg.UserSpace, dynmdl.SpaceSaveFromTsSave)
		var partitions string
		if tsRly.Template == dynmdl.UpTempCollect {
			newPart, err := s.updatePageSource(c, tsRly.PageSources[actmdl.ActTypeCollect], &pb.NativePageSource{PageId: arg.PID, Partitions: arg.Partitions, ActType: actmdl.ActTypeCollect})
			if err != nil {
				auditContent.UnsetCollectTemp()
			}
			partitions = newPart
		}
		var dynamic string
		if isFirstAudit(tsRly.State) {
			dynamic = tsRly.Dynamic
			if err := s.savePageDynamic(c, tsRly.Pid, arg.Dynamic); err == nil {
				dynamic = arg.Dynamic
			}
		}
		if isAuditing(tsRly.AuditState) && !(tsRly.AuditType == pb.TsAuditAuto && auditType == pb.TsAuditManual) {
			// 审核中的活动，更新审核内容
			atime := tsRly.AuditTime
			if auditType == pb.TsAuditManual {
				atime = auditTime()
			}
			if err := s.updateTsModule(c, tsRly.TsID, atime, arg.VideoDisplay, arg.ShareImage, lastModules); err != nil {
				return
			}
			if auditType == pb.TsAuditManual {
				_ = s.aegisDao.AegisUpdateByTsReq(c, &dynmdl.TsSendReq{
					TsID:         tsRly.TsID,
					Title:        tsRly.Title,
					BgColor:      tmpPage.BgColor,
					State:        tmpPage.State,
					Uid:          mid,
					Pid:          tsRly.Pid,
					Modules:      lastModules,
					AuditTime:    atime,
					ShareImage:   arg.ShareImage,
					Partitions:   partitions,
					AuditContent: auditContent,
					Dynamic:      dynamic,
				})
			}
			return
		}
		if isAuditing(tsRly.AuditState) && tsRly.AuditType == pb.TsAuditAuto && auditType == pb.TsAuditManual {
			// 自动过审中，修改需人工送审内容
			_ = s.natDao.UpdateTsState(c, tsRly.TsID, pb.TsCheckOffline, "重新人工送审关闭")
		}
		//add native_ts_page tspid
		tmpTsPage := &pb.NativeTsPage{Title: tsRly.Title, ForeignID: tsRly.ForeignID, State: pb.TsWaitCheck, Pid: arg.PID, VideoDisplay: arg.VideoDisplay, AuditType: auditType, AuditTime: auditTime(), ShareImage: arg.ShareImage, Template: tsRly.Template}
		tsID, e := s.tsModuleSave(c, tmpTsPage, lastModules)
		if e != nil {
			log.Error("s.tsModuleSave(%v) error(%v)", tmpTsPage, e)
			return
		}
		if auditType == pb.TsAuditAuto {
			return
		}
		// 送审
		ts := &dynmdl.TsSendReq{
			TsID:         tsID,
			Title:        tmpTsPage.Title,
			BgColor:      tmpPage.BgColor,
			State:        tmpPage.State,
			Uid:          mid,
			Pid:          tmpTsPage.Pid,
			Modules:      lastModules,
			AuditTime:    tmpTsPage.AuditTime,
			ShareImage:   arg.ShareImage,
			Partitions:   partitions,
			Template:     tsRly.Template,
			AuditContent: auditContent,
			Dynamic:      dynamic,
			IsFirstAudit: isFirstAudit(tsRly.State),
		}
		if e = s.aegisDao.AegisAddByTsReq(c, ts); e != nil {
			log.Error("s.natDao.AegisAdd %d error(%v)", ts.TsID, e)
		}
	}); err != nil {
		log.Errorc(c, "pageOnlineSave fanout.Do() failed, req=%+v error=%+v", arg, err)
	}
	return svRly, nil
}

// pageWaitSave .
func (s *Service) pageWaitSave(c context.Context, mid int64, arg *dynmdl.ParamMinePageSave, tsRly *dynmdl.TsPageRly) (*dynmdl.PageSaveRly, error) {
	lastModules, err := s.pageCommonCheck(c, mid, arg.Modules, whitemdl.WhitelistOpWaitSave)
	if err != nil {
		log.Error("s.pageCommonCheck(%d) error(%v)", mid, err)
		return nil, err
	}
	//title不可以重复
	tagTitle, tagID, err := s.tagNameValid(c, arg.Title, mid)
	if err != nil {
		return nil, err
	}
	expiryArcs := s.delInvalidResources(c, tagID, lastModules)
	tmpPage := &pb.NativePage{
		BgColor:   arg.BgColor,
		ID:        arg.PID,
		Title:     tagTitle,
		ForeignID: tagID,
	}
	if arg.BgColor != "" { //夜间模式颜色适配
		tmpPage.Attribute = tsRly.Attribute | pb.AttrIsNotNightNum
	} else {
		tmpPage.Attribute = tsRly.Attribute & (pb.AttrMaxNum - pb.AttrIsNotNightNum)
	}
	if err = s.natDao.PageWaitUpdate(c, tmpPage); err != nil {
		return nil, err
	}
	if err := s.cache.Do(c, func(c context.Context) {
		_ = s.updateSpaceFromTsOp(c, mid, arg.PID, arg.UserSpace, dynmdl.SpaceSaveFromTsSave)
		//add native_ts_page tspid
		tmpTsPage := &pb.NativeTsPage{Title: tagTitle, ForeignID: tagID, State: pb.TsWaitCheck, Pid: arg.PID, VideoDisplay: arg.VideoDisplay, AuditType: pb.TsAuditManual, AuditTime: auditTime(), ShareImage: arg.ShareImage, Template: tsRly.Template}
		if _, err := s.tsModuleSave(c, tmpTsPage, lastModules); err != nil {
			log.Error("s.tsModuleSave(%v) error(%v)", tmpTsPage, err)
			return
		}
		if tsRly.Template == dynmdl.UpTempCollect {
			_, _ = s.updatePageSource(c, tsRly.PageSources[actmdl.ActTypeCollect], &pb.NativePageSource{PageId: arg.PID, Partitions: arg.Partitions, ActType: actmdl.ActTypeCollect})
		}
		_ = s.savePageDynamic(c, arg.PID, arg.Dynamic)
	}); err != nil {
		log.Errorc(c, "pageWaitSave fanout.Do() failed, req=%+v error=%+v", arg, err)
	}
	return &dynmdl.PageSaveRly{Title: tagTitle, TopicID: tagID, PID: arg.PID, State: pb.WaitForCommit, AuditState: pb.TsWaitCheck, ExpiryArcs: expiryArcs}, nil
}

/*
*
MinePageAdd .
*评论，动态发起草稿
*话题广场直接发起活动，直接送审
*/
func (s *Service) MinePageAdd(c context.Context) error {
	return ecode.Error(
		ecode.RequestErr,
		"亲爱的UP主，感谢您一直以来对本功能内测的支持。我们将在6.47版本将此功能转移至新话题体系内，您发起的活动将自动迁移至您发起的话题内，页面的展示可能有所不同。",
	)
}

func (s *Service) tsModuleSave(c context.Context, tmpTsPage *pb.NativeTsPage, lastModules []*dynmdl.NativeTsModuleExt) (int64, error) {
	//add native_ts_page
	tsPid, err := s.natDao.TsPageSave(c, tmpTsPage)
	if err != nil {
		log.Error("s.natDao.TsPageSave(%v) error(%v)", tmpTsPage, err)
		return 0, err
	}
	//add native_mdoule
	if len(lastModules) > 0 { //remark,meta 不是必填
		if err := s.natDao.TsModuleSave(c, lastModules, tsPid); err != nil {
			log.Error("s.natDao.TsModuleSave(%d) error(%v)", tsPid, err)
			return 0, err
		}
	}
	return tsPid, nil
}

func (s *Service) updateTsModule(c context.Context, tsID, auditTime int64, videoDisplay, shareImage string, lastModules []*dynmdl.NativeTsModuleExt) error {
	if err := s.natDao.UpdateVideoDisplay(c, tsID, auditTime, videoDisplay, shareImage); err != nil {
		return err
	}
	rly, err := s.natDao.NtTsModuleIDs(c, tsID, 0, -1)
	if err != nil {
		log.Error("Fail to get NtTsModuleIDs, tsID=%+v error=%+v", tsID, err)
		return err
	}
	if rly != nil {
		if err := s.natDao.DeleteTsModule(c, rly.IDs); err != nil {
			return err
		}
	}
	if len(lastModules) > 0 {
		if err := s.natDao.TsModuleSave(c, lastModules, tsID); err != nil {
			log.Error("Fail to batch create native_ts_module, tsID=%+v error=%+v", tsID, err)
			return err
		}
	}
	return nil
}

// titleCheck .
func (s *Service) TitleCheck(c context.Context, title string) (int64, error) {
	var tagID int64
	//长度check
	if utf8.RuneCountInString(title) < 3 || utf8.RuneCountInString(title) > 20 {
		return tagID, ecode.Error(ecode.RequestErr, "活动字数最少3个字，最多20个字")
	}
	//敏感词发生未知错误，则不过敏感词服务，由审核后台人员处理
	if filrly, err := s.fliClient.Filter(c, &fligrpc.FilterReq{Area: "activity_topic", Message: title}); err == nil && (filrly != nil && filrly.Level >= 30) {
		return tagID, ecode.Error(ecode.RequestErr, "活动名称含有特殊字符")
	}
	// tag 不存在则不需要校验话题状态
	rly, err := s.tagDao.TagByName(c, title)
	if err != nil {
		log.Error("s.tagDao.TagByName(%s) error(%v)", title, err)
		return tagID, err
	}
	if rly == nil || rly.Id == 0 {
		return tagID, nil
	}
	// check tagID是否是新频道
	tagID = rly.Id
	infos, err := s.channelClient.Infos(c, &channelapi.InfosReq{Cids: []int64{tagID}})
	if err != nil {
		log.Error("Fail to get channel infos, tagID=%+v error=%+v", tagID, err)
		return 0, err
	}
	if channel, ok := infos.GetCidMap()[tagID]; ok && channel.GetCType() == showmdl.NewChannel {
		return 0, ecode.Error(ecode.RequestErr, "话题已被使用")
	}
	// check tagid是否是话题活动
	// tagID 待上线，待审核，已上线，直接查数据库
	id, err := s.natDao.NatTagIDExist(c, tagID)
	if err != nil {
		log.Error("s.natDao.RawTitleSearch(%s) error(%v)", title, err)
		return tagID, err
	}
	if id > 0 {
		return tagID, ecode.Error(ecode.RequestErr, "活动已存在")
	}
	return tagID, nil
}

func (s *Service) RemarkCheck(c context.Context, msg string) error {
	remarkLimit := 200
	if utf8.RuneCountInString(msg) > remarkLimit {
		return ecode.Errorf(ecode.RequestErr, "活动说明最多%d个字", remarkLimit)
	}
	//敏感词发生未知错误，则不过敏感词服务，由审核后台人员处理
	if filrly, err := s.fliClient.Filter(c, &fligrpc.FilterReq{Area: "activity_topic", Message: msg}); err == nil && (filrly != nil && filrly.Level >= 30) {
		return ecode.Error(ecode.RequestErr, "活动名称含有特殊字符")
	}
	return nil
}

// TsWhiteSave 功能下线 保留router.
func (s *Service) TsWhiteSave(c context.Context) (*dynmdl.TsWhiteRly, error) {
	return nil, aecode.UpActIllegal
}

func (s *Service) InlineTsWhite(c context.Context, mid int64) (*dynmdl.TsWhiteRly, error) {
	status := func() int {
		ok, err := s.natDao.CacheSponsoredUp(c, mid)
		if err == nil && ok {
			return 1
		}
		return 0
	}()
	return &dynmdl.TsWhiteRly{Status: status}, nil
}

// TsWhite
func (s *Service) TsWhite(c context.Context, mid int64) (*dynmdl.TsWhiteRly, error) {
	rly := &dynmdl.TsWhiteRly{}
	if ok, err := s.IsUpActUid(c, mid); err == nil && ok {
		rly.Status = 1
	}
	return rly, nil
}

func (s *Service) MyArchiveList(c context.Context, mid, pn, ps int64) (*dynmdl.MyArchiveListRly, error) {
	rly := &dynmdl.MyArchiveListRly{}
	req := &upclient.ArcPassedReq{Mid: mid, Pn: pn, Ps: ps}
	arcsRly, err := s.upClient.ArcPassed(c, req)
	if err != nil {
		log.Errorc(c, "Fail to get upArcs, req=%+v error=%+v", req, err)
		return nil, err
	}
	if arcsRly == nil {
		return rly, nil
	}
	rly.Total = arcsRly.Total
	if rly.Total > pn*ps {
		rly.HasMore = true
	}
	rly.List = make([]*dynmdl.ArchiveItem, 0, len(arcsRly.GetArchives()))
	for _, arc := range arcsRly.GetArchives() {
		if arc == nil {
			continue
		}
		rly.List = append(rly.List, transUpArchiveItem(arc))
	}
	return rly, nil
}

func (s *Service) ActArchiveList(c context.Context, req *dynmdl.ActArchiveListReq) (*dynmdl.ActArchiveListRly, error) {
	tagRly, err := s.tagDao.TagByName(c, req.Title)
	if err != nil {
		return nil, err
	}
	if tagRly == nil {
		return &dynmdl.ActArchiveListRly{}, nil
	}
	topicReq := &topicapi.BriefDynsReq{
		TopicId:  tagRly.GetId(),
		Sortby:   req.Sort,
		Types:    strconv.FormatInt(TagTypeVideo, 10),
		PageSize: req.Ps,
		Offset:   req.Offset,
	}
	topicRly, err := s.topicClient.BriefDyns(c, topicReq)
	if err != nil {
		log.Errorc(c, "Fail to get activity archiveList, req=%+v error=%+v", topicReq, err)
		return nil, err
	}
	aids := make([]int64, 0, len(topicRly.GetDynamics()))
	for _, v := range topicRly.GetDynamics() {
		if v.GetType() == TagTypeVideo {
			aids = append(aids, int64(v.Rid))
		}
	}
	list := make([]*dynmdl.ArchiveItem, 0, len(aids))
	if len(aids) > 0 {
		arcsRly, err := s.arcClient.Arcs(c, &arcapi.ArcsRequest{Aids: aids})
		if err != nil {
			log.Errorc(c, "Fail to get arcs, aids=%+v error=%+v", aids, err)
			return nil, err
		}
		for _, aid := range aids {
			arc, ok := arcsRly.GetArcs()[aid]
			if !ok || arc == nil {
				continue
			}
			list = append(list, trans2ArchiveItem(arc))
		}
	}
	var hasMore bool
	if topicRly.GetHasMore() == 1 {
		hasMore = true
	}
	rly := &dynmdl.ActArchiveListRly{
		HasMore: hasMore,
		Offset:  topicRly.GetOffset(),
		List:    list,
	}
	return rly, nil
}

// 过滤掉已不在该话题下的稿件
func (s *Service) delInvalidResources(c context.Context, tagID int64, modules []*dynmdl.NativeTsModuleExt) (expiryArcs bool) {
	var (
		topicRids []int64
		rids      []int64
	)
	for _, m := range modules {
		for _, r := range m.Resources {
			if r.ResourceType == pb.MixAvidType {
				rids = append(rids, r.ResourceID)
			}
			if r.ResourceFrom == dynmdl.TsResFromAct {
				topicRids = append(topicRids, r.ResourceID)
			}
		}
	}
	arcMu := sync.Mutex{}
	isIncluded := make(map[int64]bool, len(topicRids))
	validArcs := make(map[int64]struct{}, len(rids))
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		if len(topicRids) == 0 {
			return nil
		}
		rly, err := s.topicClient.IsIncluded(ctx, &topicapi.IsIncludedReq{TopicId: tagID, Rids: topicRids})
		if err != nil {
			log.Errorc(ctx, "Fail to judge rids of topic, tagID=%+v rids=%+v error=%+v", tagID, topicRids, err)
			return err
		}
		isIncluded = rly.GetIsIncluded()
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if len(rids) == 0 {
			return nil
		}
		rly, err := s.arcClient.Arcs(ctx, &arcapi.ArcsRequest{Aids: rids})
		if err != nil {
			log.Errorc(ctx, "Fail to get arcs, rids=%+v error=%+v", rids, err)
			return err
		}
		for _, v := range rly.GetArcs() {
			if !v.IsNormal() {
				continue
			}
			arcMu.Lock()
			validArcs[v.Aid] = struct{}{}
			arcMu.Unlock()
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return
	}
	for _, m := range modules {
		tmpResources := make([]*pb.NativeTsModuleResource, 0)
		for _, r := range m.Resources {
			func() {
				if _, ok := validArcs[r.ResourceID]; !ok && r.ResourceType == pb.MixAvidType {
					expiryArcs = true
					return
				}
				if r.ResourceFrom != dynmdl.TsResFromAct {
					tmpResources = append(tmpResources, r)
					return
				}
				if is, ok := isIncluded[r.ResourceID]; ok && is {
					tmpResources = append(tmpResources, r)
					return
				}
				expiryArcs = true
			}()
		}
		m.Resources = tmpResources
	}
	return
}

func (s *Service) SpaceSyncSetting(c context.Context, req *pb.SpaceSyncSettingReq) (*pb.SpaceSyncSettingReply, error) {
	var (
		spaceButton string
		page        *pb.NativePage
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		var err error
		spaceButton, err = s.spaceButton(c, req.Mid)
		if err != nil && err != aecode.ActivityNtUserLimit {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if pages, err := s.NativePages(c, []int64{req.PageId}); err == nil {
			page = pages[req.PageId]
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if page == nil || page.RelatedUid != req.Mid {
		spaceButton = ""
	}
	return &pb.SpaceSyncSettingReply{SpaceButton: spaceButton}, nil
}

func (s *Service) TsSetting(c context.Context, mid int64) (*dynmdl.TsSettingRly, error) {
	spaceButton, _ := s.spaceButton(c, mid)
	return &dynmdl.TsSettingRly{SpaceButton: spaceButton}, nil
}

func (s *Service) spaceButton(c context.Context, mid int64) (string, error) {
	var isUP, isSponsored bool
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		rly, err := s.IsUpActUid(ctx, mid)
		if err != nil {
			return err
		}
		isUP = rly
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		rly, err := s.natDao.CacheSponsoredUp(ctx, mid)
		if err == nil {
			isSponsored = rly
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return "", err
	}
	if !isUP && !isSponsored {
		return "", aecode.ActivityNtUserLimit
	}
	userTab, err := s.spaceDao.UserTab(c, mid)
	if err != nil {
		// 当前UP主空间无活动tab
		if ecode.EqualError(ecode.NothingFound, err) {
			return dynmdl.SpaceBtPersonal, nil
		}
		return "", err
	}
	// 当前UP主空间有活动tab，不是UP主发起活动
	if userTab.TabType != spacemdl.TabTypeUpAct {
		return "", nil
	}
	// 当前UP主空间有活动tab，是up主发起活动
	pageRly, err := s.NativePages(c, []int64{userTab.TabCont})
	if err != nil {
		return "", err
	}
	if page, ok := pageRly[userTab.TabCont]; ok && page.FromType == pb.PageFromUid {
		return dynmdl.SpaceBtExclusive, nil
	}
	return "", nil
}

func (s *Service) TsSpace(c context.Context, mid int64) (*dynmdl.TsSpaceRly, error) {
	userSpace, err := s.natDao.UserSpaceByMid(c, mid)
	if err != nil {
		return nil, err
	}
	if userSpace == nil {
		return nil, ecode.NothingFound
	}
	pages, err := s.NativePages(c, []int64{userSpace.PageId})
	if err != nil {
		return nil, err
	}
	title := ""
	if page, ok := pages[userSpace.PageId]; ok && page != nil {
		title = page.Title
	}
	return &dynmdl.TsSpaceRly{
		NativeUserSpace: userSpace,
		NativePage: &struct {
			Title string `json:"title"`
		}{Title: title},
	}, nil
}

func (s *Service) TsSpaceSave(c context.Context, req *dynmdl.TsSpaceSaveReq, mid int64) error {
	if req.PageID == 0 || mid == 0 {
		return ecode.Error(ecode.RequestErr, fmt.Sprintf("pageID=%d or mid=%d is empty", req.PageID, mid))
	}
	var isUP, isSponsored bool
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		rly, err := s.IsUpActUid(ctx, mid)
		if err != nil {
			return err
		}
		isUP = rly
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if _, err := s.isPageBelongUp(ctx, req.PageID, mid); err != nil {
			return err
		}
		return nil
	})
	userSpace := &pb.NativeUserSpace{}
	eg.Go(func(ctx context.Context) error {
		rly, err := s.natDao.UserSpaceByMid(ctx, mid)
		if err != nil {
			return err
		}
		userSpace = rly
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		rly, err := s.natDao.CacheSponsoredUp(ctx, mid)
		if err == nil {
			isSponsored = rly
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return err
	}
	if !isUP && !isSponsored {
		return aecode.ActivityNtUserLimit
	}
	attrs := &pb.NativeUserSpace{
		Mid:          mid,
		Title:        req.Title,
		PageId:       req.PageID,
		DisplaySpace: req.DisplaySpace,
	}
	if userSpace == nil {
		if !needHandleSpace(req.DisplaySpace, req.PageID, 0, "") {
			return nil
		}
		setUserSpaceState(req.DisplaySpace, req.From, attrs)
		id, err := s.natDao.AddUserSpace(c, attrs)
		if err != nil {
			return err
		}
		attrs.Id = id
		if ifDoOnline(req.DisplaySpace, req.From) {
			if err := s.userSpaceDoOnline(c, mid, attrs); err != nil {
				return err
			}
		}
	} else {
		if !needHandleSpace(req.DisplaySpace, req.PageID, userSpace.PageId, userSpace.State) {
			return nil
		}
		attrs.Id = userSpace.Id
		setUserSpaceState(req.DisplaySpace, req.From, attrs)
		if err := s.natDao.UpdateUserSpace(c, attrs, userSpace.PageId, userSpace.State); err != nil {
			return err
		}
		if ifDoOnline(req.DisplaySpace, req.From) {
			// 此时空间可能配置的是另一个活动，处于待上线状态，这种情况直接覆盖，另一个活动上线操作时失败处理
			if err := s.userSpaceDoOnline(c, mid, attrs); err != nil {
				return err
			}
		} else if ifDoOffline(req.DisplaySpace, req.PageID, userSpace.PageId, userSpace.State) {
			if err := s.userSpaceDoOffline(c, mid, userSpace); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) isPageBelongUp(c context.Context, pageID, mid int64) (bool, error) {
	// 可能为审核状态，此处不能获取在线
	pages, err := s.natDao.NativePages(c, []int64{pageID})
	if err != nil {
		return false, err
	}
	page, ok := pages[pageID]
	if !ok || page.State == pb.OfflineState || page.State == pb.CheckOffline {
		return false, aecode.NativePageOffline
	}
	if page.RelatedUid != mid || page.FromType != pb.PageFromUid {
		return false, aecode.ActivityNtUserLimit
	}
	return true, nil
}

func (s *Service) userSpaceDoOnline(c context.Context, mid int64, attrs *pb.NativeUserSpace) error {
	success := false
	defer func() {
		if success {
			return
		}
		_ = s.natDao.UpdateUserSpaceState(c, attrs.Id, attrs.PageId, pb.USpaceOfflineBindFail, attrs.State)
	}()
	success, err := s.spaceDao.UpActivityTab(c, mid, spacemdl.OpOnline, attrs.Title, attrs.PageId)
	if err != nil {
		success = false
		return err
	}
	return nil
}

func (s *Service) userSpaceDoOffline(c context.Context, mid int64, attrs *pb.NativeUserSpace) error {
	success := false
	defer func() {
		if success {
			return
		}
		// 下线失败，重置为上线状态
		_ = s.natDao.UpdateUserSpaceState(c, attrs.Id, attrs.PageId, pb.USpaceOnline, attrs.State)
	}()
	success, err := s.spaceDao.UpActivityTab(c, mid, spacemdl.OpOffline, "", attrs.PageId)
	if err != nil {
		success = false
		return err
	}
	return nil
}

func (s *Service) updateSpaceFromTsOp(c context.Context, mid, pid int64, rawUserSpace string, from string) error {
	if rawUserSpace == "" || rawUserSpace == "null" {
		return nil
	}
	userSpace := &pb.NativeUserSpace{}
	if err := json.Unmarshal([]byte(rawUserSpace), userSpace); err != nil {
		log.Error("Fail to unmarshal userSpace, userSpace=%+v error=%+v", rawUserSpace, err)
		return err
	}
	req := &dynmdl.TsSpaceSaveReq{From: from, Title: userSpace.Title, DisplaySpace: userSpace.DisplaySpace, PageID: userSpace.PageId}
	if from == dynmdl.SpaceSaveFromTsAdd {
		req.PageID = pid
	}
	if err := s.TsSpaceSave(c, req, mid); err != nil {
		log.Error("Fail to update user_space from ts operation, req=%+v mid=%+v error=%+v", req, mid, err)
		return err
	}
	return nil
}

func (s *Service) passAutoAddWhitelist(c context.Context, mid int64) bool {
	if s.c.WhitelistCondition == nil {
		return false
	}
	var (
		uprating *upratingGRPC.RatingReply
		account  *accountGRPC.CardReply
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		rly, err := s.upratingDao.Rating(ctx, mid)
		if err != nil {
			return err
		}
		uprating = rly
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		rly, err := s.accDao.Card3(ctx, mid)
		if err != nil {
			return err
		}
		account = rly
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Fail to execute errgroup, mid=%+v error=%+v", mid, err)
		return false
	}
	if uprating.Rating == nil || account.Card == nil {
		return false
	}
	if uprating.Rating.Level < s.c.WhitelistCondition.MinRatingLevel || account.Card.Official.Role == CompanyRole {
		return false
	}
	log.Warn("auto add whitelist, mid=%d rating={%+v} official={%+v}", mid, uprating.Rating, account.Card.Official)
	return true
}

func (s *Service) tryLockWlist(c context.Context, lockKey string) (string, error) {
	lockID, locked, err := s.natDao.Lock(c, lockKey, s.c.WhitelistCondition.LockExpire)
	if err != nil {
		return "", err
	}
	if !locked {
		log.Error("Fail to get autoAudit lock, lock has been taken")
		return "", errors.New("lock has been taken")
	}
	return lockID, nil
}

func (s *Service) updatePageSource(c context.Context, oldData *pb.NativePageSource, newData *pb.NativePageSource) (newPartitions string, err error) {
	if oldData == nil {
		if _, err := s.natDao.AddPageSource(c, newData); err != nil {
			return "", err
		}
		return newData.Partitions, nil
	}
	if oldData.Partitions == newData.Partitions {
		return newData.Partitions, nil
	}
	if err := s.natDao.UpdatePageSource(c, oldData.Id, newData.Partitions); err != nil {
		return oldData.Partitions, err
	}
	return newData.Partitions, nil
}

func (s *Service) savePageDynamic(c context.Context, pid int64, dynamic string) error {
	list, err := s.natDao.NativePagesExt(c, []int64{pid})
	if err != nil {
		return err
	}
	pageDyn, ok := list[pid]
	if !ok {
		_, err = s.natDao.AddNativePageDyn(c, pid, dynamic)
		return err
	}
	if pageDyn.Dynamic == dynamic {
		return nil
	}
	return s.natDao.UpdateNatDynDynamic(c, pageDyn.Id, dynamic)
}

func (s *Service) checkDynamic(c context.Context, dynamic string) error {
	if dynamic == "" {
		return nil
	}
	if utf8.RuneCountInString(dynamic) > dynmdl.MaxDynamicLen {
		return ecode.Error(ecode.RequestErr, fmt.Sprintf("动态最多%d个字", dynmdl.MaxDynamicLen))
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		// 动态库（包含基础库）
		rly, err := s.fliClient.Filter(ctx, &fligrpc.FilterReq{Area: "bplus_dongtai", Message: dynamic})
		if err == nil && rly.Level > s.c.FilterClass.MinDynamic {
			return ecode.Error(ecode.RequestErr, "动态内容包含敏感信息，请修改后重新提交")
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		// 动态话题（不包含基础库）
		rly, err := s.fliClient.Filter(ctx, &fligrpc.FilterReq{Area: "bplus_topic", Message: dynamic})
		if err == nil && rly.Level > s.c.FilterClass.MinDynTopic {
			return ecode.Error(ecode.RequestErr, "动态内容包含敏感信息，请修改后重新提交")
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Fail to check dynamic, dynamic=%s error=%+v", dynamic, err)
		return err
	}
	return nil
}

func transUpArchiveItem(arc *upclient.Arc) *dynmdl.ArchiveItem {
	return &dynmdl.ArchiveItem{
		Aid:      arc.GetAid(),
		Title:    arc.GetTitle(),
		Pubdate:  cardmdl.PubDataString(arc.GetPubDate().Time()),
		Duration: cardmdl.DurationString(arc.GetDuration()),
		Pic:      arc.GetPic(),
		View:     cardmdl.StatString(arc.GetStat().View, ""),
		Danmaku:  cardmdl.StatString(arc.GetStat().Danmaku, ""),
	}
}

func trans2ArchiveItem(arc *arcapi.Arc) *dynmdl.ArchiveItem {
	return &dynmdl.ArchiveItem{
		Aid:      arc.GetAid(),
		Title:    arc.GetTitle(),
		Pubdate:  cardmdl.PubDataString(arc.GetPubDate().Time()),
		Duration: cardmdl.DurationString(arc.GetDuration()),
		Pic:      arc.GetPic(),
		View:     cardmdl.StatString(arc.GetStat().View, ""),
		Danmaku:  cardmdl.StatString(arc.GetStat().Danmaku, ""),
	}
}

func setUserSpaceState(displaySpace int64, from string, userSpace *pb.NativeUserSpace) {
	// 用户操作，即时生效
	if displaySpace == 1 {
		switch from {
		case dynmdl.SpaceSaveFromTsAdd:
			userSpace.State = pb.USpaceWaitingOnline //待审核通过后生效
		default:
			userSpace.State = pb.USpaceOnline //用户操作即时生效
		}
		return
	}
	// 还未创建
	if userSpace.Id == 0 {
		userSpace.State = ""
		return
	}
	// 关闭，即时生效
	if displaySpace == 0 {
		userSpace.State = pb.USpaceOfflineNormal
	}
}

func ifDoOnline(displaySpace int64, from string) bool {
	if displaySpace == 1 && from != dynmdl.SpaceSaveFromTsAdd {
		return true
	}
	return false
}

func ifDoOffline(displaySpace int64, pid, oldPid int64, oldState string) bool {
	if displaySpace == 0 && pid == oldPid && oldState == pb.USpaceOnline {
		return true
	}
	return false
}

func needHandleSpace(displaySpace int64, pid, oldPid int64, oldState string) bool {
	if displaySpace == 1 && oldState == pb.USpaceOnline && pid == oldPid {
		return false
	}
	if displaySpace == 0 && oldState == pb.USpaceOfflineNormal {
		return false
	}
	return true
}

func auditTime() int64 {
	return time.Now().Unix()
}

func isNatAuditFail(natState int64) bool {
	return natState == pb.CheckOffline
}

func isAuditing(tsState int64) bool {
	return tsState == pb.TsWaitCheck
}

func tsComparison(tsRly *dynmdl.TsPageRly, arg *dynmdl.ParamMinePageSave, lastModules []*dynmdl.NativeTsModuleExt) (dynmdl.AuditContent, int) {
	res := func() int {
		if arg.ShareImage != "" && tsRly.ShareImage != arg.ShareImage {
			return dynmdl.TsCompareManual
		}
		if isFirstAudit(tsRly.State) && tsRly.Dynamic != arg.Dynamic {
			return dynmdl.TsCompareManual
		}
		return dynmdl.ModuleComparison(tsRly.Modules, lastModules)
	}()
	var auditContent dynmdl.AuditContent
	if res != dynmdl.TsCompareNoChange {
		auditContent.SetModule()
	}
	if canSetCollectTemp(tsRly, arg, res) {
		res = dynmdl.TsCompareManual
		auditContent.SetCollectTemp()
	}
	return auditContent, res
}

func buildCarouselExt(extParam string) (string, *dynmdl.ResourceExt, error) {
	ext := &dynmdl.ResourceExt{}
	if err := json.Unmarshal([]byte(extParam), ext); err != nil {
		log.Error("Fail to unmarshal ResourceExt, ext=%s error=%+v", extParam, err)
		return "", nil, err
	}
	newExt := &dynmdl.ResourceExt{ImgUrl: ext.ImgUrl, Length: ext.Length, Width: ext.Width}
	res, err := json.Marshal(newExt)
	if err != nil {
		log.Error("Fail to marshal CarouselImg ext, ext=%+v error=%+v", newExt, err)
		return "", nil, err
	}
	return string(res), newExt, nil
}

func extractTsResourceIDs(modules []*dynmdl.ModuleExt) (aids []int64, actPids []int64) {
	for _, m := range modules {
		for _, r := range m.Resources {
			if r.ResourceID == 0 {
				continue
			}
			switch r.ResourceType {
			case pb.MixAvidType:
				aids = append(aids, r.ResourceID)
			case pb.MixActivity:
				actPids = append(actPids, r.ResourceID)
			}
		}
	}
	return aids, actPids
}

func setArcOfResDetail(detail *dynmdl.ResourceDetail, arcs map[int64]*arcapi.Arc) bool {
	arc, ok := arcs[detail.ResourceID]
	if !ok || arc == nil || !arc.IsNormal() {
		return false
	}
	detail.Arc = &dynmdl.ResourceArc{
		Title:   arc.Title,
		Pic:     arc.Pic,
		Danmuku: arc.Stat.Danmaku,
		View:    arc.Stat.View,
	}
	return true
}

func setActivityOfResDetail(detail *dynmdl.ResourceDetail, pages map[int64]*pb.NativePage) bool {
	page, ok := pages[detail.ResourceID]
	if !ok || page == nil || page.IsOffline() {
		return false
	}
	detail.Title = page.Title
	return true
}

func isPartitionChanged(upPart string, actRly *actGRPC.ActSubProtocolReply) bool {
	if actRly == nil {
		return false
	}
	var actPart string
	if actRly.Protocol != nil {
		actPart = actRly.Protocol.Types
	}
	if rebuildPartition(upPart) == rebuildPartition(actPart) {
		return false
	}
	var author string
	if actRly.Subject != nil {
		author = actRly.Subject.Author
	}
	return author != actmdl.ActAuthorUp
}

func rebuildPartition(partitions string) string {
	if partitions == "" {
		return ""
	}
	parts := strings.Split(partitions, ",")
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func canSetCollectTemp(tsRly *dynmdl.TsPageRly, arg *dynmdl.ParamMinePageSave, comparison int) bool {
	if tsRly.Template != dynmdl.UpTempCollect {
		return false
	}
	if tsRly.Partitions != arg.Partitions {
		return true
	}
	return (tsRly.State == pb.CheckOffline || tsRly.State == pb.WaitForCheck) && comparison != dynmdl.TsCompareNoChange
}

func isFirstAudit(natState int64) bool {
	// 首次审核中、首次审核被打回
	return natState == pb.CheckOffline || natState == pb.WaitForCheck
}
