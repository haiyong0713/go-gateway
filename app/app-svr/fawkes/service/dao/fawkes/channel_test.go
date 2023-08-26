package fawkes

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	xsql "go-common/library/database/sql"

	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
)

// TestFawkesChannelList test FawkesChannelList.
func TestFawkesChannelList(t *testing.T) {
	convey.Convey("ChannelList", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.ChannelList(c, 1, 10, "y")
			convCtx.Convey("Then err should be nil.chLists should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesTxChannelAdd test FawkesTxChannelAdd.
func TestFawkesTxChannelAdd(t *testing.T) {
	convey.Convey("TxChannelAdd", t, func(convCtx convey.C) {
		var (
			tx, _         = d.BeginTran(context.Background())
			code          = "100"
			name          = "DIO"
			plate         = "mi"
			operator      = "DIO"
			status        = int8(1)
			channelStatus = int8(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxChannelAdd(tx, code, name, plate, operator, status, channelStatus)
			convCtx.Convey("Then err should be nil.r should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.TxChannelAdd(tx, code, name, plate, operator, status, channelStatus)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestFawkesTxChannelToStatic test TxChannelToStatic
func TestFawkesTxChannelToStatic(t *testing.T) {
	convey.Convey("TxChannelToStatic", t, func(convCtx convey.C) {
		var (
			tx, _     = d.BeginTran(context.Background())
			channelID = int64(1)
			operator  = "fd"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxChannelToStatic(tx, channelID, operator)
			convCtx.Convey("Then err should be nil.r should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.TxChannelToStatic(tx, channelID, operator)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestFawkesTxChannelDelete test FawkesTxChannelDelete.
func TestFawkesTxChannelDelete(t *testing.T) {
	convey.Convey("TxChannelDelete", t, func(convCtx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxChannelDelete(tx, 4, "y")
			convCtx.Convey("Then err should be nil.r should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.TxChannelDelete(tx, 4, "y")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestFawkesCheckChannelByCode test FawkesCheckChannelByCode.
func TestFawkesCheckChannelByCode(t *testing.T) {
	convey.Convey("CheckChannelByCode", t, func(convCtx convey.C) {
		var (
			c    = context.Background()
			code = "233"
			name = "fd"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CheckChannelByCode(c, code, name)
			convCtx.Convey("Then err should be nil.count should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesGetAppCountByID test FawkesGetAppCountByID.
func TestFawkesGetAppCountByID(t *testing.T) {
	convey.Convey("GetAppCountByID", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetAppCountByID(c, 233)
			convCtx.Convey("Then err should be nil.count should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesGetChannelIDByCode test FawkesGetChannelIDByCode.
func TestFawkesGetChannelIDByCode(t *testing.T) {
	convey.Convey("GetChannelIDByCode", t, func(convCtx convey.C) {
		var (
			c     = context.Background()
			code  = "233"
			name  = "233"
			plate = "233"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetChannelIDByCode(c, code, name, plate)
			convCtx.Convey("Then err should be nil.count should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesCheckChannelByID test FawkesCheckChannelByID.
func TestFawkesCheckChannelByID(t *testing.T) {
	convey.Convey("CheckChannelByID", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CheckChannelByID(c, 4)
			convCtx.Convey("Then err should be nil.count should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesAppChannelList test FawkesAppChannelList.
func TestFawkesAppChannelList(t *testing.T) {
	convey.Convey("AppChannelList", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.AppChannelList(c, "df", "y", "", "", -1, -1, -1)
			convCtx.Convey("Then err should be nil.chList should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesAppChannelAdd test FawkesAppChannelAdd.
func TestFawkesAppChannelAdd(t *testing.T) {
	convey.Convey("AppChannelAdd", t, func(convCtx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.AppChannelAdd(tx, 1, 0, "df", "y")
			convCtx.Convey("Then err should be nil.r should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.AppChannelAdd(tx, 1, 0, "df", "y")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestFawkesCheckAppChannel test FawkesCheckAppChannel.
func TestFawkesCheckAppChannel(t *testing.T) {
	convey.Convey("CheckAppChannel", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.CheckAppChannel(c, "df", 4)
			convCtx.Convey("Then err should be nil.count should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesAppChannelDelete test FawkesAppChannelDelete.
func TestFawkesAppChannelDelete(t *testing.T) {
	convey.Convey("AppChannelDelete", t, func(convCtx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.AppChannelDelete(tx, "df", 4)
			convCtx.Convey("Then err should be nil.r should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.AppChannelDelete(tx, "df", 4)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestFawkesGetChannelByID test FawkesGetChannelByID.
func TestFawkesGetChannelByID(t *testing.T) {
	convey.Convey("GetChannelByID", t, func(convCtx convey.C) {
		var c = context.Background()
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetChannelByID(c, 4)
			convCtx.Convey("Then err should be nil.chel should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesTxCustomChannelDeleteByID test FawkesTxCustomChannelDeleteByID.
func TestFawkesTxCustomChannelDeleteByID(t *testing.T) {
	convey.Convey("TxCustomChannelDeleteByID", t, func(convCtx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxCustomChannelDeleteByID(tx, 4)
			convCtx.Convey("Then err should be nil.r should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.TxCustomChannelDeleteByID(tx, 4)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestFawkesGetChannelCount test FawkesGetChannelCount.
func TestFawkesGetChannelCount(t *testing.T) {
	convey.Convey("GetChannelById", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetChannelCount(c, "")
			convCtx.Convey("Then err should be nil.chel should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesGetAppChannelCount test FawkesGetAppChannelCount.
func TestFawkesGetAppChannelCount(t *testing.T) {
	convey.Convey("GetAppChannelCount", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetAppChannelCount(c, "9n0f", "y")
			convCtx.Convey("Then err should be nil.chel should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
