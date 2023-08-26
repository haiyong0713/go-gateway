package service

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/web/job/internal/model"
)

func (s *Service) setRegionList() {
	log.Info("分区列表数据更新开始")
	var (
		list   []*model.Region
		config map[int64][]*model.RegionConfig
	)
	ctx := context.Background()
	g := errgroup.WithCancel(ctx)
	g.Go(func(ctx context.Context) error {
		var err error
		list, err = s.dao.RegionList(ctx)
		return err
	})
	g.Go(func(ctx context.Context) error {
		var err error
		config, err = s.dao.RegionConfig(ctx)
		return err
	})
	if err := g.Wait(); err != nil {
		log.Error("日志告警 分区列表获取错误,error:%+v", err)
		return
	}
	log.Info("分区列表db数据获取成功")
	children := map[string][]*model.Region{}
	for _, val := range list {
		if val.Reid == 0 {
			continue
		}
		val.Config = config[val.ID]
		key := fmt.Sprintf("%v_%v_%v", val.Plat, val.Reid, val.Language)
		children[key] = append(children[key], val)
	}
	data := map[string][]*model.Region{}
	for _, val := range list {
		if val.Reid != 0 {
			continue
		}
		val.Config = config[val.ID]
		val.Children = children[fmt.Sprintf("%v_%v_%v", val.Plat, val.Rid, val.Language)]
		key := fmt.Sprintf("%v_%v", val.Plat, val.Language)
		data[key] = append(data[key], val)
	}
	if err := retry(func() (err error) {
		return s.dao.AddCacheRegionList(ctx, data)
	}); err != nil {
		log.Error("日志告警 分区列表数据缓存错误,error:%v", err)
		return
	}
	log.Info("分区列表数据更新成功")
}
