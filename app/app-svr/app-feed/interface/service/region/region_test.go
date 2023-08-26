package region

import (
	"context"
	"flag"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/interface/model/tag"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	s *Service
)

func init() {
	dir, _ := filepath.Abs("../../cmd/app-feed-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	s = New(conf.Conf)
}

func TestService_HotTags(t *testing.T) {
	type args struct {
		c    context.Context
		mid  int64
		rid  int16
		ver  string
		plat int8
		now  time.Time
	}
	tests := []struct {
		name        string
		args        args
		wantHs      []*tag.Hot
		wantVersion string
		wantErr     error
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHs, gotVersion, err := s.HotTags(tt.args.c, tt.args.mid, tt.args.rid, tt.args.ver, tt.args.plat, tt.args.now)
			So(gotHs, ShouldEqual, tt.wantHs)
			So(gotVersion, ShouldEqual, tt.wantVersion)
			So(err, ShouldEqual, tt.wantErr)
		})
	}
}
