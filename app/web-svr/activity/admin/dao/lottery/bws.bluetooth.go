package lottery

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

const (
	_bluetoothUpLimitSQL   = "SELECT `id`,`mid`,`blue_key`,`bid`,`desc`,`ctime`,`mtime` FROM act_bws_bluetooth_ups WHERE bid=? AND del=0 LIMIT ?,?"
	_bluetoothUpCountSQL   = "SELECT count(id) FROM act_bws_bluetooth_ups WHERE bid=? AND del=0"
	_inBluetoothUpSQL      = "INSERT IGNORE INTO act_bws_bluetooth_ups (`bid`,`mid`,`blue_key`,`desc`) VALUES (?,?,?,?) ON DUPLICATE KEY UPDATE `bid`=?,`mid`=?,`blue_key`=?,`desc`=?"
	_delBluetoothUpSQL     = "UPDATE act_bws_bluetooth_ups SET del=1 WHERE bid=? AND del=0"
	_delBluetoothUpByIDSQL = "UPDATE act_bws_bluetooth_ups SET del=1 WHERE id=?"
	_upBluetoothUpSQL      = "UPDATE act_bws_bluetooth_ups SET `mid`=?,`blue_key`=?,`desc`=? WHERE id=?"
)

func (d *Dao) BluetoothUpLimit(ctx context.Context, bid int64, pn, ps int) ([]*bwsmdl.BluetoothUp, error) {
	rows, err := d.db.Query(ctx, _bluetoothUpLimitSQL, bid, pn, ps)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*bwsmdl.BluetoothUp
	for rows.Next() {
		u := &bwsmdl.BluetoothUp{}
		if err := rows.Scan(&u.Id, &u.Mid, &u.Key, &u.Bid, &u.Desc, &u.Ctime, &u.Mtime); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res = append(res, u)
	}
	err = rows.Err()
	return res, nil
}

func (d *Dao) BluetoothUpCount(ctx context.Context, bid int64) (int, error) {
	row := d.db.QueryRow(ctx, _bluetoothUpCountSQL, bid)
	var count int
	if err := row.Scan(&count); err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return count, nil
}

// InBluetoothUp .
func (d *Dao) InBluetoothUp(tx *xsql.Tx, bid, mid int64, key, desc string) error {
	_, err := tx.Exec(_inBluetoothUpSQL, bid, mid, key, desc, bid, mid, key, desc)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

// DelBluetoothUp .
func (d *Dao) DelBluetoothUp(tx *xsql.Tx, bid int64) error {
	_, err := tx.Exec(_delBluetoothUpSQL, bid)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

// DelBluetoothUpByID .
func (d *Dao) DelBluetoothUpByID(ctx context.Context, id int64) error {
	_, err := d.db.Exec(ctx, _delBluetoothUpByIDSQL, id)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

// UpBluetoothUp .
func (d *Dao) UpBluetoothUp(ctx context.Context, id, mid int64, key, desc string) error {
	if _, err := d.db.Exec(ctx, _upBluetoothUpSQL, mid, key, desc, id); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}
