package search

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/cache/memcache"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
	searchModel "go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	SEARCH_SPREAD_COUNT_PREFIX     = "tianma_search_spread_count"
	SEARCH_SPREAD_COUNT_EXPIRATION = 30 // expiration in seconds
)

func (d *Dao) SearchSpreadConfigQueryById(c context.Context, ids []int64) (ret map[int64]*searchModel.SpreadConfig, err error) {
	var obj []*searchModel.SpreadConfig
	if err = d.DB.Table("search_spread_config").
		Where("id in (?)", ids).
		Scan(&obj).
		Error; err != nil {
		return
	}

	ret = make(map[int64]*searchModel.SpreadConfig)
	for _, conf := range obj {
		ret[conf.ID] = conf
	}
	return
}

// SearchOptSpreadConfig pass/reject/hidden spread config
func (d *Dao) SearchOptSpreadConfig(c context.Context, ids []int64, option string) (err error) {
	update := map[string]interface{}{}
	switch option {
	case common.OptionBatchPass:
		update["valid_status"] = common.StatusOnline
		update["check"] = common.Pass
	case common.OptionBatchReject:
		update["valid_status"] = common.StatusDownline
		update["check"] = common.Rejecte
	case common.OptionBatchHidden:
		update["valid_status"] = common.StatusDownline
		update["check"] = common.InValid
	}
	err = d.DB.Table("search_spread_config").
		Where("id in (?)", ids).
		Update(update).Error
	return
}

// SetSearchSpreadCount set search spread count to MC
func (d *Dao) SetSearchSpreadCount(c context.Context, params *searchModel.RecomParam, count int) (err error) {
	var (
		key string
	)
	if key, err = d.getSearchSpreadCountKey(c, params); err != nil {
		return err
	}
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(strconv.Itoa(count)),
		Flags:      memcache.FlagRAW,
		Expiration: SEARCH_SPREAD_COUNT_EXPIRATION,
	}
	return d.MC.Set(c, item)
}

// GetSearchSpreadCount get search spread count from MC
func (d *Dao) GetSearchSpreadCount(c context.Context, params *searchModel.RecomParam) (count int, err error) {
	var (
		v   string
		key string
	)
	if key, err = d.getSearchSpreadCountKey(c, params); err != nil {
		return 0, err
	}
	if err = d.MC.Get(c, key).Scan(&v); err != nil {
		return 0, err
	}
	if count, err = strconv.Atoi(v); err != nil {
		return 0, err
	}
	return
}

func (d *Dao) getSearchSpreadCountKey(_ context.Context, params *searchModel.RecomParam) (key string, err error) {
	cardTypeKey, err := util.Int2AlphaString(params.CardType)
	if err != nil {
		return
	}

	return fmt.Sprintf("%s_%d_%d_%d_%d_%s",
		SEARCH_SPREAD_COUNT_PREFIX, params.StartTs, params.EndTs, params.Plat, params.Pos,
		cardTypeKey), nil
}
