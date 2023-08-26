package favorite

import (
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-intl/interface/conf"
)

const (
	_isFav  = "/x/internal/v2/fav/video/favoured"
	_addFav = "/x/internal/v2/fav/video/add"
)

// Dao is favorite dao
type Dao struct {
	client *httpx.Client
	isFav  string
	addFav string
}

// New initial favorite dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPClient),
		isFav:  c.Host.APICo + _isFav,
		addFav: c.Host.APICo + _addFav,
	}
	return
}
