package recommend

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-gateway/app/app-svr/app-car/job/conf"

	"github.com/pkg/errors"
)

const (
	_rankAllAppURL       = "/data/rank/all-app.json"
	_hotHeTongtabcardURL = "/data/rank/reco-app-remen-card-%d.json"
)

type Dao struct {
	c      *conf.Config
	client *bm.Client
	// url
	rankAllAppURL string
	hotURL        string
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPData, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		// url
		rankAllAppURL: c.Host.Data + _rankAllAppURL,
		hotURL:        c.Host.Data + _hotHeTongtabcardURL,
	}
	return d
}

// RankAppAll all rank
func (d *Dao) RankAppAll(c context.Context) ([]int64, error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	if err := d.client.Get(c, d.rankAllAppURL, "", nil, &res); err != nil {
		log.Error("recommend All rank hots url(%s) error(%v)", d.rankAllAppURL, err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		log.Error("recommend All rank hots url(%s) error(%v)", d.rankAllAppURL, res.Code)
		return nil, fmt.Errorf("recommend All rank hots api response code(%v)", res.Code)
	}
	aids := make([]int64, 0, len(res.List))
	for _, list := range res.List {
		if list.Aid != 0 {
			aids = append(aids, list.Aid)
		}
	}
	return aids, nil
}

func (d *Dao) HotHeTongTabCard(c context.Context, i int) ([]int64, error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			ID   int64  `json:"id"`
			Goto string `json:"goto"`
		} `json:"list"`
	}
	if err := d.client.Get(c, fmt.Sprintf(d.hotURL, i), "", nil, &res); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(d.hotURL, i))
	}
	if res.Code != 0 {
		return nil, errors.Wrap(ecode.Int(res.Code), fmt.Sprintf("code(%d)", res.Code))
	}
	aids := make([]int64, 0, len(res.List))
	for _, v := range res.List {
		if v.Goto != "av" || v.ID == 0 {
			continue
		}
		aids = append(aids, v.ID)
	}
	return aids, nil
}
