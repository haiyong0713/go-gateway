package common

import (
	"strings"
	"testing"

	pb "go-gateway/app/app-svr/app-gw/management/api"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster"
	metadata2 "go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"

	"github.com/stretchr/testify/assert"
)

func TestPathMetaAppendBreakerAPIs(t *testing.T) {
	var (
		bas = []*pb.BreakerAPI{
			{
				Api:     "/x/tv/homepage",
				Ratio:   100,
				Reason:  "test",
				Action:  &pb.BreakerAction{Action: nil},
				Enable:  true,
				Node:    "main.app-svr",
				Gateway: "tv-gateway",
			},
			{
				Api:     "/x/tv/hotword",
				Ratio:   100,
				Reason:  "test",
				Action:  &pb.BreakerAction{Action: nil},
				Enable:  true,
				Node:    "main.app-svr",
				Gateway: "tv-gateway",
			},
			{
				Api:     "/x/tv/third/ugc/archive",
				Ratio:   100,
				Reason:  "test",
				Action:  &pb.BreakerAction{Action: nil},
				Enable:  true,
				Node:    "main.app-svr",
				Gateway: "tv-gateway",
			},
			{
				Api:     "/x/tv/third/ugc/video",
				Ratio:   100,
				Reason:  "test",
				Action:  &pb.BreakerAction{Action: nil},
				Enable:  true,
				Node:    "main.app-svr",
				Gateway: "tv-gateway",
			},
		}
		pms = []*sdk.PathMeta{
			{
				Pattern:    "= /x/tv/third/live/list",
				ClientInfo: metadata2.ClientInfo{AppID: "tv.interface", Endpoint: "discovery://tv.interface"},
			},
			{
				Pattern:    "= /x/tv/ugc/view",
				ClientInfo: metadata2.ClientInfo{AppID: "tv.interface", Endpoint: "discovery://tv.interface"},
			},
			{
				Pattern:    "~ ^/x/tv",
				ClientInfo: metadata2.ClientInfo{AppID: "tv.interface", Endpoint: "discovery://tv.interface"},
			},
		}
	)
	p1 := PathMetaAppendBreakerAPIs(bas)
	out, err := p1(pms)
	assert.Equal(t, err, nil)
	for _, v := range out {
		if strings.HasPrefix(v.Pattern, "= ") {
			assert.Equal(t, v.GetMatcher().Name(), "exactlyMatcher")
		}
		if strings.HasPrefix(v.Pattern, "~ ") {
			assert.Equal(t, v.GetMatcher().Name(), "regexMatcher")
		}
	}
}

func TestCfgDigist(t *testing.T) {
	var (
		bas = []*pb.BreakerAPI{
			{
				Api:     "/x/tv/third/ugc/archive",
				Ratio:   100,
				Reason:  "test",
				Action:  &pb.BreakerAction{Action: nil},
				Enable:  true,
				Node:    "main.app-svr",
				Gateway: "tv-gateway",
			},
		}
		pms = []*sdk.PathMeta{
			{
				Pattern:    "= /x/tv/third/live/list",
				ClientInfo: metadata2.ClientInfo{AppID: "tv.interface", Endpoint: "discovery://tv.interface"},
			},
			{
				Pattern:    "~ ^/x/tv",
				ClientInfo: metadata2.ClientInfo{AppID: "tv.interface", Endpoint: "discovery://tv.interface"},
			},
		}
	)
	p1 := PathMetaAppendBreakerAPIs(bas)
	out, err := p1(pms)
	assert.Equal(t, err, nil)
	cfg := &sdk.Config{DynPath: out}
	templet := &sdk.Config{
		DynPath: []*sdk.PathMeta{
			{
				Pattern:           "= /x/tv/third/ugc/archive",
				ClientInfo:        metadata2.ClientInfo{AppID: "tv.interface", Endpoint: "discovery://tv.interface"},
				BackupRetryOption: sdk.BackupRetryOption{Ratio: 100},
			},
			{
				Pattern:    "= /x/tv/third/live/list",
				ClientInfo: metadata2.ClientInfo{AppID: "tv.interface", Endpoint: "discovery://tv.interface"},
			},
			{
				Pattern:    "~ ^/x/tv",
				ClientInfo: metadata2.ClientInfo{AppID: "tv.interface", Endpoint: "discovery://tv.interface"},
			},
		},
	}
	assert.Equal(t, cfg.Digest(), templet.Digest())
}
