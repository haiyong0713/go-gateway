package remix

import (
	"context"
	"testing"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/dao/rank"
	rankmdl "go-gateway/app/web-svr/activity/interface/model/rank"
	remixmdl "go-gateway/app/web-svr/activity/interface/model/remix"
	taskmdl "go-gateway/app/web-svr/activity/interface/model/task"

	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMemberCount(t *testing.T) {
	Convey("test award member count success", t, WithService(func(s *Service) {

		_, err := s.MemberCount(context.Background())
		So(err, ShouldBeNil)

	}))

}
func testGetAccountInfo2() map[int64]*accapi.Info {
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
func TestPersonal(t *testing.T) {
	Convey("test personal success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockRankDao := rank.NewMockDao(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockArc := api.NewMockArchiveClient(mockCtl)

		defer mockCtl.Finish()
		arcMap := testgetArcs()
		mockRankDao.EXPECT().GetMidRank(Any(), Any(), Any()).Return(&rankmdl.Redis{Mid: 1, Rank: 1, Score: 1, Aids: []int64{1, 2}}, nil)
		mockAcc.EXPECT().Info3(Any(), Any()).Return(&accapi.InfoReply{Info: &accapi.Info{
			Mid:  1,
			Name: "4",
		}}, nil)

		mockArc.EXPECT().Arcs(Any(), Any()).Return(&api.ArcsReply{Arcs: arcMap}, nil)
		s.accClient = mockAcc
		client.ArchiveClient = mockArc
		s.rank = mockRankDao
		expected := testPersonalExpected()
		res, err := s.Personal(context.Background(), 1)
		So(err, ShouldBeNil)
		So(res, ShouldResemble, expected)

	}))
}
func testPersonalExpected() *remixmdl.MemberActivityInfoReply {
	return &remixmdl.MemberActivityInfoReply{
		Task: []*taskmdl.MidRule{
			{Object: 1, MID: 1, State: 1},
			{Object: 2, MID: 1, State: 1},
		},
		Rank: &remixmdl.Rank{
			Rank:  1,
			Score: 1,
			Video: []*remixmdl.Video{
				{
					Aid:      1,
					TypeName: "",
					Title:    "",
					Desc:     "",
					Duration: 0,
					Pic:      "",
					View:     4,
				},
				{
					Aid:      2,
					TypeName: "",
					Title:    "",
					Desc:     "",
					Duration: 0,
					Pic:      "",
					View:     3,
				},
			},
		},
		Account: &remixmdl.Account{
			Mid:  1,
			Name: "4",
		},
	}
}

func testgetArcs() map[int64]*api.Arc {
	arcs1 := make(map[int64]*api.Arc)
	arcs1[2] = &api.Arc{
		Aid:   2,
		State: 0,
		Author: api.Author{
			Mid: 1,
		},
		Stat: api.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[1] = &api.Arc{
		Aid:   1,
		State: 0,
		Author: api.Author{
			Mid: 1,
		},
		Stat: api.Stat{
			View: 4,
			Coin: 666666,
		},
	}
	return arcs1
}

func testgetArcs2() map[int64]*api.Arc {
	arcs1 := make(map[int64]*api.Arc)
	arcs1[2] = &api.Arc{
		Aid:   2,
		State: 0,
		Author: api.Author{
			Mid: 1,
		},
		Stat: api.Stat{
			View: 2,
			Coin: 666666,
		},
	}
	arcs1[1] = &api.Arc{
		Aid:   1,
		State: 0,
		Author: api.Author{
			Mid: 1,
		},
		Stat: api.Stat{
			View: 1,
			Coin: 666666,
		},
	}
	arcs1[3] = &api.Arc{
		Aid:   3,
		State: 0,
		Author: api.Author{
			Mid: 1,
		},
		Stat: api.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[4] = &api.Arc{
		Aid:   4,
		State: 0,
		Author: api.Author{
			Mid: 1,
		},
		Stat: api.Stat{
			View: 4,
			Coin: 666666,
		},
	}
	return arcs1
}

func TestRank(t *testing.T) {
	Convey("test rank success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockRankDao := rank.NewMockDao(mockCtl)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockArc := api.NewMockArchiveClient(mockCtl)

		defer mockCtl.Finish()
		infoMap1 := testGetAccountInfo2()

		mockRankDao.EXPECT().GetRank(Any(), Any()).Return([]*rankmdl.Redis{{Mid: 1, Rank: 1, Score: 1, Aids: []int64{3, 4}}, {Mid: 2, Rank: 2, Score: 0, Aids: []int64{1, 2}}}, nil)
		arcMap := testgetArcs2()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&api.ArcsReply{Arcs: arcMap}, nil)
		mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
		s.account.AccClient = mockAcc
		client.ArchiveClient = mockArc
		s.rank = mockRankDao
		expected := testRankExpected()
		res, err := s.Rank(context.Background(), 1)
		So(err, ShouldBeNil)
		So(res, ShouldResemble, expected)

	}))
}

func testRankExpected() *remixmdl.RankReply {
	rankBatch := make([]*remixmdl.RankMember, 0)
	rankBatch = append(rankBatch,
		&remixmdl.RankMember{
			Account: &remixmdl.Account{
				Mid:  1,
				Name: "1",
			},
			Score: 1,
			Videos: []*remixmdl.Video{
				{

					Aid:      3,
					TypeName: "",
					Title:    "",
					Desc:     "",
					Duration: 0,
					Pic:      "",
					View:     3,
				},
				{
					Aid:      4,
					TypeName: "",
					Title:    "",
					Desc:     "",
					Duration: 0,
					Pic:      "",
					View:     4,
				},
			},
		},
		&remixmdl.RankMember{
			Account: &remixmdl.Account{
				Mid:  2,
				Name: "2",
			},
			Score: 0,
			Videos: []*remixmdl.Video{
				{

					Aid:      1,
					TypeName: "",
					Title:    "",
					Desc:     "",
					Duration: 0,
					Pic:      "",
					View:     1,
				},
				{
					Aid:      2,
					TypeName: "",
					Title:    "",
					Desc:     "",
					Duration: 0,
					Pic:      "",
					View:     2,
				},
			},
		},
	)
	return &remixmdl.RankReply{
		Rank: rankBatch,
	}
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
	return account
}
