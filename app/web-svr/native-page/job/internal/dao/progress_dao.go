package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/credis"
	"go-common/library/railgun.v2/message"
	"go-common/library/railgun.v2/processor/single"
	"strconv"
	"strings"
	"sync"

	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"
	"go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/app/web-svr/native-page/job/internal/model"
	v1 "go-gateway/app/web-svr/native-page/job/internal/model/nat_event"

	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	_type "git.bilibili.co/bapis/bapis-go/push/service/broadcast/type"
	v2 "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/net/rpc/warden"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/sync/pipeline/fanout"
	"go-common/library/utils/collection"
	"go-common/library/xstr"
)

type progressCfg struct {
	Client                   *warden.ClientConfig
	Token                    string
	UserToken                string
	PageProgressParamsExpire int64
}

type progressDao struct {
	cfg    *progressCfg
	db     *sql.DB
	cache  *fanout.Fanout
	redis  credis.Redis
	client v2.BroadcastAPIClient
	actDao *activityDao
	// cache
	sid2GroupParams map[int64]map[int64][]*model.ProgressParam //sid=>groupID:[]param
	pageRelations   map[int64]int64                            //子父页面关系
}

func newProgressDao(cfg *progressCfg, db *sql.DB, cache *fanout.Fanout, r credis.Redis, actDao *activityDao) *progressDao {
	d := &progressDao{
		cfg:    cfg,
		db:     db,
		cache:  cache,
		redis:  r,
		actDao: actDao,
	}
	var err error
	if d.client, err = v2.NewClient(cfg.Client); err != nil {
		panic(fmt.Sprintf("Fail to new broadcastClient, config=%+v error=%+v", cfg.Client, err))
	}
	return d
}

const (
	targetPath     = "bilibili.broadcast.message.main.NativePageEvent"
	userTargetPath = "/bilibili.broadcast.message.main.NativePage/WatchNotify"
	roomID         = "native-page://%d" //native-page://{page_id}
)

const (
	_counterSep = "_ss_"
)

var (
	progressSQL = fmt.Sprintf(
		"SELECT `id`,`native_id`,`f_id`,`width`,`ukey` FROM `native_module` WHERE `category`=%d and `state`=%d",
		model.CategoryProgress, model.Online)
	clickProgressSQL = fmt.Sprintf(
		"SELECT `id`,`module_id`,`foreign_id`,`tip`,`unfinished_image` FROM `native_click` where `type`=%d and `state`=%d",
		model.ClickTypProgress, model.Online)
	childPageSQL = fmt.Sprintf("SELECT `module_id`, `foreign_id` FROM `native_mixture_ext` WHERE `m_type`=%d AND `state`=%d",
		model.MixTypeTab, model.Online)
	parentPageSQL = fmt.Sprintf("SELECT `id`, `native_id` FROM `native_module` WHERE `category` in (%s) AND `state`=%d",
		xstr.JoinInts([]int64{model.CategoryTab, model.CategorySelect}), model.Online)
	moduleIDSQL   = "SELECT `id`,`native_id` FROM `native_module` WHERE `id` in (%s)"
	pagesByIdsSQL = "SELECT `id`,`state` FROM `native_page` WHERE `id` in (%s)"
)

func (d *progressDao) PushRoom(c context.Context, event *v1.NativePageEvent) (*v2.PushRoomReply, error) {
	body, err := ptypes.MarshalAny(event)
	if err != nil {
		log.Errorc(c, "Fail to marshal nativePageEvent, event=%+v error=%+v", event, err)
		return nil, err
	}
	req := &v2.PushRoomReq{
		Msg: &_type.Message{
			TargetPath: targetPath,
			Body:       body,
		},
		RoomId: fmt.Sprintf(roomID, event.PageID),
		Token:  d.cfg.Token,
	}
	reply, err := d.client.PushRoom(c, req)
	log.Info("progress-push-room, event=%+v reply=%+v", event, reply)
	if err != nil {
		log.Errorc(c, "Fail to push native-page-room, event=%+v error=%+v", event, err)
		return nil, err
	}
	return reply, nil
}

func (d *progressDao) PushMids(c context.Context, event *v1.NativePageEvent, mids []int64) (*v2.PushMidsReply, error) {
	body, err := ptypes.MarshalAny(event)
	if err != nil {
		log.Errorc(c, "Fail to marshal nativePageEvent, event=%+v error=%+v", event, err)
		return nil, err
	}
	req := &v2.PushMidsReq{
		Msg: &_type.Message{
			TargetPath: userTargetPath,
			Body:       body,
		},
		Mids:  mids,
		Token: d.cfg.UserToken,
	}
	reply, err := d.client.PushMids(c, req)
	log.Info("progress-push-mids, event=%+v mids=%+v reply=%+v", event, mids, reply)
	if err != nil {
		log.Errorc(c, "Fail to push native-page-mids, event=%+v mids=%+v error=%+v", event, mids, err)
		return nil, err
	}
	return reply, nil
}

func (d *progressDao) PushProgress(c context.Context, param *model.ProgressParam, progress int64, mids []int64, dimension model.ProgressDimension) (int64, error) {
	displayNum := model.ProgressStatString(progress)
	if param.Type == model.TypeClickProgress {
		displayNum = strconv.FormatInt(progress, 10)
	}
	pushReq := &v1.NativePageEvent{
		PageID: param.PageID,
		Items: []*v1.EventItem{
			{
				ItemID:     param.ID,
				Type:       param.Type,
				Num:        progress,
				DisplayNum: displayNum,
				WebKey:     param.WebKey,
				Dimension:  param.Dimension,
			},
		},
	}
	switch {
	case dimension.IsUser():
		if len(mids) == 0 {
			return 0, errors.New("mids is empty")
		}
		pushRly, err := d.PushMids(c, pushReq, mids)
		if err != nil {
			return 0, err
		}
		return pushRly.GetMsgId(), nil
	case dimension.IsTotal():
		pushRly, err := d.PushRoom(c, pushReq)
		if err != nil {
			return 0, err
		}
		return pushRly.GetMsgId(), nil
	default:
		log.Errorc(c, "Unexpected dimension=%+v", dimension)
		return 0, errors.Errorf("Unexpected dimension=%+v", dimension)
	}
}

func (d *progressDao) GetProgressParams(c context.Context) ([]*model.ProgressParam, error) {
	rows, err := d.db.Query(c, progressSQL)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*model.ProgressParam{}, nil
		}
		log.Errorc(c, "Fail to query progressSQL, sql=%s error=%+v", progressSQL, err)
		return nil, err
	}
	defer rows.Close()
	list := make([]*model.ProgressParam, 0, 100)
	for rows.Next() {
		m := &model.ProgressParam{Type: model.TypeProgress}
		if err = rows.Scan(&m.ID, &m.PageID, &m.Sid, &m.GroupID, &m.WebKey); err != nil {
			log.Errorc(c, "Fail to scan progressParam row, error=%+v", err)
			continue
		}
		list = append(list, m)
	}
	err = rows.Err()
	if err != nil {
		log.Errorc(c, "Fail to get progressParam rows, error=%+v", err)
		return nil, err
	}
	return list, nil
}

func (d *progressDao) GetProgressParamsFromClick(c context.Context) ([]*model.ProgressParam, error) {
	params, err := d.getClickProgressParams(c)
	if err != nil {
		return nil, err
	}
	moduleIDs := make([]int64, 0, len(params))
	for _, v := range params {
		moduleIDs = append(moduleIDs, v.ModuleID)
	}
	modules, err := d.getModulesByID(c, moduleIDs)
	if err != nil {
		return nil, err
	}
	list := make([]*model.ProgressParam, 0, len(params))
	for _, v := range params {
		if module, ok := modules[v.ModuleID]; !ok || module == nil {
			continue
		}
		progressParam := &model.ProgressParam{
			ID:      v.ID,
			PageID:  modules[v.ModuleID].NativeID,
			Sid:     v.Sid,
			GroupID: v.GroupID,
			WebKey:  v.WebKey,
			Type:    model.TypeClickProgress,
		}
		list = append(list, progressParam)
	}
	return list, nil
}

func (d *progressDao) getAllProgressParams(c context.Context) ([]*model.ProgressParam, error) {
	var (
		progressParams []*model.ProgressParam
		clickParams    []*model.ProgressParam
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		var err error
		if progressParams, err = d.GetProgressParams(c); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var err error
		if clickParams, err = d.GetProgressParamsFromClick(c); err != nil {
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "Fail to get all progressParams, error=%+v", err)
		return nil, err
	}
	return d.filterProgressParams(c, append(progressParams, clickParams...)), nil
}

func (d *progressDao) filterProgressParams(ctx context.Context, params []*model.ProgressParam) []*model.ProgressParam {
	pageIds := make([]int64, 0, len(params))
	for _, param := range params {
		if param == nil || param.PageID <= 0 {
			continue
		}
		pageIds = append(pageIds, param.PageID)
	}
	if len(pageIds) == 0 {
		return nil
	}
	pages, err := d.pagesByIds(ctx, pageIds)
	if err != nil {
		// 数据请求失败跳过过滤
		return params
	}
	var validParams []*model.ProgressParam
	for _, param := range params {
		if param == nil || param.PageID <= 0 {
			continue
		}
		if page, ok := pages[param.PageID]; ok && page.State == api.OnlineState {
			validParams = append(validParams, param)
		}
	}
	return validParams
}

func (d *progressDao) getChildPage(c context.Context) (map[int64][]*model.ChildPage, error) {
	rows, err := d.db.Query(c, childPageSQL)
	if err != nil {
		log.Errorc(c, "Fail to query childPageSQL, sql=%s error=%+v", childPageSQL, err)
		return nil, err
	}
	defer rows.Close()
	list := make(map[int64][]*model.ChildPage)
	for rows.Next() {
		m := &model.ChildPage{}
		err = rows.Scan(&m.ModuleID, &m.ChildPageID)
		if err != nil {
			log.Errorc(c, "Fail to scan childPage row, error=%+v", err)
			continue
		}
		list[m.ModuleID] = append(list[m.ModuleID], m)
	}
	err = rows.Err()
	if err != nil {
		log.Errorc(c, "Fail to get childPage rows, error=%+v", err)
		return nil, err
	}
	return list, nil
}

func (d *progressDao) getParentPage(c context.Context) (map[int64]*model.ParentPage, error) {
	rows, err := d.db.Query(c, parentPageSQL)
	if err != nil {
		log.Errorc(c, "Fail to query parentPageSQL, sql=%s error=%+v", parentPageSQL, err)
		return nil, err
	}
	defer rows.Close()
	list := make(map[int64]*model.ParentPage)
	for rows.Next() {
		m := &model.ParentPage{}
		err = rows.Scan(&m.ModuleID, &m.ParentPageID)
		if err != nil {
			log.Errorc(c, "Fail to scan parentPage row, error=%+v", err)
			continue
		}
		list[m.ModuleID] = m
	}
	err = rows.Err()
	if err != nil {
		log.Errorc(c, "Fail to get parentPage rows, error=%+v", err)
		return nil, err
	}
	return list, nil
}

func (d *progressDao) getClickProgressParams(c context.Context) ([]*model.ClickProgressParam, error) {
	rows, err := d.db.Query(c, clickProgressSQL)
	if err != nil {
		log.Errorc(c, "Fail to query clickProgressSQL, sql=%s error=%+v", clickProgressSQL, err)
		return nil, err
	}
	defer rows.Close()
	list := make([]*model.ClickProgressParam, 0, 100)
	for rows.Next() {
		tip := ""
		m := &model.ClickProgressParam{}
		if err = rows.Scan(&m.ID, &m.ModuleID, &m.Sid, &tip, &m.WebKey); err != nil {
			log.Errorc(c, "Fail to scan ClickProgressParam row, error=%+v", err)
			continue
		}
		clickTip := &api.ClickTip{}
		if err = json.Unmarshal([]byte(tip), clickTip); err != nil {
			log.Errorc(c, "Fail to unmarshal clickTip, clickTip=%+v error=%+v", tip, err)
			continue
		}
		m.GroupID = clickTip.GroupId
		list = append(list, m)
	}
	err = rows.Err()
	if err != nil {
		log.Errorc(c, "Fail to get ClickProgressParam rows, error=%+v", err)
		return nil, err
	}
	return list, nil
}

func (d *progressDao) getModulesByID(c context.Context, ids []int64) (map[int64]*natmdl.NatModule, error) {
	if len(ids) == 0 {
		return map[int64]*natmdl.NatModule{}, nil
	}
	querySql := fmt.Sprintf(moduleIDSQL, xstr.JoinInts(ids))
	rows, err := d.db.Query(c, querySql)
	if err != nil {
		log.Errorc(c, "Fail to query modIDSQL, sql=%s error=%+v", querySql, err)
		return nil, err
	}
	defer rows.Close()
	list := make(map[int64]*natmdl.NatModule, len(ids))
	for rows.Next() {
		m := &natmdl.NatModule{}
		err = rows.Scan(&m.ID, &m.NativeID)
		if err != nil {
			log.Errorc(c, "Fail to scan NatModule row, error=%+v", err)
			continue
		}
		list[m.ID] = m
	}
	err = rows.Err()
	if err != nil {
		log.Errorc(c, "Fail to get NatModule rows, error=%+v", err)
		return nil, err
	}
	return list, nil
}

func (d *progressDao) pagesByIds(ctx context.Context, ids []int64) (map[int64]*api.NativePage, error) {
	rows, err := d.db.Query(ctx, fmt.Sprintf(pagesByIdsSQL, collection.JoinSliceInt(ids, ",")))
	if err != nil {
		log.Errorc(ctx, "Fail to query PagesByIds, ids=%+v error=%+v", ids, err)
		return nil, err
	}
	defer rows.Close()
	items := make(map[int64]*api.NativePage, len(ids))
	for rows.Next() {
		t := &api.NativePage{}
		if err = rows.Scan(&t.ID, &t.State); err != nil {
			log.Errorc(ctx, "Fail to scan nativePage row, error=%+v", err)
			continue
		}
		items[t.ID] = t
	}
	if err = rows.Err(); err != nil {
		log.Errorc(ctx, "Fail to get nativePages rows, error=%+v", err)
		return nil, err
	}
	return items, nil
}

func (d *progressDao) doPoint(c context.Context, item interface{}, extra *single.Extra) message.Policy {
	pointMsg, ok := item.(*model.PointMsg)
	if !ok || pointMsg.Mid == 0 || pointMsg.Counter == "" {
		return message.Ignore
	}
	sid, groupID, ok := extractCounter(pointMsg.Counter)
	if !ok || sid == 0 || groupID == 0 {
		return message.Ignore
	}
	params, err := d.parsePointToProgressParams(sid, groupID)
	log.Info("parsePointToProgressParams, msg=%+v paramsCount=%+v extra(%+v)", item, len(params), extra)
	if err != nil {
		return message.Ignore
	}
	for _, v := range params {
		dimension := model.ProgressDimension(v.Dimension)
		if !dimension.IsUser() {
			continue
		}
		_, _ = d.PushProgress(c, v, pointMsg.Total, []int64{pointMsg.Mid}, dimension)
	}
	return message.Success
}

func (d *progressDao) parsePointToProgressParams(sid, groupID int64) ([]*model.ProgressParam, error) {
	groups, ok := d.sid2GroupParams[sid]
	if !ok {
		return nil, errors.Errorf("key=%+v not found in sid-to-groupIDs cache", sid)
	}
	params, ok := groups[groupID]
	if !ok {
		return nil, errors.Errorf("key=%+v not found in groupID-to-progressParams cache", sid)
	}
	return params, nil
}

func (d *progressDao) loadProgressParamsExtra(c context.Context) {
	progressParams, err := d.getAllProgressParams(c)
	log.Info("Start to load progressParamsExtra, len=%+v error=%+v", len(progressParams), err)
	if err != nil {
		log.Error("Fail to get all progressParams, error=%+v", err)
		return
	}
	// for cache
	d.setPageProgressParamsCache(c, progressParams)
	// for user push
	d.setParentIDToProgressParams(progressParams)
	progRlys, err := d.batchActivityProgress(c, progressParams, false)
	if err != nil {
		return
	}
	sid2GroupParams := make(map[int64]map[int64][]*model.ProgressParam, len(progRlys))
	paramLen := 0
	for _, v := range progressParams {
		rly, ok := progRlys[v.Sid]
		if !ok || len(rly.Groups) == 0 {
			continue
		}
		group, ok := rly.Groups[v.GroupID]
		if !ok {
			continue
		}
		if _, ok := sid2GroupParams[v.Sid]; !ok {
			sid2GroupParams[v.Sid] = make(map[int64][]*model.ProgressParam)
		}
		// 推送维度
		v.Dimension = group.Info.Dim1
		sid2GroupParams[v.Sid][v.GroupID] = append(sid2GroupParams[v.Sid][v.GroupID], v)
		paramLen++
	}
	d.sid2GroupParams = sid2GroupParams
	log.Info("loadProgressParamsExtra, len=%+v", paramLen)
}

func (d *progressDao) batchActivityProgress(c context.Context, params []*model.ProgressParam, strictMod bool) (map[int64]*actGRPC.ActivityProgressReply, error) {
	reqs := make(map[int64][]int64, len(params))
	for _, param := range params {
		if param == nil || param.Sid == 0 || param.GroupID == 0 {
			continue
		}
		reqs[param.Sid] = append(reqs[param.Sid], param.GroupID)
	}
	if len(reqs) == 0 {
		return map[int64]*actGRPC.ActivityProgressReply{}, nil
	}
	eg := errgroup.WithContext(c)
	rlys := make(map[int64]*actGRPC.ActivityProgressReply, len(reqs))
	lock := sync.Mutex{}
	for k, v := range reqs {
		sid := k
		gids := uniqueArray(v)
		eg.Go(func(ctx context.Context) error {
			rly, err := d.actDao.ActivityProgress(ctx, sid, 2, 0, gids)
			if err != nil && strictMod {
				return err
			}
			if rly == nil {
				return nil
			}
			lock.Lock()
			defer lock.Unlock()
			rlys[sid] = rly
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("Fail to batch request ActivityProgress, error=%+v", err)
		return nil, err
	}
	return rlys, nil
}

func (d *progressDao) setPageProgressParamsCache(c context.Context, params []*model.ProgressParam) {
	data := make(map[int64][]*model.ProgressParam)
	for _, v := range params {
		pageID := v.PageID
		if parentID, ok := d.GetParentPageID(v.PageID); ok {
			pageID = parentID
		}
		data[pageID] = append(data[pageID], v)
	}
	for pageID, items := range data {
		val := make([]*model.ProgressParam, len(items))
		copy(val, items)
		pid := pageID
		err := d.cache.Do(c, func(ctx context.Context) {
			if err := d.AddCachePageProgressParams(ctx, pid, val); err != nil {
				log.Errorc(ctx, "Fail to add pageProgressParams cache, pageID=%+v error=%+v", pid, err)
			}
		})
		if err != nil {
			log.Errorc(c, "Fanout fail to process, pageID=%+v error=%+v", pid, err)
		}
	}
}

func (d *progressDao) loadPageRelations(c context.Context) {
	childPages, parentPages, err := d.getRelatedPages(c)
	log.Info("Start to load pageRelations, childLen=%+v parentLen=%+v error=%+v", len(childPages), len(parentPages), err)
	if err != nil {
		log.Error("Fail to get relatedPages, error=%+v", err)
		return
	}
	// child => parent1
	relations := make(map[int64]int64, len(childPages))
	for moduleID, v := range childPages {
		if page, ok := parentPages[moduleID]; !ok || page == nil {
			continue
		}
		for _, page := range v {
			relations[page.ChildPageID] = parentPages[moduleID].ParentPageID
		}
	}
	// tab组件可以配二级tab组件
	// parent1 => parent2
	for childID, parentID := range relations {
		if _, ok := relations[parentID]; !ok {
			continue
		}
		relations[childID] = relations[parentID]
	}
	d.pageRelations = relations
	log.Info("load pageRelations success, len=%+v", len(d.pageRelations))
}

func (d *progressDao) getRelatedPages(c context.Context) (map[int64][]*model.ChildPage, map[int64]*model.ParentPage, error) {
	eg := errgroup.WithContext(c)
	var (
		childPages  map[int64][]*model.ChildPage
		parentPages map[int64]*model.ParentPage
	)
	eg.Go(func(ctx context.Context) error {
		var err error
		if childPages, err = d.getChildPage(ctx); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var err error
		if parentPages, err = d.getParentPage(c); err != nil {
			return err
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Errorc(c, "Fail to get child and parent pages, error=%+v", err)
		return nil, nil, err
	}
	if len(childPages) == 0 {
		log.Info("No childPages")
		return map[int64][]*model.ChildPage{}, map[int64]*model.ParentPage{}, nil
	}
	if len(parentPages) == 0 {
		return nil, nil, errors.Errorf("ParentPages is empty")
	}
	return childPages, parentPages, nil
}

func (d *progressDao) GetParentPageID(pageID int64) (int64, bool) {
	parentID, ok := d.pageRelations[pageID]
	return parentID, ok
}

func (d *progressDao) setParentIDToProgressParams(params []*model.ProgressParam) {
	for _, v := range params {
		parentID, ok := d.GetParentPageID(v.PageID)
		if !ok {
			continue
		}
		v.PageID = parentID
	}
}

func unpackPoint(msg message.Message) (*single.UnpackMessage, error) {
	pointMsg := &model.PointMsg{}
	if err := json.Unmarshal(msg.Payload(), pointMsg); err != nil {
		return nil, err
	}
	return &single.UnpackMessage{
		Group: pointMsg.Mid,
		Item:  pointMsg,
	}, nil
}

// nolint:gomnd
func extractCounter(counter string) (sid, groupID int64, ok bool) {
	// $sid $groupID $dim01_$dim02_${counter_key}
	parts := strings.Split(counter, _counterSep)
	if len(parts) < 3 {
		return 0, 0, false
	}
	sid, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		log.Error("Fail to parse sid, counter=%+v sid=%+v error=%+v", counter, parts[0], err)
		return 0, 0, false
	}
	groupID, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		log.Error("Fail to parse groupID, counter=%+v groupID=%+v error=%+v", counter, parts[1], err)
		return 0, 0, false
	}
	return sid, groupID, true
}

func uniqueArray(arr []int64) []int64 {
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
