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

// TestTxAddApk test TxAddApk.
func TestTxAddApk(t *testing.T) {
	convey.Convey("TxAddApk", t, func(ctx convey.C) {
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
			_, err := d.TxAddApk(tx, 123, 123, "test", "test", "test", "test", "test", "test", 123, 1)
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
			_, err := d.TxAddApk(tx, 123, 123, "test", "test", "test", "test", "test", "test", 123, 1)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}

// TestTxAddDiffPatch test TxAddDiffPatch.
func TestTxAddDiffPatch(t *testing.T) {
	convey.Convey("TxAddDiffPatch", t, func(ctx convey.C) {
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
			_, err := d.TxAddDiffPatch(tx, "test", "test", 123, 123, 123, 123, "test", "test", "test", "test", 123)
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
			_, err := d.TxAddDiffPatch(tx, "test", "test", 123, 123, 123, 123, "test", "test", "test", "test", 123)
			ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
		ctx.Reset(func() {
			tx.Rollback()
		})
	})
}
