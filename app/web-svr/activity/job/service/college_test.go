package service

import (
	arcapi "go-gateway/app/app-svr/archive/service/api"

	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"

	"testing"

	tagnewapi "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCollegeRank(t *testing.T) {
	Convey("test TestCollegeRank success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockArc := arcapi.NewMockArchiveClient(mockCtl)
		mockActPlat := actplatapi.NewMockActPlatClient(mockCtl)
		mockTag := tagnewapi.NewMockTagRPCClient(mockCtl)
		defer mockCtl.Finish()
		// mock
		// arc data mock

		mockTag.EXPECT().RidsByTag(Any(), Any()).Return(&tagnewapi.RidsByTagReply{Rids: []int64{1, 2, 3}, Hasmore: true}, nil)
		mockTag.EXPECT().RidsByTag(Any(), Any()).Return(&tagnewapi.RidsByTagReply{Rids: []int64{4, 5, 6}, Hasmore: false}, nil)
		mockTag.EXPECT().RidsByTag(Any(), Any()).Return(&tagnewapi.RidsByTagReply{Rids: []int64{4, 5, 6}, Hasmore: false}, nil)
		mockTag.EXPECT().RidsByTag(Any(), Any()).Return(&tagnewapi.RidsByTagReply{Rids: []int64{4, 5, 6}, Hasmore: false}, nil)
		arcs1 := testgetArchive()
		arcs2 := testgetArchive2()
		arcs3 := testgetArchive3()
		arcs4 := testgetArchive4()
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs1}, nil)
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs2}, nil)
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs3}, nil)
		mockArc.EXPECT().Arcs(Any(), Any()).Return(&arcapi.ArcsReply{Arcs: arcs4}, nil)
		// account mock
		mockActPlat.EXPECT().GetFormulaResult(Any(), Any()).Return(&actplatapi.GetFormulaResultResp{Result: 50}, nil)
		mockActPlat.EXPECT().GetFormulaResult(Any(), Any()).Return(&actplatapi.GetFormulaResultResp{Result: 1000}, nil)
		mockActPlat.EXPECT().GetFormulaResult(Any(), Any()).Return(&actplatapi.GetFormulaResultResp{Result: 222}, nil)

		s.arcClient = mockArc
		s.actplatClient = mockActPlat
		s.tagNewClient = mockTag

		s.CollegeRank()
	}))
}

func testgetArchive2() map[int64]*arcapi.Arc {
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

func testgetArchive() map[int64]*arcapi.Arc {
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

func testgetArchive3() map[int64]*arcapi.Arc {
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
func testgetArchive4() map[int64]*arcapi.Arc {
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
