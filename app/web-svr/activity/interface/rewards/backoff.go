package rewards

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	cacheKey4LiveARCouponIntoBackupMQ = "BNJ2021_live_AR_coupon_backup_%02d"
	cacheKey4LiveAwardSendingBackupMQ = "REWARD_SENDING_BACKOFF_%v_%02d"
)

func genBackupRandSuffix() int64 {
	rand.Seed(time.Now().UnixNano())

	return rand.Int63n(100)
}

func backoffKey4LiveARCoupon() string {
	suffix := genBackupRandSuffix()

	return fmt.Sprintf(cacheKey4LiveARCouponIntoBackupMQ, suffix)
}

func backoffKey4AwardSending(typ string, mid int64) string {
	randSuffix := mid % 20

	return fmt.Sprintf(cacheKey4LiveAwardSendingBackupMQ, typ, randSuffix)
}
