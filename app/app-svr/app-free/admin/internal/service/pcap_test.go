package service

import (
	"testing"
)

func TestService_tshark(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		s       *Service
		args    args
		wantErr bool
	}{{
		"",
		s,
		args{"/Users/xin/Downloads/Capture20190725105758_1.pcap"},
		false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData, err := tt.s.tshark(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.tshark() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("Service.tshark() = \n%v", gotData)
		})
	}
}

func TestService_matchHost(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name string
		s    *Service
		args args
	}{{
		"",
		s,
		// cn-[a-z0-9]{1,10}-[a-z]{1,10}-(v|live|bcache)-[0-9]{1,10}.bilivideo.com/
		args{"http://cn-hbyc3-dx-v-04.bilivideo.com/"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData := tt.s.matchHost(tt.args.host)
			t.Logf("Service.matchHost() = %v\n", gotData)
		})
	}
}
