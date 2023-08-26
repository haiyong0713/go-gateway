package service

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/web-svr/esports/admin/api"
	"go-gateway/app/web-svr/esports/admin/model"
)

func (s *Service) CreatePoster(c *bm.Context, req *pb.CreatePosterReq) (err error) {
	poster := &model.Poster{
		BgImage:       req.BgImage,
		ContestID:     req.ContestID,
		PositionOrder: req.Order,
		IsCenteral:    0,
		OnlineStatus:  0,
		CreatedBy:     req.CreatedBy,
	}

	if err = s.dao.CreatePoster(c, poster); err != nil {
		log.Error("create poster error 0: %s", err.Error())
		return err
	}

	return err
}

func (s *Service) EditPoster(c *bm.Context, req *pb.EditPosterReq) (err error) {
	if poster, err := s.dao.FindPosterById(c, req.Id); err != nil {
		log.Error("edit poster error 0: %s", err.Error())
		return err
	} else if poster == nil {
		err = ecode.NothingFound
		log.Error("edit poster error 1: %s", err.Error())
		return err
	} else {
		poster := &model.Poster{
			ID:            req.Id,
			BgImage:       req.BgImage,
			ContestID:     req.ContestID,
			CreatedBy:     req.CreatedBy,
			PositionOrder: req.Order,
		}

		if err = s.dao.EditPoster(c, poster); err != nil {
			log.Error("edit poster error 2: %s", err.Error())
			return err
		}
	}
	return err
}

func (s *Service) TogglePoster(c *bm.Context, req *pb.TogglePosterReq) (err error) {
	// 查找记录，并检查是否需要更改状态
	if poster, err := s.dao.FindPosterById(c, req.Id); err != nil {
		log.Error("toggle error 0: %s", err.Error())
		return err
	} else if poster == nil {
		err = ecode.NothingFound
		log.Error("toggle error 1: %s", err.Error())
		return err
	} else {
		if poster.OnlineStatus == req.OnlineStatus {
			err = ecode.Error(-703, "当前配置已是该状态，无需调整")
			log.Error("toggle error 2: %s", err.Error())
			return err
		}

		tx := s.dao.DB.Begin()

		// 修改配置上下线
		if err = s.dao.TogglePoster(c, req.Id, req.OnlineStatus); err != nil {
			log.Error("toggle error 3: %s", err.Error())
			tx.Rollback()
			return err
		}

		if req.OnlineStatus == 0 && poster.IsCenteral == 1 {
			if err = s.dao.UnCenterPoster(c, req.Id); err != nil {
				log.Error("toggle error 4: %s", err.Error())
				tx.Rollback()
				return err
			}
		}

		defer tx.Commit()
	}
	return err
}

func (s *Service) CenterPoster(c *bm.Context, req *pb.CenterPosterReq) (err error) {
	// 查找记录，并检查是否需要更改状态
	if poster, err := s.dao.FindPosterById(c, req.Id); err != nil {
		log.Error("center poster error 0: %s", err.Error())
		return err
	} else if poster == nil {
		err = ecode.NothingFound
		log.Error("center poster error 1: %s", err.Error())
		return err
	} else {
		if poster.IsCenteral == req.IsCenteral {
			err = ecode.Error(-703, "当前配置已是该状态，无需调整")
			log.Error("center poster error 2: %s", err.Error())
			return err
		}
	}

	// 取消单个定位状态
	if req.IsCenteral == 0 {
		if err = s.dao.UnCenterPoster(c, req.Id); err != nil {
			log.Error("center poster error 3: %s", err.Error())
			return err
		}
	} else if req.IsCenteral == 1 {
		tx := s.dao.DB.Begin()
		if err = s.dao.UnCenterAllPoster(c); err != nil {
			log.Error("center poster error 4: %s", err.Error())
			tx.Rollback()
			return err
		}
		if err = s.dao.CenterPoster(c, req.Id); err != nil {
			log.Error("center poster error 5: %s", err.Error())
			tx.Rollback()
			return err
		}
		defer tx.Commit()
	}

	return err
}

func (s *Service) DeletePoster(c *bm.Context, req *pb.DeletePosterReq) (err error) {
	// 查找记录，并检查是否需要更改状态
	if poster, err := s.dao.FindPosterById(c, req.Id); err != nil {
		log.Error("delete poster error 0: %s", err.Error())
		return err
	} else if poster == nil {
		err = ecode.NothingFound
		log.Error("delete poster error 1: %s", err.Error())
		return err
	} else {
		if err = s.dao.DeletePoster(c, req.Id); err != nil {
			log.Error("delete poster error 2: %s", err.Error())
			return err
		}
	}
	return err
}

func (s *Service) GetPosterList(c *bm.Context, req *pb.GetPosterListReq) (rep *pb.GetPosterListRep, err error) {
	rep = new(pb.GetPosterListRep)

	if list, err := s.dao.GetPosterList(c, req.PageNum, req.PageSize); err != nil {
		log.Error("get poster list error 0: %s", err.Error())
		return nil, err
	} else {
		items := make([]*pb.Poster, len(list))
		for i, item := range list {
			items[i] = &pb.Poster{
				Id:           item.ID,
				BgImage:      item.BgImage,
				ContestID:    item.ContestID,
				IsCenteral:   item.IsCenteral,
				OnlineStatus: item.OnlineStatus,
				CreatedBy:    item.CreatedBy,
				Order:        item.PositionOrder,
				Ctime:        item.CTime.Unix(),
			}
		}
		rep.Items = items
	}

	if total, err := s.dao.GetPosterCount(c); err != nil {
		log.Error("get poster list error 1: %s", err.Error())
		return nil, err
	} else {
		rep.Page = &pb.PosterPager{
			Total:    int32(total),
			PageSize: req.PageSize,
			PageNum:  req.PageNum,
		}
	}

	return rep, nil
}

func (s *Service) GetEffectivePosterList(c *bm.Context) (rep *pb.GetEffectivePosterListRep, err error) {
	rep = new(pb.GetEffectivePosterListRep)

	if list, err := s.dao.GetEffectivePosterList(c); err != nil {
		log.Error("get effective poster list error 0: %s", err.Error())
		return nil, err
	} else {
		items := make([]*pb.EffectivePoster, len(list))
		for i, item := range list {
			items[i] = &pb.EffectivePoster{
				Id:         item.ID,
				BgImage:    item.BgImage,
				ContestID:  item.ContestID,
				IsCenteral: item.IsCenteral,
				Order:      item.PositionOrder,
				Ctime:      item.CTime.Unix(),
			}
		}
		rep.Items = items
	}

	return
}
