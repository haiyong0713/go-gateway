package resource

import (
	"context"

	"go-gateway/app/app-svr/app-car/interface/conf"
	"go-gateway/app/app-svr/app-car/interface/model/banner"
	"go-gateway/app/app-svr/app-car/interface/model/show"
	"go-gateway/app/app-svr/app-car/interface/model/tab"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

const (
	_default = "default"
)

type Service struct {
	c *conf.Config
}

func New(c *conf.Config) *Service {
	s := &Service{
		c: c,
	}
	return s
}

func (s *Service) ShowTab(c context.Context, param *show.ShowParam) []*tab.Tab {
	items := []*tab.Tab{}
	tabConfig := s.c.Custom.TabConfigs
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.TabConfig2, &feature.OriginResutl{
		MobiApp:    param.MobiApp,
		Device:     param.Device,
		Build:      int64(param.Build),
		BuildLimit: param.Build >= 1010000,
	}) {
		tabConfig = s.c.Custom.TabConfigs2
	}
	for i, t := range tabConfig {
		item := &tab.Tab{
			ID:           t.ID,
			Name:         t.Name,
			TabID:        t.TabID,
			URI:          t.URI,
			Pos:          i + 1,
			IsDefault:    t.IsDefault,
			Icon:         t.Icon,
			IconSelected: t.IconSelected,
		}
		// 需要屏蔽的渠道
		if _, ok := t.HideChannel[param.Channel]; ok {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (s *Service) Banner(c context.Context, param *show.ShowParam) []*banner.Banner {
	items := []*banner.Banner{}
	if len(s.c.Custom.Banners) == 0 {
		return items
	}
	banners, ok := s.c.Custom.Banners[param.Channel]
	if !ok {
		banners, ok = s.c.Custom.Banners[_default]
		if !ok {
			return items
		}
	}
	for _, v := range banners {
		item := &banner.Banner{
			ID:    v.ID,
			Image: v.Image,
			URI:   v.URL,
		}
		items = append(items, item)
	}
	return items
}
