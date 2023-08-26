package service

import (
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"testing"

	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRemixEveryHour(t *testing.T) {
	Convey("test RemixEveryHour  success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		// mockRelation := relationapi.NewMockRelationClient(mockCtl)
		// mockAcc := accapi.NewMockAccountClient(mockCtl)
		defer mockCtl.Finish()
		// mock
		// arc data mock
		arcs1 := testremixgetArcs1()
		arcs2 := testremixgetArcs2()
		arcs3 := testremixgetArcs3()
		arcs4 := testremixgetArcs4()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs1}, nil)
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)
		// // // relation mock
		// statMap1 := testGetRelationMap1()
		// statMap2 := testGetRelationMap2()
		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap1}, nil)
		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)
		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)
		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)
		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)
		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)

		// // // account mock
		// infoMap1 := testGetAccountInfo()
		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
		s.arcClient = mockArc
		// s.relationClient = mockRelation
		// s.accClient = mockAcc
		s.RemixEveryHour()
	}))
}

func testremixgetArcs1() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[1] = &arcapi.Arc{
		Aid:   1,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 999999,
			Coin: 2,
			Like: 2222,
		},
	}
	arcs1[2] = &arcapi.Arc{
		Aid:   2,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 99999,
			Coin: 66666,
		},
	}
	arcs1[3] = &arcapi.Arc{
		Aid:   3,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 5,
		},
	}
	arcs1[4] = &arcapi.Arc{
		Aid:   4,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 9999,
			Coin: 7,
		},
	}
	arcs1[5] = &arcapi.Arc{
		Aid:   5,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 9999,
			Coin: 7,
		},
	}

	arcs1[6] = &arcapi.Arc{
		Aid:   6,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
		Stat: arcapi.Stat{
			View: 6,
			Coin: 7,
		},
	}
	arcs1[7] = &arcapi.Arc{
		Aid:   7,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
		Stat: arcapi.Stat{
			View: 6,
			Coin: 66666,
		},
	}
	arcs1[12] = &arcapi.Arc{
		Aid:   12,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 9999,
			Coin: 999,
		},
	}
	arcs1[13] = &arcapi.Arc{
		Aid:   5,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 9999,
			Coin: 999,
		},
	}
	arcs1[14] = &arcapi.Arc{
		Aid:   5,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 9999,
			Coin: 99,
		},
	}
	arcs1[15] = &arcapi.Arc{
		Aid:   5,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 9999,
			Coin: 7,
		},
	}
	arcs1[16] = &arcapi.Arc{
		Aid:   5,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 9999,
			Coin: 7,
		},
	}

	return arcs1
}

func testremixgetArcs2() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[8] = &arcapi.Arc{
		Aid:   8,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 1,
		},
	}
	arcs1[9] = &arcapi.Arc{
		Aid:   9,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 1,
		},
	}
	arcs1[10] = &arcapi.Arc{
		Aid:   10,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 0,
			Coin: 1,
		},
	}
	arcs1[11] = &arcapi.Arc{
		Aid:   11,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 1,
		},
	}
	return arcs1
}

func testremixgetArcs3() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[12] = &arcapi.Arc{
		Aid:   12,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 2,
		},
	}
	arcs1[13] = &arcapi.Arc{
		Aid:   13,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 1,
			Coin: 2,
		},
	}
	arcs1[14] = &arcapi.Arc{
		Aid:   14,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[15] = &arcapi.Arc{
		Aid:   15,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[16] = &arcapi.Arc{
		Aid:   16,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 1,
			Coin: 2,
		},
	}
	arcs1[17] = &arcapi.Arc{
		Aid:   17,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
		},
	}
	arcs1[18] = &arcapi.Arc{
		Aid:   18,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
		Stat: arcapi.Stat{
			View: 3,
			Coin: 666666,
			Like: 1222,
		},
	}
	return arcs1
}
func testremixgetArcs4() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[19] = &arcapi.Arc{
		Aid:   19,
		State: 0,
		Author: arcapi.Author{
			Mid: 1,
		},
		Stat: arcapi.Stat{
			View: 1,
			Coin: 8832,

			Reply:   1231,
			Danmaku: 993,
			Fav:     9832,
			Like:    2,
		},
	}
	arcs1[20] = &arcapi.Arc{
		Aid:   20,
		State: 0,
		Author: arcapi.Author{
			Mid: 2,
		},
		Stat: arcapi.Stat{
			View: 2,
			Coin: 8832,

			Reply:   1231,
			Danmaku: 993,
			Fav:     9832,
			Like:    1,
		},
	}
	arcs1[21] = &arcapi.Arc{
		Aid:   21,
		State: 0,
		Author: arcapi.Author{
			Mid: 3,
		},
		Stat: arcapi.Stat{
			View: 6,
			Coin: 8832,

			Reply:   1231,
			Danmaku: 993,
			Fav:     9832,
			Like:    1222,
		},
	}
	arcs1[22] = &arcapi.Arc{
		Aid:   22,
		State: 0,
		Author: arcapi.Author{
			Mid: 4,
		},
		Stat: arcapi.Stat{
			View: 666,
			Coin: 8832,

			Reply:   1231,
			Danmaku: 993,
			Fav:     9832,
			Like:    2333,
		},
	}
	return arcs1
}

// func TestRemixEveryHourCount(t *testing.T) {
// 	Convey("test RemixEveryHour  count score", t, WithService(func(s *Service) {
// 		mockCtl := NewController(t)
// 		mockArc := arcapi.NewMockArchiveClient(mockCtl)
// 		// mockRelation := relationapi.NewMockRelationClient(mockCtl)
// 		// mockAcc := accapi.NewMockAccountClient(mockCtl)
// 		defer mockCtl.Finish()
// 		// mock
// 		// arc data mock
// 		arcs1 := testremixgetArcs5()
// 		arcs2 := testremixgetArcs6()
// 		arcs3 := testremixgetArcs7()
// 		arcs4 := testremixgetArcs8()
// 		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs1}, nil)
// 		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
// 		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
// 		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)
// 		// // // relation mock
// 		// statMap1 := testGetRelationMap1()
// 		// statMap2 := testGetRelationMap2()
// 		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap1}, nil)
// 		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)
// 		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)
// 		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)
// 		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)
// 		// mockRelation.EXPECT().Stats(Any(), Any()).Return(&relationapi.StatsReply{StatReplyMap: statMap2}, nil)

// 		// // // account mock
// 		// infoMap1 := testGetAccountInfo()
// 		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
// 		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
// 		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
// 		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
// 		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
// 		// mockAcc.EXPECT().Infos3(Any(), Any()).Return(&accapi.InfosReply{Infos: infoMap1}, nil)
// 		s.arcClient = mockArc
// 		// s.relationClient = mockRelation
// 		// s.accClient = mockAcc

// 		s.RemixEveryHour()
// 	}))
// }

func testremixgetArcs5() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[1] = &arcapi.Arc{
		Aid:    1,
		State:  0,
		Videos: 1,
		Author: arcapi.Author{
			Mid: 1111111,
		},
		Stat: arcapi.Stat{
			View:  1000066,
			Reply: 1750,
			Fav:   29956,
			Coin:  89756,
			Like:  3333333,
		},
	}

	// arcs1[6] = &arcapi.Arc{
	// 	Aid:    6,
	// 	State:  0,
	// 	Videos: 1,

	// 	Author: arcapi.Author{
	// 		Mid: 444444,
	// 	},
	// 	Stat: arcapi.Stat{
	// 		View:  715154,
	// 		Reply: 1836,
	// 		Fav:   17307,
	// 		Coin:  60075,
	// 	},
	// }

	// arcs1[10] = &arcapi.Arc{
	// 	Aid:    10,
	// 	Videos: 1,
	// 	State:  0,
	// 	Author: arcapi.Author{
	// 		Mid: 444444,
	// 	},
	// 	Stat: arcapi.Stat{
	// 		View:  311434,
	// 		Reply: 981,
	// 		Fav:   4368,
	// 		Coin:  17320,
	// 	},
	// }

	return arcs1
}

func testremixgetArcs6() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)
	arcs1[3] = &arcapi.Arc{
		Aid:    3,
		Videos: 1,
		State:  0,
		Author: arcapi.Author{
			Mid: 222222,
		},
		Stat: arcapi.Stat{
			View:  2061349,
			Reply: 2506,
			Fav:   74066,
			Coin:  92323,
		},
	}
	arcs1[2] = &arcapi.Arc{
		Aid:    2,
		Videos: 1,
		State:  0,
		Author: arcapi.Author{
			Mid: 1111111,
		},
		Stat: arcapi.Stat{
			View:  7956518,
			Reply: 14222,
			Fav:   315000,
			Coin:  543000,
		},
	}
	arcs1[7] = &arcapi.Arc{
		Aid:    7,
		Videos: 1,
		State:  0,
		Author: arcapi.Author{
			Mid: 444444,
		},
		Stat: arcapi.Stat{
			View:  182647,
			Reply: 488,
			Fav:   1921,
			Coin:  2195,
		},
	}
	arcs1[5] = &arcapi.Arc{
		Aid:    5,
		Videos: 1,
		State:  0,
		Author: arcapi.Author{
			Mid: 333333,
		},
		Stat: arcapi.Stat{
			View:  1473806,
			Reply: 787,
			Fav:   32435,
			Coin:  20987,
		},
	}

	return arcs1
}

func testremixgetArcs7() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)

	arcs1[8] = &arcapi.Arc{
		Aid:    8,
		Videos: 1,
		State:  0,
		Author: arcapi.Author{
			Mid: 444444,
		},
		Stat: arcapi.Stat{
			View:  565097,
			Reply: 1570,
			Fav:   14211,
			Coin:  54103,
		},
	}
	arcs1[9] = &arcapi.Arc{
		Aid:    9,
		Videos: 1,
		State:  0,
		Author: arcapi.Author{
			Mid: 444444,
		},
		Stat: arcapi.Stat{
			View:  743593,
			Reply: 1840,
			Fav:   15203,
			Coin:  68975,
		},
	}

	return arcs1
}

func testremixgetArcs8() map[int64]*arcapi.Arc {
	arcs1 := make(map[int64]*arcapi.Arc)

	arcs1[4] = &arcapi.Arc{
		Aid:    4,
		Videos: 1,
		State:  0,
		Author: arcapi.Author{
			Mid: 333333,
		},
		Stat: arcapi.Stat{
			View:  3121719,
			Reply: 1425,
			Fav:   105077,
			Coin:  91988,
		},
	}

	return arcs1
}
