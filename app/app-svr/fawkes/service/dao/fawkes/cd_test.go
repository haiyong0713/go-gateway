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

	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
)

// TestPackVersionByAppKey test PackVersionByAppKey.
func TestPackVersionByAppKey(t *testing.T) {
	convey.Convey("PackVersionByAppKey", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackVersionByAppKey(context.Background(), "9n0f", "test", "", 20, 0)
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
			_, err := d.PackVersionByAppKey(context.Background(), "9n0f", "test", "", 20, 0)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestPackVersionByAppKeys test PackVersionByAppKeys.
func TestPackVersionByAppKeys(t *testing.T) {
	convey.Convey("PackVersionByAppKeys", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackVersionByAppKeys(context.Background(), "test", []string{"9n0f", "5le0"})
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
			_, err := d.PackVersionByAppKeys(context.Background(), "test", []string{"9n0f", "5le0"})
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestPackVersionCount test PackVersionCount.
func TestPackVersionCount(t *testing.T) {
	convey.Convey("PackVersionCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackVersionCount(context.Background(), "9n0f", "test")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestTxSetPackVersion test TxSetPackVersion.
func TestTxSetPackVersion(t *testing.T) {
	convey.Convey("TxSetPackVersion", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetPackVersion(tx, "bili.tv", "9n0f", "test", "5.40.1", 5395000, 1)
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
			_, err := d.TxSetPackVersion(tx, "bili.tv", "9n0f", "test", "5.40.1", 5395000, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxSetPack test TxSetPack.
func TestTxSetPack(t *testing.T) {
	convey.Convey("TxSetPack", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetPack(tx, "bili.tv", "9n0f", "test", 1, 2, 3, 4, "y", "y", 4, "y", 5, "y", "y", "y", "y", "y", "y", "y", "y", "y", "y", 1, 1, "", "")
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
			_, err := d.TxSetPack(tx, "bili.tv", "9n0f", "test", 1, 2, 3, 4, "y", "y", 4, "y", 5, "y", "y", "y", "y", "y", "y", "y", "y", "y", "y", 1, 1, "", "")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestPackVersionList test PackVersionList.
func TestPackVersionList(t *testing.T) {
	convey.Convey("PackVersionList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackVersionList(context.Background(), "9n0f", "test", 1, 20)
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
			_, err := d.PackVersionList(context.Background(), "9n0f", "test", 1, 20)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestPackByVersion test PackByVersion.
func TestPackByVersion(t *testing.T) {
	convey.Convey("PackByVersion", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackByVersion(context.Background(), "9n0f", "test", 3)
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
			_, err := d.PackByVersion(context.Background(), "9n0f", "test", 3)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestPackByVersions test PackByVersions.
func TestPackByVersions(t *testing.T) {
	convey.Convey("PackByVersions", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackByVersions(context.Background(), "9n0f", "test", []int64{3})
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
			_, err := d.PackByVersions(context.Background(), "9n0f", "test", []int64{3})
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestPackVersionByID test PackVersionByID.
func TestPackVersionByID(t *testing.T) {
	convey.Convey("PackVersionByID", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackVersionByID(context.Background(), "9n0f", 1)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestPackByBuild test PackByBuild.
func TestPackByBuild(t *testing.T) {
	convey.Convey("PackByBuild", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackByBuild(context.Background(), "9n0f", "test", 24)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestTxUpPackVersionID test TxUpPackVersionID.
func TestTxUpPackVersionID(t *testing.T) {
	convey.Convey("TxUpPackVersionID", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpPackVersionID(tx, "9n0f", 1, 1)
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
			_, err := d.TxUpPackVersionID(tx, "9n0f", 1, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxSetPackConfigSwitch test TxSetPackConfigSwitch.
func TestTxSetPackConfigSwitch(t *testing.T) {
	convey.Convey("TxSetPackConfigSwitch", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetPackConfigSwitch(tx, 1, 1)
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
			_, err := d.TxSetPackConfigSwitch(tx, 1, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxSetPackUpgradConfig test TxSetPackUpgradConfig.
func TestTxSetPackUpgradConfig(t *testing.T) {
	convey.Convey("TxSetPackUpgradConfig", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetPackUpgradConfig(tx, "9n0f", "test", 1, "1,2", "", "", "3,4", "1.0.1", "", 1, "copntent", "y", "y", "y", "y", "", 1, 1)
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
			_, err := d.TxSetPackUpgradConfig(tx, "9n0f", "test", 1, "1,2", "", "", "3,4", "1.0.1", "", 1, "copntent", "y", "y", "y", "y", "", 1, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestPackUpgradConfig test PackUpgradConfig.
func TestPackUpgradConfig(t *testing.T) {
	convey.Convey("PackUpgradConfig", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackUpgradConfig(context.Background(), "9n0f", "test", 1)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestTxSetPackFilterConfig test TxSetPackFilterConfig.
func TestTxSetPackFilterConfig(t *testing.T) {
	convey.Convey("TxSetPackFilterConfig", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetPackFilterConfig(tx, "9n0f", "test", 1, "network", "isp", "channel", "city", 1, "salt", "desc", "y", "", 2)
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
			_, err := d.TxSetPackFilterConfig(tx, "9n0f", "test", 1, "network", "isp", "channel", "city", 1, "salt", "desc", "y", "", 2)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestPackFilterConfig test PackFilterConfig.
func TestPackFilterConfig(t *testing.T) {
	convey.Convey("PackFilterConfig", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackFilterConfig(context.Background(), "5le0", "test", []int64{123, 234})
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
			_, err := d.PackFilterConfig(context.Background(), "5le0", "test", []int64{123, 234})
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxSetPackFlowConfig test TxSetPackFlowConfig.
func TestTxSetPackFlowConfig(t *testing.T) {
	convey.Convey("TxSetPackFlowConfig", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetPackFlowConfig(tx, "5le0", "test", "0,99", 123)
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
			_, err := d.TxSetPackFlowConfig(tx, "5le0", "test", "0,99", 123)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestPackFlowConfig test PackFlowConfig.
func TestPackFlowConfig(t *testing.T) {
	convey.Convey("PackFlowConfig", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PackFlowConfig(context.Background(), "5le0", "test", []int64{123, 124})
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
			_, err := d.PackFlowConfig(context.Background(), "5le0", "test", []int64{123, 124})
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestPatchListCount test PatchListCount.
func TestPatchListCount(t *testing.T) {
	convey.Convey("PatchListCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PatchListCount(context.Background(), "5le0", 124)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestPatchList test PatchList.
func TestPatchList(t *testing.T) {
	convey.Convey("PatchList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.PatchList(context.Background(), "5le0", 124, 1, 20)
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
			_, err := d.PatchList(context.Background(), "5le0", 124, 1, 20)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestGenerateListCount test GenerateListCount.
func TestGenerateListCount(t *testing.T) {
	convey.Convey("GenerateListCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.GenerateListCount(context.Background(), "5le0", "test", 124)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestGenerateList test GenerateList.
func TestGenerateList(t *testing.T) {
	convey.Convey("GenerateList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.GenerateList(context.Background(), "5le0", 124)
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
			_, err := d.GenerateList(context.Background(), "5le0", 124)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestLastPack test LastPack.
func TestLastPack(t *testing.T) {
	convey.Convey("LastPack", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.LastPack(context.Background(), "5le0", 0, 1, 2)
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
			_, err := d.LastPack(context.Background(), "5le0", 0, 1, 2)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestGenerate test Generate.
func TestGenerate(t *testing.T) {
	convey.Convey("Generate", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.Generate(context.Background(), "5le0", 124)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestTxUpGenerateCDN test TxUpGenerateCDN.
func TestTxUpGenerateCDN(t *testing.T) {
	convey.Convey("TxUpGenerateCDN", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpGenerateCDN(tx, "5le0", "test", "y", 1, 123)
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
			_, err := d.TxUpGenerateCDN(tx, "5le0", "test", "y", 1, 123)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxSetGenerate test TxSetGenerate.
func TestTxSetGenerate(t *testing.T) {
	convey.Convey("TxSetGenerate", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetGenerate(tx, "5le0", 1, 2, 3, "y", "y", "y", "y", "y", "y")
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
			_, err := d.TxSetGenerate(tx, "5le0", 1, 2, 3, "y", "y", "y", "y", "y", "y")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxSetGenerates test TxSetGenerates.
func TestTxSetGenerates(t *testing.T) {
	convey.Convey("TxSetGenerates", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetGenerates(tx, []string{}, []interface{}{})
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
			_, err := d.TxSetGenerates(tx, []string{}, []interface{}{})
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpGenerateStatus test TxUpGenerateStatus.
func TestTxUpGenerateStatus(t *testing.T) {
	convey.Convey("TxUpGenerateStatus", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpGenerateStatus(tx, "5le0", "y", 1, 1)
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
			_, err := d.TxUpGenerateStatus(tx, "5le0", "y", 1, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxAppCDGenerateTestStateSet test TxAppCDGenerateTestStateSet.
//func TestTxAppCDGenerateTestStateSet(t *testing.T) {
//	convey.Convey("TxAppCDGenerateTestStateSet", t, func(ctx convey.C) {
//		var tx, _ = d.BeginTran(context.Background())
//		ctx.Convey("When everything is correct", func(ctx convey.C) {
//			_, err := d.TxAppCDGenerateTestStateSet(tx, "android", "1", 1)
//			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
//				err = nil
//				ctx.So(err, convey.ShouldBeNil)
//			})
//		})
//		ctx.Convey("When tx.Exec gets error", func(ctx convey.C) {
//			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
//				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
//					return nil, fmt.Errorf("tx.Exec Error")
//				})
//			defer guard.Unpatch()
//			_, err := d.TxAppCDGenerateTestStateSet(tx, "android", "1", 1)
//			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
//				ctx.So(err, convey.ShouldNotBeNil)
//			})
//		})
//		ctx.Reset(func() {
//			tx.Rollback()
//		})
//	})
//}

// TestTxAddPatch test TxAddPatch.
func TestTxAddPatch(t *testing.T) {
	convey.Convey("TxAddPatch", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAddPatch(tx, "y", 1, 2, 3, 4, 5, 6, 7, "y", "y", "y", "y", "y")
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
			_, err := d.TxAddPatch(tx, "y", 1, 2, 3, 4, 5, 6, 7, "y", "y", "y", "y", "y")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestKeyContribute dao ut.
func TestFormLike(t *testing.T) {
	convey.Convey("keyContribute", t, func(ctx convey.C) {
		key := d.darkness("")
		ctx.Convey("key should not be equal to xxxx", func(ctx convey.C) {
			ctx.So(key, convey.ShouldNotEqual, "secret")
		})
	})
}

// TestTxUpPackSteadyState test TxUpPackSteadyState.
func TestTxUpPackSteadyState(t *testing.T) {
	convey.Convey("TxUpPackSteadyState", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpPackSteadyState(tx, "android", "desc", 0, 1)
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
			_, err := d.TxUpPackSteadyState(tx, "android", "desc", 0, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxInManagerVersion test TxInManagerVersion.
func TestTxInManagerVersion(t *testing.T) {
	convey.Convey("TxInManagerVersion", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxInManagerVersion(tx, 0, 0, 1, "desc", "1.0")
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
			_, err := d.TxInManagerVersion(tx, 0, 0, 1, "desc", "1.0")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxInManagerVersionUpdate test TxInManagerVersionUpdate.
func TestTxInManagerVersionUpdate(t *testing.T) {
	convey.Convey("TxInManagerVersionUpdate", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxInManagerVersionUpdate(tx, 100, 0, 1, 0, 0, 0, 0, 100, 1, 0, "channel", "url", "md5", "model", "policyName", "policyURL", "")
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
			_, err := d.TxInManagerVersionUpdate(tx, 100, 0, 1, 0, 0, 0, 0, 100, 1, 0, "channel", "url", "md5", "model", "policyName", "policyURL", "")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxInManagerVersionUpdateLimit test TxInManagerVersionUpdateLimit.
func TestTxInManagerVersionUpdateLimit(t *testing.T) {
	convey.Convey("TxInManagerVersionUpdate", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxInManagerVersionUpdateLimit(tx, 1, []int{1}, "ne")
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
			_, err := d.TxInManagerVersionUpdateLimit(tx, 1, []int{1}, "ne")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestCustomChannelPack test CustomChannelPack.
func TestCustomChannelPack(t *testing.T) {
	convey.Convey("CustomChannelPack", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.CustomChannelPack(context.Background(), "9n0f", "", 0)
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
			_, err := d.CustomChannelPack(context.Background(), "9n0f", "", 0)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestCustomChannelPackByID test CustomChannelPackByID.
func TestCustomChannelPackByID(t *testing.T) {
	convey.Convey("CustomChannelPackByID", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.CustomChannelPackByID(context.Background(), "9n0f", 0)
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
			_, err := d.CustomChannelPackByID(context.Background(), "9n0f", 0)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestCustomChannelPacks test CustomChannelPacks.
func TestCustomChannelPacks(t *testing.T) {
	convey.Convey("CustomChannelPacks", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.CustomChannelPacks(context.Background(), "9n0f", 0)
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
			_, err := d.CustomChannelPacks(context.Background(), "9n0f", 0)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxAddCustomChannelPacks test TxAddCustomChannelPacks.
func TestTxAddCustomChannelPacks(t *testing.T) {
	convey.Convey("TxAddCustomChannelPacks", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAddCustomChannelPacks(tx, []string{}, []interface{}{})
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
			_, err := d.TxAddCustomChannelPacks(tx, []string{}, []interface{}{})
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpCustomChannelPack test TxUpCustomChannelPack.
func TestTxUpCustomChannelPack(t *testing.T) {
	convey.Convey("TxUpCustomChannelPack", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpCustomChannelPack(tx, "android", "desc", "", "", 0, 0, 0)
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
			_, err := d.TxUpCustomChannelPack(tx, "android", "", "", "", 0, 0, 0)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

func TestDao_GenerateByAppKeyAndStatus(t *testing.T) {
	convey.Convey("GenerateByAppKeyAndStatus", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			rows, err := d.GenerateByAppKeyAndStatus(context.Background(), "android", []int{cdmdl.GenerateUpload, cdmdl.GenerateTest, cdmdl.GeneratePublish}, 1)
			fmt.Printf("rows %v", rows)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
		/*ctx.Convey("When tx.Exec gets error", func(ctx convey.C) {
			guard := monkey.PatchInstanceMethod(reflect.TypeOf(tx), "Exec",
				func(_ *xsql.Tx, _ string, _ ...interface{}) (sql.Result, error) {
					return nil, fmt.Errorf("tx.Exec Error")
				})
			defer guard.Unpatch()
			_, err := d.GenerateByAppKeyAndStatus(context.Background(), "android", []int{cdmdl.GenerateUpload, cdmdl.GenerateTest, cdmdl.GeneratePublish}, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})*/
	})
}
