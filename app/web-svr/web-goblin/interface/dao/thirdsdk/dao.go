package thirdsdk

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/web-goblin/interface/conf"
	"go-gateway/app/web-svr/web-goblin/interface/model/thirdsdk"

	"github.com/pkg/errors"
)

const _userBindURL = "/x/admin/archive-push/api/users/binding"

// Dao dao struct.
type Dao struct {
	// config
	c *conf.Config
	// http
	client *bm.Client
	// redis
	redis       *redis.Pool
	userBindURL string
}

// New new dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// config
		c:           c,
		client:      bm.NewClient(c.MgrClient),
		redis:       redis.NewPool(c.MgrRedis.Config),
		userBindURL: c.Host.Mgr + _userBindURL,
	}
	return
}

func (d *Dao) UserBind(ctx context.Context, mid int64) (*thirdsdk.MgrUserBind, error) {
	const _vendorID int64 = 2
	param := url.Values{}
	var res struct {
		Code int                     `json:"code"`
		Data []*thirdsdk.MgrUserBind `json:"data"`
	}
	param.Set("vendorId", strconv.FormatInt(_vendorID, 10))
	param.Set("mid", strconv.FormatInt(mid, 10))
	if err := d.client.Get(ctx, d.userBindURL, "", param, &res); err != nil {
		return nil, errors.Wrap(err, d.userBindURL+"?"+param.Encode())
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.userBindURL+"?"+param.Encode())
	}
	if len(res.Data) == 0 {
		return nil, nil
	}
	for _, val := range res.Data {
		if mid == val.Mid && val.VendorID == _vendorID {
			return val, nil
		}
	}
	return nil, nil
}
