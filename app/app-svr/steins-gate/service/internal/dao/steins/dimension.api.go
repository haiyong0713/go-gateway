package steins

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-common/library/ecode"
	"go-common/library/xstr"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/pkg/errors"
)

const (
	_bvcLimit = 50
)

// BvcDimension calls BVC's api to get HD dimension
func (d *Dao) BvcDimension(c context.Context, cid int64) (dimension *model.DimensionInfo, err error) {
	var (
		resp = new(model.DimensionReply)
		req  *http.Request
	)
	if req, err = http.NewRequest("GET", d.bvcDimensionURL+d.bvcSign(cid), nil); err != nil {
		err = errors.Wrapf(err, "url %s", d.bvcDimensionURL+d.bvcSign(cid))
		return
	}
	if err = d.httpVideoClient.Do(c, req, resp); err != nil {
		err = errors.Wrapf(err, "url %s", d.bvcDimensionURL+d.bvcSign(cid))
		return
	}
	if resp.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(resp.Code), "message %s, url %s", resp.Message, d.bvcDimensionURL+d.bvcSign(cid))
		return
	}
	if len(resp.Info) == 0 {
		err = errors.Wrapf(ecode.NothingFound, "url %s", d.bvcDimensionURL+d.bvcSign(cid))
		return
	}
	dimension = resp.Info[0]
	return
}

func (d *Dao) BvcDimensions(c context.Context, oriCids []int64) (dimensions map[int64]*model.DimensionInfo, err error) {
	var (
		dimCids = make(map[int64]struct{})
		newCids []int64
		mutex   = sync.Mutex{}
	)
	for _, cid := range oriCids { // 过滤重复的cid
		if _, ok := dimCids[cid]; ok {
			continue
		}
		newCids = append(newCids, cid)
		dimCids[cid] = struct{}{}
	}
	cidPces := splitIDs(newCids, _bvcLimit)
	eg := errgroup.WithContext(c)
	dimensions = make(map[int64]*model.DimensionInfo, len(newCids))
	for _, pce := range cidPces {
		tmp := pce
		eg.Go(func(c context.Context) (err error) {
			splitDims, err := d.bvcDimensions(c, tmp)
			mutex.Lock()
			for cid, dim := range splitDims {
				dimensions[cid] = dim
			}
			mutex.Unlock()
			return
		})
	}
	err = eg.Wait()
	return
}

// BvcDimension calls BVC's api to get HD dimension
func (d *Dao) bvcDimensions(c context.Context, cids []int64) (dimensions map[int64]*model.DimensionInfo, err error) {
	if len(cids) == 0 {
		return
	}
	var (
		resp = new(model.DimensionsReply)
		req  *http.Request
	)
	if req, err = http.NewRequest("GET", d.bvcDimensionsURL+d.bvcSignBatch(cids), nil); err != nil {
		err = errors.Wrapf(err, "url %s", d.bvcDimensionsURL+d.bvcSignBatch(cids))
		return
	}
	if err = d.httpVideoClient.Do(c, req, resp); err != nil {
		err = errors.Wrapf(err, "url %s", d.bvcDimensionsURL+d.bvcSignBatch(cids))
		return
	}
	if resp.Code != ecode.OK.Code() {
		err = errors.Wrapf(ecode.Int(resp.Code), "message %s, url %s", resp.Message, d.bvcDimensionsURL+d.bvcSignBatch(cids))
		return
	}
	if len(resp.Info) == 0 {
		err = errors.Wrapf(ecode.NothingFound, "url %s", d.bvcDimensionsURL+d.bvcSignBatch(cids))
		return
	}
	dimensions = make(map[int64]*model.DimensionInfo)
	for k, v := range resp.Info {
		if len(v) == 0 {
			continue
		}
		if v[0] == nil {
			continue
		}
		dimensions[k] = v[0]
	}
	return
}

func (d *Dao) bvcSign(cid int64) (query string) {
	return d.bvcSignTool(fmt.Sprintf("%s=%d", "cid", cid))
}

func (d *Dao) bvcSignBatch(cids []int64) (query string) {
	return d.bvcSignTool(fmt.Sprintf("%s=%s", "cids", xstr.JoinInts(cids)))
}

func (d *Dao) bvcSignTool(cidQuery string) (query string) {
	kvs := []string{cidQuery, fmt.Sprintf("%s=%d", "timestamp", time.Now().Unix())}
	kvsStr := strings.Join(kvs, "&")
	mh := md5.Sum([]byte(kvsStr + "&key=" + d.c.Bvc.Key))
	sign := hex.EncodeToString(mh[:])
	return "?" + kvsStr + "&sign=" + sign
}

func splitIDs(ids []int64, ps int) (pces [][]int64) {
	if len(ids) == 0 {
		return
	}
	var nbPce int
	if len(ids)%ps == 0 {
		nbPce = len(ids) / ps
	} else {
		nbPce = len(ids)/ps + 1
	}
	for i := 0; i < nbPce; i++ {
		if end := (i + 1) * ps; end > len(ids) {
			pces = append(pces, ids[i*ps:])
		} else {
			pces = append(pces, ids[i*ps:(i+1)*ps])
		}
	}
	return

}
