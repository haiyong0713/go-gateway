package playurl

import (
	"context"
	"flag"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-view/interface/conf"

	"github.com/stretchr/testify/assert"
)

var (
	d *Dao
)

func init() {
	flag.Set("conf", "../../cmd/app-view-test.toml")
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func ctx() context.Context {
	return context.Background()
}

func TestDao_PlayOnlineTotal(t *testing.T) {
	_, canShow := d.PlayOnlineTotal(ctx(), 520058607, 10281697)
	assert.Equal(t, canShow, true)
}
