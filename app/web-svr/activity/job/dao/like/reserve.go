package like

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/activity/job/component"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
)

const (
	_reserveListSQL         = "SELECT `id`,`mid`,`state`,`num` FROM `act_reserve_%02d` WHERE `sid` = ? AND id > ? ORDER BY id ASC LIMIT ?"
	_reserveTunnleSQL       = "SELECT `id`,`mid`,`state`,`num`, `ctime`,`platform`,`from` FROM `act_reserve_%02d` WHERE `sid` = ? AND id > ? ORDER BY id ASC LIMIT ?"
	_tunnelCntSQL           = "SELECT count(*) as c FROM act_subject_tunnel WHERE template_id>0 AND sid=?"
	_subjectFlagSQL         = "SELECT flag FROM act_subject WHERE id=?"
	_reserveTableNum        = 100
	_platform               = 3
	_stateOk                = 1
	_stateCancel            = 2
	_insert2NewReserveTable = "INSERT INTO `act_reserve_new_%02d` (`sid`,`mid`,`num`,`state`,`ipv6`,`score`,`adjust_score`,`from`,`typ`,`oid`,`platform`,`mobiapp`,`buvid`, `spmid`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	_update2NewReserveTable = "INSERT INTO `act_reserve_new_%02d` (`sid`,`mid`,`num`,`state`,`ipv6`,`score`,`adjust_score`,`from`,`typ`,`oid`,`platform`,`mobiapp`,`buvid`, `spmid`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `sid` = ?, `mid` = ?,`num` = ?,`state` = ?,`ipv6` = ?,`score` = ?,`adjust_score` = ?,`from` = ?,`typ` = ?,`oid` = ?,`platform` = ?,`mobiapp` = ?,`buvid` = ?, `spmid` = ?"
)

func (d *Dao) RawReserveList(c context.Context, sid, id, limit int64) (res []*like.Reserve, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_reserveListSQL, sid%_reserveTableNum), sid, id, limit)
	if err != nil {
		err = errors.Wrap(err, "d.db.Query()")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := new(like.Reserve)
		if err = rows.Scan(&item.ID, &item.Mid, &item.State, &item.Num); err != nil {
			err = errors.Wrap(err, "rows.Scan()")
			return
		}
		res = append(res, item)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err()")
	}
	return
}

func (d *Dao) SubjectFlag(c context.Context, sid int64) (flag int64, err error) {
	row := d.db.QueryRow(c, _subjectFlagSQL, sid)
	if err = row.Scan(&flag); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "SubjectFlag:QueryRow(%d)", sid)
		}
	}
	return
}

func (d *Dao) TunnelTemplateCnt(c context.Context, sid int64) (count int64, err error) {
	row := d.db.QueryRow(c, _tunnelCntSQL, sid)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "TunnelTemplateCnt:QueryRow(%d)", sid)
		}
	}
	return
}

func (d *Dao) AsyncSendTunnelDatabus(ctx context.Context, data *like.Reserve) (err error) {
	var state int64 = _stateOk
	if data.State != _stateOk {
		state = _stateCancel
	}
	reqParam := struct {
		Platform int64 `json:"platform"`
		Mid      int64 `json:"mid"`
		State    int64 `json:"state"`
		BizID    int64 `json:"biz_id"`
		UniqueID int64 `json:"unique_id"`
	}{_platform, data.Mid, state, d.c.Rule.TunnelPushBizID, data.Sid}
	if err = d.tunnelPub.Send(ctx, strconv.FormatInt(data.Mid, 10), reqParam); err != nil {
		log.Errorc(ctx, "d.tunnelPub.Send data(%+v) reqParam(%+v) error(%+v)", data, reqParam, err)
	}
	return
}

func (d *Dao) AsyncSendGroupDatabus(ctx context.Context, data *like.Reserve, timestap int64) (err error) {
	var state int64 = _stateOk
	if data.State != _stateOk {
		state = _stateCancel
	}
	reqParam := struct {
		Platform  string `json:"platform"`
		From      string `json:"from"`
		Mid       int64  `json:"mid"`
		Source    string `json:"source"`
		Name      string `json:"name"`
		State     int64  `json:"state"`
		Timestamp int64  `json:"timestamp"`
	}{data.Platform, data.From, data.Mid, d.c.TunnelGroup.Source, strconv.FormatInt(data.Sid, 10), state, timestap}
	key := strconv.FormatInt(data.Mid, 10)
	buf, _ := json.Marshal(reqParam)
	for i := 0; i < 3; i++ {
		if err = component.BGroupMessagePub.Send(ctx, key, buf); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		log.Errorc(ctx, "AsyncSendGroupDatabus d.tunnelGroupPub.Send data(%+v) reqParam(%+v) error(%+v)", data, reqParam, err)
	}
	return
}

func (d *Dao) TunnelReserveList(c context.Context, sid, id, limit int64) (res []*like.ReserveTunnel, err error) {
	rows, err := d.db.Query(c, fmt.Sprintf(_reserveTunnleSQL, sid%_reserveTableNum), sid, id, limit)
	if err != nil {
		err = errors.Wrap(err, "d.db.Query()")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		item := new(like.ReserveTunnel)
		if err = rows.Scan(&item.ID, &item.Mid, &item.State, &item.Num, &item.Ctime, &item.Platform, &item.From); err != nil {
			err = errors.Wrap(err, "rows.Scan()")
			return
		}
		res = append(res, item)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "rows.Err()")
	}
	return
}

func (d *Dao) Insert2NewReserveTable(ctx context.Context, data *like.ActReserveField) (err error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_insert2NewReserveTable, data.Mid%100), data.Sid, data.Mid, data.Num, data.State, data.IPV6, data.Score, data.AdjustScore, data.From, data.Typ, data.Oid, data.Platform, data.Mobiapp, data.Buvid, data.Spmid)
	if err != nil {
		err = errors.Wrap(err, "_insert2NewReserveTable Exec")
		return
	}
	if eff, _ := res.RowsAffected(); eff <= 0 {
		err = errors.Wrap(err, "_insert2NewReserveTable RowsAffected")
		return
	}
	return
}

func (d *Dao) NewReserveTableOnDuplicate(ctx context.Context, data *like.ActReserveField) (err error) {
	res, err := d.db.Exec(ctx, fmt.Sprintf(_update2NewReserveTable, data.Mid%100), data.Sid, data.Mid, data.Num, data.State, data.IPV6, data.Score, data.AdjustScore, data.From, data.Typ, data.Oid, data.Platform, data.Mobiapp, data.Buvid, data.Spmid, data.Sid, data.Mid, data.Num, data.State, data.IPV6, data.Score, data.AdjustScore, data.From, data.Typ, data.Oid, data.Platform, data.Mobiapp, data.Buvid, data.Spmid)
	if err != nil {
		err = errors.Wrap(err, "_update2NewReserveTable Exec")
		return
	}
	if eff, _ := res.RowsAffected(); eff <= 0 {
		err = errors.Wrap(err, "_update2NewReserveTable RowsAffected")
		return
	}
	return
}

func (d *Dao) SendLotteryNotify2Tunnel(ctx context.Context, data *like.LotteryReserveNotify) (err error) {
	if err = retry.WithAttempts(ctx, "", 3, netutil.DefaultBackoffConfig, func(c context.Context) error {
		err = d.upReservePushPub.Send(ctx, strconv.FormatInt(data.CardUniqueID, 10), data)
		log.Infoc(ctx, "d.upReservePushPub.Send data(%+v)", data)
		return err
	}); err != nil {
		err = errors.Wrapf(err, "SendLotteryNotify2Tunnel err data(%+v)", data)
		return
	}
	return
}
