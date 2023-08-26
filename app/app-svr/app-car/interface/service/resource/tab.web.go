package resource

import (
	"context"

	"go-gateway/app/app-svr/app-car/interface/model/show"
	"go-gateway/app/app-svr/app-car/interface/model/tab"
)

func (s *Service) ShowTabWeb(c context.Context, param *show.ShowParam) []*tab.TabWeb {
	items := []*tab.TabWeb{}
	tabConfig := s.c.Custom.TabConfigsWeb
	for i, t := range tabConfig {
		item := &tab.TabWeb{
			ID:           t.ID,
			Name:         t.Name,
			TabID:        t.TabID,
			Goto:         t.Goto,
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
