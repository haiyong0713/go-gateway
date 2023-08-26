package game

import (
	"context"
	"net/url"
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/conf"
)

const (
	_entryHost = "http://internal.api.biligame.net"
	_entryUri  = "/api/game/base/get_by_id"
)

func TestClient(t *testing.T) {
	type Res struct {
		Id      int    `json:"id"`
		Name    string `json:"name"`
		IosLink string `json:"iosLink"`
	}
	cfg := &conf.EntryGameClientConfig{
		Secret: "8f3550e0c04211e79ddafe210a2e3379",
		DesKey: "0a1c3bc7e7144b6d8d5932a7d0d26c28",
	}

	client := NewEntryClient(cfg)
	params := url.Values{}
	params.Add("game_base_id", "1")
	res := &Res{}
	if err := client.Get(context.Background(), _entryHost+_entryUri, params, res); err != nil {
		t.Errorf("client.Get error(%+v)", err)
	}
}
