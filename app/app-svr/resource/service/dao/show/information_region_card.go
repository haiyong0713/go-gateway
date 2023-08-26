package show

import (
	"context"
	"time"

	"go-common/library/log"

	resourcegrpc "go-gateway/app/app-svr/resource/service/api/v1"
)

const (
	_informationRegionCardSQL = "SELECT card_type,card_id,card_pos,is_cover,pos_index FROM information_recommend_card WHERE " +
		"stime<? AND etime>? AND audit_status=2 AND offline_status=0 AND is_deleted=0"
)

func (d *Dao) InformationRegionCard(c context.Context, now time.Time) (res []*resourcegrpc.InformationRegionCard, err error) {
	rows, err := d.db.Query(c, _informationRegionCardSQL, now, now)
	if err != nil {
		log.Error(err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var re = new(resourcegrpc.InformationRegionCard)
		if err = rows.Scan(&re.CardType, &re.Rid, &re.CardPosition, &re.IsCover, &re.PositionIdx); err != nil {
			log.Error(err.Error())
			return
		}
		res = append(res, re)
	}
	err = rows.Err()
	return
}
