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

// TestLaserCount test LaserCount.
func TestLaserCount(t *testing.T) {
	convey.Convey("LaserCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.LaserCount(context.Background(), "9n0f", "", "", "", 0, 0, 0, 0, 0)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestLaserList test LaserList.
func TestLaserList(t *testing.T) {
	convey.Convey("LaserList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.LaserList(context.Background(), "9n0f", "", "", "", 0, 0, 0, 0, 0, 0, 0)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		ctx.Convey("When db.Query gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(d.db), "Query", func(_ *xsql.DB, _ context.Context, _ string, _ ...interface{}) (*xsql.Rows, error) {
				return nil, fmt.Errorf("db.Query error")
			})
			defer guard.Unpatch()
			_, err := d.LaserList(context.Background(), "9n0f", "", "", "", 0, 0, 0, 0, 0, 0, 0)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxAddLaser test TxAddLaser.
func TestTxAddLaser(t *testing.T) {
	convey.Convey("TxAddLaser", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAddLaser(tx, "9n0f", "ios", "", "2019-01-01", "a", "", "", 123, 1, 1, 0)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		ctx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.TxAddLaser(tx, "9n0f", "ios", "", "2019-01-01", "a", "", "", 123, 1, 1, 0)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelLaser test TxDelLaser.
func TestTxDelLaser(t *testing.T) {
	convey.Convey("TxDelLaser", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxDelLaser(tx, 1)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		ctx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.TxDelLaser(tx, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpLaserStatus test TxUpLaserStatus.
func TestTxUpLaserStatus(t *testing.T) {
	convey.Convey("TxUpLaserStatus", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpLaserStatus(tx, 1, 1, "", "")
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		ctx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.TxUpLaserStatus(tx, 1, 1, "", "")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpLaserURL test TxUpLaserURL.
func TestTxUpLaserURL(t *testing.T) {
	convey.Convey("TxUpLaserURL", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpLaserURL(tx, 1, "", "", "")
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		ctx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.TxUpLaserURL(tx, 1, "", "", "")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpLaserErrorMessage test TxUpLaserErrorMessage.
func TestTxUpLaserErrorMessage(t *testing.T) {
	convey.Convey("TxUpLaserErrorMessage", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpLaserErrorMessage(tx, 1, "")
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		ctx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.TxUpLaserErrorMessage(tx, 1, "")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestLaseActiveCount test LaseActiveCount.
func TestLaseActiveCount(t *testing.T) {
	convey.Convey("LaseActiveCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.LaserActiveCount(context.Background(), "9n0f", "", 0, 0, 0, 0)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestLaserList test LaserActiveList.
func TestLaserActiveList(t *testing.T) {
	convey.Convey("LaserActiveList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.LaserActiveList(context.Background(), "9n0f", "", 0, 0, 0, 0, 0, 0)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		ctx.Convey("When db.Query gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(d.db), "Query", func(_ *xsql.DB, _ context.Context, _ string, _ ...interface{}) (*xsql.Rows, error) {
				return nil, fmt.Errorf("db.Query error")
			})
			defer guard.Unpatch()
			_, err := d.LaserActiveList(context.Background(), "9n0f", "", 0, 0, 0, 0, 0, 0)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}
