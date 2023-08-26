package bwsonline

import (
	"context"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"strconv"
	"strings"
)

const (
	GiftStockHashTag   = "hash:tag:sid:%v:gift_id:%v"
	FixStockLock       = "fix:lock"
	GiftStockPre       = "stock:gift_ver"
	UniqIdSetPre       = "uniq:set"
	BackUpGiftStockPre = "back_up:gift_stock:set"
	// 1、定制化脚本一（运营增加库存）
	IncrStockLua = `
    local goodStock
    local stockKey = KEYS[2]
    local incrStock = tonumber(ARGV[1])
    local val = redis.call('SET', KEYS[1], 1 , 'EX' , 3600 , 'NX')
    if  val == false  then
        return 0
    end
	if  val['ok'] == 'OK' then
        goodStock = redis.call('GET', stockKey)

        if goodStock == false then
		  return  redis.call('INCRBY', stockKey, incrStock)
		end

		if tonumber(goodStock) <= 0 then
          redis.call('SET', stockKey, incrStock)
		  return  incrStock
		end
        
     	return redis.call('INCRBY',stockKey, incrStock)
    else
     	return 0
    end
	`
	// 2、定制化脚本二（运营减少库存）
	DecrStockLua = `
    local val = redis.call('SET', KEYS[1], 1 , 'EX' , 3600 , 'NX')
    if val == false then 
        return 0
    end
	if  val['ok'] == 'OK' then
     	return redis.call('DECRBY', KEYS[2] , tonumber(ARGV[1]))
    else
     	return 0
    end
	`

	// 3、定制化脚本三（线上扣库存）
	ConsumerStockLua = `
    local  res = {}
    local  consumerStock = tonumber(ARGV[1])
    local  leftStock = redis.call('DECRBY', KEYS[1] , consumerStock)
    leftStock = tonumber(leftStock)
    local topNum = leftStock + consumerStock
	if topNum > 0 then
        leftStock = leftStock + 1
        if  leftStock <= 0 then 
            leftStock = 1 
        end
        local index = 1
        for i= topNum , leftStock , -1 do
            res[index] = ARGV[4]..':'..ARGV[2]..':'..ARGV[3]..':'..i
            index = index + 1
        end
        redis.call('SADD', KEYS[2] , unpack(res))
    end
    return res
    `

	// 4、定制化脚本四：从备份（防丢）缓存中获取库存ID，使用了redis.replicate_commands，请慎重使用该脚本
	ConsumerBackUpStockLua = ` 
    redis.replicate_commands()
    local  consumerStock = tonumber(ARGV[1])
    local  backUpStocks = redis.call('SRANDMEMBER', KEYS[1] , consumerStock)
	if backUpStocks and #backUpStocks > 0 then
        redis.call('SADD', KEYS[2] , unpack(backUpStocks))
        redis.call('SREM', KEYS[1] , unpack(backUpStocks))
    end
    return backUpStocks
    `
)

// getHashTag 生成redis hash tag
func getHashTag(sid string, giftId int64) string {
	return fmt.Sprintf(GiftStockHashTag, sid, giftId)
}

func getStockKey(sid string, giftId int64) string {
	hashTag := getHashTag(sid, giftId)
	stockKey := "{" + buildKey(hashTag) + "}" + separator + GiftStockPre
	return stockKey
}

func getUniqSetKey(sid string, giftId int64) string {
	hashTag := getHashTag(sid, giftId)
	uniqSetKey := "{" + buildKey(hashTag) + "}" + separator + UniqIdSetPre
	return uniqSetKey
}

func getBackUpSetKey(sid string, giftId int64) string {
	hashTag := getHashTag(sid, giftId)
	uniqSetKey := "{" + buildKey(hashTag) + "}" + separator + BackUpGiftStockPre
	return uniqSetKey
}

func getLockKey(sid string, giftId, giftVer int64) string {
	hashTag := getHashTag(sid, giftId)
	LoackKey := "{" + buildKey(hashTag) + "}" + separator + FixStockLock + separator + fmt.Sprint(giftVer)
	return LoackKey
}

// IncrStock 支持:并发、幂等
func (d *Dao) IncrStock(ctx context.Context, sid string, giftId, giftVer int64, incr int) (replay int64, err error) {
	var (
		stockKey = getStockKey(sid, giftId)
	)
	if replay, err = redis.Int64(d.redis.Do(ctx, "EVAL", IncrStockLua, 2, getLockKey(sid, giftId, giftVer), stockKey, incr)); err != nil {
		log.Errorc(ctx, "YyincrStock conn.Do(key:%s) error(%v)", stockKey, err)
		return
	}
	log.Infoc(ctx, "YyincrStock begin , stockKey:%v , replay:%v", stockKey, replay)
	return
}

// DecrStock 支持:并发、幂等
func (d *Dao) DecrStock(ctx context.Context, sid string, giftId, giftVer int64, decr int) (decrNum int64, err error) {
	var (
		stockKey = getStockKey(sid, giftId)
	)
	log.Infoc(ctx, "YyDecrStock begin , stockKey:%v", stockKey)
	if decrNum, err = redis.Int64(d.redis.Do(ctx, "EVAL", DecrStockLua, 2, getLockKey(sid, giftId, giftVer), stockKey, decr)); err != nil {
		log.Errorc(ctx, "YyDecrStock conn.Do(key:%s) error(%v)", stockKey, err)
		return 0, err
	}
	return decrNum, nil
}

func (d *Dao) ConsumerStock(ctx context.Context, sid string, giftId, giftVer int64, num int) (uniqIds []string, err error) {
	var (
		stockKey     = getStockKey(sid, giftId)
		uniqSetKey   = getUniqSetKey(sid, giftId)
		backUpSetKey = getBackUpSetKey(sid, giftId)
		reply        []string
		cnt          int
	)
	log.Infoc(ctx, "user consumer stock begin , stockKey:%v , uniqSetKey:%v , backUpSetKey:%v", stockKey, uniqSetKey, backUpSetKey)
	// 1、首先尝试从备份库取数据
	if cnt, err = d.GetBackUpStock(ctx, sid, giftId); err == nil && cnt > 0 && num > 0 {
		log.Infoc(ctx, "Consumer backUp stock backUpSetKey:%v SCARD:%v , num:%v", backUpSetKey, cnt, num)
		if reply, err = redis.Strings(d.redis.Do(ctx, "EVAL", ConsumerBackUpStockLua, 2, backUpSetKey, uniqSetKey, num)); err == nil {
			uniqIds = append(uniqIds, reply...)
		}
		num = num - len(uniqIds)
	}
	// 2、再次尝试从正式库取数据
	if num > 0 {
		log.Infoc(ctx, "Consumer real stock stockKey:%v num:%v", stockKey, num)
		if reply, err = redis.Strings(d.redis.Do(ctx, "EVAL", ConsumerStockLua, 2, stockKey, uniqSetKey, num, giftId, giftVer, sid)); err == nil {
			uniqIds = append(uniqIds, reply...)
		}
	}

	if err != nil {
		log.Errorc(ctx, "user consumer stock err , stockKey:%v  , err:%v", stockKey, err)
	}
	return
}

// GetGiftStocks 批量获取奖品库存
func (d *Dao) GetGiftStocks(ctx context.Context, sid string, giftIds []int64) (stocks map[int64]int, err error) {
	stockKeys := redis.Args{}
	for _, giftId := range giftIds {
		stockKeys = stockKeys.Add(getStockKey(sid, giftId))
	}

	log.Infoc(ctx, "GetGiftStocks stockKeys:%v", stockKeys)
	if len(stockKeys) > 0 {
		var stockArr []int
		if stockArr, err = redis.Ints(d.redis.Do(ctx, "MGET", stockKeys...)); err != nil {
			log.Errorc(ctx, "GetGiftStocks  keys:%v , err :%v", stockKeys, err)
			if err == redis.ErrNil {
				err = nil
			}
			return
		}
		log.Infoc(ctx, "GetGiftStocks mget , value:%v", stockArr)
		stocks = make(map[int64]int)
		for k, v := range stockArr {
			if k >= 0 && k < len(giftIds) {
				stocks[giftIds[k]] = v
			}
		}
	}

	return
}

// RandGetStockNo 随机获取已经发放的库存ID
func (d *Dao) RandGetStockNo(ctx context.Context, sid string, giftId int64, num int) (stocks []string, err error) {
	var (
		uniqSetKey = getUniqSetKey(sid, giftId)
	)

	if stocks, err = redis.Strings(d.redis.Do(ctx, "SRANDMEMBER", uniqSetKey, num)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(ctx, "RandGetStockNo SRANDMEMBER:%s error(%+v)", uniqSetKey, err)
	}
	return
}

// AckStockNo 确认库存ID已经正确处理后，移除处理
func (d *Dao) AckStockNo(ctx context.Context, sid string, giftId int64, stocks []string) (replay int64, err error) {
	var (
		uniqSetKey = getUniqSetKey(sid, giftId)
	)
	args := redis.Args{}
	args = append(args, uniqSetKey)
	for _, v := range stocks {
		args = append(args, v)
	}

	if replay, err = redis.Int64(d.redis.Do(ctx, "SREM", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(ctx, "AckStockNo SREM:%s error(%+v)", uniqSetKey, err)
	}
	return
}

// SmoveStockNo 将超时丢失的订单号，重新找回来
func (d *Dao) SmoveStockNo(ctx context.Context, sid string, giftId int64, stockNo string) (replay int, err error) {
	var (
		uniqSetKey   = getUniqSetKey(sid, giftId)
		backUpSetKey = getBackUpSetKey(sid, giftId)
	)

	if replay, err = redis.Int(d.redis.Do(ctx, "SMOVE", uniqSetKey, backUpSetKey, stockNo)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
	}
	log.Infoc(ctx, "SmoveStockNo stockNo:%s , source:%s , destination:%s , replay:%d ,  error(%+v)", stockNo, uniqSetKey, backUpSetKey, replay, err)
	return
}

// GetBackUpStock 获取备份库存数量
func (d *Dao) GetBackUpStock(ctx context.Context, sid string, giftId int64) (stock int, err error) {
	var backUpSetKey = getBackUpSetKey(sid, giftId)
	if stock, err = redis.Int(d.redis.Do(ctx, "SCARD", backUpSetKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
	}
	return
}

func ParseSerialNo(uniqIds []string) (res map[string]int, err error) {
	res = make(map[string]int)
	for _, order := range uniqIds {
		item := strings.Split(order, ":")
		if len(item) > 0 {
			var reserveNo int
			if reserveNo, err = strconv.Atoi(item[len(item)-1]); err != nil {
				return
			}
			res[order] = reserveNo
		}
	}
	return
}
