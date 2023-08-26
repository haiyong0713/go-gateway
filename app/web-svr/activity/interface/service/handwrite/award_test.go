package handwrite

import (
	"context"
	"testing"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/dao/handwrite"
	"go-gateway/app/web-svr/activity/interface/dao/rank"
	hwMdl "go-gateway/app/web-svr/activity/interface/model/handwrite"
	rankMdl "go-gateway/app/web-svr/activity/interface/model/rank"

	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAwardMemberCount(t *testing.T) {
	Convey("test award member count success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().GetAwardCount(Any()).Return(&hwMdl.AwardCount{God: 1, Tired: 0, New: 1}, nil)
		s.handwrite = mockHandwriteDao

		_, err := s.AwardMemberCount(context.Background())
		So(err, ShouldBeNil)

	}))
	Convey("test award member count err", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().GetAwardCount(Any()).Return(nil, ecode.ActivityWriteHandActivityMemberErr)
		s.handwrite = mockHandwriteDao

		_, err := s.AwardMemberCount(context.Background())
		So(err, ShouldEqual, ecode.ActivityWriteHandActivityMemberErr)

	}))
	Convey("test award member count nil", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().GetAwardCount(Any()).Return(nil, nil)
		s.handwrite = mockHandwriteDao

		_, err := s.AwardMemberCount(context.Background())
		So(err, ShouldBeNil)

	}))
}

func TestRank(t *testing.T) {
	Convey("test rank success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockRankDao := rank.NewMockDao(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		defer mockCtl.Finish()

		rankRes := testRankReturn1()

		mockRankDao.EXPECT().GetRank(Any(), Any()).Return(rankRes, nil)
		infoMap1 := testGetAccountInfo()
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)

		s.rank = mockRankDao
		s.accClient = mockAcc
		expected := expectedRank()
		res, err := s.Rank(context.Background())
		So(err, ShouldBeNil)
		So(res, ShouldResemble, expected)

	}))

	Convey("test rank get rank error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)

		mockRankDao := rank.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockRankDao.EXPECT().GetRank(Any(), Any()).Return(nil, ecode.ActivityWriteHandActivityMemberErr)

		s.rank = mockRankDao
		_, err := s.Rank(context.Background())
		So(err, ShouldNotBeNil)

	}))

	Convey("test rank get mid info error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockRankDao := rank.NewMockDao(mockCtl)
		defer mockCtl.Finish()
		rankRes := testRankReturn1()

		mockRankDao.EXPECT().GetRank(Any(), Any()).Return(rankRes, nil)
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(nil, ecode.ActivityWriteHandActivityMemberErr)

		s.accClient = mockAcc
		s.rank = mockRankDao
		_, err := s.Rank(context.Background())
		So(err, ShouldNotBeNil)

	}))
}

func expectedRank() *hwMdl.RankReply {
	res := &hwMdl.RankReply{}
	rank := make([]*hwMdl.RankMember, 0)
	rank = append(rank, &hwMdl.RankMember{
		Account: &hwMdl.Account{
			Mid:  1,
			Name: "1",
		}, Score: 1111,
	}, &hwMdl.RankMember{
		Account: &hwMdl.Account{
			Mid:  3,
			Name: "3",
		}, Score: 111,
	}, &hwMdl.RankMember{
		Account: &hwMdl.Account{
			Mid:  2,
			Name: "2",
		}, Score: 111,
	}, &hwMdl.RankMember{
		Account: &hwMdl.Account{
			Mid:  4,
			Name: "4",
		}, Score: 11,
	})
	res.Rank = rank
	return res
}
func testGetAccountInfo() map[int64]*accapi.Info {
	account := make(map[int64]*accapi.Info)
	account[1] = &accapi.Info{
		Mid:  1,
		Name: "1",
	}
	account[2] = &accapi.Info{
		Mid:  2,
		Name: "2",
	}
	account[3] = &accapi.Info{
		Mid:  3,
		Name: "3",
	}
	account[4] = &accapi.Info{
		Mid:  4,
		Name: "4",
	}
	return account
}

func testRankReturn1() []*rankMdl.Redis {
	return []*rankMdl.Redis{{
		Mid:   1,
		Rank:  2,
		Score: 1111,
	}, {
		Mid:   3,
		Rank:  2,
		Score: 111,
	}, {
		Mid:   2,
		Rank:  3,
		Score: 111,
	}, {
		Mid:   4,
		Rank:  4,
		Score: 11,
	}}
}

func TestPersonal(t *testing.T) {
	Convey("test personal success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)

		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().GetMidAward(Any(), Any()).Return(&hwMdl.MidAward{God: 4, Tired: 1, New: 1, Rank: 1, Score: 1}, nil)
		mockHandwriteDao.EXPECT().GetAwardCount(Any()).Return(&hwMdl.AwardCount{God: 10, Tired: 10, New: 1}, nil)
		mockAcc.EXPECT().Info3(Any(), Any()).Return(&accapi.InfoReply{Info: &accapi.Info{
			Mid:  4,
			Name: "4",
		}}, nil)

		s.handwrite = mockHandwriteDao
		s.accClient = mockAcc

		expected := testExpectedPersonal()
		res, err := s.Personal(context.Background(), 1)
		So(err, ShouldBeNil)
		So(res, ShouldResemble, expected)

	}))

	Convey("test personal error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)

		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().GetMidAward(Any(), Any()).Return(nil, ecode.ActivityWriteHandActivityMemberErr)
		mockHandwriteDao.EXPECT().GetAwardCount(Any()).Return(nil, ecode.ActivityWriteHandActivityMemberErr)
		mockAcc.EXPECT().Info3(Any(), Any()).Return(nil, ecode.ActivityWriteHandActivityMemberErr)

		s.handwrite = mockHandwriteDao
		s.accClient = mockAcc

		_, err := s.Personal(context.Background(), 1)
		So(err, ShouldNotBeNil)

	}))
}

func testExpectedPersonal() *hwMdl.PersonalReply {
	res := &hwMdl.PersonalReply{}
	res.Money = 10000000
	res.Rank = 1
	res.Score = 1
	account := &hwMdl.Account{}
	account.Mid = 4
	account.Name = "4"
	res.Account = account
	return res
}
