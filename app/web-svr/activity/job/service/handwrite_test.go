package service

import (
	"testing"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/job/dao/handwrite"
	"go-gateway/app/web-svr/activity/job/dao/rank"
	mdlRank "go-gateway/app/web-svr/activity/job/model/rank"

	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHandWriteMemberScoreSuccess(t *testing.T) {
	Convey("test handwrite member score success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockRelation := relationapi.NewMockRelationClient(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		defer mockCtl.Finish()
		// mock
		// arc data mock
		arcs1 := testgetArcs1()
		// arcs2 := testgetArcs2()
		// arcs3 := testgetArcs3()
		// arcs4 := testgetArcs4()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs1}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)

		// relation mock
		statMap1 := testGetRelationMap1()
		statMap2 := testGetRelationMap2()
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap1}, nil)
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)

		// account mock
		infoMap1 := testGetAccountInfo()
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)

		s.arcClient = mockArc
		s.relationClient = mockRelation
		s.accClient = mockAcc

		s.HandWriteMemberScore()
	}))
}

func TestHandWriteMemberScoreArchiveError(t *testing.T) {
	Convey("test handwrite member score archive error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		defer mockCtl.Finish()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		s.arcClient = mockArc
		s.HandWriteMemberScore()
	}))
}

func TestHandWriteMemberDeleteAid(t *testing.T) {
	Convey("test handwrite member delete aid", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockRelation := relationapi.NewMockRelationClient(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		defer mockCtl.Finish()
		// mock
		// arc data mock
		arcs5 := testgetArcs5()
		// arcs2 := testgetArcs2()
		// arcs3 := testgetArcs3()
		// arcs4 := testgetArcs4()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs5}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)

		// relation mock
		statMap1 := testGetRelationMap1()
		statMap2 := testGetRelationMap2()
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap1}, nil)
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)

		// account mock
		infoMap1 := testGetAccountInfo()
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)

		s.arcClient = mockArc
		s.relationClient = mockRelation
		s.accClient = mockAcc

		s.HandWriteMemberScore()
	}))
}
func TestFandWriteMemberScoreMidDistinctError(t *testing.T) {
	Convey("test handwrite member mid distinct error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockHandWriteDao := handwrite.NewMockDao(mockCtl)
		mockRelation := relationapi.NewMockRelationClient(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)

		defer mockCtl.Finish()

		arcs1 := testgetArcs1()
		// arcs2 := testgetArcs2()
		// arcs3 := testgetArcs3()
		// arcs4 := testgetArcs4()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs1}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)

		mockHandWriteDao.EXPECT().MidListDistinct(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		mockHandWriteDao.EXPECT().GetActivityMember(Any()).Return([]int64{1, 2}, nil)
		// mockHandWriteDao.EXPECT().SetMidInitFans(Any(), Any()).Return(nil)

		mockRelation.EXPECT().Stats(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		mockRelation.EXPECT().Stats(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)

		infoMap1 := testGetAccountInfo()
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)

		s.arcClient = mockArc
		s.handWrite = mockHandWriteDao
		s.accClient = mockAcc
		s.arcClient = mockArc
		s.relationClient = mockRelation
		s.HandWriteMemberScore()
	}))
}

func TestHandWriteMemberScoreSetFansError(t *testing.T) {
	Convey("test handwrite member mid distinct error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockHandWriteDao := handwrite.NewMockDao(mockCtl)
		mockRelation := relationapi.NewMockRelationClient(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)

		defer mockCtl.Finish()

		arcs1 := testgetArcs1()
		// arcs2 := testgetArcs2()
		// arcs3 := testgetArcs3()
		// arcs4 := testgetArcs4()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs1}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)

		mockHandWriteDao.EXPECT().MidListDistinct(Any(), Any()).Return(nil, nil)
		mockHandWriteDao.EXPECT().GetActivityMember(Any()).Return([]int64{1, 2}, nil)
		// mockHandWriteDao.EXPECT().SetMidInitFans(Any(), Any()).Return(ecode.ActivityWriteHandFansErr)

		mockRelation.EXPECT().Stats(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		mockRelation.EXPECT().Stats(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)

		infoMap1 := testGetAccountInfo()
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)

		mockHandWriteDao.EXPECT().AddMidAward(Any(), Any()).Return(nil)
		mockHandWriteDao.EXPECT().SetAwardCount(Any(), Any()).Return(nil)

		s.accClient = mockAcc
		s.arcClient = mockArc
		s.handWrite = mockHandWriteDao
		s.relationClient = mockRelation
		s.HandWriteMemberScore()
	}))
}

func TestHandWriteMemberScoreSaveError(t *testing.T) {
	Convey("test handwrite save db error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockRelation := relationapi.NewMockRelationClient(mockCtl)
		mockRankDao := rank.NewMockDao(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockHandWriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()
		// mock
		// arc data mock
		arcs1 := testgetArcs1()
		// arcs2 := testgetArcs2()
		// arcs3 := testgetArcs3()
		// arcs4 := testgetArcs4()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs1}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)

		// relation mock
		statMap1 := testGetRelationMap1()
		statMap2 := testGetRelationMap2()
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap1}, nil)
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)

		mockHandWriteDao.EXPECT().MidListDistinct(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		mockHandWriteDao.EXPECT().GetActivityMember(Any()).Return([]int64{1, 2}, nil)
		mockHandWriteDao.EXPECT().SetMidInitFans(Any(), Any()).Return(nil)

		// account mock
		infoMap1 := testGetAccountInfo()
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)

		// dao mock
		mockRankDao.EXPECT().GetRank(Any(), Any()).Return(nil, nil)
		mockRankDao.EXPECT().SetRank(Any(), Any(), Any()).Return(ecode.ActivityWriteHandFansErr)
		mockRankDao.EXPECT().BatchAddRank(Any(), Any()).Return(ecode.ActivityWriteHandFansErr)

		s.handWrite = mockHandWriteDao
		s.arcClient = mockArc
		s.relationClient = mockRelation
		s.accClient = mockAcc
		s.rank = mockRankDao

		s.HandWriteMemberScore()
	}))
}

func TestFandWriteMemberScoreInfo3Error(t *testing.T) {
	Convey("test handwrite member info3 error", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockHandWriteDao := handwrite.NewMockDao(mockCtl)
		mockRelation := relationapi.NewMockRelationClient(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockRankDao := rank.NewMockDao(mockCtl)

		defer mockCtl.Finish()

		arcs1 := testgetArcs1()
		// arcs2 := testgetArcs2()
		// arcs3 := testgetArcs3()
		// arcs4 := testgetArcs4()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs1}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)

		mockHandWriteDao.EXPECT().MidListDistinct(Any(), Any()).Return(nil, nil)
		mockHandWriteDao.EXPECT().GetActivityMember(Any()).Return([]int64{1, 2}, nil)
		// mockHandWriteDao.EXPECT().SetMidInitFans(Any(), Any()).Return(nil)

		mockRelation.EXPECT().Stats(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		mockRelation.EXPECT().Stats(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)

		mockAcc.EXPECT().Infos3(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)

		mockRankDao.EXPECT().SetRank(Any(), Any(), Any()).Return(nil)
		mockRankDao.EXPECT().GetRank(Any(), Any()).Return(nil, nil)
		mockHandWriteDao.EXPECT().AddMidAward(Any(), Any()).Return(nil)
		mockHandWriteDao.EXPECT().SetAwardCount(Any(), Any()).Return(nil)

		s.arcClient = mockArc
		s.handWrite = mockHandWriteDao
		s.accClient = mockAcc
		s.arcClient = mockArc
		s.relationClient = mockRelation
		s.rank = mockRankDao

		s.HandWriteMemberScore()
	}))
}

func TestHandWriteMemberScoreOrderByHistory(t *testing.T) {
	Convey("test handwrite set", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockRelation := relationapi.NewMockRelationClient(mockCtl)
		mockRankDao := rank.NewMockDao(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockHandWriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()
		// mock
		// arc data mock
		arcs2 := testgetArcs2()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)

		// relation mock
		statMap1 := testGetRelationMap1()
		statMap2 := testGetRelationMap2()
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap1}, nil)
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)

		mockHandWriteDao.EXPECT().MidListDistinct(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		mockHandWriteDao.EXPECT().GetActivityMember(Any()).Return([]int64{1, 2}, nil)
		mockHandWriteDao.EXPECT().SetMidInitFans(Any(), Any()).Return(nil)

		// account mock
		infoMap1 := testGetAccountInfo()
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)

		// dao mock
		rankReturn := testRankReturn1()
		mockRankDao.EXPECT().GetRank(Any(), Any()).Return(rankReturn, nil)
		mockRankDao.EXPECT().SetRank(Any(), Any(), Any()).Return(nil)
		mockRankDao.EXPECT().BatchAddRank(Any(), Any()).Return(nil)

		s.handWrite = mockHandWriteDao
		s.arcClient = mockArc
		s.relationClient = mockRelation
		s.accClient = mockAcc
		s.rank = mockRankDao

		s.HandWriteMemberScore()
	}))
}

func TestHandWriteMemberScoreCountAward(t *testing.T) {
	Convey("test handwrite member count award", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockRelation := relationapi.NewMockRelationClient(mockCtl)
		mockRankDao := rank.NewMockDao(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockHandWriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()
		// mock
		// arc data mock
		arcs3 := testgetArcs3()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)

		// relation mock
		statMap1 := testGetRelationMap1()
		statMap2 := testGetRelationMap2()
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap1}, nil)
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)

		mockHandWriteDao.EXPECT().MidListDistinct(Any(), Any()).Return(nil, nil)
		mockHandWriteDao.EXPECT().GetActivityMember(Any()).Return([]int64{1, 2}, nil)
		mockHandWriteDao.EXPECT().SetMidInitFans(Any(), Any()).Return(nil)

		// account mock
		infoMap1 := testGetAccountInfo()
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)

		// dao mock
		rankReturn := testRankReturn1()
		mockRankDao.EXPECT().GetRank(Any(), Any()).Return(rankReturn, nil)
		mockRankDao.EXPECT().SetRank(Any(), Any(), Any()).Return(nil)
		mockRankDao.EXPECT().BatchAddRank(Any(), Any()).Return(nil)
		mockHandWriteDao.EXPECT().AddMidAward(Any(), Any()).Return(nil)
		mockHandWriteDao.EXPECT().SetAwardCount(Any(), Any()).Return(nil)
		s.handWrite = mockHandWriteDao
		s.arcClient = mockArc
		s.relationClient = mockRelation
		s.accClient = mockAcc
		s.rank = mockRankDao

		s.HandWriteMemberScore()
	}))
}

func TestHandWriteMemberScoreCount(t *testing.T) {
	Convey("test handwrite member score count", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockRelation := relationapi.NewMockRelationClient(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockRankDao := rank.NewMockDao(mockCtl)
		mockHandWriteDao := handwrite.NewMockDao(mockCtl)

		defer mockCtl.Finish()
		// mock
		// arc data mock
		arcs4 := testgetArcs4()
		// arcs2 := testgetArcs2()
		// arcs3 := testgetArcs3()
		// arcs4 := testgetArcs4()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
		// mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)
		mockHandWriteDao.EXPECT().MidListDistinct(Any(), Any()).Return(nil, ecode.ActivityWriteHandFansErr)
		mockHandWriteDao.EXPECT().GetActivityMember(Any()).Return([]int64{1, 2}, nil)
		mockHandWriteDao.EXPECT().SetMidInitFans(Any(), Any()).Return(nil)

		// relation mock
		statMap1 := testGetRelationMap1()
		statMap2 := testGetRelationMap2()
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap1}, nil)
		mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)

		rankReturn := testRankReturn1()
		mockRankDao.EXPECT().GetRank(Any(), Any()).Return(rankReturn, nil)
		mockRankDao.EXPECT().SetRank(Any(), Any(), Any()).Return(nil)
		mockRankDao.EXPECT().BatchAddRank(Any(), Any()).Return(nil)

		// account mock
		infoMap1 := testGetAccountInfo()
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
		s.rank = mockRankDao
		s.arcClient = mockArc
		s.relationClient = mockRelation
		s.accClient = mockAcc
		s.handWrite = mockHandWriteDao

		s.HandWriteMemberScore()
	}))
}
func testRankReturn1() []*mdlRank.Redis {
	return []*mdlRank.Redis{{
		Mid:  3,
		Rank: 2,
	}, {
		Mid:  2,
		Rank: 3,
	}, {
		Mid:  1,
		Rank: 4,
	}}
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
func testGetRelationMap1() map[int64]*relationapi.StatReply {
	relation := make(map[int64]*relationapi.StatReply)
	relation[1] = &relationapi.StatReply{
		Mid:      1,
		Follower: 111,
	}
	relation[2] = &relationapi.StatReply{
		Mid:      2,
		Follower: 3,
	}
	relation[3] = &relationapi.StatReply{
		Mid:      3,
		Follower: 10,
	}
	relation[4] = &relationapi.StatReply{
		Mid:      4,
		Follower: 4,
	}
	return relation
}
func testGetRelationMap2() map[int64]*relationapi.StatReply {
	relation := make(map[int64]*relationapi.StatReply)
	relation[2] = &relationapi.StatReply{
		Mid:      2,
		Follower: 3,
	}
	relation[3] = &relationapi.StatReply{
		Mid:      4,
		Follower: 4,
	}
	relation[4] = &relationapi.StatReply{
		Mid:      4,
		Follower: 4,
	}
	return relation
}
func testgetArcs1() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[480041427] = &arcapi.Arc{
		Aid:   480041427,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 1,
			Coin: 2,
		},
	}
	arcs1[800121568] = &arcapi.Arc{
		Aid:   800121568,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 2,
			Coin: 3,
		},
	}
	arcs1[920004623] = &arcapi.Arc{
		Aid:   920004623,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
		Stat: arcapi.Stat{
			View: 4,
			Coin: 5,
		},
	}
	arcs1[440112197] = &arcapi.Arc{
		Aid:   440112197,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
		Stat: arcapi.Stat{
			View: 6,
			Coin: 7,
		},
	}
	arcs1[920111156] = &arcapi.Arc{
		Aid:   920111156,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
		Stat: arcapi.Stat{
			View: 6,
			Coin: 7,
		},
	}
	arcs1[600022189] = &arcapi.Arc{
		Aid:   600022189,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
		Stat: arcapi.Stat{
			View: 6,
			Coin: 7,
		},
	}
	arcs1[960121700] = &arcapi.Arc{
		Aid:   960121700,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
		Stat: arcapi.Stat{
			View: 6,
			Coin: 66666,
		},
	}
	return arcs1
}

func testgetArcs2() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[960054204] = &arcapi.Arc{
		Aid:   960054204,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 1,
		},
	}
	arcs1[640011253] = &arcapi.Arc{
		Aid:   640011253,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 1,
		},
	}
	arcs1[360045380] = &arcapi.Arc{
		Aid:   360045380,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 1,
		},
	}
	arcs1[680042159] = &arcapi.Arc{
		Aid:   680042159,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 1,
		},
	}
	return arcs1
}

func testgetArcs3() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[640114618] = &arcapi.Arc{
		Aid:   640114618,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[920111156] = &arcapi.Arc{
		Aid:   920111156,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[600022189] = &arcapi.Arc{
		Aid:   600022189,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[960121700] = &arcapi.Arc{
		Aid:   960121700,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[480041427] = &arcapi.Arc{
		Aid:   480041427,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[240011993] = &arcapi.Arc{
		Aid:   240011993,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[480003774] = &arcapi.Arc{
		Aid:   480003774,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	return arcs1
}
func testgetArcs4() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[480041427] = &arcapi.Arc{
		Aid:   480041427,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 333233,
			Coin: 8832,

			Reply:   1231,
			Danmaku: 993,
			Fav:     9832,
			Like:    9312,
		},
	}
	arcs1[240011993] = &arcapi.Arc{
		Aid:   240011993,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
	}
	arcs1[480003774] = &arcapi.Arc{
		Aid:   480003774,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
	}
	arcs1[600067439] = &arcapi.Arc{
		Aid:   600067439,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
	}
	return arcs1
}

func testgetArcs5() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[480041427] = &arcapi.Arc{
		Aid:   480041427,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 333233,
			Coin: 8832,

			Reply:   1231,
			Danmaku: 993,
			Fav:     9832,
			Like:    9312,
		},
	}
	arcs1[240011993] = &arcapi.Arc{
		Aid:   240011993,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
	}
	arcs1[480003774] = &arcapi.Arc{
		Aid:   480003774,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
	}

	return arcs1
}
