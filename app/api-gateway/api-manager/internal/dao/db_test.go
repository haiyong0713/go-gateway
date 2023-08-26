package dao

import (
	"testing"

	pb "go-gateway/app/api-gateway/api-manager/api"
	"go-gateway/app/api-gateway/api-manager/internal/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoTidb(t *testing.T) {
	Convey("TestDaoTidb", t, func() {
		apis := &model.ApiRawInfo{
			ID:          1,
			DiscoveryID: "coin",
			Protocol:    0,
			ApiService:  "Coin",
			ApiPath:     "/main/coin",
			ApiHeader:   "",
			ApiParams:   "",
			FormBody:    "",
			JsonBody:    "{1}",
			Output:      "{2}",
			State:       0,
			Description: "测试",
		}
		err := d.AddApi(ctx, apis)
		So(err, ShouldBeNil)
		res, err := d.GetHttpApis(ctx)
		So(err, ShouldBeNil)
		So(len(res), ShouldEqual, 0)
		res, err = d.GetGrpcApis(ctx, "coin")
		So(err, ShouldBeNil)
		So(res[0].ApiPath, ShouldEqual, "/main/coin")
		resH, err := d.GetHttpApisByPath(ctx, []string{"/main/coin"})
		So(err, ShouldBeNil)
		So(resH, ShouldResemble, map[string]*pb.ApiInfo{"/main/coin": {
			Input:  "{1}",
			Output: "{2}",
		}})
		resS, err := d.GetServiceName(ctx, []string{"coin"})
		So(err, ShouldBeNil)
		So(resS, ShouldResemble, map[string][]string{"coin": {"Coin"}})

		protos := &model.ProtoInfo{
			ID:          1,
			FilePath:    "/community/service/coin/api.proto",
			GoPath:      "/community/service/coin",
			DiscoveryID: "community.service.coin",
			Alias:       "community.service.coin.api.proto",
			Package:     "community.service.coin.v1",
			File:        "",
		}

		err = d.AddProto(ctx, protos)
		So(err, ShouldBeNil)
		res2, err := d.GetAllProtos(ctx)
		So(err, ShouldBeNil)
		So(res2[0].GoPath, ShouldEqual, "/community/service/coin")
		res2, err = d.GetProto(ctx, "community.service.coin")
		So(err, ShouldBeNil)
		So(res2[0].GoPath, ShouldEqual, "/community/service/coin")
		resP, err := d.GetProtoByDis(ctx, []string{"community.service.coin"})
		So(err, ShouldBeNil)
		So(resP, ShouldResemble, map[string]*pb.ApiInfo{"community.service.coin": {
			PbAlias: "community.service.coin.api.proto",
			PbPath:  "/community/service/coin",
		}})
	})
}
