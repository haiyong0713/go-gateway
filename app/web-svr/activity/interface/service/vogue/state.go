package vogue

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	accountAPI "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) State(c context.Context, mid int64) (res *model.State, err error) {
	res = new(model.State)
	eg := errgroup.WithContext(c)
	var selection *model.Selection
	eg.GOMAXPROCS(5)
	eg.Go(func(ctx context.Context) (err error) {
		var (
			task        *model.Task
			goods       *model.Goods
			cost, today int64
		)
		if task, err = s.dao.Task(ctx, mid); err != nil {
			log.Error("s.dao.Task(%v)", err)
			return
		}
		if task == nil { // 没有参加活动
			return
		}
		if goods, err = s.dao.Goods(ctx, task.Goods); err != nil {
			log.Error("s.dao.Goods(%v)", err)
			return
		}
		stock := goods.Stock - goods.Send
		if stock < 0 {
			stock = 0
		}
		selection = &model.Selection{
			Name:       goods.Name,
			Type:       "real",
			Cost:       goods.Score,
			Picture:    goods.Picture,
			Status:     task.GoodsState,
			Stock:      stock,
			TotalStock: goods.Stock,
		}
		if goods.AttrVal(model.GoodsAttrReal) == 0 {
			selection.Type = "virtual"
		}
		if selection.Status == 1 {
			if stock <= 0 {
				selection.Status = 4
			}
			if goods.AttrVal(model.GoodsAttrSellOut) == 1 {
				selection.Status = 5
			}
		}
		if cost, today, err = s.getScore(c, mid); err != nil {
			log.Error("s.getScore(%v)", err)
			return err
		}
		res.User = &model.User{
			Score:     cost,
			Today:     today,
			Token:     model.TokenDecode(mid),
			Selection: selection,
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		var start, end int64
		start, end, err = s.actTime(c)
		res.Duration = &model.Duration{
			Start: start,
			End:   end,
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		var (
			conf                              string
			data                              = make([]*model.ScoreItem, 0, 0)
			inviteScore, todayLimit, maxScore int64
			start, end                        int64
		)
		if start, end, err = s.secondDoubleTime(c); err != nil {
			log.Error("s.doubleTime(%v)", err)
			return err
		}
		res.DoubleDuration = &model.Duration{
			Start: start,
			End:   end,
		}
		if conf, err = s.dao.Config(ctx, "score_list"); err != nil {
			log.Error("s.dao.Config(%v)", err)
			return err
		}
		if conf == "" {
			return
		}
		if err = json.Unmarshal([]byte(conf), &data); err != nil {
			log.Error("json.Unmarshal(%v)", err)
			return err
		}
		if inviteScore, err = s.inviteScore(c); err != nil {
			log.Error("s.inviteScore(%v)", err)
			return err
		}
		if todayLimit, err = s.todayLimit(c); err != nil {
			log.Error("s.todayLimit(%v)", err)
			return err
		}
		if maxScore, err = s.viewScore(c); err != nil {
			log.Error("s.viewScore(%v)", err)
			return err
		}
		res.ScoreList = &model.ScoreList{
			FirstBonusTime: make([]int64, 0, 2),
			InviteScore:    inviteScore,
			TodayLimit:     todayLimit,
			MaxScore:       maxScore,
		}
		for _, n := range data {
			if n.Max-1 > res.ScoreList.MaxScore {
				res.ScoreList.MaxScore = n.Max - 1
			}
			if n.Show && len(res.ScoreList.FirstBonusTime) < 2 {
				res.ScoreList.FirstBonusTime = append(res.ScoreList.FirstBonusTime, n.Num)
			}
		}
		now := time.Now().Unix()
		if start <= now && now <= end {
			res.ScoreList.MaxScore = res.ScoreList.MaxScore * 2
			res.ScoreList.InviteScore = res.ScoreList.InviteScore * 2
			res.ScoreList.TodayLimit = res.ScoreList.TodayLimit * 2
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		res.PlayList = []*model.PlayListItem{}
		var data string
		if data, err = s.dao.Config(ctx, "play_list"); err != nil {
			log.Error("s.dao.Config(%v)", err)
			return err
		}
		if data == "" {
			return
		}
		if err = json.Unmarshal([]byte(data), &res.PlayList); err != nil {
			log.Error("json.Unmarshal(%v)", err)
			return err
		}
		return err
	})
	eg.Go(func(ctx context.Context) (err error) {
		var account *accountAPI.InfoReply
		if account, err = s.accClient.Info3(c, &accountAPI.MidReq{Mid: mid}); err != nil {
			log.Error("s.accClient.Info3(%v) error(%v)", account, err)
			return err
		}
		res.UserInfo = &model.UserInfo{
			Name:    account.Info.GetName(),
			Picture: account.Info.GetFace(),
		}
		return err
	})
	if err = eg.Wait(); err != nil {
		log.Error("state err %v", err)
		return nil, err
	}
	return
}

func (s *Service) getScore(c context.Context, mid int64) (res, today int64, err error) {
	return
}

func (s *Service) doubleTime(c context.Context) (start, end int64, err error) {
	var startStr, endStr string
	if startStr, err = s.dao.Config(c, "act_double_start"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return
	}
	if start, err = strconv.ParseInt(startStr, 10, 64); err != nil {
		log.Error("strconv.ParseInt(%v)", err)
		return
	}
	if endStr, err = s.dao.Config(c, "act_double_end"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return
	}
	if end, err = strconv.ParseInt(endStr, 10, 64); err != nil {
		log.Error("strconv.ParseInt(%v)", err)
		return
	}
	return
}

// hardcode
func (s *Service) secondDoubleTime(c context.Context) (start, end int64, err error) {
	var startStr, endStr string
	if startStr, err = s.dao.Config(c, "act_second_double_start"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return
	}
	if start, err = strconv.ParseInt(startStr, 10, 64); err != nil {
		log.Error("strconv.ParseInt(%v)", err)
		return
	}
	if endStr, err = s.dao.Config(c, "act_second_double_end"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return
	}
	if end, err = strconv.ParseInt(endStr, 10, 64); err != nil {
		log.Error("strconv.ParseInt(%v)", err)
		return
	}
	return
}

func (s *Service) inviteScore(c context.Context) (inviteScore int64, err error) {
	var score string
	if score, err = s.dao.Config(c, "invite_score"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return 0, err
	}
	if inviteScore, err = strconv.ParseInt(score, 10, 64); err != nil {
		log.Error("strconv.ParseInt(%v)", err)
		return 0, err
	}
	return
}

func (s *Service) viewScore(c context.Context) (viewScore int64, err error) {
	var score string
	if score, err = s.dao.Config(c, "view_score"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return 0, err
	}
	if viewScore, err = strconv.ParseInt(score, 10, 64); err != nil {
		log.Error("strconv.ParseInt(%v)", err)
		return 0, err
	}
	return
}

func (s *Service) todayLimit(c context.Context) (todayLimit int64, err error) {
	var limit string
	if limit, err = s.dao.Config(c, "today_limit"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return 0, err
	}
	if todayLimit, err = strconv.ParseInt(limit, 10, 64); err != nil {
		log.Error("strconv.ParseInt(%v)", err)
		return 0, err
	}
	return
}

func (s *Service) actTime(c context.Context) (start, end int64, err error) {
	var startStr, endStr string
	if startStr, err = s.dao.Config(c, "act_start"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return 0, 0, err
	}
	if start, err = strconv.ParseInt(startStr, 10, 64); err != nil {
		log.Error("strconv.ParseInt(%v)", err)
		return 0, 0, err
	}
	if endStr, err = s.dao.Config(c, "act_end"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return 0, 0, err
	}
	if end, err = strconv.ParseInt(endStr, 10, 64); err != nil {
		log.Error("strconv.ParseInt(%v)", err)
		return 0, 0, err
	}
	return
}
