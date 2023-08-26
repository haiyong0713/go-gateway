package native

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"
	tagecode "go-main/app/community/tag/ecode"

	"go-gateway/app/web-svr/native-page/admin/dao"
	admdl "go-gateway/app/web-svr/native-page/admin/model"
	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"
	xecode "go-gateway/app/web-svr/native-page/ecode"
	"go-gateway/app/web-svr/native-page/interface/api"
)

func (s *Service) ReOnline(c context.Context, arg *natmdl.OnlineParam) error {
	//结束时间
	if arg.Etime < arg.Stime || arg.Etime < time.Now().Unix() {
		return ecode.Error(ecode.RequestErr, "结束时间不合法")
	}
	pagesInfo, err := s.dao.FindPageByIds(c, []int64{arg.ID})
	if err != nil {
		return err
	}
	if len(pagesInfo) == 0 || pagesInfo[0] == nil {
		return ecode.Error(ecode.NothingFound, "活动配置不存在")
	}
	// 话题活动下线后支持修改stime和etime
	if pagesInfo[0].Type != natmdl.TopicType || pagesInfo[0].State != natmdl.OfflineState {
		return ecode.Error(ecode.RequestErr, "活动不支持重新上线")
	}
	if pagesInfo[0].SkipUrl == "" {
		//检查组件数量
		var mous []*natmdl.NatModule
		if mous, err = s.dao.ModulesInfo(c, pagesInfo[0].ID, []int{0}); err != nil {
			log.Error("s.dao.ModulesInfo error(%v)", err)
			return err
		}
		if len(mous) == 0 {
			err = ecode.Error(ecode.RequestErr, "该话题下没有一个有效的组件")
			return err
		}
	}
	if pageRes, err := s.dao.PageByFID(c, pagesInfo[0].ForeignID, natmdl.TopicType); err != nil || len(pageRes) > 0 {
		log.Error("s.dao.PageByFID(%d) error(%v)", pagesInfo[0].ForeignID, err)
		return ecode.Error(ecode.RequestErr, "已存在同名话题")
	}
	// 程序上防并发，与up主发起活动
	var ok bool
	if ok, err = s.dao.NtTsTitleUnique(c, pagesInfo[0].Title); err != nil {
		return err
	}
	if !ok {
		return ecode.Error(ecode.RequestErr, "该话题已存在")
	}
	req := make(map[string]interface{})
	req["stime"] = time.Unix(arg.Stime, 0).Format("2006-01-02 15:04:05")
	req["etime"] = time.Unix(arg.Etime, 0).Format("2006-01-02 15:04:05")
	req["state"] = natmdl.WaitForOnline
	return s.dao.ModifyPage(c, arg.ID, req)
}

// ModifyPage .
func (s *Service) ModifyPage(c context.Context, arg *natmdl.ModifyParam) error {
	pagesInfo, err := s.dao.FindPageByIds(c, []int64{arg.ID})
	if err != nil {
		return err
	}
	if len(pagesInfo) == 0 || pagesInfo[0] == nil {
		return ecode.NothingFound
	}
	req := make(map[string]interface{})
	req["act_type"] = arg.ActType
	req["hot"] = arg.Hot
	req["dynamic_id"] = arg.DynamicID
	req["attribute"] = arg.Attribute
	req["operator"] = arg.UserName
	if pagesInfo[0].FromType != 1 { //up主发起活动不允许修改uid
		req["related_uid"] = arg.RelatedUid
	}
	req["bg_color"] = arg.BgColor
	var confSort *natmdl.ConfSet
	if pagesInfo[0].Type == natmdl.MenuType {
		if arg.BgType == natmdl.BgTypeColor {
			confSort = &natmdl.ConfSet{
				BgType:         arg.BgType,
				TabTopColor:    arg.TabTopColor,
				TabMiddleColor: arg.TabMiddleColor,
				TabBottomColor: arg.TabBottomColor,
				FontColor:      arg.FontColor,
				BarType:        arg.BarType,
			}
		} else if arg.BgType == natmdl.BgTypeImage {
			confSort = &natmdl.ConfSet{
				BgType:    arg.BgType,
				BgImage1:  arg.BgImage1,
				BgImage2:  arg.BgImage2,
				FontColor: arg.FontColor,
				BarType:   arg.BarType,
			}
		}
	}
	if confSort != nil {
		var conStr []byte
		if conStr, err = json.Marshal(confSort); err != nil {
			return err
		}
		if len(conStr) > natmdl.MaxLen {
			return errors.Errorf("conf_set 字段过长,请联系开发处理")
		}
		req["conf_set"] = string(conStr)
	} else {
		req["conf_set"] = ""
	}
	err = s.dao.ModifyPage(c, arg.ID, req)
	if err != nil {
		log.Error("Fail to modify page, err=%+v", err)
		return err
	}
	//保存动态广场相关数据
	func() {
		if arg.ActType != natmdl.ActTypeBiz {
			return
		}
		req := &natmdl.EditParam{
			ID:          arg.ID,
			Validity:    arg.Validity,
			ValidStime:  arg.ValidStime,
			SquareTitle: arg.SquareTitle,
			SmallCard:   arg.SmallCard,
			BigCard:     arg.BigCard,
			Tids:        arg.Tids,
		}
		err := s.dao.UpDynExt(c, req)
		if err != nil {
			log.Error("Fail to modify native_page_dyn, err=%+v req=%+v", err, req)
		}
	}()
	return nil
}

// PageSave .
func (s *Service) PageSave(c context.Context, req *natmdl.AddPageParam) (pageID *natmdl.AddReID, err error) {
	var (
		fid     int64
		PageRes []*natmdl.PageParam
		ok      bool
		upgPage *natmdl.PageParam
	)
	if req.Type == natmdl.TopicType {
		fid, err = s.addTagIfNotExisted(c, req.Title)
		if err != nil {
			log.Error("Fail to add tag, title=%s error=%+v", req.Title, err)
			return
		}
		if PageRes, err = s.dao.PageByFID(c, fid, req.Type); err != nil {
			log.Error("[EditPage] s.dao.FindPageByFID() error(%v)", err)
			return
		}
		// 程序上防并发，与up主发起活动
		if ok, err = s.dao.NtTsTitleUnique(c, req.Title); err != nil {
			return
		}
		if !ok {
			err = ecode.Error(ecode.RequestErr, "该话题已存在")
			return
		}
		for _, v := range PageRes {
			if !api.IsFromTopicUpg(int32(v.FromType)) {
				err = ecode.Error(ecode.RequestErr, "该话题已存在")
				return
			}
			upgPage = v
		}
	}
	if upgPage != nil {
		if err = s.dao.DelPage(c, upgPage.ID, "system", "运营发起自动升级类型话题"); err != nil {
			log.Error("日志告警 自动升级类型话题发起失败-运营，pageID=%+v error=%+v", upgPage.ID, err)
			return nil, err
		}
	}
	natPage := &natmdl.PageParam{Title: req.Title, Creator: req.UserName, Operator: req.UserName, Type: req.Type, ForeignID: fid, RelatedUid: req.RelatedUid, ActType: req.ActType, FromType: natmdl.PageFromSystem, FirstPid: req.FirstPid}
	if req.Type == api.BottomType || req.Type == natmdl.MenuType || req.Type == natmdl.OgvType || req.Type == natmdl.PlayerType ||
		req.Type == natmdl.UgcType || req.Type == natmdl.SpaceType || req.Type == api.LiveTabType {
		natPage.State = 1 // menu tab的状态保存及上线
	}
	if pageID, err = s.dao.AddPage(c, natPage); err != nil {
		log.Error("[PageTitle] s.dao.AddPage() error(%v)", err)
		if upgPage != nil {
			log.Error("日志告警 自动升级类型话题发起失败-运营，pageID=%+v error=%+v", upgPage.ID, err)
		}
		return
	}
	func() {
		if req.ActType != natmdl.ActTypeBiz {
			return
		}
		req := &natmdl.EditParam{ID: pageID.ID, Validity: req.Validity, ValidStime: req.ValidStime}
		err := s.dao.UpDynExt(c, req)
		if err != nil {
			log.Error("Fail to save native_page_dyn, err=%+v req=%+v", err, req)
		}
	}()
	if upgPage != nil {
		if err := s.cache.Do(c, func(ctx context.Context) {
			_ = s.addao.ThirdRecord(ctx, upgPage.ID, dao.DataTypeNatSource, dao.ActTypeState2Operator, fmt.Sprintf("运营发起（%d）", pageID.ID), req.UserName)
		}); err != nil {
			log.Error("third_record fanout.Do() failed, req=%+v error=%+v", req, err)
		}
	}
	return
}

func (s *Service) addTagIfNotExisted(c context.Context, tagName string) (int64, error) {
	tagID, err := s.dao.NatTagID(c, tagName)
	if err == nil {
		return tagID, nil
	}
	if !ecode.EqualError(tagecode.TagNotExist, err) {
		log.Error("Fail to get tag, tagName=%s error=%+v", tagName, err)
		return 0, err
	}
	tag, err := s.dao.AddTag(c, tagName)
	if err != nil {
		log.Error("Fail to add tag, tagName=%s error=%+v", tagName, err)
		return 0, err
	}
	if tag == nil {
		return 0, errors.New("tag is nil")
	}
	return tag.GetId(), nil
}

// DelPage .
func (s *Service) DelPage(c context.Context, id int64, user, offReason string) (err error) {
	// menu下的page不支持删除
	if err = s.dao.DelPage(c, id, user, offReason); err != nil {
		log.Error("[DelPage] s.dao.DelPage() ID(%d) error(%v)", id, err)
		return
	}
	// 通知空间解绑
	if err = s.unbindSpace(c, id); err != nil {
		return
	}
	return
}

func (s *Service) unbindSpace(c context.Context, id int64) error {
	page, err := s.dao.FindPageById(c, id)
	if err != nil {
		return err
	}
	if page.RelatedUid == 0 || page.FromType != api.PageFromUid {
		return nil
	}
	userSpace, err := s.dao.UserSpaceByMid(c, page.RelatedUid)
	if err != nil {
		return err
	}
	// 用户可能先关闭展示
	if userSpace == nil || userSpace.State != api.USpaceOnline {
		log.Info("Space has already unbind, userSpace=%+v", userSpace)
		return nil
	}
	if userSpace.PageId != id {
		log.Info("Skip unbinding user_space, page is not equal, pageID=%+v spacePageID=%+v", id, userSpace.PageId)
		return nil
	}
	success := false
	defer func() {
		_ = s.dao.ResetUserSpace(c, page.RelatedUid, id, api.USpaceOfflineNormal, userSpace.State)
		if !success {
			log.Error("日志告警 Native活动下线解绑失败，userSpace=%+v error=%+v", userSpace, err)
			return
		}
		_ = s.sendSpaceOfflineLetter(c, page.RelatedUid)
	}()
	success, err = s.dao.UpActivityTab(c, page.RelatedUid, 0, userSpace.Title, id)
	if err != nil {
		return err
	}
	if !success {
		return xecode.SpaceUnbindFail
	}
	return nil
}

// EditPage .
func (s *Service) EditPage(c context.Context, natPage *natmdl.EditParam, pType int) (err error) {
	oldPage, err := s.judge(c, natPage)
	if err != nil {
		return
	}
	if oldPage == nil {
		return ecode.NothingFound
	}
	err = s.dao.UpdatePage(c, natPage, pType)
	if err != nil {
		return
	}
	if oldPage.Type == natmdl.TopicType { //保存广场信息
		if err = s.dao.UpDynExt(c, natPage); err != nil {
			log.Error("s.dao.UpDynExt %v error(%v)", natPage, err)
			return
		}
	}
	// 保存白名单对应数据源
	func() {
		if natPage.WhiteValue == "" {
			return
		}
		extReq := &natmdl.PageExt{Pid: natPage.ID, WhiteValue: natPage.WhiteValue}
		err := s.dao.UpPageExt(c, extReq)
		if err != nil {
			log.Error("Fail to modify native_page_ext, err=%+v req=%+v", err, natPage)
		}
	}()
	return
}

// SearchPage native page .
func (s *Service) SearchPage(c context.Context, param *natmdl.SearchParam) (res *natmdl.SearchRes, err error) {
	if res, err = s.dao.SearchPage(c, param); err != nil {
		log.Error("[SearchPage] s.dao.SearchPage() error(%v)", err)
	}
	return
}

// UpPage native page .
func (s *Service) UpPage(c context.Context, param *natmdl.UpParam) (*natmdl.UpReply, error) {
	res, err := s.dao.SearchPage(c, &natmdl.SearchParam{ActOrigin: param.ActOrigin, Pn: param.Pn, Ps: param.Ps, States: []int{1}, Ptypes: []int{1}, PageParam: natmdl.PageParam{Title: param.Title, RelatedUid: param.Uid}, FromTypes: []int{1}})
	if err != nil {
		log.Error("s.dao.SearchPage(%v) error(%+v)", param, err)
		return nil, err
	}
	if res == nil {
		return &natmdl.UpReply{Page: natmdl.Cfg{Total: 0}}, nil
	}
	var uids, tagIDs []int64
	for _, v := range res.Item {
		if v == nil {
			continue
		}
		if v.RelatedUid > 0 {
			uids = append(uids, v.RelatedUid)
		}
		if v.ForeignID > 0 {
			tagIDs = append(tagIDs, v.ForeignID)
		}
	}
	var users map[int64]*acccli.Info
	eg := errgroup.WithContext(c)
	// 批量获取用户信息
	if len(uids) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			users, e = s.dao.Infos3(ctx, uids)
			if e != nil {
				log.Error("s.dao.Infos3(%v) error(%v)", uids, e)
				e = nil
			}
			return
		})
	}
	var tags map[int64]*tagrpc.Tag
	//批量获取tag状态
	if len(tagIDs) > 0 {
		eg.Go(func(ctx context.Context) (e error) {
			tags, e = s.dao.Tags(ctx, tagIDs)
			if e != nil {
				log.Error("s.dao.Tags(%v) error(%v)", uids, e)
				e = nil
			}
			return
		})
	}
	if err = eg.Wait(); err != nil {
		return nil, err
	}
	rly := &natmdl.UpReply{}
	rly.Page = natmdl.Cfg{Total: res.Page.Total, Num: res.Page.Num, Size: res.Page.Size}
	for _, v := range res.Item {
		if v == nil {
			continue
		}
		tmp := &natmdl.UpItem{UID: v.RelatedUid, PageID: v.ID, TagID: v.ForeignID, TagName: v.Title, ActOrigin: v.ActOrigin}
		if uVal, ok := users[v.RelatedUid]; ok && uVal != nil {
			tmp.Name = uVal.Name
		}
		if tVal, ok := tags[v.ForeignID]; ok && tVal != nil {
			tmp.TagType = tVal.Type
		}
		rly.Item = append(rly.Item, tmp)
	}
	return rly, nil
}

// FindPage native page .
func (s *Service) FindPage(c context.Context, title string, id int64) (*natmdl.NatPageExt, error) {
	var (
		tagID int64
		err   error
	)
	if title != "" {
		if tagID, err = s.dao.NatTagID(c, title); err != nil {
			return nil, err
		}
	}
	res, err := s.dao.FindPageByID(c, tagID, id, natmdl.TopicType, []int64{natmdl.WaitForOnline, natmdl.OnlineState})
	if err != nil {
		log.Error("[FindPage] s.dao.FindPage() error(%v)", err)
		return nil, err
	}
	if res == nil {
		return nil, ecode.NothingFound
	}
	pageDyn, err := s.dao.DynExtByPid(c, res.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error("[FindPage]  s.dao.DynExtByPid(%d) error(%v)", res.ID, err)
		return nil, err
	}
	var topicInfo []*natmdl.TopicInfo
	if pageDyn != nil && pageDyn.Tids != "" {
		func() {
			//获取name
			tidStr := strings.Split(pageDyn.Tids, ",")
			var tids []int64
			for _, v := range tidStr {
				tid, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					continue
				}
				if tid > 0 {
					tids = append(tids, tid)
				}
			}
			tagInfo, err := s.dao.Tags(c, tids)
			if err != nil {
				log.Error("s.dao.Tags(%v) error(%v)", tids, err)
				return
			}
			for _, v := range tids {
				if tv, ok := tagInfo[v]; !ok || tv == nil {
					continue
				}
				topicInfo = append(topicInfo, &natmdl.TopicInfo{Tid: v, Name: tagInfo[v].Name})
			}
		}()
	}
	extPage, err := s.dao.FindExtByPid(c, res.ID)
	if err != nil {
		log.Error("s.dao.FindExtByPid(%v) error(%v)", res.ID, err)
		return nil, err
	}
	return &natmdl.NatPageExt{NatPage: res, PageDyn: &natmdl.PageDynRly{PageDyn: pageDyn, TopicInfo: topicInfo}, Ext: extPage}, nil
}

// Module .
func (s *Service) Module(c context.Context, param *natmdl.ModuleParam) (*natmdl.SaveModuleReply, error) {
	var (
		portion = &natmdl.JsonData{}
	)
	if err := json.Unmarshal([]byte(param.Data), portion); err != nil {
		log.Error("[Module] json.Unmarshal() json(%s) error(%v)", param.Data, err)
		return nil, err
	}
	// 至少有一个组件或者有一个跳转链接
	if (len(portion.Modules) == 0 || portion.Structure == nil || len(portion.Structure.Root.Children) == 0) &&
		(portion.Base == nil || len(portion.Base.Modules) == 0) {
		return nil, ecode.Error(ecode.RequestErr, "应该至少有一个组件")
	}
	ver := fmt.Sprintf("%d-%s", time.Now().UnixNano()/1e6, "admin")
	if err := s.dao.SaveModule(c, param.NativeID, portion, ver); err != nil {
		log.Error("[module] s.dao.SaveModule() error(%v)", err)
		return nil, err
	}
	return &natmdl.SaveModuleReply{Ver: ver}, nil
}

// TsOnline .
// nolint:gocognit
func (s *Service) TsOnline(c context.Context, arg *natmdl.TsOnlineReq) error {
	// tsid数据获取
	tsid, err := strconv.ParseInt(arg.Oid, 10, 64)
	if err != nil || tsid <= 0 {
		log.Error("TsOnline oid not int %s", arg.Oid)
		return ecode.RequestErr
	}
	// 获取pid信息
	pageInfo, err := s.dao.FindPage(c, "", arg.Pid, natmdl.TopicType, []int64{natmdl.WaitForCheck, natmdl.OnlineState, natmdl.WaitForOnline})
	if err != nil {
		// 发生未知错误
		log.Error("TsOnline s.dao.FindPage(%d) error(%v)", arg.Pid, err)
		return err
	}
	//只能修改up主发起的活动
	if pageInfo == nil || pageInfo.ID == 0 || pageInfo.FromType != natmdl.PageFromUid {
		log.Error("TsOnline page not exist or not from up")
		return ecode.RequestErr
	}
	//获取tsid
	tsInfo, err := s.dao.FindTsPage(c, tsid)
	if err != nil {
		// 发生未知错误
		log.Error("TsOnline s.dao.FindTsPage(%d) error(%v)", tsid, err)
		return err
	}
	if ifAuditContentChanged(arg.AuditTime, tsInfo.AuditTime) {
		log.Error("Fail to handle tsOnline, audit has changed, realTime=%+v arg=%+v", tsInfo.AuditTime, arg)
		return xecode.NaAuditChanged
	}
	if tsInfo == nil || tsInfo.Pid != arg.Pid || tsInfo.State != natmdl.TsWaitOnline {
		log.Error("TsOnline ts page not exist or not right pid")
		return ecode.RequestErr
	}
	if tsInfo.AuditType == natmdl.TsAutoAudit {
		return s.CommitTsModule(c, tsid, pageInfo, tsInfo)
	}
	if arg.State == 1 {
		if err := s.handleTemplate(c, pageInfo, arg.AuditContent); err != nil {
			return err
		}
	}
	//if arg.AuditContent > 0 && !arg.AuditContent.IsModule() {
	//	return nil
	//}
	// 首次审核通过,兜底，需要check titile
	var upgPage *natmdl.PageParam
	if arg.State == 1 && pageInfo.State == natmdl.WaitForCheck {
		var fidInfos []*natmdl.PageParam
		if fidInfos, err = s.dao.PageByFID(c, pageInfo.ForeignID, natmdl.TopicType); err != nil {
			log.Error("[EditPage] s.dao.PageByFID error(%v)", err)
			return err
		}
		isNotExist := true
		for _, v := range fidInfos {
			if api.IsFromTopicUpg(int32(v.FromType)) {
				upgPage = v
				continue
			}
			if v.ID != pageInfo.ID { //排除自己,已有上线的同名话题
				isNotExist = false
				break
			}
		}
		if !isNotExist {
			arg.State = 2
			arg.Reason = "已存在相同活动"
		}
		// 程序上防并发，与up主发起活动
		var tOK bool
		if tOK, err = s.dao.NtTsTitleUnique(c, pageInfo.Title); err != nil {
			return err
		}
		if !tOK {
			return ecode.Error(ecode.RequestErr, "话题并发，请稍后重试")
		}
	}
	//上报
	beginDate, _ := strconv.ParseInt(time.Now().Format("20060102"), 10, 64)
	applyDate, _ := strconv.ParseInt(tsInfo.Ctime.Time().Format("20060102"), 10, 64)
	infoParam := natmdl.NtCloudInfo{Ctime: int64(tsInfo.Ctime), Mid: pageInfo.RelatedUid, BeginDate: int32(beginDate), ApplyDate: int32(applyDate), ActivityName: pageInfo.Title, TopicID: pageInfo.ForeignID}
	// 发私信 SenderUID-bilibili活动
	lotReq := &admdl.LetterParam{RecverIDs: []uint64{uint64(pageInfo.RelatedUid)}, SenderUID: s.c.Up.ActSenderUid, MsgType: 10}
	if arg.State != 1 { //审核不通过
		infoParam.Status = "failed"
		//首次审核不通过
		if pageInfo.State == natmdl.WaitForCheck {
			infoParam.Status = "finished"
			// 更page
			if err = s.dao.ModifyPage(c, arg.Pid, map[string]interface{}{"state": natmdl.CheckOffline}); err != nil {
				log.Error("TsOnline s.dao.ModifyPage(%d) offline error(%v)", arg.Pid, err)
				return err
			}
		}
		// 更新ts_module
		if err = s.dao.ModifyTsPage(c, tsid, map[string]interface{}{"state": natmdl.TsOffline, "msg": arg.Reason}); err != nil {
			log.Error("TsOnline s.dao.ModifyTsPage(%d) offline error(%v)", arg.Pid, err)
			return err
		}
		// 发私信
		lotReq.NotifyCode = s.c.Up.NotifyCodeNotPass
		lotReq.Params = dao.BuildNotifyParams([]string{tsInfo.Title, arg.Reason})
		lotReq.JumpUri = fmt.Sprintf("https://www.bilibili.com/blackboard/up-sponsor.html?act_from=message&act_id=%d", pageInfo.ID)
		if pageInfo.State == natmdl.OnlineState {
			lotReq.NotifyCode = s.c.Up.NotifyCodeNotPassEdit
			lotReq.JumpUri = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d", pageInfo.ID)
			lotReq.JumpUri2 = fmt.Sprintf("https://www.bilibili.com/blackboard/up-sponsor.html?act_from=message&act_id=%d", pageInfo.ID)
		}
		// 更新空间配置
		_ = s.dao.ResetUserSpace(c, pageInfo.RelatedUid, arg.Pid, api.USpaceOfflineAuditFail, api.USpaceWaitingOnline)
	} else {
		infoParam.Status = "pass"
		lotReq.NotifyCode = s.c.Up.NotifyCodePass
		lotReq.Params = dao.BuildNotifyParams([]string{tsInfo.Title})
		lotReq.JumpUri = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d", pageInfo.ID)
		// 首次审核通过，话题自动升级处理
		if pageInfo.State == natmdl.OnlineState {
			lotReq.NotifyCode = s.c.Up.NotifyCodePassEdit
		} else {
			if err := s.handleUpgPage(c, upgPage, pageInfo); err != nil {
				return err
			}
		}
		// 获取modules info 非必填
		if err = s.CommitTsModule(c, tsid, pageInfo, tsInfo); err != nil {
			if upgPage != nil {
				log.Error("日志告警 自动升级类型话题发起失败-UP主，pageID=%+v error=%+v", upgPage.ID, err)
			}
			return err
		}
	}
	//上报数据
	if err := s.cache.Do(c, func(c context.Context) {
		s.infocSave(infoParam)
	}); err != nil {
		log.Error("infoc fanout.Do failed, arg=%+v error=%+v", arg, err)
	}
	// 发私信
	if _, err = s.addao.SendLetter(c, lotReq); err != nil {
		log.Error("UpActEdit s.dao.SendLetter error(%v)", err)
	}
	return err
}

func (s *Service) CommitTsModule(c context.Context, tsid int64, pageInfo *natmdl.NatPage, tsInfo *natmdl.NatTsPage) error {
	modus, err := s.dao.GetTsModules(c, tsid)
	if err != nil {
		log.Error("TsOnline s.dao.GetTsModules(%d)error(%v)", tsid, err)
		return err
	}
	if err = s.dao.CommitTsModule(c, pageInfo, modus, tsid, tsInfo); err != nil {
		log.Error("TsOnline  s.initPage(%d)error(%v)", pageInfo.ID, err)
		return err
	}
	return nil
}

func (s *Service) handleTemplate(c context.Context, page *natmdl.NatPage, auditContent natmdl.AuditContent) error {
	if !auditContent.IsCollectTemp() {
		return nil
	}
	pageSources, err := s.dao.PageSourcesByPid(c, page.ID)
	if err != nil {
		return err
	}
	source, ok := pageSources[admdl.ActTypeCollect]
	if !ok {
		log.Warn("native_page_source not found, act_type=%d", admdl.ActTypeCollect)
		return nil
	}
	if source.Sid != 0 {
		return s.addao.UpdateActSubject(c, source.Sid, &admdl.AddActSubjectReq{Author: admdl.ActAuthorUp, Types: source.Partitions})
	}
	sid, err := s.addao.AddActSubject(c, &admdl.AddActSubjectReq{
		Type:   admdl.ActTypeCollect,
		Stime:  time.Now(),
		Etime:  time.Unix(2147356800, 0),
		Author: admdl.ActAuthorUp,
		Name:   page.Title,
		Types:  source.Partitions,
		Tags:   page.Title,
	})
	if err != nil {
		return err
	}
	return s.dao.UpdatePageSource(c, source.Id, sid)
}

// SearchModule .
func (s *Service) SearchModule(c context.Context, param *natmdl.SearchModule) (res *natmdl.ModuleRes, err error) {
	if res, err = s.dao.SearchModule(c, param); err != nil {
		log.Error("[SearchModule] s.dao.SearchModule() error(%v)", err)
	}
	return
}

// PageSkipUrl .
func (s *Service) PageSkipUrl(c context.Context, param *natmdl.EditParam, pType int) (err error) {
	oldPage, err := s.judge(c, param)
	if err != nil {
		return
	}
	if oldPage == nil {
		return ecode.NothingFound
	}
	if err = s.dao.PageSkipUrl(c, param, pType); err != nil {
		log.Error("[PageTitle] s.dao.AddPage() error(%v)", err)
		return
	}
	if oldPage.Type == natmdl.TopicType { //保存广场信息
		if err = s.dao.UpDynExt(c, param); err != nil {
			log.Error("s.dao.UpDynExt %v error(%v)", param, err)
			return
		}
	}
	return
}

func (s *Service) judge(c context.Context, natPage *natmdl.EditParam) (pageRes *natmdl.FindRes, err error) {
	//结束时间
	if natPage.Etime < natPage.Stime || natPage.Etime < time.Now().Unix() {
		err = ecode.Error(ecode.RequestErr, "结束时间不合法")
		return
	}
	if pageRes, err = s.dao.PageByID(c, natPage.ID); err != nil {
		log.Error("[EditPage]  s.dao.FindPage() error(%v)", err)
		return
	}
	// 已上线则不允许修改活动开始时间
	if pageRes != nil {
		if pageRes.State == 1 && int64(pageRes.Stime) != natPage.Stime {
			err = ecode.Error(ecode.RequestErr, "该话题是上线状态，不允许修改活动开始时间")
		}
		return
	}
	return
}

// nolint:gocognit
func (s *Service) SaveTab(c context.Context, req *natmdl.SaveTabReq, loginUser string) (rly *natmdl.SaveTabRly, err error) {
	rly = &natmdl.SaveTabRly{}
	// pid 唯一校验
	if len(req.TabModules) != 0 {
		uniqErr := ecode.Error(ecode.RequestErr, "每个底栏的页面应唯一")
		pidMap := make(map[int32]struct{})
		pids := make([]int32, 0, len(req.TabModules))
		for _, tabModule := range req.TabModules {
			if tabModule.Category == natmdl.CategoryPage && tabModule.Pid != 0 {
				if _, exist := pidMap[tabModule.Pid]; exist {
					return nil, uniqErr
				}
				pidMap[tabModule.Pid] = struct{}{}
				pids = append(pids, tabModule.Pid)
			}
		}
		if len(pids) > 0 {
			existModules, err := s.dao.GetTabModuleByPids(c, pids, natmdl.CategoryPage)
			if err != nil {
				return nil, err
			}
			if len(existModules) > 0 {
				for _, existModule := range existModules {
					if existModule.TabId != req.ID {
						return nil, uniqErr
					}
				}
			}
		}
	}
	// tab
	if req.Etime != 0 && req.Stime != 0 && req.Etime < req.Stime {
		err = ecode.Error(ecode.RequestErr, "下线时间应大于上线时间")
		return
	}
	tabData := natmdl.TabData{
		Title:         req.Title,
		Stime:         xtime.Time(req.Stime),
		Etime:         xtime.Time(req.Etime),
		State:         natmdl.TabStateValid,
		Operator:      loginUser,
		BgType:        req.BgType,
		BgImg:         req.BgImg,
		BgColor:       req.BgColor,
		IconType:      req.IconType,
		ActiveColor:   req.ActiveColor,
		InactiveColor: req.InactiveColor,
	}
	tx := s.dao.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Error("SaveTab panic: %v\n%s", r, buf)
		}
		if err != nil {
			if err1 := tx.Rollback().Error; err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("tx.Commit() error(%v)", err)
			return
		}
	}()
	tabId := req.ID
	if tabId == 0 {
		// create
		tab := natmdl.Tab{TabData: tabData, Creator: loginUser}
		tabId, err = s.dao.CreateTab(c, tx, &tab)
	} else {
		// update
		var stime, etime string
		if tabData.Stime != 0 {
			stime = time.Unix(int64(tabData.Stime), 0).Format("2006-01-02 15:04:05")
		}
		if tabData.Etime != 0 {
			etime = time.Unix(int64(tabData.Etime), 0).Format("2006-01-02 15:04:05")
		}
		tabMap := map[string]interface{}{
			"title":          tabData.Title,
			"stime":          stime,
			"etime":          etime,
			"operator":       tabData.Operator,
			"bg_type":        tabData.BgType,
			"bg_img":         tabData.BgImg,
			"bg_color":       tabData.BgColor,
			"icon_type":      tabData.IconType,
			"active_color":   tabData.ActiveColor,
			"inactive_color": tabData.InactiveColor,
		}
		err = s.dao.UpdateTabById(c, tx, tabId, tabMap)
	}
	if err != nil {
		return
	}
	// tab_module
	tabModules, err := s.dao.GetTabModuleByTabIds(c, []int32{tabId})
	if err != nil {
		return
	}
	// 先更新为无效
	if err = s.dao.UpdateTabModulesByTabId(c, tx, tabId, map[string]interface{}{"state": natmdl.TabModuleStateInValid}); err != nil {
		return
	}
	// 更新数据、插入数据
	id2TabModule := make(map[int32]*natmdl.TabModule)
	for _, tabModule := range tabModules {
		id2TabModule[tabModule.ID] = tabModule
	}
	for _, tm := range req.TabModules {
		tabModuleData := natmdl.TabModuleData{
			Title:       tm.Title,
			TabId:       tabId,
			State:       natmdl.TabModuleStateValid,
			Operator:    loginUser,
			ActiveImg:   tm.ActiveImg,
			InactiveImg: tm.InactiveImg,
			Category:    tm.Category,
			Pid:         tm.Pid,
			Url:         tm.Url,
			Rank:        tm.Rank,
		}
		if _, exist := id2TabModule[tm.ID]; !exist {
			// 新增：ID未传
			if tm.ID != 0 {
				continue
			}
			tabModule := natmdl.TabModule{TabModuleData: tabModuleData}
			if _, err = s.dao.CreateTabModule(c, tx, &tabModule); err != nil {
				return
			}
		} else {
			// 更新：ID已传并存在
			tabModuleMap := map[string]interface{}{
				"title":        tabModuleData.Title,
				"operator":     tabModuleData.Operator,
				"state":        natmdl.TabModuleStateValid,
				"active_img":   tabModuleData.ActiveImg,
				"inactive_img": tabModuleData.InactiveImg,
				"category":     tabModuleData.Category,
				"pid":          tabModuleData.Pid,
				"url":          tabModuleData.Url,
				"rank":         tabModuleData.Rank,
			}
			if err = s.dao.UpdateTabModulesById(c, tx, tm.ID, tabModuleMap); err != nil {
				return
			}
		}
	}
	rly.ID = tabId
	return
}

func (s *Service) EditTab(c context.Context, id int32, stime, etime int64) (err error) {
	if stime != 0 && etime != 0 && stime > etime {
		err = ecode.Error(ecode.RequestErr, "下线时间应大于上线时间")
		return
	}
	var sTime, eTime string
	if stime != 0 {
		sTime = time.Unix(stime, 0).Format("2006-01-02 15:04:05")
	}
	if etime != 0 {
		eTime = time.Unix(etime, 0).Format("2006-01-02 15:04:05")
	}
	tabMap := map[string]interface{}{"stime": sTime, "etime": eTime}
	err = s.dao.UpdateTabById(c, nil, id, tabMap)
	return
}

func (s *Service) SearchTab(c context.Context, req *natmdl.SearchTabReq) (rly *natmdl.SearchTabRly, err error) {
	rly = &natmdl.SearchTabRly{
		Total: 0,
		List:  make([]*natmdl.SearchTabItem, 0),
	}
	// 获取tab
	tabList, err := s.dao.SearchTab(c, req)
	if err != nil || tabList.Total == 0 {
		return
	}
	tabs := tabList.List
	// 获取tab_id
	tabIds := make([]int32, 0, len(tabs))
	for _, tab := range tabs {
		tabIds = append(tabIds, tab.ID)
	}
	// 获取tab_module
	tabModules, err := s.dao.GetTabModuleByTabIds(c, tabIds)
	if err != nil {
		return
	}
	// 获取page
	id2TapModule := make(map[int32][]*natmdl.TabModule)
	pageMap := make(map[int64]*natmdl.NatPage)
	if len(tabModules) != 0 {
		// 获取pid
		pids := make([]int64, 0, len(tabModules))
		for _, tabModule := range tabModules {
			id2TapModule[tabModule.TabId] = append(id2TapModule[tabModule.TabId], tabModule)
			if tabModule.Category == natmdl.CategoryPage {
				pids = append(pids, int64(tabModule.Pid))
			}
		}
		// 获取pid list
		pages, err := s.dao.FindPageByIds(c, pids)
		if err != nil {
			return rly, err
		}
		for _, page := range pages {
			pageMap[page.ID] = page
		}
	}
	// 拼接数据
	nowTime := time.Now().Unix()
	for _, tab := range tabs {
		// check state
		listItem := natmdl.SearchTabItem{
			Tab:        *tab,
			TabModules: make([]*natmdl.SearchTabModuleItem, 0, len(id2TapModule[tab.ID])),
		}
		if listItem.Tab.State == natmdl.TabStateValid && int64(listItem.Tab.Stime) > 0 && int64(listItem.Tab.Stime) <= nowTime && (int64(listItem.Tab.Etime) <= 0 || int64(listItem.Tab.Etime) >= nowTime) {
			listItem.Tab.State = natmdl.TabStateValid
		} else {
			listItem.Tab.State = natmdl.TabStateInvalid
		}
		if tabModules, exist := id2TapModule[tab.ID]; exist {
			for _, tabModule := range tabModules {
				// topicName
				topicName := ""
				if page, exist := pageMap[int64(tabModule.Pid)]; exist && tabModule.Category == natmdl.CategoryPage {
					topicName = page.Title
				}
				listItem.TabModules = append(listItem.TabModules, &natmdl.SearchTabModuleItem{
					TabModule: *tabModule,
					TopicName: topicName,
				})
			}
		}
		rly.List = append(rly.List, &listItem)
	}
	rly.Total = tabList.Total
	return
}

func (s *Service) GetTabOfPage(c context.Context, pid int32) (*natmdl.SearchTabItem, error) {
	// 获取tab、tab_modules
	targetModule, err := s.dao.GetTabModuleByPid(c, pid)
	if err != nil {
		return nil, err
	}
	tab, err := s.dao.GetTabById(c, targetModule.TabId)
	if err != nil {
		return nil, err
	}
	tabModules, err := s.dao.GetTabModuleByTabIds(c, []int32{tab.ID})
	if err != nil {
		return nil, err
	}
	// 获取page
	pageMap := make(map[int64]*natmdl.NatPage)
	if len(tabModules) != 0 {
		// 获取pid
		pids := make([]int64, 0, len(tabModules))
		for _, tabModule := range tabModules {
			if tabModule.Category == natmdl.CategoryPage {
				pids = append(pids, int64(tabModule.Pid))
			}
		}
		// 获取pid list
		pages, err := s.dao.FindPageByIds(c, pids)
		if err != nil {
			return nil, err
		}
		for _, page := range pages {
			pageMap[page.ID] = page
		}
	}
	// 拼接数据
	searchTabItem := natmdl.SearchTabItem{
		Tab:        *tab,
		TabModules: make([]*natmdl.SearchTabModuleItem, 0, len(tabModules)),
	}
	for _, tabModule := range tabModules {
		// topicName
		topicName := ""
		if page, exist := pageMap[int64(tabModule.Pid)]; exist && tabModule.Category == natmdl.CategoryPage {
			topicName = page.Title
		}
		searchTabItem.TabModules = append(searchTabItem.TabModules, &natmdl.SearchTabModuleItem{
			TabModule: *tabModule,
			TopicName: topicName,
		})
	}
	return &searchTabItem, nil
}

func (s *Service) SpaceOffline(c context.Context, req *natmdl.SpaceOfflineReq) error {
	switch req.TabType {
	case natmdl.TabTypeUpAct:
		return s.spaceOfflineUpAct(c, req)
	case natmdl.TabTypeNapage:
		return s.spaceOfflineNapage(c, req)
	}
	return nil
}

func (s *Service) spaceOfflineUpAct(c context.Context, req *natmdl.SpaceOfflineReq) error {
	userSpace, err := s.dao.UserSpaceByMid(c, req.Mid)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if userSpace == nil {
		return ecode.NothingFound
	}
	if userSpace.PageId != req.PageID {
		return xecode.UpBindOtherPage
	}
	if userSpace.State != api.USpaceOnline {
		return xecode.NotOnline
	}
	if err := s.dao.UpdateUserSpaceState(c, userSpace.Id, userSpace.PageId, api.USpaceOfflineNormal, api.USpaceOnline); err != nil {
		return err
	}
	// 发私信
	_ = s.sendSpaceOfflineLetter(c, req.Mid)
	return nil
}

func (s *Service) spaceOfflineNapage(c context.Context, req *natmdl.SpaceOfflineReq) error {
	return s.sendSpaceOfflineLetter(c, req.Mid)
}

func (s *Service) sendSpaceOfflineLetter(c context.Context, mid int64) error {
	// 发私信
	letterReq := &admdl.LetterParam{
		RecverIDs:  []uint64{uint64(mid)},
		SenderUID:  s.c.Up.ActSenderUid,
		MsgType:    10,
		NotifyCode: s.c.Up.NotifyCodeSpaceOff,
	}
	_, err := s.addao.SendLetter(c, letterReq)
	if err != nil {
		log.Error("Fail to send space_offline_letter, mid=%+v error=%+v", mid, err)
		return err
	}
	return nil
}

func (s *Service) handleUpgPage(c context.Context, upgPage *natmdl.PageParam, page *natmdl.NatPage) error {
	if upgPage == nil {
		return nil
	}
	if err := s.dao.DelPage(c, upgPage.ID, "system", "UP主发起自动升级类型话题"); err != nil {
		log.Error("日志告警 自动升级类型话题发起失败-UP主，pageID=%+v error=%+v", upgPage.ID, err)
		return err
	}
	_ = s.cache.Do(c, func(ctx context.Context) {
		username := strconv.FormatInt(page.RelatedUid, 10)
		if user, _ := s.dao.Info3(ctx, page.RelatedUid); user != nil {
			username = user.Name
		}
		_ = s.addao.ThirdRecord(ctx, upgPage.ID, dao.DataTypeNatSource, dao.ActTypeState2Up, fmt.Sprintf("UP主发起（%d）", page.ID), fmt.Sprintf("UP主：%s", username))
	})
	return nil
}

func ifAuditContentChanged(reqTime string, auditTime int64) bool {
	// 兼容老数据
	if reqTime == "" || reqTime == "0" || auditTime == 0 {
		return false
	}
	compareTime, err := strconv.ParseInt(reqTime, 10, 64)
	if err != nil {
		log.Error("Fail to parse reqTime, reqTime=%+v error=%+v", reqTime, err)
		return true
	}
	return compareTime != auditTime
}
