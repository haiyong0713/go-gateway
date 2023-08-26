package poll

import (
	"time"

	"go-common/library/database/sql"
	model "go-gateway/app/web-svr/activity/interface/model/poll"
)

// TxAddVote is
func (d *Dao) TxAddVote(tx *sql.Tx, arg *model.PollVote) error {
	if arg.VoteAt <= 0 {
		arg.VoteAt = time.Now().Unix()
	}
	if _, err := tx.Exec(
		`INSERT INTO act_poll_vote(poll_id,mid,poll_option_id,ticket_count,vote_at) VALUES (?,?,?,?,?)`,
		arg.PollId, arg.Mid, arg.PollOptionId, arg.TicketCount, arg.VoteAt); err != nil {
		return err
	}
	return nil
}

// TxIncrOptionStat is
func (d *Dao) TxIncrOptionStat(tx *sql.Tx, pollID int64, optionID int64, ticketCount int64) error {
	res, err := tx.Exec(
		`UPDATE act_poll_option_stat SET ticket_sum=ticket_sum+?, vote_sum=vote_sum+1 WHERE poll_option_id=? AND poll_id=? LIMIT 1`,
		ticketCount, optionID, pollID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}

	if _, err := tx.Exec(
		`INSERT INTO act_poll_option_stat(poll_id,poll_option_id,ticket_sum,vote_sum) VALUES(?,?,?,1) ON DUPLICATE KEY UPDATE ticket_sum=ticket_sum+VALUES(ticket_sum), vote_sum=vote_sum+1`,
		pollID, optionID, ticketCount,
	); err != nil {
		return err
	}
	return nil
}
