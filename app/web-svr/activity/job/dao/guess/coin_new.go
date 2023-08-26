package guess

import (
	"context"
	"fmt"

	"go-common/library/database/sql"
)

const (
	sql4UserGuessLog = `
UPDATE act_guess_user_%v
SET status = 1, income = ?
WHERE status = 0
	AND id = ?
`
	sql4UserGuessArchive = `
UPDATE act_guess_user_log
SET total_success = total_success + ?, success_rate = CASE 
	WHEN total_guess = 0 THEN 0
	ELSE convert(total_success / total_guess, decimal(10, 2)) * 100
END, total_income = total_income + ?
WHERE mid = ?
	AND business = ?
`
	sql4UpdateUserGuessArchive = `
UPDATE act_guess_user_log
SET total_guess = ?, total_success = ?, success_rate = CASE 
	WHEN total_guess = 0 THEN 0
	ELSE convert(total_success / total_guess, decimal(10, 2)) * 100
END, total_income = ?
WHERE mid = ?
	AND business = ?
`
	sql4UserGuessAggregation = `
SELECT COUNT(1) AS total_guess
	, SUM(if(income > 0, 1, 0)) AS total_success
	, SUM(income) AS total_income
FROM act_guess_user_%v
WHERE mid = ?
`
)

func (d *Dao) FetchUserGuessStats(ctx context.Context, mid int64) (totalGuess, totalSuccess, totalIncome int64, err error) {
	row := d.db.QueryRow(ctx, fmt.Sprintf(sql4UserGuessAggregation, userHit(mid)), mid)
	err = row.Scan(
		&totalGuess,
		&totalSuccess,
		&totalIncome)

	return
}

func (d *Dao) UpdateUserGuessStats(ctx context.Context, mid, totalGuess, totalSuccess, totalIncome, business int64) error {
	_, err := d.db.Exec(ctx, sql4UpdateUserGuessArchive, totalGuess, totalSuccess, totalIncome, mid, business)

	return err
}

func (d *Dao) RepairUserGuessLogs(ctx context.Context, mid int64, m map[int64]float64) (err error) {
	if len(m) == 0 {
		return
	}

	var tx *sql.Tx
	tx, err = d.db.Begin(ctx)
	if err != nil {
		return
	}

	for k, v := range m {
		_, err = tx.Exec(fmt.Sprintf(sql4UpdateUserGuessLogCoins, userHit(mid)), v*_int, k)
		if err != nil {
			break
		}
	}

	if err == nil {
		err = tx.Commit()
	}

	if err != nil {
		err = tx.Rollback()
	}

	return
}

// Atomic update user guess log and archive
func (d *Dao) UpdateUserGuessRelations(ctx context.Context, mid, business, guessLogID int64, income float64) (err error) {
	var tx *sql.Tx

	tx, err = d.db.Begin(ctx)
	if err != nil {
		return
	}

	_, err = tx.Exec(fmt.Sprintf(sql4UserGuessLog, userHit(mid)), income*_int, guessLogID)
	if err == nil {
		switch income > 0 {
		case true:
			_, err = tx.Exec(sql4UserGuessArchive, 1, income*_int, mid, business)
		default:
			_, err = tx.Exec(sql4UserGuessArchive, 0, 0, mid, business)
		}
	}

	if err == nil {
		err = tx.Commit()
	} else {
		tmpErr := err
		err = tx.Rollback()
		if err == nil {
			err = tmpErr
		}
	}

	return
}
