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

// TestAppInfo test AppInfo.
func TestAppInfo(t *testing.T) {
	convey.Convey("AppInfo", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppInfo(context.Background(), "y", 1)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestAppPass test AppPass.
func TestAppPass(t *testing.T) {
	convey.Convey("AppPass", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppPass(context.Background(), "y")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestTxUpApp test TxUpApp.
func TestTxUpApp(t *testing.T) {
	convey.Convey("TxUpApp", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpApp(tx, 1, "y", "y", "y", "", "", "", "", "", "")
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
			_, err := d.TxUpApp(tx, 1, "y", "y", "y", "", "", "", "", "", "")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestAppFollow test AppFollow.
func TestAppFollow(t *testing.T) {
	convey.Convey("AppFollow", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppFollow(context.Background(), "y")
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
			_, err := d.AppFollow(context.Background(), "y")
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAppsPass test AppsPass.
func TestAppsPass(t *testing.T) {
	convey.Convey("AppsPass", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppsPass(context.Background(), []string{}, "", 0)
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
			_, err := d.AppsPass(context.Background(), []string{}, "", 0)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxInAppFollow test TxInAppFollow.
func TestTxInAppFollow(t *testing.T) {
	convey.Convey("TxInAppFollow", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxInAppFollow(tx, "y", "y")
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
			_, err := d.TxInAppFollow(tx, "y", "y")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelAppFollow test TxDelAppFollow.
func TestTxDelAppFollow(t *testing.T) {
	convey.Convey("TxDelAppFollow", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxDelAppFollow(tx, "y", "y")
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
			_, err := d.TxDelAppFollow(tx, "y", "y")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxAppAdd test TxAppAdd.
func TestTxAppAdd(t *testing.T) {
	convey.Convey("TxAppAdd", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAppAdd(tx, "y", "y", "y", "y", "y", "y", "y", "y", "1", "")
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
			_, err := d.TxAppAdd(tx, "y", "y", "y", "y", "y", "y", "y", "y", "1", "")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxAppAttributeAdd test TxAppAttributeAdd.
func TestTxAppAttributeAdd(t *testing.T) {
	convey.Convey("TxAppAttributeAdd", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAppAttributeAdd(tx, 1, 1, 1, "y", "y", "", "")
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
			_, err := d.TxAppAttributeAdd(tx, 1, 1, 1, "y", "y", "", "")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestAppsAudit test AppsAudit.
func TestAppsAudit(t *testing.T) {
	convey.Convey("AppsAudit", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppsAudit(context.Background())
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
			_, err := d.AppsAudit(context.Background())
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxUpAppAudit test TxUpAppAudit.
func TestTxUpAppAudit(t *testing.T) {
	convey.Convey("TxUpAppAudit", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpAppAudit(tx, "y", "", 1, 1)
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
			_, err := d.TxUpAppAudit(tx, "y", "", 1, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestAppBasicInfo Test AppBasicInfo
func TestAppBasicInfo(t *testing.T) {
	convey.Convey("AppBasicInfo", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, _, _, err := d.AppBasicInfo(context.Background(), "y")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestAppRobotWebhook test AppRobotWebhook.
func TestAppRobotWebhook(t *testing.T) {
	convey.Convey("AppRobotWebhook", t, func(ctx convey.C) {
		_, err := d.AppRobotWebhook(context.Background(), "y")
		ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
			err = nil
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

// TestAppNotificationList test AppNotificationList
func TestAppNotificationList(t *testing.T) {
	convey.Convey("AppNotificationList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AppNotificationList(context.Background(), "y", "y", 1)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestAppNotificationUpdate test TxAppNotificationUpdate
func TestAppNotificationUpdate(t *testing.T) {
	convey.Convey("TxAppNotificationUpdate", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAppNotificationUpdate(tx, 1, "a", "ios", "d", "e", "f", "g", 1, 1, 1, 1, "1", "i", "j")
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
			_, err := d.TxAppNotificationUpdate(tx, 1, "a", "ios", "d", "e", "f", "g", 1, 1, 1, 1, "1", "i", "j")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxAppNotificationAdd test TxAppNotificationAdd
func TestTxAppNotificationAdd(t *testing.T) {
	convey.Convey("TxAppNotificationAdd", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAppNotificationAdd(tx, "a", "ios", "a", "a", "a", "a", 1, 1, 1, 1, "a", "a", "a")
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
			_, err := d.TxAppNotificationAdd(tx, "a", "ios", "a", "a", "a", "a", 1, 1, 1, 1, "a", "a", "a")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}
