package fawkes

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	xsql "go-common/library/database/sql"

	cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"

	"github.com/bouk/monkey"
	"github.com/smartystreets/goconvey/convey"
)

// TestFawkesBuildPacks test FawkesBuildPacks.
func TestFawkesBuildPacks(t *testing.T) {
	convey.Convey("BuildPacks", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			appKey     = "56ba"
			pn         = int(1)
			ps         = int(20)
			pkgType    = int(1)
			status     = int(1)
			gitType    = int(1)
			gitKeyword = "master"
			operator   = "zhangyuhang"
			order      = "status"
			sort       = "desc"
		)
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.BuildPacks(c, appKey, pn, ps, pkgType, status, gitType, gitKeyword, operator, order, sort, 0, 1, "", false)
			convCtx.Convey("Then err should be nil.buildPacks should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("When db.Query gets error", func(convCtx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(d.db), "Query", func(_ *xsql.DB, _ context.Context, _ string, _ ...interface{}) (*xsql.Rows, error) {
				return nil, fmt.Errorf("db.Query error")
			})
			defer guard.Unpatch()
			_, err := d.BuildPacks(c, appKey, pn, ps, pkgType, status, gitType, gitKeyword, operator, order, sort, 0, 1, "", false)
			convCtx.Convey("Error should not be nil", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestFawkesBuildPack test FawkesBuildPack.
func TestFawkesBuildPack(t *testing.T) {
	convey.Convey("BuildPack", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			appKey  = "56ba"
			buildID = int64(1)
		)
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.BuildPack(c, appKey, buildID)
			convCtx.Convey("Then err should be nil.r should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesBuildPackAppKey test FawkesBuildPackAppKey.
func TestFawkesBuildPackAppKey(t *testing.T) {
	convey.Convey("BuildPackAppKey", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.BuildPackById(c, buildID)
			convCtx.Convey("Then err should be nil.appKey should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesBuildPacksCount test FawkesBuildPacksCount.
func TestFawkesBuildPacksCount(t *testing.T) {
	convey.Convey("BuildPacksCount", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			appKey     = "56ba"
			pkgType    = int(1)
			status     = int(1)
			gitType    = int(1)
			gitKeyword = "ctime"
			operator   = "DESC"
		)
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.BuildPacksCount(c, appKey, pkgType, status, gitType, gitKeyword, operator, 0, 1, "", false)
			convCtx.Convey("Then err should be nil.r should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesBuildPacksShouldRefresh test BuildPacksShouldRefresh
func TestFawkesBuildPacksShouldRefresh(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("BuildPacksShouldRefresh", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.BuildPacksShouldRefresh(c)
			convCtx.Convey("Then err should be nil.buildPacks should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
		convCtx.Convey("When db.Query gets error", func(convCtx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(d.db), "Query", func(_ *xsql.DB, _ context.Context, _ string, _ ...interface{}) (*xsql.Rows, error) {
				return nil, fmt.Errorf("db.Query error")
			})
			defer guard.Unpatch()
			_, err := d.BuildPacksShouldRefresh(c)
			convCtx.Convey("Error should not be nil", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxInsertBuildPack test TxInsertBuildPack.
func TestTxInsertBuildPack(t *testing.T) {
	convey.Convey("TxInsertBuildPack", t, func(ctx convey.C) {
		var tx, err = d.BeginTran(context.Background())
		if err != nil {
			tx, _ = d.BeginTran(context.Background())
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxInsertBuildPack(tx, "9n0f", "y", "y", "y", 1, 1, 1, "y", "y", "y", 1, 1, "y")
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxInsertBuildPackCreate test TxInsertBuildPackCreate.
func TestTxInsertBuildPackCreate(t *testing.T) {
	convey.Convey("TxInsertBuildPackCreate", t, func(ctx convey.C) {
		var tx, err = d.BeginTran(context.Background())
		if err != nil {
			tx, _ = d.BeginTran(context.Background())
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxInsertBuildPackCreate(tx, "9n0f", "y", "", "y", "y", 1, 1, "y", "y", "{}", "", "", 1, 1)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpdateBuildPackBaseInfo test TxUpdateBuildPackBaseInfo.
func TestTxUpdateBuildPackBaseInfo(t *testing.T) {
	convey.Convey("TxUpdateBuildPackBaseInfo", t, func(ctx convey.C) {
		var tx, err = d.BeginTran(context.Background())
		if err != nil {
			tx, _ = d.BeginTran(context.Background())
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpdateBuildPackBaseInfo(tx, 0, 3, "y", "y", 1, 1)
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
			_, err := d.TxUpdateBuildPackBaseInfo(tx, 0, 3, "y", "y", 1, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpdateBuildPackInfo test TxUpdateBuildPackInfo.
func TestTxUpdateBuildPackInfo(t *testing.T) {
	convey.Convey("TxUpdateBuildPackInfo", t, func(ctx convey.C) {
		var tx, err = d.BeginTran(context.Background())
		if err != nil {
			tx, _ = d.BeginTran(context.Background())
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpdateBuildPackInfo(tx, 0, 3, "y", "", "y", "y", "y", "y", "y", "y", "", false)
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
			_, err := d.TxUpdateBuildPackInfo(tx, 0, 3, "y", "", "y", "y", "y", "y", "y", "y", "", false)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpdateBuildPackStatus test TxUpdateBuildPackStatus.
func TestTxUpdateBuildPackStatus(t *testing.T) {
	convey.Convey("TxUpdateBuildPackStatus", t, func(ctx convey.C) {
		var tx, err = d.BeginTran(context.Background())
		if err != nil {
			tx, _ = d.BeginTran(context.Background())
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpdateBuildPackStatus(tx, 0, 3)
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
			_, err := d.TxUpdateBuildPackStatus(tx, 0, 3)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelBuild test TxDelBuild.
func TestTxDelBuild(t *testing.T) {
	convey.Convey("TxDelBuild", t, func(ctx convey.C) {
		var tx, err = d.BeginTran(context.Background())
		if err != nil {
			tx, _ = d.BeginTran(context.Background())
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxDelBuild(tx, 0)
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
			_, err := d.TxDelBuild(tx, 0)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TxBuildPackDidPush test TxBuildPackDidPush.
func TestTxBuildPackDidPush(t *testing.T) {
	convey.Convey("TxBuildPackDidPush", t, func(ctx convey.C) {
		var tx, err = d.BeginTran(context.Background())
		if err != nil {
			tx, _ = d.BeginTran(context.Background())
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxBuildPackDidPush(tx, 0)
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
			_, err := d.TxBuildPackDidPush(tx, 0)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpdateTestStatus test TxUpdateTestStatus.
func TestTxUpdateTestStatus(t *testing.T) {
	convey.Convey("TxUpdateTestStatus", t, func(ctx convey.C) {
		buildPack := &cimdl.BuildPack{TestStatus: 1, TaskIds: "160,161", GitlabJobID: 2366359}
		var tx, err = d.BeginTran(context.Background())
		if err != nil {
			tx, _ = d.BeginTran(context.Background())
		}
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpdateTestStatus(tx, buildPack)
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
			_, err := d.TxUpdateTestStatus(tx, buildPack)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}
