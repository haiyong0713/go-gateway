package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model/archive"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_ArcWithStat(t *testing.T) {
	Convey("ArcWithStat", t, func() {
		arc, err := s.ArcWithStat(context.TODO(), &api.ArcRequest{Aid: 800007088})
		So(err, ShouldBeNil)
		Printf("%+v\n", arc)
		bs, _ := json.Marshal(arc)
		Printf("%s", bs)
	})
}

func Test_ArchiveWithPlayer(t *testing.T) {
	Convey("ArchiveWithPlayer", t, func() {
		arcs, err := s.ArchivesWithPlayer(context.TODO(), &archive.ArgPlayer{
			Aids:     []int64{10111001},
			Qn:       32,
			Platform: "android",
			RealIP:   "121.31.246.238",
			Fnver:    0,
			Fnval:    16,
			Session:  "",
		}, true)
		So(err, ShouldBeNil)
		Printf("%+v\n", arcs[10111001])
	})
}

func Test_Archives3(t *testing.T) {
	Convey("Archives3", t, func() {
		as, err := s.Archives3(context.TODO(), []int64{10112614}, 12, "iphone", "phone")
		So(err, ShouldBeNil)
		for _, a := range as {
			Printf("%+v\n\n", a)
			bs, _ := json.Marshal(a)
			Printf("%s\n\n", bs)
		}
	})
}

func Test_ArcsWithPlayurl(t *testing.T) {
	Convey("ArcsWithPlayurl", t, func() {
		arcs, err := s.ArcsWithPlayurl(context.TODO(), &api.ArcsWithPlayurlRequest{
			Aids:     []int64{10111001},
			Qn:       32,
			Platform: "android",
			Ip:       "121.31.246.238",
			Fnver:    0,
			Fnval:    16,
			Session:  "",
		})
		So(err, ShouldBeNil)
		Printf("%+v\n", arcs[10111001])
	})
}

func TestService_ArcsPlayerSvr(t *testing.T) {
	//正常返回 aid + cid（多个）+ 不返回秒开数据
	video1 := api.PlayAv{}
	video1.Aid = 880072935
	video1.PlayVideos = append(video1.PlayVideos, &api.PlayVideo{
		Cid: 10282899,
	}, &api.PlayVideo{
		Cid: 10282898,
	})
	//aid + cid
	video2 := api.PlayAv{}
	video2.Aid = 10318733
	video2.PlayVideos = append(video2.PlayVideos, &api.PlayVideo{
		Cid: 10211337, //非首p
	}, &api.PlayVideo{
		Cid: 10211338, //正好是首p
	})
	//aid -没有cid 要返回秒开
	video3 := api.PlayAv{}
	video3.Aid = 400103150
	//aid -不要秒开
	video4 := api.PlayAv{}
	video4.Aid = 880021381
	video4.NoPlayer = true //不要秒开
	//aid -没有cid 不返回秒开，需要返回extra数据
	video5 := api.PlayAv{}
	video5.Aid = 880084745
	//aid pgc
	video6 := api.PlayAv{}
	video6.Aid = 440105542
	//aid pgc
	video7 := api.PlayAv{}
	video7.Aid = 10113230

	req := api.ArcsPlayerRequest{
		BatchPlayArg: &api.BatchPlayArg{
			Build:  61900105,
			Device: "phone",
			//Qn:      32,
			MobiApp: "android",
			//Fnval:   16,
			Ip:             "10.23.16.20",
			ShowPgcPlayurl: true,
		},
	}
	//req.PlayAvs = append(req.PlayAvs, &video1, &video2, &video3, &video4, &video5, &video6, &video7)
	req.PlayAvs = append(req.PlayAvs, &video1)
	Convey("ArcsPlayerSvr", t, func() {
		a, e := s.ArcsPlayerSvr(context.TODO(), &req)
		fmt.Println(a, e)
		//So(err, ShouldBeNil)
		//Printf("%+v\n", arcs[10111001])
	})
}

func TestService_Views3(t *testing.T) {
	a, _ := s.View3(context.TODO(), 200066015)
	fmt.Println(a)
}

func Test_Creators(t *testing.T) {
	Convey("creators", t, func() {
		arcs, _ := s.Creators(context.TODO(), []int64{10111001})
		Printf("%+v\n", arcs[10111001])
	})
}

func TestService_ArcRedirectPolicyAddSrv(t *testing.T) {
	req := &api.ArcRedirectPolicyAddRequest{
		Aid:            560115889,
		RedirectType:   1,
		RedirectTarget: "https://www.bilibili.com/bangumi/play/ep140371?theme=movie",
		PolicyType:     1,
		PolicyId:       13123,
	}
	Convey("redirect_policy_add", t, func() {
		err := s.ArcRedirectPolicyAddSrv(context.TODO(), req)
		Printf("%+v\n", err)
	})
}

func TestService_ArcsRedirectPolicy(t *testing.T) {
	req := []int64{560115889}
	Convey("redirect_policy_add", t, func() {
		data, err := s.ArcsRedirectPolicy(context.TODO(), req)
		Printf("%+v\n%+v", data, err)
	})
}
