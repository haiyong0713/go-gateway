package act

import (
	"context"
	"sync"

	actmdl "go-gateway/app/app-svr/app-show/interface/model/act"
	natapi "go-gateway/app/web-svr/native-page/interface/api"

	actapi "git.bilibili.co/bapis/bapis-go/activity/service"
	api "git.bilibili.co/bapis/bapis-go/bilibili/app/show/gateway/v1"
	v1 "git.bilibili.co/bapis/bapis-go/bilibili/broadcast/message/main"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
)

func (s *Service) GetActProgress(c context.Context, req *api.GetActProgressReq) (*api.GetActProgressReply, error) {
	progressParams, err := s.actDao.NatProgressParams(c, req.PageID)
	if err != nil {
		log.Error("Fail to get natProgressParams from dao, pageID=%+v error=%+v", req.PageID, err)
		return nil, err
	}
	progRlys, err := s.getProgresses(c, progressParams, req.Mid)
	if err != nil {
		return nil, err
	}
	event := &v1.NativePageEvent{
		PageID: req.PageID,
		Items:  []*v1.EventItem{},
	}
	for _, param := range progressParams {
		progRly, ok := progRlys[param.Sid]
		if !ok || len(progRly.Groups) == 0 {
			continue
		}
		group, ok := progRly.Groups[param.GroupID]
		if !ok {
			continue
		}
		item := progressParam2EventItem(param, group.Total)
		if item == nil {
			continue
		}
		event.Items = append(event.Items, item)
	}
	return &api.GetActProgressReply{Event: event}, nil
}

func (s *Service) getProgresses(c context.Context, progressParams []*natapi.ProgressParam, mid int64) (map[int64]*actapi.ActivityProgressReply, error) {
	progReqs := make(map[int64][]int64, len(progressParams))
	for _, v := range progressParams {
		progReqs[v.Sid] = append(progReqs[v.Sid], v.GroupID)
	}
	eg := errgroup.WithContext(c)
	progRlys := make(map[int64]*actapi.ActivityProgressReply, len(progReqs))
	lock := sync.Mutex{}
	for k, v := range progReqs {
		sid := k
		gids := v
		eg.Go(func(ctx context.Context) error {
			rly, err := s.actDao.ActivityProgress(ctx, sid, 2, mid, gids)
			if err != nil {
				return nil
			}
			lock.Lock()
			progRlys[sid] = rly
			lock.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("Fail to batch request ActivityProgress, error=%+v", err)
		return nil, err
	}
	return progRlys, nil
}

func progressParam2EventItem(param *natapi.ProgressParam, progress int64) *v1.EventItem {
	if param == nil {
		return nil
	}
	return &v1.EventItem{
		ItemID:     param.Id,
		Type:       param.Type,
		Num:        progress,
		DisplayNum: actmdl.ProgressStatString(progress),
		WebKey:     param.WebKey,
		Dimension:  param.Dimension,
	}
}
