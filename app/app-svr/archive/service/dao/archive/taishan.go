package archive

import (
	"context"
	"encoding/json"

	"go-common/library/database/taishan"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"

	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/archive/service/model"
	"go-gateway/app/app-svr/archive/service/model/archive"

	"github.com/pkg/errors"
)

func (d *Dao) newGetReq(key string) *taishan.GetReq {
	req := &taishan.GetReq{
		Table: d.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: d.Taishan.tableCfg.Token,
		},
		Record: &taishan.Record{
			Key: []byte(key),
		},
	}
	return req
}

func (d *Dao) newBatchGetReq(keys []string) *taishan.BatchGetReq {
	req := &taishan.BatchGetReq{
		Table: d.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: d.Taishan.tableCfg.Token,
		},
	}
	records := make([]*taishan.Record, 0, len(keys))
	for _, key := range keys {
		records = append(records, &taishan.Record{
			Key: []byte(key),
		})
	}
	req.Records = records
	return req
}

func checkRecord(r *taishan.Record) error {
	if r == nil || r.Status == nil {
		return errors.New("record is nil")
	}
	errNo := int32(404)
	if r.Status.ErrNo == errNo {
		return ecode.NothingFound
	}
	if r.Columns == nil || len(r.Columns) == 0 || r.Columns[0] == nil {
		return errors.New("Record.Colums is nil")
	}
	return nil
}

func (d *Dao) getFromTaishan(c context.Context, key string) ([]byte, error) {
	req := d.newGetReq(key)
	resp, err := d.Taishan.client.Get(c, req)
	if err != nil {
		return nil, err
	}
	if err = checkRecord(resp.Record); err != nil {
		return nil, err
	}
	return resp.Record.Columns[0].Value, nil
}

func (d *Dao) batchGetFromTaishan(c context.Context, keys []string) ([][]byte, error) {
	req := d.newBatchGetReq(keys)
	resp, err := d.Taishan.client.BatchGet(c, req)
	if err != nil {
		return nil, err
	}
	if resp.AllFailed {
		var err error
		for _, rec := range resp.Records {
			if rec.Status.ErrNo != 0 {
				err = errors.Wrapf(err, "failed to batch get from taishan, key:%s err:%+v", rec.Key, rec.Status)
			}
		}
		return nil, err
	}
	var bss [][]byte
	for _, rec := range resp.Records {
		if rec.Status.ErrNo != 0 {
			log.Error("failed to get from taishan in batch key:%s err:%+v", rec.Key, rec.Status)
			continue
		}
		if rec.Columns == nil || rec.Columns[0] == nil {
			log.Error("failed to get from taishan in batch, Columns is nil key:%s colums:%+v", rec.Key, rec.Columns)
			continue
		}
		bss = append(bss, rec.Columns[0].Value)
	}
	if len(bss) == 0 {
		return nil, ecode.NothingFound
	}
	return bss, nil
}

func (d *Dao) getArcFromTaishan(c context.Context, aid int64) (*api.Arc, error) {
	bs, err := d.getFromTaishan(c, model.ArcKey(aid))
	if err != nil {
		return nil, err
	}
	a := &api.Arc{}
	if err = a.Unmarshal(bs); err != nil {
		return nil, err
	}
	return a, nil
}

func (d *Dao) batchGetArcFromTaishan(c context.Context, aids []int64) (map[int64]*api.Arc, []int64, error) {
	var (
		missed []int64
		keys   []string
		keyMap = make(map[int64]struct{}, len(aids))
	)
	for _, aid := range aids {
		if _, ok := keyMap[aid]; ok {
			continue
		}
		keyMap[aid] = struct{}{}
		keys = append(keys, model.ArcKey(aid))
	}
	bss, err := d.batchGetFromTaishan(c, keys)
	if err != nil {
		return nil, aids, err
	}
	am := make(map[int64]*api.Arc, len(bss))
	for _, bs := range bss {
		a := &api.Arc{}
		if err := a.Unmarshal(bs); err != nil {
			continue
		}
		am[a.Aid] = a
		delete(keyMap, a.Aid)
	}
	for aid := range keyMap {
		missed = append(missed, aid)
	}
	return am, missed, nil
}

func (d *Dao) getPagesFromTaishan(c context.Context, aid int64) ([]*api.Page, error) {
	bs, err := d.getFromTaishan(c, model.PageKey(aid))
	if err != nil {
		return nil, err
	}
	vs := &api.AidVideos{}
	if err = vs.Unmarshal(bs); err != nil {
		return nil, err
	}
	return vs.Pages, nil
}

func (d *Dao) batchGetPagesFromTaishan(c context.Context, aids []int64) (map[int64][]*api.Page, []int64, error) {
	var (
		missed []int64
		keys   []string
		keyMap = make(map[int64]struct{}, len(aids))
	)
	for _, aid := range aids {
		if _, ok := keyMap[aid]; ok {
			continue
		}
		keyMap[aid] = struct{}{}
		keys = append(keys, model.PageKey(aid))
	}
	bss, err := d.batchGetFromTaishan(c, keys)
	if err != nil {
		return nil, aids, err
	}
	apm := make(map[int64][]*api.Page, len(bss))
	for _, bs := range bss {
		a := &api.AidVideos{}
		if err := a.Unmarshal(bs); err != nil {
			continue
		}
		apm[a.Aid] = a.Pages
		delete(keyMap, a.Aid)
	}
	for aid := range keyMap {
		missed = append(missed, aid)
	}
	return apm, missed, nil
}

func (d *Dao) getSimpleArcFromTaishan(c context.Context, aid int64) (*api.SimpleArc, error) {
	bs, err := d.getFromTaishan(c, model.SimpleArcKey(aid))
	if err != nil {
		return nil, err
	}
	sa := &api.SimpleArc{}
	if err = sa.Unmarshal(bs); err != nil {
		return nil, err
	}
	return sa, nil
}

func (d *Dao) batchGetSimpleArcFromTaishan(c context.Context, aids []int64) (map[int64]*api.SimpleArc, error) {
	var (
		keys   []string
		keyMap = make(map[int64]struct{}, len(aids))
	)
	for _, aid := range aids {
		if _, ok := keyMap[aid]; ok {
			continue
		}
		keyMap[aid] = struct{}{}
		keys = append(keys, model.SimpleArcKey(aid))
	}
	bss, err := d.batchGetFromTaishan(c, keys)
	if err != nil {
		return nil, err
	}
	am := make(map[int64]*api.SimpleArc, len(bss))
	for _, bs := range bss {
		a := &api.SimpleArc{}
		if err := a.Unmarshal(bs); err != nil {
			continue
		}
		am[a.Aid] = a
	}
	return am, nil
}

func (d *Dao) getDescFromTaishanV2(c context.Context, aid int64) (*archive.Addit, error) {
	bs, err := d.getFromTaishan(c, model.DescKeyV2(aid))
	if err != nil {
		return nil, err
	}
	addit := &archive.Addit{}
	if err = json.Unmarshal(bs, addit); err != nil {
		return nil, err
	}
	return addit, nil
}

func (d *Dao) getVideoFromTaishan(c context.Context, aid, cid int64) (*api.Page, error) {
	bs, err := d.getFromTaishan(c, model.VideoKey(aid, cid))
	if err != nil {
		return nil, err
	}
	p := &api.Page{}
	if err = p.Unmarshal(bs); err != nil {
		return nil, err
	}
	return p, nil
}

func (d *Dao) batchGetVideoFromTaishan(c context.Context, aidCids map[int64][]int64) (map[int64][]*api.Page, []int64, error) {
	keys := []string{}
	cached := make(map[int64][]*api.Page)
	tmpCidAid := make(map[int64]int64)

	for aid, cids := range aidCids {
		if aid == 0 || len(cids) == 0 {
			continue
		}
		for _, cid := range cids {
			if cid == 0 {
				continue
			}
			keys = append(keys, model.VideoKey(aid, cid))
			tmpCidAid[cid] = aid
		}
	}

	if len(keys) == 0 {
		missCids := []int64{}
		for key := range tmpCidAid {
			missCids = append(missCids, key)
		}
		d.infoProm.Incr("batchGetVideoFromTaishan_no_args")
		return cached, missCids, nil
	}

	bss, err := d.batchGetFromTaishan(c, keys)
	if err != nil {
		log.Error("batchGetVideoFromTaishan keys(%+v) error(%+v)", keys, err)
		return nil, nil, err
	}

	cachedCids := sets.NewInt64()
	for index, bs := range bss {
		if bs == nil {
			continue
		}
		vs := &api.Page{}
		if err = vs.Unmarshal(bs); err != nil {
			log.Error("batchGetVideoFromTaishan aidCids(%+v) index(%+v) Unmarshal error(%+v)", aidCids, index, err)
			continue
		}
		if aid, ok := tmpCidAid[vs.Cid]; ok {
			cached[aid] = append(cached[aid], vs)
			cachedCids.Insert(vs.Cid)
		}
	}

	missCids := make([]int64, 0, len(tmpCidAid))
	for aid, cids := range aidCids {
		if aid == 0 || len(cids) == 0 {
			continue
		}
		for _, cid := range cids {
			if cid == 0 {
				continue
			}
			if !cachedCids.Has(cid) {
				missCids = append(missCids, cid)
			}
		}
	}
	return cached, missCids, nil
}

func (d *Dao) batchGetDescV2FromTaishan(c context.Context, aids []int64) (map[int64]*api.DescriptionReply, []int64) {
	keys := []string{}
	miss := make(map[int64]interface{})
	noFoundCacheAid := []int64{}
	descMulti := make(map[int64]*api.DescriptionReply)
	//拼装key
	for _, aid := range aids {
		if aid == 0 {
			continue
		}
		keys = append(keys, model.DescKeyV2(aid))
		miss[aid] = struct{}{}
	}
	//批量从taishan获取数据
	bss, err := d.batchGetFromTaishan(c, keys)
	if err != nil {
		return descMulti, aids
	}
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		desc := &archive.Addit{}
		if err = json.Unmarshal(bs, desc); err != nil {
			log.Error("batchGetDescFromTaishan Unmarshal error(%+v)", err)
			continue
		}
		//如果存在则从miss中删除，否则为没有找到缓存数据
		delete(miss, desc.Aid)
		//赋值
		descMulti[desc.Aid] = &api.DescriptionReply{
			Desc:        desc.Description,
			DescV2Parse: d.GetDescV2Params(desc.DescV2),
		}
	}
	if len(miss) > 0 {
		for aid := range miss {
			noFoundCacheAid = append(noFoundCacheAid, aid)
		}
	}
	return descMulti, noFoundCacheAid
}

func (d *Dao) setTaishan(c context.Context, key, val []byte) error {
	req := &taishan.PutReq{
		Table: d.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: d.Taishan.tableCfg.Token,
		},
		Record: &taishan.Record{
			Key: key,
			Columns: []*taishan.Column{
				{
					Value: val,
				},
			},
		},
	}
	resp, err := d.Taishan.client.Put(c, req)
	if err != nil {
		return err
	}
	if resp.GetStatus() == nil {
		return errors.New("response status is invalid")
	}
	if resp.GetStatus().ErrNo != 0 {
		return errors.Errorf("key: %+v, errno: %+v, errmsg: %+v", string(key), resp.Status.ErrNo, resp.Status.Msg)
	}
	return nil
}

func (d *Dao) batchPutTaishan(c context.Context, kvMap map[string][]byte) error {
	if len(kvMap) == 0 {
		return nil
	}

	req := &taishan.BatchPutReq{
		Table: d.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: d.Taishan.tableCfg.Token,
		},
	}

	records := make([]*taishan.Record, 0, len(kvMap))
	for key, value := range kvMap {
		records = append(records, &taishan.Record{
			Key: []byte(key),
			Columns: []*taishan.Column{
				{
					Value: value,
				},
			},
		})
	}
	req.Records = records

	resp, err := d.Taishan.client.BatchPut(c, req)
	if err != nil {
		log.Error("failed to batch set taishan kvMap:%+v err:%+v", kvMap, err)
		return err
	}

	for _, rec := range resp.Records {
		if rec.Status.ErrNo != 0 {
			err = errors.Wrapf(err, "failed to batch set taishan, key:%s err:%+v", rec.Key, rec.Status)
		}
	}

	return err
}

func (d *Dao) delTaishan(c context.Context, key, val []byte) error {
	req := &taishan.DelReq{
		Table: d.Taishan.tableCfg.Table,
		Auth: &taishan.Auth{
			Token: d.Taishan.tableCfg.Token,
		},
		Record: &taishan.Record{
			Key: key,
			Columns: []*taishan.Column{
				{
					Value: val,
				},
			},
		},
	}
	resp, err := d.Taishan.client.Del(c, req)
	if err != nil {
		return err
	}
	if resp.GetStatus() == nil {
		return errors.New("response status is invalid")
	}
	if resp.GetStatus().ErrNo != 0 {
		return errors.Errorf("key: %+v, errno: %+v, errmsg: %+v", string(key), resp.Status.ErrNo, resp.Status.Msg)
	}
	return nil
}

func (d *Dao) batchGetRedirectTaishan(c context.Context, aids []int64) (map[int64]*api.RedirectPolicy, []int64) {
	keys := []string{}
	miss := make(map[int64]interface{})
	noFoundCacheAid := []int64{}
	res := make(map[int64]*api.RedirectPolicy)
	//拼装key
	for _, aid := range aids {
		keys = append(keys, model.RedirectKey(aid))
		miss[aid] = struct{}{}
	}
	//批量从taishan获取数据
	bss, err := d.batchGetFromTaishan(c, keys)
	if err != nil {
		return res, aids
	}
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		redirect := &archive.ArcRedirect{}
		if err = json.Unmarshal(bs, redirect); err != nil {
			log.Error("batchGetRedirectTaishan Unmarshal error(%+v)", err)
			continue
		}
		//如果存在则从miss中删除，否则为没有找到缓存数据
		delete(miss, redirect.Aid)
		//赋值
		res[redirect.Aid] = &api.RedirectPolicy{
			Aid:            redirect.Aid,
			RedirectType:   redirect.RedirectType,
			RedirectTarget: redirect.RedirectTarget,
			PolicyType:     redirect.PolicyType,
			PolicyId:       redirect.PolicyId,
		}
	}
	if len(miss) > 0 {
		for aid := range miss {
			noFoundCacheAid = append(noFoundCacheAid, aid)
		}
	}
	return res, noFoundCacheAid
}

func (d *Dao) setVideoToTaishan(c context.Context, aid, cid int64, p *api.Page) error {
	val, err := p.Marshal()
	if err != nil {
		log.Error("d.setVideoToTaishan Marshal error(%v)", err)
		return err
	}
	if err = d.setTaishan(c, []byte(model.VideoKey(aid, cid)), val); err != nil {
		log.Error("d.setVideoToTaishan error(%v)", err)
		return err
	}
	return nil
}

func (d *Dao) batchPutVideoToTaishan(c context.Context, vs map[int64][]*api.Page) {
	if len(vs) == 0 {
		return
	}

	kvMap := make(map[string][]byte, len(vs))
	for aid, values := range vs {
		for _, v := range values {
			res, err := v.Marshal()
			if err != nil {
				log.Error("d.batchPutVideoToTaishan Marshal(%+v) error(%+v)", v, err)
				continue
			}
			kvMap[model.VideoKey(aid, v.Cid)] = res
		}
	}

	if err := d.batchPutTaishan(c, kvMap); err != nil {
		log.Error("d.batchPutVideoToTaishan(%+v) error(%+v)", kvMap, err)
	}
}
