package poll

import (
	"context"
	"sort"
	"time"

	model "go-gateway/app/web-svr/activity/interface/model/poll"

	"go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
)

// PollOptionStatTop is
func (d *Dao) PollOptionStatTop(ctx context.Context, pollID int64, count int64) ([]*model.PollOptionStat, error) {
	opts, ok := d.ListPollOption(ctx, pollID)
	if !ok {
		return []*model.PollOptionStat{}, nil
	}

	optIDs := make(map[int64]struct{}, len(opts))
	for _, o := range opts {
		optIDs[o.Id] = struct{}{}
	}

	rows, err := d.db.Query(ctx, `SELECT id,poll_id,poll_option_id,ticket_sum,vote_sum FROM act_poll_option_stat WHERE poll_id=?`, pollID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*model.PollOptionStat{}
	for rows.Next() {
		item := &model.PollOptionStat{}
		if err := rows.Scan(&item.Id, &item.PollId, &item.PollOptionId, &item.TicketSum, &item.VoteSum); err != nil {
			log.Error("Failed to scan poll option stat: %+v", err)
			continue
		}
		out = append(out, item)
	}

	filtered := []*model.PollOptionStat{}
	for _, os := range out {
		if _, ok := optIDs[os.PollOptionId]; !ok {
			continue
		}
		filtered = append(filtered, os)
	}

	result := make([]*model.PollOptionStat, 0, len(filtered))
	sort.Slice(filtered, func(i int, j int) bool {
		return filtered[i].TicketSum > filtered[j].TicketSum
	})
	for i, s := range filtered {
		if int64(i) >= count {
			break
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// RawPollVoteUserStatByDate is
func (d *Dao) RawPollVoteUserStatByDate(ctx context.Context, mid int64, pollID int64, date time.Time) (*model.PollVoteUserStat, error) {
	inDate := xtime.Time(asDate(date).Unix())
	row := d.db.QueryRow(ctx,
		`SELECT id,mid,poll_id,date,vote_count FROM act_poll_vote_user_stat WHERE mid=? AND poll_id=? AND date=?`,
		mid, pollID, inDate,
	)
	out := &model.PollVoteUserStat{}
	if err := row.Scan(&out.Id, &out.Mid, &out.PollId, &out.Date, &out.VoteCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return out, nil
}

// RawLastPollVoteUserStat is
func (d *Dao) RawLastPollVoteUserStat(ctx context.Context, mid int64, pollID int64) (*model.PollVoteUserStat, error) {
	row := d.db.QueryRow(ctx,
		`SELECT id,mid,poll_id,date,vote_count FROM act_poll_vote_user_stat WHERE mid=? AND poll_id=? ORDER BY id DESC LIMIT 1`,
		mid, pollID,
	)
	out := &model.PollVoteUserStat{}
	if err := row.Scan(&out.Id, &out.Mid, &out.PollId, &out.Date, &out.VoteCount); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return out, nil
}

// InitPollVoteUserStat is
func (d *Dao) InitPollVoteUserStat(ctx context.Context, mid int64, pollID int64, date time.Time) bool {
	inDate := xtime.Time(asDate(date).Unix())
	res, err := d.db.Exec(ctx, `INSERT IGNORE INTO act_poll_vote_user_stat(mid,poll_id,date) VALUES(?,?,?)`, mid, pollID, inDate)
	if err != nil {
		log.Error("Failed to init poll vote user stat: %+v", err)
		return false
	}
	affectCount, err := res.RowsAffected()
	if err != nil {
		log.Error("Failed to get init poll result: %+v", err)
		return false
	}
	return affectCount > 0
}

func asDate(in time.Time) time.Time {
	return time.Date(in.Year(), in.Month(), in.Day(), 0, 0, 0, 0, in.Location())
}

// TxIncrDailyUserVoteStat is
func (d *Dao) TxIncrDailyUserVoteStat(tx *sql.Tx, mid int64, pollID int64, date time.Time, maxVoteCount int64) (bool, error) {
	inDate := xtime.Time(asDate(date).Unix())
	res, err := tx.Exec(
		`UPDATE act_poll_vote_user_stat SET vote_count=vote_count+1 WHERE mid=? AND poll_id=? AND date=? AND vote_count<?`,
		mid, pollID, inDate, maxVoteCount,
	)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected > 0 {
		return true, nil
	}
	return false, nil
}
