package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/robfig/cron"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
)

func (s *Service) FlushCache() {
	var err error
	c := cron.New()
	// 每10秒刷新一下缓存
	err = c.AddFunc("*/10 * * * *", func() {
		if err = s.ugctab.UpdateCache(context.Background()); err != nil {
			log.Error("ugctab.service FlushCache.UpadateCache error: %s", err)
		}
	})
	if err != nil {
		log.Error("ugctab.service FlushCache.UpadateCache error: %s", err)
	}
	c.Start()
}

// UgcTabV2-404优化
func (s *Service) UgcTabV2(c context.Context, req *pb.UgcTabReq) (*pb.UgcTabV2Reply, error) {
	rly, err := s.UgcTab(c, req)
	if err != nil {
		if err == ecode.NothingFound {
			return &pb.UgcTabV2Reply{}, nil
		}
		return nil, err
	}
	return &pb.UgcTabV2Reply{Item: rly}, nil
}

// nolint: gocognit
func (s *Service) UgcTab(c context.Context, req *pb.UgcTabReq) (reply *pb.UgcTabReply, err error) {
	var (
		//nowTime  int64
		ugctab       []*model.UgcTabItem
		tagMap       map[string]bool
		upMap        map[string]bool
		arcMap       map[string]bool
		isTagMatch   = false
		isUpMatch    = false
		isArcMatch   = false
		isAvMatch    = false
		isBuildMatch = false
		ugcType      bool // 是否是全量投放，true为全量，false为投放部分
	)
	// 判断时间
	//nowTime = time.Now().Unix()
	if ugctab, err = s.ugctab.GetEffectiveUgcTab(c); err != nil {
		log.Error("get ugctab  error: %s", err)
		return nil, err
	}
	if len(ugctab) == 0 {
		log.Error("not found effective ugctab: %s", err)
		return nil, ecode.NothingFound
	}

	// 判断平台版本
	var builds []model.BuildLimit
	// 判断条件
	state := &model.UgcTabItem{}
	for _, ugc := range ugctab {
		if err = json.Unmarshal([]byte(ugc.Builds), &builds); err != nil {
			log.Error("get ugctab list error: %s", err)
			return nil, err
		}
		for _, p := range builds {
			if p.Plat != req.Plat {
				continue
			}
			switch p.Conditions {
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
			// Bug
			if isBuildMatch {
				break
			}
		}
		// 版本未匹配
		if !isBuildMatch {
			continue
		}

		// 判断tag，稿件，avid
		tagMap = make(map[string]bool)
		upMap = make(map[string]bool)
		arcMap = make(map[string]bool)

		tagString := strings.Split(ugc.Tagid, ",")
		upString := strings.Split(ugc.Upid, ",")
		arcString := strings.Split(ugc.Arctype, ",")

		for _, v := range tagString {
			tagMap[v] = true
		}
		for _, v := range upString {
			upMap[v] = true
		}
		for _, v := range arcString {
			arcMap[v] = true
		}
		if _, isArcMatch = arcMap[strconv.FormatInt(req.Tid, 10)]; isArcMatch {
			state = ugc
			break
		}
		for _, v := range req.Tag {
			// BUG:
			if _, isTagMatch = tagMap[strconv.FormatInt(v, 10)]; isTagMatch {
				state = ugc
				break
			}
		}
		if isTagMatch {
			break
		}
		for _, v := range req.UpId {
			// BUG:
			if _, isUpMatch = upMap[strconv.FormatInt(v, 10)]; isUpMatch {
				state = ugc
				break
			}
		}
		if isUpMatch {
			break
		}

		avidStr := strconv.FormatInt(req.AvId, 10)
		if ugc.AvidMap != nil {
			if _, isAvMatch = ugc.AvidMap[avidStr]; isAvMatch {
				state = ugc
				break
			}
		}

		// 判断是否是全量投放
		if ugc.UgcType == 1 {
			state = ugc
			ugcType = true
		}
	}
	if !isBuildMatch && !ugcType {
		return nil, ecode.NothingFound
	}
	if !ugcType && !isTagMatch && !isUpMatch && !isAvMatch && !isArcMatch {
		return nil, ecode.NothingFound
	}

	reply = &pb.UgcTabReply{
		Id: state.ID,
		// Tab样式,1-文字，2-图片
		TabType: state.TabType,
		// Tab内容，当type为1时tab为文字，为2时为图片地址
		Tab: state.Tab,
		// LinkTab,1-H5链接，2-Native ID
		LinkType: state.LinkType,
		// Link内容,当link_type为1时为H5链接，为2时为Native ID
		Link: state.Link,
		// 背景
		Bg: state.Bg,
		// tab字体颜色选中状态
		Selected: state.Selected,
		// tab字体颜色
		Color: state.Color,
	}

	return reply, err
}

func (s *Service) UgcTabBatch(c context.Context, req *pb.UgcTabBatchReq) (reply *pb.UgcTabBatchReply, err error) {
	eg, _ := errgroup.WithContext(c)
	lock := sync.Mutex{}
	reply = new(pb.UgcTabBatchReply)
	reply.Tabs = make(map[int64]*pb.UgcTabReply)

	for avid, info := range req.Arcs {
		singleReq := &pb.UgcTabReq{
			Tid:   info.Tid,
			Tag:   info.Tag,
			UpId:  info.UpId,
			AvId:  avid,
			Plat:  req.Plat,
			Build: req.Build,
		}
		eg.Go(func() error {
			var (
				tab *pb.UgcTabReply
			)
			tab, _ = s.UgcTab(c, singleReq)
			lock.Lock()
			reply.Tabs[singleReq.AvId] = tab
			lock.Unlock()
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("UgcTabBatch error: %s", err)
	}
	return reply, err
}
