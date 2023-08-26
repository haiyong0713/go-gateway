package account

import (
	"context"
	"time"

	model "go-gateway/app/app-svr/app-interface/interface-legacy/model/account"

	passportuser "git.bilibili.co/bapis/bapis-go/passport/service/user"
)

var configs = &model.StatisticsConfigs{
	Info: []*model.StatisticsInfo{
		{
			Type: "用户网络身份标识和鉴权信息",
			Ids:  []int64{2, 3, 4, 5, 6, 34, 33, 7},
		},
		{
			Type: "用户基本信息",
			Ids:  []int64{8, 9, 10, 11},
		},
		{
			Type: "用户身份证明",
			Ids:  []int64{12, 13, 14, 15},
		},
		{
			Type: "个人财产信息",
			Ids:  []int64{16, 17},
		},
		{
			Type: "内容制作与发布",
			Ids:  []int64{18, 19, 20, 21, 22, 23, 24},
		},
		{
			Type: "互动与服务",
			Ids:  []int64{25, 26, 27, 28, 29, 30, 31, 32},
		},
	},
}

func pureDate(in time.Time) time.Time {
	year, month, day := in.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, in.Location())
}

func (s *Service) ExportStatistics(ctx context.Context, mid int64, sel string) (*model.ExportedStatistics, error) {
	card, err := s.accDao.Card(ctx, mid)
	if err != nil {
		return nil, err
	}

	ppUser, err := s.accDao.UserDetail(ctx, &passportuser.UserDetailReq{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}

	dateAt := time.Now().Add(-time.Duration(39 * time.Hour))
	stat, err := s.accDao.ExportStatistics(ctx, mid, dateAt, sel)
	if err != nil {
		return nil, err
	}
	out := &model.ExportedStatistics{}
	out.Date = pureDate(dateAt).Unix()
	out.Statistics = append(out.Statistics, model.ResolveShuangQiongStats(stat, card, ppUser)...)
	out.Configs = configs
	return out, nil
}
