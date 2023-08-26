package service

import (
	"testing"

	xtime "go-common/library/time"
	"go-gateway/app/app-svr/app-job/job/model/space"
)

func TestService_contributeConsumeproc(t *testing.T) {
	tests := []struct {
		name string
		s    *Service
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.contributeConsumeproc()
		})
	}
}

func TestService_contributeCache(t *testing.T) {
	type args struct {
		vmid  int64
		attrs *space.Attrs
		ctime xtime.Time
		ip    string
	}
	tests := []struct {
		name    string
		s       *Service
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.contributeCache(tt.args.vmid, tt.args.attrs, tt.args.ctime, tt.args.ip, false, false, ""); (err != nil) != tt.wantErr {
				t.Errorf("Service.contributeCache() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
