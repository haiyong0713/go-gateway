package service

import (
	"context"
	"sort"
	"strings"

	"go-common/library/log"
	v1 "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_iconFixType = "fix"
)

// IndexIcon get index icons
func (s *Service) IndexIcon() (res *model.IndexIcon) {
	return s.indexIcon
}

func (s *Service) loadIndexIcon() {
	if s.indexIconRunning {
		return
	}
	s.indexIconRunning = true
	defer func() {
		s.indexIconRunning = false
	}()
	data, err := s.resgrpc.IndexIconNew(context.Background(), &v1.NoArgRequest{})
	if err != nil || data == nil {
		log.Error("s.res.IndexIcon error(%v)", err)
		return
	}
	iconReply, ok := data.GetIndexIcon()[_iconFixType]
	if !ok || len(iconReply.GetIndexIconItem()) == 0 {
		log.Error("s.res.IndexIcon data icons not found")
		return
	}
	icons := iconReply.GetIndexIconItem()
	sort.Slice(icons, func(i, j int) bool { return icons[i].Weight > icons[j].Weight })
	tmp := make([]*model.IndexIcon, 0, len(icons))
	for _, v := range icons {
		v.Icon = strings.Replace(v.Icon, "http://", "//", 1)
		tmp = append(tmp, &model.IndexIcon{
			ID:     v.Id,
			Title:  v.Title,
			Links:  v.Links,
			Icon:   v.Icon,
			Weight: v.Weight,
		})
	}
	s.indexIcons = tmp
}

func (s *Service) randomIndexIcon() {
	if s.randomIconRunning {
		return
	}
	s.randomIconRunning = true
	defer func() {
		s.randomIconRunning = false
	}()
	var (
		total, weight int
	)
	tempIcons := make([]*model.IndexIcon, len(s.indexIcons))
	copy(tempIcons, s.indexIcons)
	length := len(tempIcons)
	if length == 0 {
		s.indexIcon = new(model.IndexIcon)
		return
	}
	for _, v := range tempIcons {
		if v.Weight == 0 {
			total++
		} else {
			total += int(v.Weight)
		}
	}
	if total == length {
		item := tempIcons[s.r.Intn(length)]
		s.indexIcon = &model.IndexIcon{ID: item.ID, Title: item.Title, Links: item.Links, Icon: item.Icon, Weight: item.Weight}
		return
	}
	randWeight := s.r.Intn(total)
	for _, v := range tempIcons {
		if v.Weight == 0 {
			weight++
		} else {
			weight += int(v.Weight)
		}
		if weight > randWeight {
			item := v
			s.indexIcon = &model.IndexIcon{ID: item.ID, Title: item.Title, Links: item.Links, Icon: item.Icon, Weight: item.Weight}
			return
		}
	}
}
