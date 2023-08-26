package dao

import (
	"context"

	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"
)

const (
	_inviteSQL       = "SELECT id,uid,mid,score,ctime FROM act_vogue_user_invite WHERE uid=? AND id < ? ORDER BY id DESC LIMIT 20"
	_insertInviteSQL = "INSERT IGNORE INTO act_vogue_user_invite (uid,mid,score) VALUE (?,?,?)"
)

func (d *Dao) RawInviteList(c context.Context, uid int64, id int64) (res []*model.Invite, err error) {
	res = make([]*model.Invite, 0, 0)
	rows, err := d.db.Query(c, _inviteSQL, uid, id)
	if err != nil {
		log.Error("dmReader.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &model.Invite{}
		if err = rows.Scan(&r.Id, &r.Uid, &r.Mid, &r.Score, &r.Ctime); err != nil {
			log.Error("row.Scan() error(%v)", err)
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}

func (d *Dao) InsertInvite(c context.Context, uid, mid, score int64) (affect int64, err error) {
	res, err := d.db.Exec(c, _insertInviteSQL, uid, mid, score)
	if err != nil {
		log.Error("d.InsertTask(%v,%v,%v) error(%v)", uid, mid, score, err)
		return
	}
	return res.RowsAffected()
}
