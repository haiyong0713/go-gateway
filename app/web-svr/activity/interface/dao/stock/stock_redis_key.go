package stock

import (
	"fmt"
	"go-gateway/app/web-svr/activity/interface/model/stock"
)

const (
	fixStockLock       = "fix:lock"
	giftStockPre       = "stock:gift_ver"
	uniqIdSetPre       = "uniq:set"
	backUpGiftStockPre = "back_up:gift_stock:set"
	giftStockHashTag   = "{hash:tag:stock_id:%v}"
	retryStockPre      = "retry:request:pre"
	userStockLimitPre  = "user:stock:limit"
)

// getHashTag 生成redis hash tag
func getRedisKeyPre(stockId int64, limitKey string) string {
	hashTag := fmt.Sprintf(giftStockHashTag, stockId)
	if limitKey == "" {
		return hashTag
	}
	return hashTag + separator + limitKey
}

func getStockKey(sbq *stock.StockReq) string {
	return buildKey(getRedisKeyPre(sbq.StockId, sbq.LimitKey), giftStockPre)
}

func getUniqSetKey(sbq *stock.StockBaseReq) string {
	return buildKey(getRedisKeyPre(sbq.StockId, sbq.LimitKey), uniqIdSetPre)
}

func getBackUpSetKey(sbq *stock.StockBaseReq) string {
	return buildKey(getRedisKeyPre(sbq.StockId, sbq.LimitKey), backUpGiftStockPre)
}

func getLockKey(sbq *stock.StockBaseReq, giftVer int64) string {
	return buildKey(getRedisKeyPre(sbq.StockId, sbq.LimitKey), fixStockLock, fmt.Sprint(giftVer))
}

func getRetryRequestKey(sbq *stock.StockBaseReq, retryId string) string {
	return buildKey(getRedisKeyPre(sbq.StockId, sbq.LimitKey), retryId, retryStockPre)
}

func getUserStockLimitKey(sbq *stock.StockBaseReq, mid int64) string {
	return buildKey(getRedisKeyPre(sbq.StockId, sbq.LimitKey), mid, userStockLimitPre)
}
