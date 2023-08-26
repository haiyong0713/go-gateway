package bubble

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"

	bubblemdl "go-gateway/app/app-svr/app-resource/interface/model/bubble"
)

const (
	_bubbleSQL = "SELECT id,position,icon,`desc`,url,stime,etime,operator,state,white_list FROM bubble WHERE state=1 AND stime<? AND etime>?"
)

func (d *Dao) Bubble(c context.Context) (res map[int64]*bubblemdl.Bubble, err error) {
	rows, err := d.db.Query(c, _bubbleSQL, time.Now(), time.Now())
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	res = make(map[int64]*bubblemdl.Bubble)
	for rows.Next() {
		var (
			position string
			pos      []*bubblemdl.Postion
		)
		re := &bubblemdl.Bubble{}
		if err = rows.Scan(&re.ID, &position, &re.Icon, &re.Desc, &re.URL, &re.STime, &re.ETime, &re.Operator, &re.State, &re.WhiteList); err != nil {
			log.Error("Bubble %v", err)
		}
		if err = json.Unmarshal([]byte(position), &pos); err != nil {
			log.Error("%v", err)
			return
		}
		for _, p := range pos {
			res[p.PositionID] = re
		}
	}
	return
}
