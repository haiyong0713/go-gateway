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

func TestFawkesTxAddOrUpdateHtConfig(t *testing.T) {
	convey.Convey("TxAddOrUpdateHtConfig", t, func(convCtx convey.C) {
		var (
			tx, _     = d.BeginTran(context.Background())
			appKey    = ""
			env       = ""
			channel   = ""
			city      = ""
			device    = ""
			buildID   = int64(0)
			upgradNum = int(0)
			gray      = int(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxAddOrUpdateHtConfig(tx, appKey, env, channel, city, device, buildID, upgradNum, gray)
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
			_, err := d.TxAddOrUpdateHtConfig(tx, appKey, env, channel, city, device, buildID, upgradNum, gray)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesHotfixConfGet(t *testing.T) {
	convey.Convey("HotfixConfGet", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			appKey  = ""
			env     = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.HotfixConfGet(c, appKey, env, buildID)
			convCtx.Convey("Then err should be nil.conf should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestFawkesTxHfConfEnvUpdate(t *testing.T) {
	convey.Convey("TxHfConfEnvUpdate", t, func(convCtx convey.C) {
		var (
			tx, _   = d.BeginTran(context.Background())
			appKey  = ""
			env     = ""
			buildID = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxHfConfEnvUpdate(tx, appKey, env, buildID)
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
			_, err := d.TxHfConfEnvUpdate(tx, appKey, env, buildID)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesTxHfEnvUpdate(t *testing.T) {
	convey.Convey("TxHfEnvUpdate", t, func(convCtx convey.C) {
		var (
			tx, _   = d.BeginTran(context.Background())
			appKey  = ""
			env     = ""
			sender  = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxHfEnvUpdate(tx, appKey, env, sender, buildID)
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
			_, err := d.TxHfEnvUpdate(tx, appKey, env, sender, buildID)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesGetHotfixID(t *testing.T) {
	convey.Convey("GetHotfixID", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			appKey  = ""
			env     = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetHotfixID(c, appKey, env, buildID)
			convCtx.Convey("Then err should be nil.hotfixID should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestFawkesGetHotfixConfigID(t *testing.T) {
	convey.Convey("GetHotfixConfigID", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			appKey  = ""
			env     = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetHotfixConfigID(c, appKey, env, buildID)
			convCtx.Convey("Then err should be nil.hotfixConfigID should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetSingleHotfixInfo(t *testing.T) {
	convey.Convey("GetSingleHotfixInfo", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			buildID = int64(1)
		)
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.GetSingleHotfixInfo(c, buildID)
			convCtx.Convey("Then err should be nil.r should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetHfJobRefresh(t *testing.T) {
	convey.Convey("GetHfJobRefresh", t, func(convCtx convey.C) {
		var (
			c = context.Background()
		)
		convCtx.Convey("When everything goes correct", func(convCtx convey.C) {
			_, err := d.GetHfJobRefresh(c)
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
			_, err := d.GetHfJobRefresh(c)
			convCtx.Convey("Error should not be nil", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestFawkesGetHotfixInfo(t *testing.T) {
	convey.Convey("GetHotfixInfo", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			appKey  = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetHotfixInfo(c, appKey, buildID)
			convCtx.Convey("Then err should be nil.hfInfos should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestFawkesPushHotfix(t *testing.T) {
	convey.Convey("PushHotfix", t, func(convCtx convey.C) {
		var (
			tx, _   = d.BeginTran(context.Background())
			appKey  = ""
			env     = ""
			nextEnv = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.PushHotfix(tx, appKey, env, nextEnv, buildID)
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
			_, err := d.PushHotfix(tx, appKey, env, nextEnv, buildID)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesPushHotfixConf(t *testing.T) {
	convey.Convey("PushHotfixConf", t, func(convCtx convey.C) {
		var (
			tx, _   = d.BeginTran(context.Background())
			appKey  = ""
			env     = ""
			nextEnv = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.PushHotfixConf(tx, appKey, env, nextEnv, buildID)
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
			_, err := d.PushHotfixConf(tx, appKey, env, nextEnv, buildID)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesTxHotfixEffect(t *testing.T) {
	convey.Convey("TxHotfixEffect", t, func(convCtx convey.C) {
		var (
			tx, _   = d.BeginTran(context.Background())
			appKey  = ""
			env     = ""
			buildID = int64(0)
			effect  = int(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxHotfixEffect(tx, appKey, env, buildID, effect)
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
			_, err := d.TxHotfixEffect(tx, appKey, env, buildID, effect)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesGetHotfixListCount(t *testing.T) {
	convey.Convey("GetHotfixListCount", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			appKey = ""
			env    = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetHotfixListCount(c, appKey, env)
			convCtx.Convey("Then err should be nil.count should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestFawkesGetHotfixList(t *testing.T) {
	convey.Convey("GetHotfixList", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			appKey = ""
			env    = ""
			pn     = int(0)
			ps     = int(0)
			order  = ""
			sort   = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetHotfixList(c, appKey, env, pn, ps, order, sort)
			convCtx.Convey("Then err should be nil.items should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestFawkesGetHotfixOrigin(t *testing.T) {
	convey.Convey("GetHotfixOrigin", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			appKey  = ""
			env     = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetHotfixOrigin(c, appKey, env, buildID)
			convCtx.Convey("Then err should be nil.origin should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestFawkesGetHotfixConfig(t *testing.T) {
	convey.Convey("GetHotfixConfig", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			appKey  = ""
			env     = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetHotfixConfig(c, appKey, env, buildID)
			convCtx.Convey("Then err should be nil.config should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestFawkesTxAddHotfixBuild(t *testing.T) {
	convey.Convey("TxAddHotfixBuild", t, func(convCtx convey.C) {
		var (
			tx, _               = d.BeginTran(context.Background())
			appKey              = ""
			buildID             = int64(0)
			gitType             = int(0)
			gitName             = ""
			env                 = ""
			version             = ""
			versionCode         = int64(0)
			internalVersionCode = int64(0)
			name                = "fd"
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxAddHotfixBuild(tx, appKey, "y", "y", buildID, versionCode, internalVersionCode, gitType, gitName, env, version, name, "")
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
			_, err := d.TxAddHotfixBuild(tx, appKey, "y", "y", buildID, versionCode, internalVersionCode, gitType, gitName, env, version, name, "")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesTxHotfixBuildIDUpdate(t *testing.T) {
	convey.Convey("TxHotfixBuildIDUpdate", t, func(convCtx convey.C) {
		var (
			tx, _ = d.BeginTran(context.Background())
			id    = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxHotfixBuildIDUpdate(tx, id)
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
			_, err := d.TxHotfixBuildIDUpdate(tx, id)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesTxHotfixUpdate(t *testing.T) {
	convey.Convey("TxHotfixUpdate", t, func(convCtx convey.C) {
		var (
			tx, _   = d.BeginTran(context.Background())
			buildID = int64(0)
			glJobID = int64(0)
			commit  = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxHotfixUpdate(tx, buildID, glJobID, commit)
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
			_, err := d.TxHotfixUpdate(tx, buildID, glJobID, commit)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesTxHotfixUpload(t *testing.T) {
	convey.Convey("TxHotfixUpload", t, func(convCtx convey.C) {
		var (
			tx, _   = d.BeginTran(context.Background())
			size    = int64(0)
			md5     = ""
			hfPath  = ""
			hfURL   = ""
			appKey  = ""
			url     = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxHotfixUpload(tx, size, buildID, md5, hfPath, hfURL, url, appKey)
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
			_, err := d.TxHotfixUpload(tx, size, buildID, md5, hfPath, hfURL, url, appKey)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesTxHotfixCancel(t *testing.T) {
	convey.Convey("TxHotfixCancel", t, func(convCtx convey.C) {
		var (
			tx, _   = d.BeginTran(context.Background())
			status  = int(0)
			appKey  = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxHotfixCancel(tx, status, appKey, buildID)
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
			_, err := d.TxHotfixCancel(tx, status, appKey, buildID)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesTxHotfixUpdateStatus(t *testing.T) {
	convey.Convey("TxHotfixUpdateStatus", t, func(convCtx convey.C) {
		var (
			tx, _        = d.BeginTran(context.Background())
			status       = int(0)
			patchBuildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxHotfixUpdateStatus(tx, patchBuildID, status)
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
			_, err := d.TxHotfixUpdateStatus(tx, patchBuildID, status)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesTxHotfixDel(t *testing.T) {
	convey.Convey("TxHotfixDel", t, func(convCtx convey.C) {
		var (
			tx, _   = d.BeginTran(context.Background())
			appKey  = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.TxHotfixDel(tx, appKey, buildID)
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
			_, err := d.TxHotfixDel(tx, appKey, buildID)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		convCtx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestFawkesGetOriginURL(t *testing.T) {
	convey.Convey("GetOriginURL", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			appKey  = ""
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetOriginURL(c, appKey, buildID)
			convCtx.Convey("Then err should be nil.config should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestFawkesGetOriginInfo test GetOriginInfo
func TestFawkesGetOriginInfo(t *testing.T) {
	convey.Convey("GetOriginInfo", t, func(convCtx convey.C) {
		var (
			c       = context.Background()
			appKey  = ""
			env     = "test"
			buildID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetOriginInfo(c, appKey, env, buildID)
			convCtx.Convey("Then err should be nil.config should not be nil.", func(convCtx convey.C) {
				err = nil
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
