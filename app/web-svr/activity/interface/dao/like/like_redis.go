package like

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/tool"

	"github.com/pkg/errors"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

func keyLikeCheck(mid, sid int64) string {
	return fmt.Sprintf("like_check_%d_%d", sid, mid)
}

func keyActOnGoing(typeIds []int64) string {
	return fmt.Sprintf("act_subject_on_going_%s", xstr.JoinInts(typeIds))
}

func entUpRankKey(sid int64) string {
	return fmt.Sprintf("ent_up_%d", sid)
}

func likeTypeCountKey(sid int64) string {
	return fmt.Sprintf("like_type_cnt_%d", sid)
}

func keySubjectRules(sid int64) string {
	return fmt.Sprintf("sub_rule_%d", sid)
}

func activityArchivesKey(sid, mid int64) string {
	return fmt.Sprintf("act_arc_%d_%d", sid, mid)
}

func actRelationInfoKey(id int64) string {
	return fmt.Sprintf("act_relation_info_%d", id)
}

func hotActRelationInfoKey() string {
	return "hot_act_relation_info"
}

func hotActSubjectInfoKey() string {
	return "hot_act_subject_info"
}

func hotActSubjectReserveIDsInfoKey() string {
	return "hot_act_subject_reserve_ids_info"
}

func keyYellowGreenPeriod(yellowSid, GreenSid int64) string {
	return fmt.Sprintf("yingyuan_vote_%d_%d", yellowSid, GreenSid)
}

// CacheLikeCheck .
func (d *Dao) CacheLikeCheck(c context.Context, mid, sid int64) (res *like.Item, err error) {
	var (
		bs  []byte
		key = keyLikeCheck(mid, sid)
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	res = new(like.Item)
	if err = json.Unmarshal(bs, res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheLikeCheck .
func (d *Dao) AddCacheLikeCheck(c context.Context, mid int64, data *like.Item, sid int64) (err error) {
	var (
		bs   []byte
		key  = keyLikeCheck(mid, sid)
		conn = component.GlobalRedis.Conn(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s,%s) error(%v)", key, string(bs), err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.likeTotalExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

func (d *Dao) CacheActSubjectsOnGoing(c context.Context, typeIds []int64) (res []int64, err error) {
	var (
		key = keyActOnGoing(typeIds)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheActSubjectsOnGoing(%s) return nil", key)
		} else {
			log.Error("CacheActSubjectsOnGoing conn.Do(GET key(%s)) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (d *Dao) AddCacheActSubjectsOnGoing(c context.Context, typeIds []int64, list []int64) (err error) {
	var (
		key = keyActOnGoing(typeIds)
		bs  []byte
	)
	if bs, err = json.Marshal(list); err != nil {
		log.Error("AddCacheActSubjectsOnGoing json.Marshal(%v) error (%v)", list, err)
		return
	}
	if _, err = component.GlobalRedis.Do(c, "SETEX", key, d.onGoingActExpire, bs); err != nil {
		log.Error("conn.Do(SETEX, %s, %d, %d)", key, d.onGoingActExpire, bs)
	}
	return
}

func (d *Dao) DelCacheSubjectRulesBySid(c context.Context, sid int64) error {
	key := keySubjectRules(sid)
	_, err := component.GlobalRedis.Do(c, "DEL", key)
	return err
}

func (d *Dao) CacheSubjectRulesBySid(c context.Context, sid int64) ([]*like.SubjectRule, error) {
	key := keySubjectRules(sid)
	bs, err := redis.Bytes(component.GlobalRedis.Do(c, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheSubjectRulesBySid(%s) return nil", key)
			return nil, nil
		}
		log.Error("CacheSubjectRulesBySid conn.Do(GET key(%s)) error(%v)", key, err)
		return nil, err
	}
	var res []*like.SubjectRule
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return res, nil
}

func (d *Dao) AddCacheSubjectRulesBySid(c context.Context, sid int64, data []*like.SubjectRule) error {
	key := keySubjectRules(sid)
	bs, err := json.Marshal(data)
	if err != nil {
		log.Error("AddCacheSubjectRulesBySid json.Marshal(%v) error (%v)", data, err)
		return err
	}
	if _, err = component.GlobalRedis.Do(c, "SETEX", key, d.subRuleExpire, bs); err != nil {
		log.Error("AddCacheSubjectRulesBySid conn.Do(SETEX, %s, %d, %d)", key, d.subRuleExpire, bs)
	}
	return nil
}

func (d *Dao) CacheSubjectRulesBySids(c context.Context, sids []int64) (map[int64][]*like.SubjectRule, error) {
	args := redis.Args{}
	for _, sid := range sids {
		args = args.Add(keySubjectRules(sid))
	}
	bss, err := redis.ByteSlices(component.GlobalRedis.Do(c, "MGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return nil, nil
		}
		log.Error("CacheSubjectRulesBySids conn.Do(MGET) error(%v)", err)
		return nil, err
	}
	res := make(map[int64][]*like.SubjectRule)
	for _, bs := range bss {
		if bs == nil {
			continue
		}
		var rules []*like.SubjectRule
		if err = json.Unmarshal(bs, &rules); err != nil {
			log.Error("CacheSubjectRulesBySids json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		if len(rules) > 0 && rules[0] != nil && rules[0].Sid > 0 {
			res[rules[0].Sid] = rules
		}
	}
	return res, nil
}

func (d *Dao) AddCacheSubjectRulesBySids(c context.Context, data map[int64][]*like.SubjectRule) error {
	if len(data) == 0 {
		return nil
	}
	args := redis.Args{}
	var keys []string
	for sid, rules := range data {
		bs, err := json.Marshal(rules)
		if err != nil {
			log.Error("AddCacheSubjectRulesBySids json.Marshal(%v) error(%v)", rules, err)
			continue
		}
		key := keySubjectRules(sid)
		keys = append(keys, key)
		args = args.Add(key).Add(bs)
	}
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	if err := conn.Send("MSET", args...); err != nil {
		log.Error("AddCacheSubjectRulesBySids MSET error(%v)", err)
		return err
	}
	count := 1
	for _, v := range keys {
		if err := conn.Send("EXPIRE", v, d.subRuleExpire); err != nil {
			log.Error("AddCacheSubjectRulesBySids conn.Send(Expire, %s, %d) error(%v)", v, d.subRuleExpire, err)
			return err
		}
		count++
	}
	if err := conn.Flush(); err != nil {
		log.Error("AddCacheSubjectRulesBySids Flush error(%v)", err)
		return err
	}
	for i := 0; i < count; i++ {
		if _, err := conn.Receive(); err != nil {
			log.Error("AddCacheSubjectRulesBySids conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) EntCache(c context.Context, sid int64, start, end int64) (list []*like.LidLikeRes, err error) {
	key := entUpRankKey(sid)
	values, err := redis.Values(component.GlobalRedisStore.Do(c, "ZREVRANGE", key, start, end, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZREVRANGE,%s,%d,%d) error(%v)", key, 0, -1, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var (
		bs []byte
		ts int64
	)
	for len(values) > 0 {
		if values, err = redis.Scan(values, &bs, &ts); err != nil {
			log.Error("redis.Scan() error(%v)", err)
			return
		}
		item := new(like.LidLikeRes)
		if e := json.Unmarshal(bs, &item); e != nil {
			log.Error("json.Unmarshal bs(%s) error(%v)", string(bs), e)
			continue
		}
		list = append(list, item)
	}
	return
}

func (dao *Dao) CacheLikeTypeCount(ctx context.Context, sid int64) (map[int64]int64, error) {
	key := likeTypeCountKey(sid)
	values, err := redis.Int64Map(component.GlobalRedisStore.Do(ctx, "HGETALL", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		err = errors.Wrapf(err, "conn.Do(HGETALL) key(%s)", key)
		return nil, err
	}
	res := make(map[int64]int64, len(values))
	for k, v := range values {
		field, e := strconv.ParseInt(k, 10, 64)
		if e != nil {
			log.Warn("CacheLikeTypeCount field(%s) strconv.ParseInt error(%v)", k, e)
			continue
		}
		res[field] = v
	}
	return res, nil
}

func (d *Dao) CacheActivityArchives(ctx context.Context, sid, mid int64) ([]*like.Item, error) {
	key := activityArchivesKey(sid, mid)
	bs, err := redis.Bytes(component.GlobalRedis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
		}
		return nil, err
	}
	out := []*like.Item{}
	if err = json.Unmarshal(bs, &out); err != nil {
		return nil, errors.WithStack(err)
	}
	return out, nil
}

func (d *Dao) AddCacheActivityArchives(ctx context.Context, sid int64, arcs []*like.Item, mid int64) error {
	key := activityArchivesKey(sid, mid)
	bs, err := json.Marshal(arcs)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = component.GlobalRedis.Do(ctx, "SETEX", key, d.actArcsExpire, bs)
	return err
}

func (d *Dao) DelCacheActivityArchives(ctx context.Context, sid, mid int64) error {
	key := activityArchivesKey(sid, mid)
	_, err := component.GlobalRedis.Do(ctx, "DEL", key)
	return err
}

func (d *Dao) CacheGetActRelationInfo(ctx context.Context, id int64) (*like.ActRelationInfo, error) {
	key := actRelationInfoKey(id)
	bs, err := redis.Bytes(component.GlobalRedis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
		}
		return nil, err
	}
	out := &like.ActRelationInfo{}
	if err = json.Unmarshal(bs, &out); err != nil {
		return nil, errors.WithStack(err)
	}
	return out, nil
}

func (d *Dao) AddCacheGetActRelationInfo(ctx context.Context, id int64, arcs *like.ActRelationInfo) error {
	key := actRelationInfoKey(id)
	bs, err := json.Marshal(arcs)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = component.GlobalRedis.Do(ctx, "SETEX", key, tool.ExpiredRedisKeyAtDayEarly(), bs)
	return err
}

func (d *Dao) HotAddActRelationInfoSet(ctx context.Context, jsonData string) (err error) {
	key := hotActRelationInfoKey()
	_, err = component.GlobalRedis.Do(ctx, "SET", key, jsonData)
	return err
}

func (d *Dao) HotGetActRelationInfoSet(ctx context.Context) (str string, err error) {
	key := hotActRelationInfoKey()
	str, err = redis.String(component.GlobalRedis.Do(ctx, "GET", key))
	return str, err
}

func (d *Dao) HotAddActSubjectInfoSet(ctx context.Context, jsonData string) (err error) {
	key := hotActSubjectInfoKey()
	_, err = component.GlobalRedis.Do(ctx, "SET", key, jsonData)
	return err
}

func (d *Dao) HotGetActSubjectInfoSet(ctx context.Context) (str string, err error) {
	key := hotActSubjectInfoKey()
	str, err = redis.String(component.GlobalRedis.Do(ctx, "GET", key))
	return str, err
}

func (d *Dao) AddCacheGetActSubjectInfo(ctx context.Context, id int64, item *like.SubjectItem) error {
	var err error
	for i := 0; i < 3; i++ {
		err = d.AddCacheActSubject(ctx, id, item)
		if err == nil {
			break
		}
	}
	return err
}

func (d *Dao) HotAddActSubjectReserveIDsInfoSet(ctx context.Context, jsonData string) (err error) {
	key := hotActSubjectReserveIDsInfoKey()
	_, err = component.GlobalRedis.Do(ctx, "SET", key, jsonData)
	return err
}

func (d *Dao) HotGetActSubjectReserveIDsInfoSet(ctx context.Context) (str string, err error) {
	key := hotActSubjectReserveIDsInfoKey()
	str, err = redis.String(component.GlobalRedis.Do(ctx, "GET", key))
	return str, err
}

func (d *Dao) DelCacheGetActRelationInfo(ctx context.Context, id int64) error {
	key := actRelationInfoKey(id)
	_, err := component.GlobalRedis.Do(ctx, "DEL", key)
	if err != nil && err != redis.ErrNil {
		return err
	}
	return nil
}

func (d *Dao) CacheYellowGreenVote(ctx context.Context, period *like.YellowGreenPeriod) (*like.YgVote, error) {
	if period == nil {
		return nil, nil
	}
	key := keyYellowGreenPeriod(period.YellowYingYuanSid, period.GreenYingYuanSid)
	bs, err := redis.Bytes(component.GlobalRedis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
		}
		return nil, err
	}
	out := &like.YgVote{}
	if err = json.Unmarshal(bs, &out); err != nil {
		return nil, errors.WithStack(err)
	}
	return out, nil
}

// 读缓存
func (d *Dao) GetVoteTotalBySid(c context.Context, sid int64) (reply []*like.LIDWithVote, err error) {
	key := keyVoteTotalBySID(sid)
	var res []byte
	res, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key))
	if err != nil && err != redis.ErrNil {
		err = errors.Errorf("GetVoteTotalBySid Err conn.Do(GET %s) error(%v)", key, err)
		return
	}

	var body []*like.LIDWithVote
	err = json.Unmarshal(res, &body)
	if err != nil {
		return
	}

	return body, nil
}
