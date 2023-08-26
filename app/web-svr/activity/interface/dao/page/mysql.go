package page

import (
	"context"
	"go-common/library/database/sql"
	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/interface/model/page"
)

func (d *Dao) RawGetPageByID(c context.Context, id int64) (res *model.ActPage, err error) {
	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,` + "`desc`" + `,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id ` +
		`FROM  act_page ` +
		`WHERE id = ?`

	a := model.ActPage{}

	err = d.db.QueryRow(c, sqlstr, id).Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorc(c, "ActPageByID QueryRow err: %v id: %v", err, id)
		return nil, err
	}
	return &a, nil
}
