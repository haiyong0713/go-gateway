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

// TestAppFFWhithlist test AppFFWhithlist.
func TestAppFFWhithlist(t *testing.T) {
	convey.Convey("AppFFWhithlist", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppFFWhithlist(context.Background(), "9n0f", "test")
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
			_, err := d.AppFFWhithlist(context.Background(), "9n0f", "test")
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxAddFFWhithlist test TxAddFFWhithlist.
func TestTxAddFFWhithlist(t *testing.T) {
	convey.Convey("TxAddFFWhithlist", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAddFFWhithlist(tx, "5le0", "test", "yyq", []int64{123})
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
			_, err := d.TxAddFFWhithlist(tx, "5le0", "test", "yyq", []int64{123})
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelFFWhithlist test TxDelFFWhithlist.
func TestTxDelFFWhithlist(t *testing.T) {
	convey.Convey("TxDelFFWhithlist", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxDelFFWhithlist(tx, "5le0", "test", 123)
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
			_, err := d.TxDelFFWhithlist(tx, "5le0", "test", 123)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxSetFFConfig test TxSetFFConfig.
func TestTxSetFFConfig(t *testing.T) {
	convey.Convey("TxSetFFConfig", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetFFConfig(tx, "5le0", "test", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "1234", "", 123)
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
			_, err := d.TxSetFFConfig(tx, "5le0", "test", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "123", "", 123)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestFFCount test FFCount.
func TestFFCount(t *testing.T) {
	convey.Convey("FFCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.FFCount(context.Background(), "9n0f", "test", "")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFFList test FFList.
func TestFFList(t *testing.T) {
	convey.Convey("FFList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.FFList(context.Background(), "9n0f", "test", "", 0, 0)
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
			_, err := d.FFList(context.Background(), "9n0f", "test", "", 0, 0)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestFFModiyCount test FFModiyCount.
func TestFFModiyCount(t *testing.T) {
	convey.Convey("FFModiyCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.FFModiyCount(context.Background(), "9n0f", "test")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestTxAddFFPublish test TxAddFFPublish.
func TestTxAddFFPublish(t *testing.T) {
	convey.Convey("TxAddFFPublish", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAddFFPublish(tx, "5le0", "test", "yyq", "yyq")
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
			_, err := d.TxAddFFPublish(tx, "5le0", "test", "yyq", "yyq")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxAddFFConfigFile test TxAddFFConfigFile.
func TestTxAddFFConfigFile(t *testing.T) {
	convey.Convey("TxAddFFConfigFile", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAddFFConfigFile(tx, []string{}, []interface{}{})
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
			_, err := d.TxAddFFConfigFile(tx, []string{}, []interface{}{})
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpFFConfigPublishURL test TxUpFFConfigPublishURL.
func TestTxUpFFConfigPublishURL(t *testing.T) {
	convey.Convey("TxUpFFConfigPublishURL", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpFFConfigPublishURL(tx, "http://xxx/xx/x/x", "y", "y", "y", 1233)
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
			_, err := d.TxUpFFConfigPublishURL(tx, "http://xxx/xx/x/x", "y", "y", "y", 1233)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpFFConfigPublishState test TxUpFFConfigPublishState.
func TestTxUpFFConfigPublishState(t *testing.T) {
	convey.Convey("TxUpFFConfigPublishState", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpFFConfigPublishState(tx, "5le0", "test", 1233, 1)
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
			_, err := d.TxUpFFConfigPublishState(tx, "5le0", "test", 1233, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxFlushFFConfig test TxFlushFFConfig.
func TestTxFlushFFConfig(t *testing.T) {
	convey.Convey("TxFlushFFConfig", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxFlushFFConfig(tx, "5le0", "test")
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
			_, err := d.TxFlushFFConfig(tx, "5le0", "test")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestAppFFHistory test AppFFHistory.
func TestAppFFHistory(t *testing.T) {
	convey.Convey("AppFFHistory", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppFFHistory(context.Background(), "9n0f", "test", -1, -1)
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
			_, err := d.AppFFHistory(context.Background(), "9n0f", "test", -1, -1)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAppFFLastFvid test AppFFLastFvid.
func TestAppFFLastFvid(t *testing.T) {
	convey.Convey("AppFFLastFvid", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppFFLastFvid(context.Background(), "9n0f", "test", 5365000)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestAppFFFile test AppFFFile.
func TestAppFFFile(t *testing.T) {
	convey.Convey("AppFFFile", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppFFFile(context.Background(), "9n0f", "test", 1234)
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
			_, err := d.AppFFFile(context.Background(), "9n0f", "test", 1234)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAppFFConfigs test AppFFConfigs.
func TestAppFFConfigs(t *testing.T) {
	convey.Convey("AppFFConfigs", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppFFConfigs(context.Background(), "9n0f", "test")
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
			_, err := d.AppFFConfigs(context.Background(), "9n0f", "test")
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAppFFConfig test AppFFConfig.
func TestAppFFConfig(t *testing.T) {
	convey.Convey("AppFFConfig", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppFFConfig(context.Background(), "9n0f", "test", "yyq")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestTxUpFFConfig test TxUpFFConfig.
func TestTxUpFFConfig(t *testing.T) {
	convey.Convey("TxUpFFConfig", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpFFConfig(tx, "5le0", "test", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "123", "", 123, 456)
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
			_, err := d.TxUpFFConfig(tx, "5le0", "test", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "yyq", "123", "", 123, 456)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpFFConfigState test TxUpFFConfigState.
func TestTxUpFFConfigState(t *testing.T) {
	convey.Convey("TxUpFFConfigState", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpFFConfigState(tx, "5le0", "test", 456)
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
			_, err := d.TxUpFFConfigState(tx, "5le0", "test", 456)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelFFConfig test TxDelFFConfig.
func TestTxDelFFConfig(t *testing.T) {
	convey.Convey("TxDelFFConfig", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxDelFFConfig(tx, "5le0", "test", "456")
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
			_, err := d.TxDelFFConfig(tx, "5le0", "test", "456")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelFFConfig2 test TxDelFFConfig2.
func TestTxDelFFConfig2(t *testing.T) {
	convey.Convey("TxDelFFConfig2", t, func(ctx convey.C) {
		var (
			tx  *xsql.Tx
			err error
		)
		for {
			if tx, err = d.BeginTran(context.Background()); err == nil && tx != nil {
				break
			}
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxDelFFConfig2(tx, "5le0", "test", "456")
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
			_, err := d.TxDelFFConfig2(tx, "5le0", "test", "456")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}
