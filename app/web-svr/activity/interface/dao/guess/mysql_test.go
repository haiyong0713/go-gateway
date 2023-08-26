package guess

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_RawUserGuessList(t *testing.T) {
	convey.Convey("RawUserGuessList", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			mid = int64(10000)
			b   = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawUserGuessList(c, mid, b)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
				bs, _ := json.Marshal(res)
				fmt.Println(string(bs))
			})
		})
	})
}

func TestDao_AddGuess(t *testing.T) {
	convey.Convey("AddGuess", t, func(convCtx convey.C) {
		var (
			tx        *sql.Tx
			c         = context.Background()
			mainID    = int64(1)
			buesiness = int64(1)
			err       error
		)
		if tx, err = d.db.Begin(c); err != nil {
			return
		}
		p := &api.GuessUserAddReq{}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.AddGuess(c, tx, mainID, p, buesiness)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_AddMainGuess(t *testing.T) {
	convey.Convey("AddGuess", t, func(convCtx convey.C) {
		var (
			tx        *sql.Tx
			c         = context.Background()
			business  = int64(1)
			oid       = int64(37)
			maxStake  = int64(10)
			stakeType = int64(1)
			title     = "竞猜标题1"
			stime     = int64(1562135632)
			etime     = int64(1572762832)
			err       error
		)
		if tx, err = d.db.Begin(c); err != nil {
			return
		}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.AddMainGuess(c, tx, business, oid, maxStake, stakeType, title, stime, etime, 1)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_UpGuess(t *testing.T) {
	convey.Convey("UpGuess", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p := &api.GuessEditReq{}
			res, err := d.UpGuess(c, p)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_BatchAddDetail(t *testing.T) {
	convey.Convey("BatchAddDetail", t, func(convCtx convey.C) {
		var (
			tx        *sql.Tx
			c         = context.Background()
			mainID    = int64(1)
			err       error
			detailAdd []*api.GuessDetailAdd
		)
		if tx, err = d.db.Begin(c); err != nil {
			return
		}
		detailAdd = append(detailAdd, &api.GuessDetailAdd{Option: "选项1", TotalStake: 10})
		detailAdd = append(detailAdd, &api.GuessDetailAdd{Option: "选项2", TotalStake: 10})
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.BatchAddDetail(c, tx, mainID, detailAdd)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_DelGroup(t *testing.T) {
	convey.Convey("DelGroup", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			mainID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.DelGroup(c, mainID)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_UpGuessResult(t *testing.T) {
	convey.Convey("UpGuessResult", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			mainID   = int64(1)
			detailID = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UpGuessResult(c, mainID, detailID)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_RawGuessMain(t *testing.T) {
	convey.Convey("RawGuessMDs", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			mainID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawGuessMain(c, mainID)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_RawUserStat(t *testing.T) {
	convey.Convey("RawUserLog", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			mid       = int64(10000)
			stakeType = int64(1)
			business  = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawUserStat(c, mid, stakeType, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_UserStatUp(t *testing.T) {
	convey.Convey("UserStatUp", t, func(convCtx convey.C) {
		var (
			tx       *sql.Tx
			c        = context.Background()
			business = int64(1)
			err      error
		)
		if tx, err = d.db.Begin(c); err != nil {
			return
		}
		defer tx.Commit()
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p := &api.GuessUserAddReq{Mid: 1, MainID: 1, DetailID: 1, StakeType: 1, Stake: 5}
			res, err := d.UserStatUp(c, tx, business, p)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestDao_UpMainCount(t *testing.T) {
	convey.Convey("UpMainCount", t, func(convCtx convey.C) {
		var (
			tx  *sql.Tx
			c   = context.Background()
			id  = int64(1)
			err error
		)
		if tx, err = d.db.Begin(c); err != nil {
			return
		}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UpMainCount(c, tx, id)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_UpDetailTotal(t *testing.T) {
	convey.Convey("UpDetailTotal", t, func(convCtx convey.C) {
		var (
			tx    *sql.Tx
			c     = context.Background()
			stake = int64(5)
			id    = int64(1)
			err   error
		)
		if tx, err = d.db.Begin(c); err != nil {
			return
		}
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.UpDetailTotal(c, tx, stake, id)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_HaveGuess(t *testing.T) {
	convey.Convey("HaveGuess", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			mid      = int64(1)
			detailID = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.HaveGuess(c, mid, detailID)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDao_RawUserGuess(t *testing.T) {
	convey.Convey("RawUserGuess", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			mid     = int64(1)
			mainIDs = []int64{1, 3}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawUserGuess(c, mainIDs, mid)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDao_RawOidsMIDs(t *testing.T) {
	convey.Convey("RawOidsMIDs", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			oids     = []int64{83}
			business = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawOidsMIDs(c, oids, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDao_RawOidMIDs(t *testing.T) {
	convey.Convey("RawOidMIDs", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			oid      = int64(83)
			business = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawOidMIDs(c, oid, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDao_RawMDsResult(t *testing.T) {
	convey.Convey("RawMDsResult", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			ids      = []int64{1}
			business = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawMDsResult(c, ids, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestDao_RawMDResult(t *testing.T) {
	convey.Convey("RawMDResult", t, func(convCtx convey.C) {
		var (
			c        = context.Background()
			id       = int64(1)
			business = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawMDResult(c, id, business)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.So(res, convey.ShouldNotBeNil)
			})
		})
	})
}
