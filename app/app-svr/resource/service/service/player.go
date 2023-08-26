package service

import (
	"context"
	"encoding/json"

	"github.com/robfig/cron"
	"go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
)

type effectivePanel struct {
	//MinuteStamp int64
	Config *pb.PlayerPanel
	TidMap map[int64]bool
}

type tag struct {
	ID   int64  `json:"id" `
	Name string `json:"name"`
}

func (s *Service) startFetchPlayerData() {
	// 循环读取当前数据
	var (
		dataErr error
		err     error
	)
	c := cron.New()

	if s.panelsInCache, dataErr = s.refreshPanelData(context.Background()); dataErr != nil {
		panic(dataErr)
	}
	err = c.AddFunc("*/5 * * * *", func() {
		if s.panelsInCache, dataErr = s.refreshPanelData(context.Background()); dataErr != nil {
			log.Error("refresh panel in cache fail!")
		}
		log.Info("refresh panel in cache success!")
	})
	if err != nil {
		panic(err)
	}
	c.Start()
}

// 刷新内存中最新的配置数据
func (s *Service) refreshPanelData(ctx context.Context) (current []*effectivePanel, err error) {
	// 获取当前有效的入口配置
	if panels, err := s.player.GetEffectivePanels(ctx); err != nil {
		log.Error("get current panel error 0: %s", err)
		if err.Error() != "-404" {
			return nil, err
		} else {
			return nil, nil
		}
	} else {
		if len(panels) == 0 {
			//log.Warn("get current panel nil")
			return nil, nil
		}
		current = make([]*effectivePanel, len(panels))

		for i, p := range panels {
			temp := &effectivePanel{
				TidMap: make(map[int64]bool),
				Config: &pb.PlayerPanel{
					Id: p.ID,
					// 按钮素材
					BtnImg: p.BtnImg,
					// 按钮文案
					BtnText: p.BtnText,
					// 字体颜色
					TextColor: p.TextColor,
					// 跳转链接
					Link: p.Link,
					// 面板文案
					Label: p.Label,
					// 展现阶段
					DisplayStage: p.DisplayStage,
					// 运营商
					Operator: p.Operator,
				},
			}

			if len(p.Tids) != 0 {
				var tids []*tag
				if err = json.Unmarshal([]byte(p.Tids), &tids); err != nil {
					log.Error("get current panel error 1: %s", err)
					return nil, err
				}
				for _, tid := range tids {
					temp.TidMap[tid.ID] = true
				}
			}

			current[i] = temp
		}

	}
	return current, nil
}

// 获取当前有效的入口配置-404优化
func (s *Service) GetPlayerCustomizedPanelV2(c context.Context, req *pb.GetPlayerCustomizedPanelReq) (*pb.GetPlayerCustomizedPanelV2Rep, error) {
	rly, err := s.GetPlayerCustomizedPanel(c, req)
	if err != nil {
		if err == ecode.NothingFound {
			return &pb.GetPlayerCustomizedPanelV2Rep{}, nil
		}
		return nil, err
	}
	return &pb.GetPlayerCustomizedPanelV2Rep{Item: rly}, nil
}

// 获取当前有效的入口配置
func (s *Service) GetPlayerCustomizedPanel(_ context.Context, req *pb.GetPlayerCustomizedPanelReq) (reply *pb.GetPlayerCustomizedPanelRep, err error) {
	// 用内存扛实时获取的内容
	if s.panelsInCache == nil || len(s.panelsInCache) == 0 {
		return nil, ecode.NothingFound
	}

	resultMap := make(map[string]bool)
	resultMap["before_play"] = false
	resultMap["after_free_play"] = false

	for _, panel := range s.panelsInCache {
		var target *pb.PlayerPanel

		if len(panel.TidMap) == 0 {
			target = panel.Config
		} else {
			for _, tid := range req.Tids {
				if isExisted, ok := panel.TidMap[tid]; ok && isExisted {
					target = panel.Config
					break
				}
			}
		}

		if target != nil {
			if reply == nil {
				reply = &pb.GetPlayerCustomizedPanelRep{}
			}

			if reply.Id == 0 && target.DisplayStage == "before_play" && target.Operator == "" {
				reply.Id = target.Id
				reply.BtnImg = target.BtnImg
				reply.BtnText = target.BtnText
				reply.TextColor = target.TextColor
				reply.Link = target.Link
			}

			if v, ok := resultMap[target.DisplayStage]; ok {
				//log.Warn("resultMap set true 1: %v", ok)
				if v {
					continue
				}
			} else {
				continue
			}
			reply.Panels = append(reply.Panels, target)
			resultMap[target.DisplayStage] = true
		}
	}
	if reply == nil {
		err = ecode.NothingFound
	}
	return reply, err
}
