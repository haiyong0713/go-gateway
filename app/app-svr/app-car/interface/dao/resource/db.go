package resource

import (
	"context"

	"go-gateway/app/app-svr/app-car/interface/model/vip"
)

const (
	_insertVipReceivedSQL = "INSERT IGNORE INTO car_vip(`mid`,`buvid`,`channel`,`batch_token`,`order_no`,`state`) VALUES (?,?,?,?,?,?)"
	_delVipReceivedSQL    = "DELETE FROM car_vip WHERE mid=?"
)

func (d *Dao) InVipReceived(ctx context.Context, arg *vip.VipReceived) (int64, error) {
	res, err := d.db.Exec(ctx, _insertVipReceivedSQL, arg.MID, arg.Buvid, arg.Channel, arg.BatchToken, arg.Channel, arg.State)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) DelVipReceived(ctx context.Context, mid int64) (int64, error) {
	res, err := d.db.Exec(ctx, _delVipReceivedSQL, mid)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
