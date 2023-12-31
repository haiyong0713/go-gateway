package tag

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
)

func TestNew(t *testing.T) {
	type args struct {
		c *conf.Config
	}
	tests := []struct {
		name  string
		args  args
		wantD *Dao
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			gotD := New(tt.args.c)
			So(gotD, ShouldEqual, tt.wantD)
		})
	}
}
