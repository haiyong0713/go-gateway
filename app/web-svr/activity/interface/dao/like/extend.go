package like

import (
	"context"
	"fmt"

	xsql "go-common/library/database/sql"
	"go-common/library/xstr"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_extendsSQL = "SELECT `lid`,`like` FROM `like_extend` WHERE `lid` IN (%s)"
)

// LikeActSums get like_action likes sum data .
func (dao *Dao) LikeActSums(c context.Context, lids []int64) (res map[int64]int64, err error) {
	var rows *xsql.Rows
	if rows, err = dao.db.Query(c, fmt.Sprintf(_extendsSQL, xstr.JoinInts(lids))); err != nil {
		err = errors.Wrap(err, "LikeActSums:Query")
		return
	}
	defer rows.Close()
	res = make(map[int64]int64, 0)
	for rows.Next() {
		a := &likemdl.LidLikeSum{}
		if err = rows.Scan(&a.Likes, &a.Lid); err != nil {
			err = errors.Wrap(err, "LikeActSums:Scan")
			return
		}
		res[a.Lid] = a.Likes
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "LikeActSums:rows.Err()")
	}
	return
}
