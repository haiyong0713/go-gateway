package page

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
	model "go-gateway/app/web-svr/activity/interface/model/page"

	api "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
)

func (s *Service) UgcURL(c context.Context, ID int64) (*model.ResUgcURL, error) {
	res := new(model.ResUgcURL)
	if ID > 10000000 {
		ID = ID % 10000000
		// native页面
		reply, err := client.NaPageClient.NativePageCards(c, &api.NativePageCardsReq{
			Pids: []int64{ID},
		})
		if err != nil {
			log.Errorc(c, "UgcURL client.NaPageClient.NativePageCards(c, %v) err[%v]", ID, err)
			return nil, err
		}
		if p, ok := reply.List[ID]; ok {
			res.Name = p.Title
			res.PcURL = p.PcURL
			res.H5URL = p.SkipURL
		}
	} else {
		// h5单页应用
		p, err := s.dao.GetPageByID(c, ID)
		if err != nil {
			return nil, err
		}
		if p != nil {
			res.Name = p.Name
			res.PcURL = p.PcURL
			res.H5URL = p.H5URL
		}
	}
	if res.H5URL == "" && res.PcURL == "" {
		// 兜底逻辑
		res.PcURL = "https://www.bilibili.com/blackboard/activity-w-MXWECuR.html"
		res.H5URL = res.PcURL
	} else {
		// 容错逻辑
		if res.H5URL == "" {
			res.H5URL = res.PcURL
		}
		if res.PcURL == "" {
			res.PcURL = res.H5URL
		}
	}
	if res.Name == "" {
		res.Name = "查看原活动"
	}
	return res, nil
}
