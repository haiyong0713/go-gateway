package bwsonline

import (
	"context"
	"fmt"
	"go-common/library/database/sql"
	"go-common/library/xstr"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/bwsonline"
	"strings"

	"github.com/pkg/errors"
)

const (
	_actInterReserveByDateSQL        = "SELECT id  ,act_type ,act_title , act_img  ,act_begin_time ,act_end_time , vip_reserve_begin_time , vip_reserve_end_time , reserve_begin_time , reserve_end_time  ,describe_info ,vip_ticket_num  ,standard_ticket_num ,screen_date ,ctime ,mtime ,display_index FROM %s WHERE screen_date IN(%s) and is_del = 0 "
	_actInterReserveByReserveTime    = "SELECT id  ,act_type ,act_title , act_img  ,act_begin_time ,act_end_time , vip_reserve_begin_time , vip_reserve_end_time , reserve_begin_time , reserve_end_time  ,describe_info ,vip_ticket_num  ,standard_ticket_num ,screen_date ,ctime ,mtime ,display_index FROM %s WHERE reserve_begin_time < ? and reserve_end_time > ? and is_del = 0 "
	_actInterReserveByVipReserveTime = "SELECT id  ,act_type ,act_title , act_img  ,act_begin_time ,act_end_time , vip_reserve_begin_time , vip_reserve_end_time , reserve_begin_time , reserve_end_time  ,describe_info ,vip_ticket_num  ,standard_ticket_num ,screen_date ,ctime ,mtime ,display_index FROM %s WHERE vip_reserve_begin_time < ? and vip_reserve_end_time > ? and is_del = 0 "
	_actInterReserveByIdSQL          = "SELECT id  ,act_type ,act_title , act_img  ,act_begin_time ,act_end_time , vip_reserve_begin_time , vip_reserve_end_time , reserve_begin_time , reserve_end_time  ,describe_info ,vip_ticket_num  ,standard_ticket_num ,screen_date ,ctime ,mtime ,display_index FROM %s WHERE id = ? and is_del = 0 "
	_actInterReserveByIdsSQL         = "SELECT id  ,act_type ,act_title , act_img  ,act_begin_time ,act_end_time , vip_reserve_begin_time , vip_reserve_end_time , reserve_begin_time , reserve_end_time  ,describe_info ,vip_ticket_num  ,standard_ticket_num ,screen_date ,ctime ,mtime ,display_index FROM %s WHERE id IN(%s) and is_del = 0 "
)

func getInterReserveTableName(year int) string {
	return fmt.Sprintf("act_bws_online_inter_reserve_%v", year)
}

func (d *Dao) RawInterReserveByDate(ctx context.Context, screenDate []int64, year int) (res []*pb.ActInterReserve, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, fmt.Sprintf(_actInterReserveByDateSQL, getInterReserveTableName(year), xstr.JoinInts(screenDate)))
	if err != nil {
		return nil, errors.Wrap(err, "RawInterReserveByDate Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := new(pb.ActInterReserve)
		if err = rows.Scan(&r.ID, &r.ActType, &r.ActTitle, &r.ActImg, &r.ActBeginTime, &r.ActEndTime, &r.VipReserveBeginTime, &r.VipReserveEndTime, &r.ReserveBeginTime, &r.ReserveEndTime, &r.DescribeInfo,
			&r.VipTicketNum, &r.StandardTicketNum, &r.ScreenDate, &r.Ctime, &r.Mtime, &r.DisplayIndex); err != nil {
			return nil, errors.Wrap(err, "RawInterReserveByDate Scan")
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawInterReserveByDate rows")
	}
	return res, nil
}

func (d *Dao) RawInterReserveByTime(ctx context.Context, beginTime, endTime int64, year int, isVip bool) (res []*pb.ActInterReserve, err error) {
	var rows *sql.Rows
	sql := _actInterReserveByReserveTime
	if isVip {
		sql = _actInterReserveByVipReserveTime
	}
	rows, err = d.db.Query(ctx, fmt.Sprintf(sql, getInterReserveTableName(year)), beginTime, endTime)

	if err != nil {
		return nil, errors.Wrap(err, "RawInterReserveByTime Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := new(pb.ActInterReserve)
		if err = rows.Scan(&r.ID, &r.ActType, &r.ActTitle, &r.ActImg, &r.ActBeginTime, &r.ActEndTime, &r.VipReserveBeginTime, &r.VipReserveEndTime, &r.ReserveBeginTime, &r.ReserveEndTime, &r.DescribeInfo,
			&r.VipTicketNum, &r.StandardTicketNum, &r.ScreenDate, &r.Ctime, &r.Mtime, &r.DisplayIndex); err != nil {
			return nil, errors.Wrap(err, "RawInterReserveByTime Scan")
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawInterReserveByTime rows")
	}
	return res, nil
}

func (d *Dao) RawInterReserveById(ctx context.Context, id int64, year int) (r *pb.ActInterReserve, err error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(_actInterReserveByIdSQL, getInterReserveTableName(year)), id)
	r = new(pb.ActInterReserve)
	if err := row.Scan(&r.ID, &r.ActType, &r.ActTitle, &r.ActImg, &r.ActBeginTime, &r.ActEndTime, &r.VipReserveBeginTime, &r.VipReserveEndTime, &r.ReserveBeginTime, &r.ReserveEndTime, &r.DescribeInfo,
		&r.VipTicketNum, &r.StandardTicketNum, &r.ScreenDate, &r.Ctime, &r.Mtime, &r.DisplayIndex); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawInterReserveById Scan")
	}
	return r, nil
}

func (d *Dao) RawInterReserveByIds(ctx context.Context, ids []int64, year int) (res []*pb.ActInterReserve, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, fmt.Sprintf(_actInterReserveByIdsSQL, getInterReserveTableName(year), xstr.JoinInts(ids)))
	if err != nil {
		return nil, errors.Wrap(err, "RawInterReserveByIds Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := new(pb.ActInterReserve)
		if err = rows.Scan(&r.ID, &r.ActType, &r.ActTitle, &r.ActImg, &r.ActBeginTime, &r.ActEndTime, &r.VipReserveBeginTime, &r.VipReserveEndTime, &r.ReserveBeginTime, &r.ReserveEndTime, &r.DescribeInfo,
			&r.VipTicketNum, &r.StandardTicketNum, &r.ScreenDate, &r.Ctime, &r.Mtime, &r.DisplayIndex); err != nil {
			return nil, errors.Wrap(err, "RawInterReserveByIds Scan")
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawInterReserveByIds rows")
	}
	return res, nil
}

const (
	_actInterReserveOrderByMidInterReserveIDsSQL = "SELECT id ,mid ,ticket_no ,inter_reserve_id ,order_no ,is_checked,ctime ,mtime FROM %s WHERE mid = ? and inter_reserve_id in (%s) and is_del = 0 "
	_actInterReserveOrderByMidSQL                = "SELECT id ,mid ,ticket_no ,inter_reserve_id ,order_no ,reserve_no,is_checked,ctime ,mtime FROM %s WHERE mid = ? and is_del = 0 "
	_actInterReserveOrderByMidInterReserveIDSQL  = "SELECT id ,mid ,ticket_no ,inter_reserve_id ,order_no ,reserve_no,is_checked,ctime ,mtime FROM %s WHERE mid = ? and inter_reserve_id = ? and is_del = 0 "
	_actInterReserveOrderCheckedSQL              = "UPDATE %s SET is_checked = 1 WHERE id = ?"
	_addInterReserveOrderSQL                     = "INSERT INTO %s (mid ,ticket_no ,inter_reserve_id ,order_no , reserve_no)VALUES(?,?,?,?,?) "
	_actInterReserveOrderByOrderNos              = "SELECT id ,mid ,ticket_no ,inter_reserve_id ,order_no ,reserve_no,is_checked,ctime ,mtime FROM %s WHERE order_no IN (%s)"
)

func getInterReserveOrderTableName(year int) string {
	return fmt.Sprintf("act_bws_online_inter_reserve_order_%v", year)
}

// CheckReserveByID 核销
func (d *Dao) CheckReserveByID(ctx context.Context, id int64, year int) (int64, error) {
	row, err := d.db.Exec(ctx, fmt.Sprintf(_actInterReserveOrderCheckedSQL, getInterReserveOrderTableName(year)), id)
	if err != nil {
		return 0, errors.Wrap(err, "UpUserAward")
	}
	return row.RowsAffected()
}

// RawMidInterReserveID 查看用户预约情况
func (d *Dao) RawMidInterReserveID(ctx context.Context, mid int64, interReserveId int64, year int) (data *bwsonline.InterReserveOrder, err error) {
	data = new(bwsonline.InterReserveOrder)
	row := d.db.QueryRow(ctx, fmt.Sprintf(_actInterReserveOrderByMidInterReserveIDSQL, getInterReserveOrderTableName(year)), mid, interReserveId)
	if err := row.Scan(&data.Id, &data.Mid, &data.TicketNo, &data.InterReserveId, &data.OrderNo, &data.ReserveNo, &data.IsChecked, &data.Ctime, &data.Mtime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawDress:QueryRow")
	}
	return data, nil
}

func (d *Dao) RawInterReserveOrderByMid(ctx context.Context, mid int64, year int) (orders []*bwsonline.InterReserveOrder, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, fmt.Sprintf(_actInterReserveOrderByMidSQL, getInterReserveOrderTableName(year)), mid)
	if err != nil {
		return nil, errors.Wrap(err, "RawInterReserveOrderByMid Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsonline.InterReserveOrder)
		if err = rows.Scan(&r.Id, &r.Mid, &r.TicketNo, &r.InterReserveId, &r.OrderNo, &r.ReserveNo, &r.IsChecked, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawInterReserveOrderByMid Scan")
		}
		orders = append(orders, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawInterReserveOrderByMid rows")
	}
	return orders, nil
}

func (d *Dao) RawMidInterReserveIDs(ctx context.Context, mid int64, interReserveIds []int64, year int) (orders []*bwsonline.InterReserveOrder, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, fmt.Sprintf(_actInterReserveOrderByMidInterReserveIDsSQL, getInterReserveOrderTableName(year), xstr.JoinInts(interReserveIds)), mid)
	if err != nil {
		return nil, errors.Wrap(err, "RawMidInterReserveIDs Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsonline.InterReserveOrder)
		if err = rows.Scan(&r.Id, &r.Mid, &r.TicketNo, &r.InterReserveId, &r.OrderNo, &r.IsChecked, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawMidInterReserveIDs Scan")
		}
		orders = append(orders, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawMidInterReserveIDs rows")
	}
	return orders, nil
}

func (d *Dao) AddInterReserveOrder(ctx context.Context, mid int64, ticketNo string, interReserveId int64, orderNo string, year int, reserveNo int) (int64, error) {
	row, err := d.db.Exec(ctx, fmt.Sprintf(_addInterReserveOrderSQL, getInterReserveOrderTableName(year)), mid, ticketNo, interReserveId, orderNo, reserveNo)
	if err != nil {
		return 0, errors.Wrap(err, "AddInterReserveOrder")
	}
	return row.RowsAffected()
}

func (d *Dao) RawInterReserveOrderByOrderNos(ctx context.Context, orderNos []string, year int) (orders []*bwsonline.InterReserveOrder, err error) {
	var rows *sql.Rows
	var newOrderNos []string
	for _, v := range orderNos {
		newOrderNos = append(newOrderNos, "\""+v+"\"")
	}
	rows, err = d.db.Query(ctx, fmt.Sprintf(_actInterReserveOrderByOrderNos, getInterReserveOrderTableName(year), strings.Join(newOrderNos, ",")))
	if err != nil {
		return nil, errors.Wrap(err, "RawInterReserveByIds Query")
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsonline.InterReserveOrder)
		if err = rows.Scan(&r.Id, &r.Mid, &r.TicketNo, &r.InterReserveId, &r.OrderNo, &r.ReserveNo, &r.IsChecked, &r.Ctime, &r.Mtime); err != nil {
			return nil, errors.Wrap(err, "RawInterReserveOrderByMid Scan")
		}
		orders = append(orders, r)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawInterReserveOrderByMid rows")
	}
	return orders, nil
}
