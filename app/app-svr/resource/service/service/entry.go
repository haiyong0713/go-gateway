package service

import (
	"context"
	"encoding/json"

	"go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"

	"github.com/robfig/cron"
)

type effectiveEntry struct {
	//MinuteStamp int64
	State     *pb.GetAppEntryStateRep
	Platforms map[int32]*pb.EffectivePlatform
}

func (s *Service) startFetchEntryData() {
	// 循环读取当前数据
	var (
		dataErr error
		err     error
	)
	c := cron.New()

	if s.entriesInCache, dataErr = s.refreshEntryData(context.Background()); dataErr != nil {
		panic(dataErr)
	}
	err = c.AddFunc("*/5 * * * *", func() {
		if entries, dataErr := s.refreshEntryData(context.Background()); dataErr != nil {
			log.Error("refresh entry in cache fail!")
			return
		} else {
			s.entriesInCache = entries
			log.Info("refresh entry in cache success!")
		}
	})
	if err != nil {
		panic(err)
	}
	c.Start()
}

// 刷新内存中最新的配置数据
func (s *Service) refreshEntryData(ctx context.Context) (current []*effectiveEntry, err error) {
	// 获取当前有效的入口配置
	if entries, err := s.entry.GetEffectiveEntries(ctx); err != nil {
		if err.Error() != "-404" {
			log.Error("get current entry error 0: %s", err)
			return nil, err
		} else {
			return nil, nil
		}
	} else {
		for _, e := range entries {
			entry := e
			// 获取已推送的最新时间配置
			if lastST, err := s.entry.GetPushedTimeSettingsByEntryID(ctx, entry.ID); err != nil {
				if err.Error() != "-404" {
					log.Error("get current entry error 1: %s", err)
					return nil, err
				}
				continue
			} else {
				if lastST == nil {
					//log.Warn("get lastST nil")
					continue
				}
				// 根据时间配置中的state寻找配置内容
				var state *model.AppEntryState
				if state, err = s.entry.GetStateByID(ctx, lastST.StateID); err != nil {
					if err.Error() != "-404" {
						log.Error("get current entry error 2: %s, %v", err, lastST.StateID)
						return nil, err
					}
					continue
				}
				//返回数据为空则丢弃
				if state == nil {
					continue
				}

				// 获取版本信息
				var plats []*pb.EffectivePlatform
				if err = json.Unmarshal([]byte(entry.Platforms), &plats); err != nil {
					log.Error("get entry list error 3: %s", err)
					return nil, err
				}
				platforms := map[int32]*pb.EffectivePlatform{}
				for _, p := range plats {
					platforms[p.Platform] = p
				}

				// 设置有效配置
				current = append(current, &effectiveEntry{
					State: &pb.GetAppEntryStateRep{
						ID:          lastST.ID,
						EntryName:   entry.EntryName,
						StateID:     lastST.StateID,
						StateName:   state.StateName,
						StaticIcon:  state.StaticIcon,
						DynamicIcon: state.DynamicIcon,
						Url:         state.Url,
						LoopCount:   lastST.SentLoop,
						STime:       lastST.STime,
						ETime:       entry.ETime,
					},
					Platforms: platforms,
				})
			}
		}
		return current, nil
	}
}

// 获取当前有效的入口配置-404优化
func (s *Service) GetAppEntryStateV2(c context.Context, req *pb.GetAppEntryStateReq) (*pb.GetAppEntryStateV2Rep, error) {
	rly, err := s.GetAppEntryState(c, req)
	if err != nil {
		if err == ecode.NothingFound {
			return &pb.GetAppEntryStateV2Rep{}, nil
		}
		return nil, err
	}
	return &pb.GetAppEntryStateV2Rep{Item: rly}, nil
}

// 获取当前有效的入口配置
func (s *Service) GetAppEntryState(_ context.Context, req *pb.GetAppEntryStateReq) (reply *pb.GetAppEntryStateRep, err error) {
	if req.Build <= 0 || (req.Plat != 0 && req.Plat != 1 && req.Plat != 2 && req.Plat != 20 && req.Plat != 5 && req.Plat != 8) {
		return nil, ecode.RequestErr
	}
	// 用内存扛实时获取的内容
	if len(s.entriesInCache) == 0 {
		return nil, ecode.NothingFound
	}

	// 对版本信息进行匹配

	for _, entry := range s.entriesInCache {
		if p, ok := entry.Platforms[req.Plat]; !ok {
			continue
		} else {
			var isBuildMatch bool
			switch p.Condition {
			case "gt":
				{
					isBuildMatch = req.Build > p.Build
					break
				}
			case "lt":
				{
					isBuildMatch = req.Build < p.Build
					break
				}
			case "eq":
				{
					isBuildMatch = req.Build == p.Build
					break
				}
			case "ne":
				{
					isBuildMatch = req.Build != p.Build
					break
				}
			default:
				isBuildMatch = false
			}
			if isBuildMatch {
				// 如果已经有匹配的平台和版本，则返回结果
				reply = entry.State
				return reply, err
			}
		}
	}

	return nil, ecode.NothingFound
}
