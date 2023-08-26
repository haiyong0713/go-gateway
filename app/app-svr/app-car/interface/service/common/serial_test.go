package common

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	comm "go-gateway/app/app-svr/app-car/interface/model/common"
)

func Test_serialInfoIntegrate(t *testing.T) {
	convey.Convey("Test_serialInfoIntegrate", t, WithService(func(s *Service) {
		var (
			c   = context.Background()
			req = comm.SerialInfosReq{
				FmCommonIds: []int64{333, 444}, // FM合集ID
				VideoIds:    []int64{111, 222}, // 视频合集ID
			}
		)
		resp, err := s.serialInfoIntegrate(c, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp, convey.ShouldNotBeNil)
		bytes, _ := json.Marshal(resp)
		_, _ = convey.Printf("res:%+v", string(bytes))
	}))
}

func Test_serialOidsByPageIntegrate(t *testing.T) {
	convey.Convey("Test_serialOidsByPageIntegrate", t, WithService(func(s *Service) {
		var (
			c         = context.Background()
			reqNoPage = comm.SerialArcsReq{ // 初次进入
				FmCommon: []*comm.SerialArcReq{
					{
						SerialId:      333,
						SerialPageReq: comm.SerialPageReq{Ps: 2},
					},
					{
						SerialId: 444,
					},
				},
				Video: []*comm.SerialArcReq{
					{
						SerialId:      111,
						SerialPageReq: comm.SerialPageReq{Ps: 2},
					},
					{
						SerialId: 222,
					},
				},
			}
			reqPageNext = comm.SerialArcsReq{ // 往下翻页，大小2
				FmCommon: []*comm.SerialArcReq{
					{
						SerialId:      333,
						SerialPageReq: comm.SerialPageReq{PageNext: &comm.SerialPageInfo{Oid: 240029469}, Ps: 2}, // 第3个稿件，共5个；预期到底了，has_more false
					},
					{
						SerialId:      444,
						SerialPageReq: comm.SerialPageReq{PageNext: &comm.SerialPageInfo{Oid: 960044310, Ps: 2}}, // 第3个稿件，共6个；预期没到底，has_more true
					},
				},
				Video: []*comm.SerialArcReq{
					{
						SerialId:      111,
						SerialPageReq: comm.SerialPageReq{PageNext: &comm.SerialPageInfo{Oid: 280087205}, Ps: 2}, // 第7个稿件，共12个；预期没到底，has_more true
					},
					{
						SerialId:      222,
						SerialPageReq: comm.SerialPageReq{PageNext: &comm.SerialPageInfo{Oid: 600063736, Ps: 2}}, // 第2个稿件，共3个；预期到底了，has_more false
					},
				},
			}
			reqPagePre = comm.SerialArcsReq{ // 往上翻页，大小2
				FmCommon: []*comm.SerialArcReq{
					{
						SerialId:      333,
						SerialPageReq: comm.SerialPageReq{PagePre: &comm.SerialPageInfo{Oid: 240029469}, Ps: 2}, // 第3个稿件，共5个；预期到顶了，has_more false
					},
					{
						SerialId:      444,
						SerialPageReq: comm.SerialPageReq{PagePre: &comm.SerialPageInfo{Oid: 360080117, Ps: 2}}, // 第4个稿件，共6个；预期没到顶，has_more true
					},
				},
				Video: []*comm.SerialArcReq{
					{
						SerialId:      111,
						SerialPageReq: comm.SerialPageReq{PagePre: &comm.SerialPageInfo{Oid: 280087205}, Ps: 2}, // 第7个稿件，共12个；预期没到顶，has_more true
					},
					{
						SerialId:      222,
						SerialPageReq: comm.SerialPageReq{PagePre: &comm.SerialPageInfo{Oid: 600063736, Ps: 2}}, // 第2个稿件，共3个；预期到顶了，has_more false
					},
				},
			}
			reqPageAll = comm.SerialArcsReq{ // 上下都要（冷启、携带历史记录时选此）
				FmCommon: []*comm.SerialArcReq{
					{
						SerialId:      333,
						SerialPageReq: comm.SerialPageReq{PagePre: &comm.SerialPageInfo{Oid: 240029469}, PageNext: &comm.SerialPageInfo{Oid: 240029469, WithCurrent: true}, Ps: 2}, // 第3个稿件，共5个；预期第1-5个稿件
					},
					{
						SerialId:      444,
						SerialPageReq: comm.SerialPageReq{PagePre: &comm.SerialPageInfo{Oid: 360080117}, PageNext: &comm.SerialPageInfo{Oid: 360080117, WithCurrent: true}, Ps: 2}, // 第4个稿件，共6个；预期第2-5个稿件
					},
				},
				Video: []*comm.SerialArcReq{
					{
						SerialId:      111,
						SerialPageReq: comm.SerialPageReq{PagePre: &comm.SerialPageInfo{Oid: 280087205}, PageNext: &comm.SerialPageInfo{Oid: 280087205, WithCurrent: true}, Ps: 2}, // 第7个稿件，共12个；预期第5-8个稿件
					},
					{
						SerialId:      222,
						SerialPageReq: comm.SerialPageReq{PagePre: &comm.SerialPageInfo{Oid: 600063736, Ps: 2}, PageNext: &comm.SerialPageInfo{Oid: 600063736, WithCurrent: true}, Ps: 2}, // 第2个稿件，共3个；预期第1-3个稿件
					},
				},
			}
		)
		// case1 无翻页
		resp1, err := s.serialOidsByPageIntegrate(c, reqNoPage)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp1, convey.ShouldNotBeNil)
		bytes, _ := json.Marshal(resp1)
		_, _ = convey.Printf("case1 reqNoPage res:%+v\n", string(bytes))

		// case2 下一页
		resp2, err := s.serialOidsByPageIntegrate(c, reqPageNext)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp2, convey.ShouldNotBeNil)
		bytes, _ = json.Marshal(resp2)
		_, _ = convey.Printf("case2 reqPageNext res:%+v\n", string(bytes))

		// case3 上一页
		resp3, err := s.serialOidsByPageIntegrate(c, reqPagePre)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp3, convey.ShouldNotBeNil)
		bytes, _ = json.Marshal(resp3)
		_, _ = convey.Printf("case3 reqPagePre res:%+v\n", string(bytes))

		// case4 上下都要（冷启、携带历史记录时选此）
		resp4, err := s.serialOidsByPageIntegrate(c, reqPageAll)
		convey.So(err, convey.ShouldBeNil)
		convey.So(resp4, convey.ShouldNotBeNil)
		bytes, _ = json.Marshal(resp4)
		_, _ = convey.Printf("case4 reqPageAll res:%+v", string(bytes))
	}))
}
