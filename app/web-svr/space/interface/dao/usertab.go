package dao

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"

	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"

	midrpc "git.bilibili.co/bapis/bapis-go/account/service"
	napagerpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
)

const (
	_usertabSQL = `SELECT tab_type,tab_name,tab_order,tab_cont, mid, is_default, limits, h5_link FROM space_usertab
                    WHERE mid = ? AND deleted = 0 AND online=1`
	_queryOnlineUserTabSQL = `SELECT id, tab_type,tab_name,tab_order,tab_cont, mid, stime, etime ,is_default
                               FROM space_usertab WHERE mid = ? AND deleted = 0 AND online=1`
	_queryUsertabSQL = `select tab_type,tab_name,tab_order,tab_cont,mid,stime,etime,is_default 
                         from space_usertab where mid = ? and deleted = 0 order by etime desc`
	_queryUsertabByIdSQL = `SELECT tab_type,tab_name,tab_order,tab_cont, mid,stime,etime,is_default 
                            FROM space_usertab WHERE id = ? AND deleted = 0 order by etime desc`
	_insertUserTabSQL = `insert into space_usertab (tab_type,mid,tab_name,tab_cont,stime,etime,online,tab_order,is_sync,is_default) 
                         values (?,?,?,?,?,?,?,?,1,?)`
	_updateUserTabSQL = `update space_usertab set tab_type=?,mid=?,tab_name=?,tab_cont=?,online=?,tab_order=?,is_default= ?
                         where id = ?`
	_onlineUserTabSQL = `update space_usertab set online = ?, stime=?, etime=? where id = ?`
)

// RawUserTab
func (d *Dao) RawUserTab(c context.Context, req *pb.UserTabReq) (res *model.UserTab, err error) {
	res = &model.UserTab{}
	row := d.db.QueryRow(c, _usertabSQL, req.Mid)
	if err = row.Scan(&res.TabType, &res.TabName, &res.TabOrder, &res.TabCont, &res.Mid, &res.IsDefault, &res.Limits, &res.H5Link); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
			return
		}
		log.Error("RawUserTab.Scan mid(%d) error(%v)", req.Mid, err)
		return
	}
	return
}

// SpaceUserTab Add
func (d *Dao) SpaceUserTabAdd(arg *model.UserTab) (err error) {
	if _, err = d.db.Exec(context.Background(), _insertUserTabSQL, arg.TabType, arg.Mid, arg.TabName,
		arg.TabCont, arg.Stime, arg.Etime, arg.Online, arg.TabOrder, arg.IsDefault); err != nil {
		log.Error("dao.SpaceUserTabAdd value(%v) error(%v)", arg, err)
		return
	}
	return
}

// SpaceUserTab Modify
func (d *Dao) SpaceUserTabModify(arg *model.UserTab) (err error) {
	if _, err = d.db.Exec(context.Background(), _updateUserTabSQL, arg.TabType, arg.Mid, arg.TabName,
		arg.TabCont, arg.Online, arg.TabOrder, arg.IsDefault, arg.ID); err != nil {
		log.Error("dao.SpaceUserTabModify value(%v) error(%v)", arg, err)
		return
	}
	return
}

func (d *Dao) SpaceUserTabOnline(id int64, arg *model.UserTab) (err error) {
	if _, err = d.db.Exec(context.Background(), _onlineUserTabSQL, arg.Online, arg.Stime, arg.Etime, id); err != nil {
		log.Error("dao.SpaceUserTabOnline online(%v) error(%v)", id, err)
		return
	}
	return
}

func (d *Dao) ValidUserTabFindByMid(mid int64) (ret *model.UserTab, err error) {
	ret = &model.UserTab{}
	row := d.db.QueryRow(context.Background(), _queryOnlineUserTabSQL, mid)
	if err = row.Scan(&ret.ID, &ret.TabType, &ret.TabName, &ret.TabOrder, &ret.TabCont, &ret.Mid, &ret.Stime, &ret.Etime, &ret.IsDefault); err != nil {
		log.Error("RawUserTab.Scan mid(%d) error(%v)", mid, err)
		return
	}
	return
}

func (d *Dao) EtimeUserTabFindByMid(arg *model.UserTab) (bool, error) {
	var tmp []*model.UserTab

	rows, err := d.db.Query(context.Background(), _queryUsertabSQL, arg.Mid)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		spaceTab := &model.UserTab{}
		if err = rows.Scan(&spaceTab.TabType, &spaceTab.TabName, &spaceTab.TabOrder, &spaceTab.TabCont, &spaceTab.Mid, &spaceTab.Stime,
			&spaceTab.Etime, &spaceTab.IsDefault); err != nil {
			log.Error("dao.EtimeUserTabFindByMid query(%+v) error(%+v)", _queryUsertabSQL, err)
			return false, err
		}
		tmp = append(tmp, spaceTab)
	}
	if err = rows.Err(); err != nil {
		log.Error("dao.RawCommonActivities rows.err error(+%v)", err)
		return false, err
	}
	if len(tmp) == 0 {
		return true, nil
	}
	tmp = append(tmp, &model.UserTab{
		ID:    -1,
		Etime: model.Sentinel,
	})
	for _, row := range tmp {
		if arg.ID == row.ID {
			continue
		}
		if arg.Stime > row.Etime {
			return true, nil
		} else if arg.Etime < row.Stime {
			continue
		} else if arg.Etime > row.Stime {
			return false, nil
		}
	}
	return false, nil
}

func (d *Dao) SpaceUserTabFindById(id int64) (ret *model.UserTab, err error) {
	ret = &model.UserTab{}
	row := d.db.QueryRow(context.Background(), _queryUsertabByIdSQL, id)
	if err = row.Scan(&ret.TabType, &ret.TabName, &ret.TabOrder, &ret.TabCont, &ret.Mid, &ret.Stime, &ret.Etime, &ret.IsDefault); err != nil {
		if err == sql.ErrNoRows {
			return
		}
		log.Error("SpaceUserTabFindById.Scan mid(%d) error(%v)", id, err)
		return
	}
	return
}

func (d *Dao) MidInfoReply(c context.Context, mid int64) (res *midrpc.CardReply, err error) {
	var midinfo *midrpc.CardReply
	arg := &midrpc.MidReq{
		Mid: mid,
	}
	midinfo, err = d.midClient.Card3(c, arg)
	if err != nil {
		err = fmt.Errorf("Get MidInfo error")
		log.Error("MidInfoReply req(%v) err(%v)", mid, err)
		return nil, err
	}
	if midinfo == nil {
		err = fmt.Errorf("无效Mid(%v)", mid)
		return nil, err
	}
	res = midinfo
	return
}

func (d *Dao) CheckNaPage(pid int64) (err error) {
	req := &napagerpc.NativePageReq{
		Pid: pid,
	}
	ret, err := d.naPageClient.NativePage(context.Background(), req)
	if err != nil {
		log.Error("dao.CheckNaPage check pid(%+v) error(%+v)", pid, err)
		return
	}
	if ret.Item == nil {
		err = ecode.Error(-400, "无效Native ID")
		return
	}
	// up主发起类型
	if ret.Item.FromType == 1 && ret.Item.RelatedUid != 0 {
		return
	}
	err = ecode.Error(-400, "非UP主发起Native ID")
	return
}

func (d *Dao) FlushCache(c context.Context, arg *model.UserTab) (err error) {
	var (
		conn  = d.redis.Get(c)
		exist = true
	)
	defer conn.Close()
	key := fmt.Sprintf("usertab_%d", arg.Mid)
	if arg.Online == model.OFFLINE {
		// 查看这个key是否存在
		if _, err = conn.Do("GET", key); err != nil {
			if err == redis.ErrNil {
				exist = false
				err = nil
			} else {
				log.Error("flushCache conn.Do(GET,%s) error(%v)", key, err)
				return
			}
		}
		// 如果存在，则删除
		if exist {
			if _, err = conn.Do("DEL", key); err != nil {
				log.Error("flushCache conn.Do(GET,%s) error(%v)", key, err)
				return
			}
		}
		return
	}
	req := &pb.UserTabReq{
		Mid: arg.Mid,
	}
	res := &model.UserTab{
		Mid:      arg.Mid,
		TabType:  arg.TabType,
		TabCont:  arg.TabCont,
		TabName:  arg.TabName,
		TabOrder: arg.TabOrder,
	}
	if err = d.AddCacheUserTab(c, req, res); err != nil {
		log.Error("flushCache AddCacheUsertab(%+v) error(%v)", req, err)
		return
	}
	return
}
