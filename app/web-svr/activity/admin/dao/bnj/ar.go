package bnj

import (
	"context"

	"go-gateway/app/web-svr/activity/admin/component"
	model "go-gateway/app/web-svr/activity/admin/model/bnj"
)

const (
	sql4AddARSetting = `
INSERT INTO bnj_ar_setting (setting)
VALUES (?);
`
	sql4AddScore2Coupon = `
INSERT INTO bnj_ar_exchange_rule (score, coupon)
VALUES (?, ?);
`
	sql4UpdateScore2Coupon = `
UPDATE bnj_ar_exchange_rule
SET score = ?, coupon = ?
WHERE id = ?;
`
	sql4DeleteScore2Coupon = `
UPDATE bnj_ar_exchange_rule
SET is_deleted = 1
WHERE id = ?;
`
)

func AddARSetting(ctx context.Context, setting string) (err error) {
	_, err = component.GlobalDB.Exec(ctx, sql4AddARSetting, setting)

	return
}

func UpsertARScore2Coupon(ctx context.Context, rule *model.Score2CouponRule) (err error) {
	if rule.ID > 0 {
		_, err = component.GlobalDB.Exec(ctx, sql4UpdateScore2Coupon, rule.Score, rule.Coupon, rule.ID)
	} else {
		_, err = component.GlobalDB.Exec(ctx, sql4AddScore2Coupon, rule.Score, rule.Coupon)
	}

	return
}

func DeleteARScore2Coupon(ctx context.Context, rule *model.Score2CouponRule) (err error) {
	if rule.ID > 0 {
		_, err = component.GlobalDB.Exec(ctx, sql4DeleteScore2Coupon, rule.ID)
	}

	return
}
