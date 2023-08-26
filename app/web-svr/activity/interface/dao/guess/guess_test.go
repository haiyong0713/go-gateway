package guess

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/interface/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_AddMatchGuess(t *testing.T) {
	var (
		c       = context.Background()
		groups  []*api.GuessGroup
		details []*api.GuessDetailAdd
	)
	details = append(details, &api.GuessDetailAdd{Option: "aaa", TotalStake: 10})
	group := &api.GuessGroup{
		Title:     "BLG VS IG第一局",
		DetailAdd: details,
	}
	groups = append(groups, group)
	req := &api.GuessAddReq{Business: 1, Oid: 1, MaxStake: 10, StakeType: 1, Groups: groups}
	convey.Convey("AddMatchGuess", t, func(ctx convey.C) {
		err := d.AddMatchGuess(c, req)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			println(err)
		})
	})
}

func TestDao_UserAddGuess(t *testing.T) {
	var (
		c        = context.Background()
		business = int64(1)
		mainID   = int64(6)
	)
	req := &api.GuessUserAddReq{Mid: 100, MainID: mainID, DetailID: 6, StakeType: 1, Stake: 5}
	convey.Convey("AddMatchGuess", t, func(ctx convey.C) {
		_, err := d.UserAddGuess(c, business, mainID, req)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			println(err)
		})
	})
}
