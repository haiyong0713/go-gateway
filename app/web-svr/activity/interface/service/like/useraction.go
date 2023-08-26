package like

import (
	"context"
	"encoding/json"
	"fmt"
	xcode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/ecode"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/client"
	actmdl "go-gateway/app/web-svr/activity/interface/model/actplat"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"strconv"
	"sync"
	"time"

	api "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
)

const (
	_taskBusinessID = 1
)

func (s *Service) SyncUserState(c context.Context, r *pb.SyncUserStateReq) error {
	us, err := s.taskDao.RawTaskUserState(c, []int64{r.TaskID}, r.MID, _taskBusinessID, r.SID, time.Now().Unix())
	if err != nil {
		return err
	}
	if len(us) == 0 {
		return s.taskDao.TaskUserStateAdd(c, r.MID, _taskBusinessID, r.TaskID, r.SID, 0, r.Count, 0, 0, 0)
	}
	return s.taskDao.TaskUserStateUp(c, r.MID, r.TaskID, 0, r.Count, 0, 0, r.SID, 0)
}

func (s *Service) SyncUserScore(c context.Context, r *pb.SyncUserScoreReq) error {
	reserve, err := s.dao.RawReserveOnly(c, r.SID, r.MID)
	if err != nil {
		return err
	}
	if reserve == nil {
		return ecode.ActivityReserveFirst
	}
	s.dao.UpReserve(c, &like.ActReserve{
		Sid:   r.SID,
		Mid:   r.MID,
		Num:   reserve.Num,
		State: reserve.State,
		Score: r.Score,
	})
	return nil
}

type Rule struct {
	Name string `json:"name"`
}
type Dim02Content struct {
	Rule *Rule `json:"rule"`
}

// CounterInfo
type CounterInfo struct {
	Dim02Content *Dim02Content `json:"dim02_content"`
}

// SendPoints
func (s *Service) SendPoints(ctx context.Context, mid int64, sid int64, groupId int64) (err error) {
	var groupInfo map[int64]*like.ReserveCounterGroupItem
	groupInfo, err = s.dao.GetReserveCounterGroupInfoByGid(ctx, []int64{groupId})
	if err != nil {
		log.Errorc(ctx, "s.dao.GetReserveCounterGroupInfoByGid(c, %v) err[%v]", groupId, err)
	}
	if len(groupInfo) > 0 {
		for _, g := range groupInfo {
			if g.ID == groupId {
				if g.Sid != sid {
					err = ecode.ActivityTaskNotExist
					return
				}
				counterInfo := &CounterInfo{}
				if err = json.Unmarshal([]byte(g.CounterInfo), counterInfo); err != nil {
					err = ecode.ActivityTaskNotExist
					return
				}
				if counterInfo == nil || counterInfo.Dim02Content == nil || counterInfo.Dim02Content.Rule == nil {
					err = ecode.ActivityTaskNotExist
					return
				}
				t := time.Now().Unix()
				activityPoints := &actmdl.ActivityPoints{
					Timestamp: t,
					Mid:       mid,
					Source:    mid,
					Activity:  strconv.FormatInt(sid, 10),
					Business:  counterInfo.Dim02Content.Rule.Name,
				}
				err = s.actDao.Send(ctx, mid, activityPoints)
				if err == nil {
					return
				}
			}
		}
	}
	return ecode.ActivityTaskNotExist

}

func (s *Service) GetReserveCounterGroupIDBySid(ctx context.Context, sid int64) (res []int64, err error) {
	res, err = s.dao.GetReserveCounterGroupIDBySid(ctx, sid)
	if err != nil {
		log.Errorc(ctx, "[GetReserveCounterGroupIDBySid][GetReserveCounterGroupIDBySid][Error], err:%+v", err)
		return
	}
	return
}

func (s *Service) ActivityProgress(c context.Context, r *pb.ActivityProgressReq) (*pb.ActivityProgressReply, error) {
	if len(r.Gids) == 0 || r.Type == 1 {
		// 获取节点组id信息
		res, err := s.dao.GetReserveCounterGroupIDBySid(c, r.Sid)
		if err != nil {
			log.Errorc(c, "s.dao.GetReserveCounterGroupIDBySid(c, %d) err[%v]", r.Sid, err)
			return nil, err
		}
		r.Gids = res
	}

	if len(r.Gids) == 0 {
		return nil, xcode.Error(xcode.RequestErr, "gid列表为空")
	}

	if len(r.Gids) > 20 {
		return nil, xcode.Error(xcode.RequestErr, "gid数量过多")
	}

	var groupInfo map[int64]*like.ReserveCounterGroupItem
	var nodeInfo map[int64][]*like.ReserveCounterNodeItem
	eg := errgroup.WithContext(c)
	// 获取节点组详细信息
	eg.Go(func(ctx context.Context) error {
		var err error
		groupInfo, err = s.dao.GetReserveCounterGroupInfoByGid(ctx, r.Gids)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetReserveCounterGroupInfoByGid(c, %v) err[%v]", r.Gids, err)
		}
		return err
	})
	// 获取节点详细信息
	eg.Go(func(ctx context.Context) error {
		var err error
		nodeInfo, err = s.dao.GetReserveCounterNodeByGid(ctx, r.Gids)
		if err != nil {
			log.Errorc(ctx, "s.dao.GetReserveCounterNodeByGid(c, %v) err[%v]", r.Gids, err)
		}
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	// 请求进度信息
	eg = errgroup.WithContext(c)
	stat := map[int64]int64{}
	lock := sync.Mutex{}
	for _, tmp := range groupInfo {
		info := tmp
		eg.Go(func(ctx context.Context) error {
			rsp, err := client.ActPlatClient.PointsUnlockGetCounterStatistic(ctx, &api.GetPointsUnlockCounterStatisticReq{
				Activity:  fmt.Sprint(info.Sid),
				GroupName: fmt.Sprint(info.ID),
				Dim01:     int32(info.Dim1),
				Dim02:     int32(info.Dim2),
				Mid:       r.Mid,
				Time:      r.Time,
			})
			if err != nil {
				log.Errorc(ctx, "client.ActPlatClient.PointsUnlockGetCounterStatistic(c, %v) err[%v]", *info, err)
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			stat[info.ID] = rsp.Total
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	// 拼装返回
	res := new(pb.ActivityProgressReply)
	res.Sid = r.Sid
	res.Groups = make(map[int64]*pb.ActivityProgressGroup)
	for gid, g := range groupInfo {
		gInfo := &pb.ActivityProgressGroup{
			Total: stat[gid],
			Nodes: make([]*pb.ActivityProgressNodeInfo, 0, len(nodeInfo[gid])),
			Info: &pb.ActivityProgressGroupInfo{
				Gid:       gid,
				GroupName: g.GroupName,
				Dim1:      g.Dim1,
				Dim2:      g.Dim2,
				Threshold: g.Threshold,
				CountInfo: g.CounterInfo,
			},
		}
		for _, n := range nodeInfo[gid] {
			gInfo.Nodes = append(gInfo.Nodes, &pb.ActivityProgressNodeInfo{
				Nid:  n.ID,
				Desc: n.NodeName,
				Val:  n.NodeVal,
			})
		}
		res.Groups[gid] = gInfo
	}
	return res, nil
}
