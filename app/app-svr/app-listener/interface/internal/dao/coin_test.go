package dao

import (
	"context"
	"testing"

	"go-common/component/metadata/network"
	"go-common/library/conf/paladin.v2"
)

var coinConf = `
  [CoinHTTP]
    Host = "http://api.bilibili.co"
    [CoinHTTP.Config]
      key     = "0e9b9fcce22daaf1"
      secret  = "76aaccc1e756ac1c5b2ec135e6bd6b39"
      dial    = "50ms"
      timeout = "400ms"
`

func buildCoinDao() *dao {
	d := &dao{}
	daoConf := &paladin.TOML{}
	_ = daoConf.UnmarshalText([]byte(coinConf))
	d.coinHTTP = newBmClient(d, "coinHTTP", daoConf)
	return d
}

func TestCoinNums(t *testing.T) {
	d := buildCoinDao()
	resp, err := d.CoinNums(context.TODO(), CoinNumsOpt{
		Business: ThumbUpBusinessAudio, Oids: []int64{2492587},
		Net: &network.Network{RemoteIP: "1.1.1.1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}
