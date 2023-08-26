package manager

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/database/sql"
	"sync"
	"time"

	xecode "go-common/library/ecode"
	pb "go-gateway/app/app-svr/resource/service/api/v2"

	model "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
	"go-gateway/app/app-svr/app-feed/ecode"
)

// cache
var (
	FrontPageBaseDefaultConfig *model.Config
	FrontPageOnlineConfigs     = make(map[int64][]*model.Config)
	frontPageOnlineConfigsLock sync.Mutex
)

const (
	getAllConfigsSQL        = "SELECT id, config_name, contract_id, resource_id, pic, litpic, url, rule, weight, agency, price, state, atype, stime, etime, ctime, cuser, mtime, muser, is_split_layer, split_layer, loc_policy_group_id FROM frontpage ORDER BY stime DESC"
	getOnlineConfigsSQL     = "SELECT id, config_name, contract_id, resource_id, pic, litpic, url, rule, weight, agency, price, state, atype, stime, etime, ctime, cuser, mtime, muser, is_split_layer, split_layer, loc_policy_group_id FROM frontpage WHERE resource_id = ? AND stime <= ? AND etime >= ? AND state = ? ORDER BY stime DESC"
	getBaseDefaultConfigSQL = "SELECT id, config_name, contract_id, resource_id, pic, litpic, url, rule, weight, agency, price, state, atype, stime, etime, ctime, cuser, mtime, muser, is_split_layer, split_layer, loc_policy_group_id FROM frontpage WHERE id = ?"
)

func (d *Dao) FrontpageGetAllConfigs(ctx context.Context) (res []*model.Config, err error) {
	res = make([]*model.Config, 0)

	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, getAllConfigsSQL); err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		}
		return nil, err
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	for rows.Next() {
		row := new(model.Config)
		if err = rows.Scan(
			&row.ID, &row.ConfigName, &row.ContractID, &row.ResourceID, &row.Pic, &row.LitPic, &row.URL, &row.Rule, &row.Weight, &row.Agency, &row.Price, &row.State, &row.Atype, &row.STime, &row.ETime, &row.CTime, &row.CUser, &row.MTime, &row.MUser, &row.IsSplitLayer, &row.SplitLayer, &row.LocPolicyGroupID); err != nil {
			return nil, err
		}
		res = append(res, row)
	}

	if err = rows.Err(); err != nil {
		res = nil
	}
	for _, config := range res {
		config.Position = 1
	}

	return
}

func (d *Dao) FrontpageGetOnlineConfigs(ctx context.Context, resourceID int64) (res []*model.Config, err error) {
	res = make([]*model.Config, 0)

	now := time.Now()
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, getOnlineConfigsSQL, resourceID, now, now, pb.State_Normal); err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			err = nil
		}
		return nil, err
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	for rows.Next() {
		row := new(model.Config)
		if err = rows.Scan(
			&row.ID, &row.ConfigName, &row.ContractID, &row.ResourceID, &row.Pic, &row.LitPic, &row.URL, &row.Rule, &row.Weight, &row.Agency, &row.Price, &row.State, &row.Atype, &row.STime, &row.ETime, &row.CTime, &row.CUser, &row.MTime, &row.MUser, &row.IsSplitLayer, &row.SplitLayer, &row.LocPolicyGroupID); err != nil {
			return nil, err
		}
		// 兜底配置总不算在线上配置里
		if row.ID != model.DefaultConfigID {
			res = append(res, row)
		}
	}

	if err = rows.Err(); err != nil {
		res = nil
		return
	}
	for _, config := range res {
		config.Position = 1
	}

	return
}

func (d *Dao) FrontpageGetBaseDefaultConfig(ctx context.Context) (res *model.Config, err error) {
	res = &model.Config{}
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, getBaseDefaultConfigSQL, model.DefaultConfigID); err != nil {
		if xecode.EqualError(xecode.NothingFound, err) {
			res = nil
			err = nil
		}
		return nil, err
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
		if err != nil {
			err = errors.Wrapf(ecode.FrontPageConfigNotFound, "基础兜底配置(ID:%d)未找到", model.DefaultConfigID)
		}
	}()

	if rows.Next() {
		if err = rows.Scan(
			&res.ID, &res.ConfigName, &res.ContractID, &res.ResourceID, &res.Pic, &res.LitPic, &res.URL, &res.Rule, &res.Weight, &res.Agency, &res.Price, &res.State, &res.Atype, &res.STime, &res.ETime, &res.CTime, &res.CUser, &res.MTime, &res.MUser, &res.IsSplitLayer, &res.SplitLayer, &res.LocPolicyGroupID); err != nil {
			if xecode.EqualError(xecode.NothingFound, err) {
				res = nil
			}
			return nil, err
		}
	} else {
		res = nil
		err = ecode.FrontPageConfigNotFound
		return
	}

	if err = rows.Err(); err != nil {
		res = nil
		return
	}
	res.Position = 1

	return
}

func (d *Dao) FrontpageCacheSetBaseDefaultConfig(config *model.Config) {
	FrontPageBaseDefaultConfig = config
}

func (d *Dao) FrontpageCacheGetBaseDefaultConfig() (res *model.Config) {
	return FrontPageBaseDefaultConfig
}

func (d *Dao) FrontpageCacheSetOnlineConfigs(resourceID int64, configs []*model.Config) {
	frontPageOnlineConfigsLock.Lock()
	if FrontPageOnlineConfigs == nil {
		FrontPageOnlineConfigs = make(map[int64][]*model.Config)
	}
	FrontPageOnlineConfigs[resourceID] = configs
	frontPageOnlineConfigsLock.Unlock()
}

func (d *Dao) FrontpageCacheGetOnlineConfigs(resourceID int64) (res []*model.Config) {
	frontPageOnlineConfigsLock.Lock()
	res = FrontPageOnlineConfigs[resourceID]
	frontPageOnlineConfigsLock.Unlock()
	return
}
