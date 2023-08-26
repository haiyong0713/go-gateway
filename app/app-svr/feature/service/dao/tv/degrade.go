package tv

import (
	"context"

	degrademdl "go-gateway/app/app-svr/feature/service/model/degrade"
)

const (
	_selectDecode    = "SELECT code,decode_type,auto_launch FROM tv_fawkes_channel WHERE is_deleted=0"
	_selectLimitType = "SELECT display_type,limit_type,limit_list,rules,rank,direction FROM display_limit WHERE deleted=0"
)

func (d *Dao) DisplayLimit(c context.Context) (map[string][]*degrademdl.DisplayLimitRes, error) {
	rows, err := d.db.Query(c, _selectLimitType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[string][]*degrademdl.DisplayLimitRes)
	for rows.Next() {
		tmp := &degrademdl.DisplayLimitDB{}
		if err = rows.Scan(&tmp.DisplayType, &tmp.LimitType, &tmp.LimitList, &tmp.Rules, &tmp.Rank, &tmp.Direction); err != nil {
			return nil, err
		}
		dis := tmp.ToRes()
		if dis != nil {
			res[tmp.DisplayType] = append(res[tmp.DisplayType], dis)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) ChannelFeature(c context.Context) (map[string]degrademdl.ChannelFeatrue, error) {
	rows, err := d.db.Query(c, _selectDecode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[string]degrademdl.ChannelFeatrue)
	for rows.Next() {
		var (
			code       string
			ty         int64
			autoLaunch int32
		)
		if err = rows.Scan(&code, &ty, &autoLaunch); err != nil {
			return nil, err
		}
		res[code] = degrademdl.ChannelFeatrue{
			DecodeType: ty,
			AutoLaunch: autoLaunch,
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
