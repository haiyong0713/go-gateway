package like

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/activity/interface/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestService_GuessAdd(t *testing.T) {
	Convey("GuessAdd", t, WithService(func(svr *Service) {
		var (
			c                = context.Background()
			groups           []*api.GuessGroup
			detail1, detail2 []*api.GuessDetailAdd
		)
		Oid := int64(37)
		detail1 = append(detail1, &api.GuessDetailAdd{Option: "第一组选项1内容", TotalStake: 10})
		detail1 = append(detail1, &api.GuessDetailAdd{Option: "第一组选项2内容", TotalStake: 10})
		detail2 = append(detail2, &api.GuessDetailAdd{Option: "第二组选项1内容", TotalStake: 10})
		detail2 = append(detail2, &api.GuessDetailAdd{Option: "第二组选项2内容", TotalStake: 10})
		detail2 = append(detail2, &api.GuessDetailAdd{Option: "第二组选项3内容", TotalStake: 10})
		groups = append(groups, &api.GuessGroup{Title: "第一组标题", DetailAdd: detail1})
		groups = append(groups, &api.GuessGroup{Title: "第二组标题", DetailAdd: detail2})
		_, err := svr.GuessAdd(c, &api.GuessAddReq{Business: int64(api.GuessBusiness_esportsType), Oid: Oid, MaxStake: 10, StakeType: int64(api.StakeType_coinType), Stime: 1561621307, Etime: 1564213306, Groups: groups})
		So(err, ShouldBeNil)
	}))
}

func TestService_GuessEdit(t *testing.T) {
	Convey("GuessEdit", t, WithService(func(svr *Service) {
		var (
			c = context.Background()
		)
		business := int64(1)
		oid := int64(456)
		stime := int64(2)
		etime := int64(7)
		rs, err := svr.GuessEdit(c, &api.GuessEditReq{Business: business, Oid: oid, Stime: stime, Etime: etime})
		So(err, ShouldBeNil)
		println(rs)
	}))
}

func TestService_GuessList(t *testing.T) {
	Convey("GuessList", t, WithService(func(svr *Service) {
		var (
			c = context.Background()
		)
		Oid := int64(37)
		rs, err := svr.GuessList(c, &api.GuessListReq{Business: int64(api.GuessBusiness_esportsType), Oid: Oid})
		So(err, ShouldBeNil)
		So(rs, ShouldNotBeNil)
	}))
}

func TestService_UserAddGuess(t *testing.T) {
	Convey("UserAddGuess", t, WithService(func(svr *Service) {
		var (
			c = context.Background()
		)
		mid := int64(100)
		mainID := int64(6)
		detailID := int64(13)
		stake := int64(7)
		rs, err := svr.UserAddGuess(c, &api.GuessUserAddReq{Mid: mid, MainID: mainID, DetailID: detailID, StakeType: int64(api.StakeType_coinType), Stake: stake})
		So(err, ShouldBeNil)
		So(rs, ShouldNotBeNil)
	}))
}

func TestService_UserGuessList(t *testing.T) {
	Convey("UserAddGuess", t, WithService(func(svr *Service) {
		var (
			c = context.Background()
		)
		req := &api.UserGuessListReq{
			Mid:      10000,
			Business: 1,
			Pn:       1,
			Ps:       10,
		}
		rs, _, err := svr.UserGuessList(c, req)
		So(err, ShouldBeNil)
		So(rs, ShouldNotBeNil)
		bs, _ := json.Marshal(rs)
		fmt.Println(string(bs))
	}))
}
