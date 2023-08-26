package reward_conf

import (
	"context"
	"go-common/library/log"
	xtime "go-common/library/time"
)

// AddOneRewardConf .
func (d *Dao) AddOneRewardConf(ctx context.Context, record *AwardConfigData) error {
	// 写库
	err := d.InsertAwardConfigData(ctx, record)
	if err != nil {
		log.Errorc(ctx, "AddOneRewardConf InsertAwardConfigData err ,err is (%v).", err)
		return err
	}
	return nil
}

// UpdateOneRewardByID .
func (d *Dao) UpdateOneRewardByID(ctx context.Context, id int64, record map[string]interface{}) (err error) {
	_, err = d.mapUpdateAwardConfigDataByID(ctx, uint64(id), record)
	return
}

const _selListByDateSQL = "SELECT id,award_id,stock_id,cost_type,cost_value,show_time,`order`,creator,`status`,ctime,mtime,activity_id,end_time from award_config_data where `status`=1 and activity_id = ? and show_time <= ? and end_time >= ? order by `order` limit ?,?"
const _selListByDateAndCTSQL = "SELECT id,award_id,stock_id,cost_type,cost_value,show_time,`order`,creator,`status`,ctime,mtime,activity_id,end_time from award_config_data where `status`=1 and activity_id = ? and show_time <= ? and end_time >= ? and cost_type=? order by `order` limit ?,?"

func (d *Dao) Search(ctx context.Context, activityId string, sTime, eTime xtime.Time, costType, pn, ps int) (list []*AwardConfigData, err error) {
	offset := (pn - 1) * ps
	limit := ps
	if costType != 0 {
		list, err = d.queryAwardConfigDataRows(ctx, _selListByDateAndCTSQL, activityId, sTime, eTime, costType, offset, limit)
	} else {
		list, err = d.queryAwardConfigDataRows(ctx, _selListByDateSQL, activityId, sTime, eTime, offset, limit)
	}

	return
}
