package dao

import (
	"context"
	"net/http"

	"go-common/library/xstr"
)

func (d *dao) SetLivePlaybackWhitelist(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodGet, d.livePlaybackWhitelistURL, nil)
	if err != nil {
		return err
	}
	bs, err := d.httpClient.Raw(ctx, req)
	if err != nil {
		return err
	}
	vals, err := xstr.SplitInts(string(bs))
	if err != nil {
		return err
	}
	whitelist := map[int64]struct{}{}
	for _, val := range vals {
		whitelist[val] = struct{}{}
	}
	return d.setCacheLivePlaybackWhitelist(ctx, whitelist)
}
