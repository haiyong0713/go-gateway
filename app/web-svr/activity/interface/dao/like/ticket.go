package like

import (
	"context"
	"database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const _sqlTicketByID = "SELECT id, ticket, state, ctime, mtime FROM act_electronic_ticket WHERE ticket = ? LIMIT 1"

func (d *Dao) GetTicketByCode(c context.Context, id string) (ti *like.Ticket, err error) {
	ti = new(like.Ticket)
	row := d.db.QueryRow(c, _sqlTicketByID, id)
	if err = row.Scan(&ti.ID, &ti.Ticket, &ti.State, &ti.Ctime, &ti.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "GetTicketByCode:row.Scan error(%v)", err)
		}
	}
	return
}

const _sqlTicketUpdateState = "UPDATE act_electronic_ticket SET state = ? WHERE id = ?"

func (d *Dao) UpdateTicketState(c context.Context, id int64, state uint8) (int64, error) {
	if res, err := d.db.Exec(c, _sqlTicketUpdateState, state, id); err != nil {
		log.Error("UpdateTicketState: db.Exec(%d,%d) error(%v)", state, id, err)
		return 0, err
	} else {
		return res.RowsAffected()
	}
}
