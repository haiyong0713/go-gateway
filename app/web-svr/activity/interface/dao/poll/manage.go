package poll

import (
	"context"

	"go-common/library/log"

	model "go-gateway/app/web-svr/activity/interface/model/poll"
)

const (
	_pollOptions       = "SELECT id,poll_id,title,image,`group`,is_deleted FROM act_poll_option"
	_pollOptionsDelete = `UPDATE act_poll_option SET is_deleted=1 WHERE id=? LIMIT 1`
	_pollOptionsAdd    = "INSERT INTO act_poll_option(poll_id,title,image,`group`) VALUES(?,?,?,?)"
	_pollOptionsUpdate = "UPDATE act_poll_option SET title=?, image=?, `group`=? WHERE id=? LIMIT 1"
)

// PollOptions is
func (d *Dao) PollOptions(ctx context.Context, pollID int64) ([]*model.PollOption, error) {
	rows, err := d.db.Query(ctx, _pollOptions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.PollOption{}
	for rows.Next() {
		item := &model.PollOption{}
		if err := rows.Scan(&item.Id, &item.PollId, &item.Title, &item.Image, &item.Group, &item.IsDeleted); err != nil {
			log.Warn("Failed to scan poll option: %+v", err)
			continue
		}
		if item.PollId != pollID {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// PollOptionsDelete is
func (d *Dao) PollOptionsDelete(ctx context.Context, pollOptionID int64) error {
	_, err := d.db.Exec(ctx, _pollOptionsDelete, pollOptionID)
	if err != nil {
		return err
	}
	return nil
}

// PollOptionsUpdate is
func (d *Dao) PollOptionsUpdate(ctx context.Context, pollOptionID int64, title string, image string, group string) error {
	_, err := d.db.Exec(ctx, _pollOptionsUpdate, title, image, group, pollOptionID)
	if err != nil {
		return err
	}
	return nil
}

// PollOptionsAdd is
func (d *Dao) PollOptionsAdd(ctx context.Context, pollID int64, title string, image string, group string) error {
	_, err := d.db.Exec(ctx, _pollOptionsAdd, pollID, title, image, group)
	if err != nil {
		return err
	}
	return nil
}
