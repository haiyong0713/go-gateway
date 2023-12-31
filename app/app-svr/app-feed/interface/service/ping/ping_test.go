package ping

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/app-feed/interface/conf"
	arcdao "go-gateway/app/app-svr/app-feed/interface/dao/archive"
)

func TestNew(t *testing.T) {
	type args struct {
		c *conf.Config
	}
	tests := []struct {
		name  string
		args  args
		wantS *Service
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		Convey(tt.name, t, func() {
			gotS := New(tt.args.c)
			So(gotS, ShouldEqual, tt.wantS)
		})
	}
}

func TestService_Ping(t *testing.T) {
	type fields struct {
		arcDao *arcdao.Dao
	}
	type args struct {
		c context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				arcDao: tt.fields.arcDao,
			}
			if err := s.Ping(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("Service.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
