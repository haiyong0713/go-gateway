package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/time"
	"go-gateway/app/web-svr/web/ecode"
	"go-gateway/app/web-svr/web/interface/model"
)

const (
	_notRobot     = -1
	_rsOk         = "000000"
	_hlKey        = "hl_%s"
	_hlVersionKey = "hl_%s_%d"
	_hdKey        = "hd_%s_%d_%d_%d"
	_hdVersionKey = "hd_%s_%d_%d_%d_%d"
	_newVersion   = 1
)

// HelpList get help list.
func (d *Dao) HelpList(c context.Context, pTypeID string, version int) (data []*model.HelpList, err error) {
	var (
		req    *http.Request
		params = url.Values{}
	)
	params.Set("parentTypeId", pTypeID)
	params.Set("robotFlag", strconv.Itoa(_notRobot))
	listURL := d.helpListURL + "?" + params.Encode()
	if version == _newVersion {
		listURL = d.helpListV2URL + "?" + params.Encode()
	}
	if req, err = http.NewRequest(http.MethodGet, listURL, nil); err != nil {
		log.Error("Help http.NewRequest(%s) error(%v)", listURL, err)
		return
	}
	var res struct {
		Code string            `json:"retCode"`
		Data []*model.HelpList `json:"items"`
	}
	err = d.httpHelp.Do(c, req, &res)
	if err != nil {
		log.Error("Help d.httpHelp.Do(%s) error(%v)", listURL, err)
		return
	}
	if res.Code != _rsOk {
		log.Error("Help dao.httpHelp.Do(%s) error(%v)", listURL, err)
		err = ecode.HelpListError
		return
	}
	data = res.Data
	return
}

func keyHl(pTypeID string, version int) string {
	if version > 0 {
		return fmt.Sprintf(_hlVersionKey, pTypeID, version)
	}
	return fmt.Sprintf(_hlKey, pTypeID)
}

func keyHd(qTypeID string, keyFlag, pn, ps, version int) string {
	if version > 0 {
		return fmt.Sprintf(_hdVersionKey, qTypeID, keyFlag, pn, ps, version)
	}
	return fmt.Sprintf(_hdKey, qTypeID, keyFlag, pn, ps)
}

// SetHlCache set help list  to cache.
func (d *Dao) SetHlCache(c context.Context, pTypeID string, version int, Hl []*model.HelpList) (err error) {
	conn := d.redisBak.Get(c)
	defer conn.Close()
	count := 0
	key := keyHl(pTypeID, version)
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	var bs []byte
	for _, list := range Hl {
		if bs, err = json.Marshal(list); err != nil {
			log.Error("json.Marshal(%v) error (%v)", list, err)
			return
		}
		if err = conn.Send("ZADD", key, list.SortNo, bs); err != nil {
			log.Error("conn.Send(ZADD, %s, %s) error(%v)", key, string(bs), err)
			return
		}
		count++
	}
	if err = conn.Send("EXPIRE", key, d.redisHelpBakExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.redisHelpBakExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// HlCache get help list from cache.
func (d *Dao) HlCache(c context.Context, pTypeID string, version int) (res []*model.HelpList, err error) {
	key := keyHl(pTypeID, version)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZREVRANGE", key, 0, -1))
	if err != nil {
		log.Error("conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() err(%v)", err)
		return
	}
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		list := &model.HelpList{}
		if err = json.Unmarshal(bs, list); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, list)
	}
	return
}

// HelpDetail get help detail.
func (d *Dao) HelpDetail(c context.Context, qTypeID string, keyFlag, pn, ps, version int) (data []*model.HelpDeatil, total int, err error) {
	var (
		req    *http.Request
		params = url.Values{}
	)
	params.Set("questionTypeId", qTypeID)
	params.Set("keyFlag", strconv.Itoa(keyFlag))
	params.Set("keyWords", "")
	params.Set("pageNo", strconv.Itoa(pn))
	params.Set("pageSize", strconv.Itoa(ps))
	params.Set("robotFlag", strconv.Itoa(_notRobot))
	searchURL := d.helpSearchURL + "?" + params.Encode()
	if version == _newVersion {
		searchURL = d.helpSearchV2URL + "?" + params.Encode()
	}
	if req, err = http.NewRequest(http.MethodGet, searchURL, nil); err != nil {
		log.Error("Help http.NewRequest(%s) error(%v)", searchURL, err)
		return
	}
	var res struct {
		Code  string              `json:"retCode"`
		Data  []*model.HelpDeatil `json:"items"`
		Total int                 `json:"totalCount"`
	}
	err = d.httpHelp.Do(c, req, &res)
	if err != nil {
		log.Error("Help d.httpHelp.Do(%s) error(%v)", searchURL, err)
		return
	}
	if res.Code != _rsOk {
		log.Error("Help dao.httpHelp.Do(%s) error(%v)", searchURL, err)
		err = ecode.HelpDetailError
		return
	}
	total = res.Total
	data = res.Data
	return
}

// HelpSearch get help search.
func (d *Dao) HelpSearch(c context.Context, pTypeID, keyWords string, keyFlag, pn, ps, version int) (data []*model.HelpDeatil, total int, err error) {
	var (
		req    *http.Request
		params = url.Values{}
	)
	params.Set("questionTypeId", pTypeID)
	params.Set("keyWords", keyWords)
	params.Set("keyFlag", strconv.Itoa(keyFlag))
	params.Set("pageNo", strconv.Itoa(pn))
	params.Set("pageSize", strconv.Itoa(ps))
	params.Set("robotFlag", strconv.Itoa(_notRobot))
	searchURL := d.helpSearchURL + "?" + params.Encode()
	if version == _newVersion {
		searchURL = d.helpSearchV2URL + "?" + params.Encode()
	}
	if req, err = http.NewRequest(http.MethodGet, searchURL, nil); err != nil {
		log.Error("Help http.NewRequest(%s) error(%v)", searchURL, err)
		return
	}
	var res struct {
		Code  string              `json:"retCode"`
		Data  []*model.HelpDeatil `json:"items"`
		Total int                 `json:"totalCount"`
	}
	err = d.httpHelp.Do(c, req, &res)
	if err != nil {
		log.Error("Help d.httpHelp.Do(%s) error(%v)", searchURL, err)
		return
	}
	if res.Code != _rsOk {
		log.Error("Help dao.httpHelp.Do(%s) error(%v)", searchURL, err)
		err = ecode.HelpSearchError
		return
	}
	total = res.Total
	data = res.Data
	return
}

// SetDetailCache  set help detail  to cache.
func (d *Dao) SetDetailCache(c context.Context, qTypeID string, keyFlag, pn, ps, version, total int, data []*model.HelpDeatil) (err error) {
	conn := d.redisBak.Get(c)
	defer conn.Close()
	count := 0
	key := keyHd(qTypeID, keyFlag, pn, ps, version)
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	var bs []byte
	for _, detail := range data {
		if bs, err = json.Marshal(detail); err != nil {
			log.Error("json.Marshal(%v) error (%v)", detail, err)
			return
		}
		if err = conn.Send("ZADD", key, combineHd(detail.UpdateTime, total), bs); err != nil {
			log.Error("conn.Send(ZADD, %s, %s) error(%v)", key, string(bs), err)
			return
		}
		count++
	}
	if err = conn.Send("EXPIRE", key, d.redisHelpBakExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.redisHelpBakExpire, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// DetailCache  get help detail  to cache.
func (d *Dao) DetailCache(c context.Context, qTypeID string, keyFlag, pn, ps, version int) (res []*model.HelpDeatil, count int, err error) {
	conn := d.redisBak.Get(c)
	defer conn.Close()
	key := keyHd(qTypeID, keyFlag, pn, ps, version)
	values, err := redis.Values(conn.Do("ZREVRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() err(%v)", err)
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		detail := &model.HelpDeatil{}
		if err = json.Unmarshal(bs, detail); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, detail)
	}
	count = fromHd(num)
	return
}

func fromHd(i int64) int {
	return int(i & 0xffff)
}

func combineHd(create time.Time, count int) int64 {
	return create.Time().Unix()<<16 | int64(count)
}
