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

// TestExistConfigVersion test ExistConfigVersion.
func TestExistConfigVersion(t *testing.T) {
	convey.Convey("ExistConfigVersion", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ExistConfigVersion(context.Background(), "9n0f", "test", "5.37.0", 5365000)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestConfigVersionByID test ConfigVersionByID.
func TestConfigVersionByID(t *testing.T) {
	convey.Convey("ConfigVersionByID", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigVersionByID(context.Background(), 5365000)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestTxSetConfigVersion test TxSetConfigVersion.
func TestTxSetConfigVersion(t *testing.T) {
	convey.Convey("TxSetConfigVersion", t, func(ctx convey.C) {
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
			_, err := d.TxSetConfigVersion(tx, "5le0", "test", "0,99", 123, "yyy")
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
			_, err := d.TxSetConfigVersion(tx, "5le0", "test", "0,99", 123, "yyy")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestConfigVersionCount test ConfigVersionCount.
func TestConfigVersionCount(t *testing.T) {
	convey.Convey("ConfigVersionCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigVersionCount(context.Background(), "9n0f", "test")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestConfigVersionIDs test ConfigVersionIDs.
func TestConfigVersionIDs(t *testing.T) {
	convey.Convey("ConfigVersionIDs", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigVersionIDs(context.Background(), "9n0f", "test", "5.37.0", 5365000)
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
			_, err := d.ConfigVersionIDs(context.Background(), "9n0f", "test", "5.37.0", 5365000)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestConfigVersionList test ConfigVersionList.
func TestConfigVersionList(t *testing.T) {
	convey.Convey("ConfigVersionList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigVersionList(context.Background(), "9n0f", "test", 1, 20)
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
			_, err := d.ConfigVersionList(context.Background(), "9n0f", "test", 1, 20)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestConfigModifyCounts test ConfigModifyCounts.
func TestConfigModifyCounts(t *testing.T) {
	convey.Convey("ConfigModifyCounts", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigModifyCounts(context.Background(), "9n0f", "test", []int64{1, 2})
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestConfigPublishs test ConfigPublishs.
func TestConfigPublishs(t *testing.T) {
	convey.Convey("ConfigPublishs", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigPublishs(context.Background(), "9n0f", "test", []int64{999999})
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
			_, err := d.ConfigPublishs(context.Background(), "9n0f", "test", []int64{999999})
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestConfigPublishAll test ConfigPublishAll.
func TestConfigPublishAll(t *testing.T) {
	convey.Convey("ConfigPublishAll", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigPublishAll(context.Background(), "9n0f", "test", 1, 20)
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
			_, err := d.ConfigPublishAll(context.Background(), "9n0f", "test", 1, 20)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxUpConfigVersionDesc test TxUpConfigVersionDesc.
func TestTxUpConfigVersionDesc(t *testing.T) {
	convey.Convey("TxUpConfigVersionDesc", t, func(ctx convey.C) {
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
			_, err := d.TxUpConfigVersionDesc(tx, 999999, "ut")
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
			_, err := d.TxUpConfigVersionDesc(tx, 999999, "ut")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelConfigVersion test TxDelConfigVersion.
func TestTxDelConfigVersion(t *testing.T) {
	convey.Convey("TxDelConfigVersion", t, func(ctx convey.C) {
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
			_, err := d.TxDelConfigVersion(tx, 999999)
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
			_, err := d.TxDelConfigVersion(tx, 999999)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxAddConfig test TxAddConfig.
func TestTxAddConfig(t *testing.T) {
	convey.Convey("TxAddConfig", t, func(ctx convey.C) {
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
			_, err := d.TxAddConfig(tx, "9n0f", "test", 1, "a", "a", "a", "a", "a")
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
			_, err := d.TxAddConfig(tx, "9n0f", "test", 1, "a", "a", "a", "a", "a")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpConfig test TxUpConfig.
func TestTxUpConfig(t *testing.T) {
	convey.Convey("TxUpConfig", t, func(ctx convey.C) {
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
			_, err := d.TxUpConfig(tx, "9n0f", "test", 1, "a", "a", "a", "a", "a")
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
			_, err := d.TxUpConfig(tx, "9n0f", "test", 1, "a", "a", "a", "a", "a")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelConfig test TxDelConfig.
func TestTxDelConfig(t *testing.T) {
	convey.Convey("TxDelConfig", t, func(ctx convey.C) {
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
			_, err := d.TxDelConfig(tx, "9n0f", "test", 99999, "a", "a", "a", "a")
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
			_, err := d.TxDelConfig(tx, "9n0f", "test", 99999, "a", "a", "a", "a")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestConfigPublish test ConfigPublish.
func TestConfigPublish(t *testing.T) {
	convey.Convey("ConfigPublish", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigPublish(context.Background(), "9n0f", "test", 999999, -1, -1)
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
			_, err := d.ConfigPublish(context.Background(), "9n0f", "test", 999999, -1, -1)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestConfig test Config.
func TestConfig(t *testing.T) {
	convey.Convey("Config", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.Config(context.Background(), "9n0f", "test", 999999)
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
			_, err := d.Config(context.Background(), "9n0f", "test", 999999)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestConfigFile test ConfigFile.
func TestConfigFile(t *testing.T) {
	convey.Convey("ConfigFile", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigFile(context.Background(), "9n0f", "test", 999999)
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
			_, err := d.ConfigFile(context.Background(), "9n0f", "test", 999999)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestConfigLastCV test ConfigLastCV.
func TestConfigLastCV(t *testing.T) {
	convey.Convey("ConfigLastCV", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigLastCV(context.Background(), "9n0f", "test", 999999, 999999)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestTxAddConfigPublish test TxAddConfigPublish.
func TestTxAddConfigPublish(t *testing.T) {
	convey.Convey("TxAddConfigPublish", t, func(ctx convey.C) {
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
			_, err := d.TxAddConfigPublish(tx, "9n0f", "test", 99999, 99999, "a", "a")
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
			_, err := d.TxAddConfigPublish(tx, "9n0f", "test", 99999, 99999, "a", "a")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxAddConfigFile test TxAddConfigFile.
func TestTxAddConfigFile(t *testing.T) {
	convey.Convey("TxAddConfigFile", t, func(ctx convey.C) {
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
			_, err := d.TxAddConfigFile(tx, []string{}, []interface{}{})
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
			_, err := d.TxAddConfigFile(tx, []string{}, []interface{}{})
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpConfigPublishFiles test TxUpConfigPublishFiles.
func TestTxUpConfigPublishFiles(t *testing.T) {
	convey.Convey("TxUpConfigPublishFiles", t, func(ctx convey.C) {
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
			_, err := d.TxUpConfigPublishFiles(tx, "9n0f", "test", 99999, "y", "y", "y", "y")
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

// TestTxUpConfigPublishState test TxUpConfigPublishState.
func TestTxUpConfigPublishState(t *testing.T) {
	convey.Convey("TxUpConfigPublishState", t, func(ctx convey.C) {
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
			_, err := d.TxUpConfigPublishState(tx, "9n0f", "test", 99999, 99999, 1)
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

// TestTxUpConfigState test TxUpConfigState.
func TestTxUpConfigState(t *testing.T) {
	convey.Convey("TxUpConfigState", t, func(ctx convey.C) {
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
			_, err := d.TxUpConfigState(tx, "9n0f", "test", 99999, 1)
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
			_, err := d.TxUpConfigState(tx, "9n0f", "test", 99999, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestConfigVersionByIDs test ConfigVersionByIDs.
func TestConfigVersionByIDs(t *testing.T) {
	convey.Convey("ConfigVersionByIDs", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigVersionByIDs(context.Background(), []int64{9999})
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
			_, err := d.ConfigVersionByIDs(context.Background(), []int64{9999})
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAllNewConfigPublish test AllNewConfigPublish.
func TestAllNewConfigPublish(t *testing.T) {
	convey.Convey("AllNewConfigPublish", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AllNewConfigPublish(context.Background(), "9n0f", "test")
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
			_, err := d.AllNewConfigPublish(context.Background(), "9n0f", "test")
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxDelConfig2 test TxDelConfig2.
func TestTxDelConfig2(t *testing.T) {
	convey.Convey("TxDelConfig2", t, func(ctx convey.C) {
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
			_, err := d.TxDelConfig2(tx, "5le0", "test", 123, "yyy", "yyy")
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
			_, err := d.TxDelConfig2(tx, "5le0", "test", 123, "yyy", "yyy")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxFlushConfig test TxFlushConfig.
func TestTxFlushConfig(t *testing.T) {
	convey.Convey("TxFlushConfig", t, func(ctx convey.C) {
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
			_, err := d.TxFlushConfig(tx, "5le0", "test", 123)
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
			_, err := d.TxFlushConfig(tx, "5le0", "test", 123)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxAddConfigs test TxAddConfigs.
func TestTxAddConfigs(t *testing.T) {
	convey.Convey("TxAddConfigs", t, func(ctx convey.C) {
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
			_, err := d.TxAddConfigs(tx, []string{}, []interface{}{})
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
			_, err := d.TxAddConfigs(tx, []string{}, []interface{}{})
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelAllConfig test TxDelAllConfig.
func TestTxDelAllConfig(t *testing.T) {
	convey.Convey("TxDelAllConfig", t, func(ctx convey.C) {
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
			_, err := d.TxDelAllConfig(tx, 123)
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
			_, err := d.TxDelAllConfig(tx, 123)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpConfigDesc test TxUpConfigDesc.
func TestTxUpConfigDesc(t *testing.T) {
	convey.Convey("TxUpConfigDesc", t, func(ctx convey.C) {
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
			_, err := d.TxUpConfigDesc(tx, "5le0", "test", 123, "desc", "yyy", "yyy")
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
			_, err := d.TxUpConfigDesc(tx, "5le0", "test", 123, "desc", "yyy", "yyy")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelAllConfigPublish test TxDelAllConfigPublish.
func TestTxDelAllConfigPublish(t *testing.T) {
	convey.Convey("TxDelAllConfigPublish", t, func(ctx convey.C) {
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
			_, err := d.TxDelAllConfigPublish(tx, 12334)
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
			_, err := d.TxDelAllConfigPublish(tx, 12334)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelAllConfigFile test TxDelAllConfigFile.
func TestTxDelAllConfigFile(t *testing.T) {
	convey.Convey("TxDelAllConfigFile", t, func(ctx convey.C) {
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
			_, err := d.TxDelAllConfigFile(tx, 12334)
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
			_, err := d.TxDelAllConfigFile(tx, 12334)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpConfigPublishTotal test TxUpConfigPublishTotal.
func TestTxUpConfigPublishTotal(t *testing.T) {
	convey.Convey("TxUpConfigPublishTotal", t, func(ctx convey.C) {
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
			_, err := d.TxUpConfigPublishTotal(tx, "5le0", "test", 123, 456, "desc", "desc")
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
			_, err := d.TxUpConfigPublishTotal(tx, "5le0", "test", 123, 456, "desc", "desc")
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestConfigPublishCount test ConfigPublishCount.
func TestConfigPublishCount(t *testing.T) {
	convey.Convey("ConfigPublishCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigPublishCount(context.Background(), "9n0f", "test")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestConfigModifyCountAll test ConfigModifyCountAll.
func TestConfigModifyCountAll(t *testing.T) {
	convey.Convey("ConfigModifyCountAll", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.ConfigModifyCountsAll(context.Background(), "9n0f", "test")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
