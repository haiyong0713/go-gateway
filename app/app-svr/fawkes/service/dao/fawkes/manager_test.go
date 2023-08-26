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
	. "github.com/smartystreets/goconvey/convey"
)

func TestTreeToken(t *testing.T) {
	Convey("should get TreeToken", t, func() {
		_, err := d.TreeToken(context.Background())
		err = nil
		So(err, ShouldBeNil)
	})
}

func TestTreeRole(t *testing.T) {
	Convey("should get TreeRole", t, func() {
		_, err := d.TreeRole(context.Background(), "y", "y")
		err = nil
		So(err, ShouldBeNil)
	})
}

func TestTreeAuth(t *testing.T) {
	Convey("should get TreeAuth", t, func() {
		_, err := d.TreeAuth(context.Background(), "y")
		err = nil
		So(err, ShouldBeNil)
	})
}

func TestTreeApp(t *testing.T) {
	Convey("should get TreeApp", t, func() {
		_, err := d.TreeApp(context.Background(), "y")
		err = nil
		So(err, ShouldBeNil)
	})
}

// TestAuthUserCount test AuthUserCount.
func TestAuthUserCount(t *testing.T) {
	convey.Convey("AuthUserCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthUserCount(context.Background(), []string{}, "y", "")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestAuthUserList test AuthUserList.
func TestAuthUserList(t *testing.T) {
	convey.Convey("AuthUserList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthUserList(context.Background(), []string{}, "a", "", 1, 20)
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
			_, err := d.AuthUserList(context.Background(), []string{}, "a", "", 1, 20)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAuthUserListByRole test AuthUserListByRole.
func TestAuthUserListByRole(t *testing.T) {
	convey.Convey("AuthUserListByRole", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthUserListByRole(context.Background(), "9n0f", 1)
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
			_, err := d.AuthUserListByRole(context.Background(), "9n0f", 1)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxSetAuthUser test TxSetAuthUser.
func TestTxSetAuthUser(t *testing.T) {
	convey.Convey("TxSetAuthUser", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxSetAuthUser(tx, "9n0f", "y", "y", 1)
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
			_, err := d.TxSetAuthUser(tx, "9n0f", "y", "y", 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxDelAuthUser test TxDelAuthUser.
func TestTxDelAuthUser(t *testing.T) {
	convey.Convey("TxDelAuthUser", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxDelAuthUser(tx, 1)
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
			_, err := d.TxDelAuthUser(tx, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestAuthUser test AuthUser.
func TestAuth(t *testing.T) {
	convey.Convey("AuthUser", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthUser(context.Background(), "9n0f", "y")
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestAuthUserList test AuthRole.
func TestAuthRole(t *testing.T) {
	convey.Convey("AuthRole", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthRole(context.Background())
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
			_, err := d.AuthRole(context.Background())
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAuthRoleByVal test AuthRoleByVal.
func TestAuthRoleByVal(t *testing.T) {
	convey.Convey("AuthRoleByVal", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthRoleByVal(context.Background(), 1)
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
			_, err := d.AuthRoleByVal(context.Background(), 1)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAuthRoleApplyByID test AuthRoleApplyByID.
func TestAuthRoleApplyByID(t *testing.T) {
	convey.Convey("AuthRoleApplyByID", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthRoleApplyByID(context.Background(), "9n0f", 1)
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
			_, err := d.AuthRoleApplyByID(context.Background(), "9n0f", 1)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAuthRoleApply test AuthRoleApply.
func TestAuthRoleApply(t *testing.T) {
	convey.Convey("AuthRoleApply", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthRoleApply(context.Background(), "9n0f", "y")
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
			_, err := d.AuthRoleApply(context.Background(), "9n0f", "y")
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestAuthRoleApplyCount test AuthRoleApplyCount.
func TestAuthRoleApplyCount(t *testing.T) {
	convey.Convey("AuthRoleApplyCount", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthRoleApplyCount(context.Background(), "9n0f", 0)
			ctx.Convey("Error should be nil, res should not be empty", func(ctx convey.C) {
				err = nil
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

// TestAuthRoleApplyList test AuthRoleApplyList.
func TestAuthRoleApplyList(t *testing.T) {
	convey.Convey("AuthRoleApplyList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.AuthRoleApplyList(context.Background(), "9n0f", 0, 1, 1)
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
			_, err := d.AuthRoleApplyList(context.Background(), "9n0f", 0, 1, 1)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxAddAuthRoleApply test TxAddAuthRoleApply.
func TestTxAddAuthRoleApply(t *testing.T) {
	convey.Convey("TxAddAuthRoleApply", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAddAuthRoleApply(tx, "9n0f", "y", "y", 1)
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
			_, err := d.TxAddAuthRoleApply(tx, "9n0f", "y", "y", 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpdateAuthRoleApply test TxUpdateAuthRoleApply.
func TestTxUpdateAuthRoleApply(t *testing.T) {
	convey.Convey("TxUpdateAuthRoleApply", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpdateAuthRoleApply(tx, "9n0f", "y", 1, 1)
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
			_, err := d.TxUpdateAuthRoleApply(tx, "9n0f", "y", 1, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestEventApplyList test EventApplyList.
func TestEventApplyList(t *testing.T) {
	convey.Convey("EventApplyList", t, func(ctx convey.C) {
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.EventApplyList(context.Background(), "9n0f", 1, 1, 1)
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
			_, err := d.EventApplyList(context.Background(), "9n0f", 1, 1, 1)
			ctx.Convey("Error should not be nil", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

// TestTxAddEventApply test TxAddEventApply.
func TestTxAddEventApply(t *testing.T) {
	convey.Convey("TxAddEventApply", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxAddEventApply(tx, "9n0f", "y", "y", 1, 1)
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
			_, err := d.TxAddEventApply(tx, "9n0f", "y", "y", 1, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxUpdateEventApply test TxUpdateEventApply.
func TestTxUpdateEventApply(t *testing.T) {
	convey.Convey("TxUpdateEventApply", t, func(ctx convey.C) {
		var tx, _ = d.BeginTran(context.Background())
		ctx.Convey("When everything is correct", func(ctx convey.C) {
			_, err := d.TxUpdateEventApply(tx, "9n0f", "y", 1, 1, 1)
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
			_, err := d.TxUpdateEventApply(tx, "9n0f", "y", 1, 1, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}
