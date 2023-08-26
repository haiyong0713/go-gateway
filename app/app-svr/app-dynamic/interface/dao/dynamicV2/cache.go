package dynamicV2

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

const (
	_schoolKey = "dc_key"
)

func (d *Dao) SchoolCache(c context.Context, start, end int) ([]*mdlv2.ItemRedis, error) {
	conn := d.redisSchool.Conn(c)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZREVRANGE", _schoolKey, start, end))
	if err != nil {
		if err == redis.ErrNil {
			return []*mdlv2.ItemRedis{}, nil
		}
		log.Error("redis (ZREVRANGE,%s,%d,%d) error(%v)", _schoolKey, start, end, err)
		return nil, err
	}
	if len(values) == 0 {
		return []*mdlv2.ItemRedis{}, nil
	}
	list := make([]*mdlv2.ItemRedis, 0, len(values))
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("redis.Scan vs(%v) error(%v)", values, err)
			return nil, err
		}
		l := &mdlv2.ItemRedis{}
		if err := json.Unmarshal([]byte(bs), &l); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		list = append(list, l)
	}
	return list, nil
}

const (
	_redisKeyExist = 1
	_redisOKStr    = "OK"
)

func (d *Dao) IsAbTestSwitchOn(ctx context.Context, testName string) bool {
	const (
		_abTestSwitchPrefix = "ab:switch:"
	)
	result, err := redis.Int(d.redisExclusive.Do(ctx, "EXISTS", _abTestSwitchPrefix+testName))
	if err != nil {
		return false
	}
	return result == _redisKeyExist
}

func (d *Dao) IsAbTestMidHit(ctx context.Context, testName string, mid int64) (bool, error) {
	const (
		_abTestMid = "ab:%s:%d"
	)
	result, err := redis.Int(d.redisExclusive.Do(ctx, "EXISTS", fmt.Sprintf(_abTestMid, testName, mid)))
	if err != nil {
		return false, err
	}
	return result == _redisKeyExist, nil
}

type AbTestResult struct {
	// 实验变量名
	AbTestVarName string `json:"abTestVarName,omitempty"`
	// 实验变量值
	AbTestVal string `json:"abTestVal,omitempty"`
	// 命中的实验分组 逗号分割的字符串
	AbTestGroups string `json:"abTestGroups,omitempty"`
}

func (d *Dao) AbTestResultCache(ctx context.Context, testName string, mid int64) (*AbTestResult, error) {
	const (
		_abTestResult = "ab:result:%s:%d"
	)
	res, err := redis.String(d.redisExclusive.Do(ctx, "GET", fmt.Sprintf(_abTestResult, testName, mid)))
	if err != nil && err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	ret := new(AbTestResult)
	err = json.Unmarshal([]byte(res), ret)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal abTestResultCache: %v", err)
	}
	return ret, nil
}

func (d *Dao) CacheAbTestResult(ctx context.Context, testName string, mid int64, res *AbTestResult, ttl time.Duration) error {
	const (
		_abTestResult = "ab:result:%s:%d"
	)
	marshaled, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("error marshal abTestResult: %v", err)
	}
	setArgs := []interface{}{
		fmt.Sprintf(_abTestResult, testName, mid), string(marshaled),
	}
	if ttl > 0 {
		setArgs = append(setArgs, "PX", ttl.Milliseconds())
	}
	resp, err := redis.String(d.redisExclusive.Do(ctx, "SET", setArgs...))
	if err != nil {
		return err
	}
	if resp != _redisOKStr {
		return fmt.Errorf("redis error SET %s %s ttl %s: %s", fmt.Sprintf(_abTestResult, testName, mid), string(marshaled), ttl, resp)
	}
	return nil
}
