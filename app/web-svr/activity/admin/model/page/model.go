package page

// bcurd -dsn='test_3306:UJPZaGKjpb2ylFx3HNhmLuwOYft4MCAi@tcp(172.22.34.101:3306)/bilibili_lottery?parseTime=true'  -schema=bilibili_lottery -table=act_page -tmpl=bilibili_log.tmpl > act_page.go

import (
	"context"
	"go-gateway/app/web-svr/activity/admin/model/stime"

	"go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
)

// ActPage represents a row from 'act_page'.
type ActPage struct {
	ID         int64      `json:"id"`          // 自增ID, 无意义
	State      int8       `json:"state"`       // 活动状态 0-正常，1-关闭评论
	Stime      stime.Time `json:"stime"`       // 开始时间
	Etime      stime.Time `json:"etime"`       // 结束时间
	Ctime      stime.Time `json:"ctime"`       // record create timestamp
	Mtime      stime.Time `json:"mtime"`       // record update/modify timestamp
	Name       string     `json:"name"`        // 活动名称
	Author     string     `json:"author"`      // 活动作者
	PcURL      string     `json:"pc_url"`      // 活动地址
	Rank       uint32     `json:"rank"`        // 排序接口
	H5URL      string     `json:"h5_url"`      // h5地址
	PcCover    string     `json:"pc_cover"`    // pc封面
	H5Cover    string     `json:"h5_cover"`    // h5封面
	PageName   string     `json:"page_name"`   // 自定义上传名
	Plat       int8       `json:"plat"`        // 平台 1,web,2app,3web and app
	Desc       string     `json:"desc"`        // 活动描述
	Click      uint64     `json:"click"`       // 点击量
	Type       int32      `json:"type"`        // 分区id
	Mold       uint8      `json:"mold"`        // 模式
	Series     uint32     `json:"series"`      // 系列
	Dept       uint32     `json:"dept"`        // 部门 0默认
	ReplyID    int32      `json:"reply_id"`    // 评论id
	TpID       int32      `json:"tp_id"`       // 模板id
	Ptime      stime.Time `json:"ptime"`       // 发布时间
	Catalog    int32      `json:"catalog"`     // 目录id
	Creator    string     `json:"creator"`     // 创建人姓名
	SpmID      string     `json:"spm_id"`      // spm id
	RelatedUID uint32     `json:"related_uid"` // 关联uid
}

// TableName ...
func (ActPage) TableName() string {
	return "act_page"
}

// Insert  Insert a record
func (a *ActPage) Insert(ctx context.Context, db *sql.DB) error {
	var err error

	const sqlstr = `INSERT INTO  act_page (` +
		` state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid` +
		`) VALUES (` +
		` ?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?` +
		`)`

	result, err := db.Exec(ctx, sqlstr, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
	if err != nil {
		log.Error("ActPage Insert Exec err: %v", err)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	a.ID = int64(id)

	return nil
}

// ActPageDelete Delete by primary key:`id`
func ActPageDelete(ctx context.Context, db *sql.DB, id int64) error {
	var err error

	const sqlstr = `DELETE FROM  act_page WHERE id = ?`

	_, err = db.Exec(ctx, sqlstr, id)
	if err != nil {
		log.Error("ActPageDelete err: %v", err)
		return err
	}
	return nil
}

// ActPageByID   Select a record by primary key:`id`
func ActPageByID(ctx context.Context, db *sql.DB, id int64) (*ActPage, error) {
	var err error

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE id = ?`

	a := ActPage{}

	err = db.QueryRow(ctx, sqlstr, id).Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		log.Error("ActPageByID QueryRow err: %v id: %v", err, id)
		return nil, err
	}

	return &a, nil
}

// Update Update a record
func (a *ActPage) Update(ctx context.Context, db *sql.DB) error {
	var err error

	const sqlstr = `UPDATE act_page SET ` +
		` state = ? ,stime = ? ,etime = ? ,ctime = ? ,mtime = ? ,name = ? ,author = ? ,pc_url = ? ,rank = ? ,h5_url = ? ,pc_cover = ? ,h5_cover = ? ,page_name = ? ,plat = ? ,desc = ? ,click = ? ,type = ? ,mold = ? ,series = ? ,dept = ? ,reply_id = ? ,tp_id = ? ,ptime = ? ,catalog = ? ,creator = ? ,spm_id = ? ,related_uid = ? ` +
		` WHERE id = ?`

	_, err = db.Exec(ctx, sqlstr, a.State, a.Stime, a.Etime, a.Ctime, a.Mtime, a.Name, a.Author, a.PcURL, a.Rank, a.H5URL, a.PcCover, a.H5Cover, a.PageName, a.Plat, a.Desc, a.Click, a.Type, a.Mold, a.Series, a.Dept, a.ReplyID, a.TpID, a.Ptime, a.Catalog, a.Creator, a.SpmID, a.RelatedUID, a.ID)
	if err != nil {
		log.Error("ActPage Update err: %v", err)
	}
	return err
}

// ActPageByState Select by index name:`state`
func ActPageByState(ctx context.Context, db *sql.DB, state int8) ([]*ActPage, error) {

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE state = ? `
	q, err := db.Query(ctx, sqlstr, state)
	if err != nil {
		log.Error("ActPageByState Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res := []*ActPage{}
	for q.Next() {
		a := ActPage{}

		err = q.Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
		if err != nil {
			log.Error("ActPageByState Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Error("ActPageByState Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

// ActPageByRank Select by index name:`rank`
func ActPageByRank(ctx context.Context, db *sql.DB, rank uint32) ([]*ActPage, error) {

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE rank = ? `
	q, err := db.Query(ctx, sqlstr, rank)
	if err != nil {
		log.Error("ActPageByRank Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res := []*ActPage{}
	for q.Next() {
		a := ActPage{}

		err = q.Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
		if err != nil {
			log.Error("ActPageByRank Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Error("ActPageByRank Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

// ActPageByPlat Select by index name:`plat`
func ActPageByPlat(ctx context.Context, db *sql.DB, plat int8) ([]*ActPage, error) {

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE plat = ? `
	q, err := db.Query(ctx, sqlstr, plat)
	if err != nil {
		log.Error("ActPageByPlat Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res := []*ActPage{}
	for q.Next() {
		a := ActPage{}

		err = q.Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
		if err != nil {
			log.Error("ActPageByPlat Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Error("ActPageByPlat Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

// ActPageByClick Select by index name:`click`
func ActPageByClick(ctx context.Context, db *sql.DB, click uint64) ([]*ActPage, error) {

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE click = ? `
	q, err := db.Query(ctx, sqlstr, click)
	if err != nil {
		log.Error("ActPageByClick Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res := []*ActPage{}
	for q.Next() {
		a := ActPage{}

		err = q.Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
		if err != nil {
			log.Error("ActPageByClick Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Error("ActPageByClick Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

// ActPageByType Select by index name:`type`
func ActPageByType(ctx context.Context, db *sql.DB, _type int32) ([]*ActPage, error) {

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE type = ? `
	q, err := db.Query(ctx, sqlstr, _type)
	if err != nil {
		log.Error("ActPageByType Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res := []*ActPage{}
	for q.Next() {
		a := ActPage{}

		err = q.Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
		if err != nil {
			log.Error("ActPageByType Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Error("ActPageByType Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

// ActPageByMold Select by index name:`mold`
func ActPageByMold(ctx context.Context, db *sql.DB, mold uint8) ([]*ActPage, error) {

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE mold = ? `
	q, err := db.Query(ctx, sqlstr, mold)
	if err != nil {
		log.Error("ActPageByMold Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res := []*ActPage{}
	for q.Next() {
		a := ActPage{}

		err = q.Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
		if err != nil {
			log.Error("ActPageByMold Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Error("ActPageByMold Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

// ActPageBySeries Select by index name:`series`
func ActPageBySeries(ctx context.Context, db *sql.DB, series uint32) ([]*ActPage, error) {

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE series = ? `
	q, err := db.Query(ctx, sqlstr, series)
	if err != nil {
		log.Error("ActPageBySeries Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res := []*ActPage{}
	for q.Next() {
		a := ActPage{}

		err = q.Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
		if err != nil {
			log.Error("ActPageBySeries Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Error("ActPageBySeries Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

// ActPageByDept Select by index name:`dept`
func ActPageByDept(ctx context.Context, db *sql.DB, dept uint32) ([]*ActPage, error) {

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE dept = ? `
	q, err := db.Query(ctx, sqlstr, dept)
	if err != nil {
		log.Error("ActPageByDept Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res := []*ActPage{}
	for q.Next() {
		a := ActPage{}

		err = q.Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
		if err != nil {
			log.Error("ActPageByDept Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Error("ActPageByDept Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}

// ActPageByStime Select by index name:`idx_stime`
func ActPageByStime(ctx context.Context, db *sql.DB, stime xtime.Time) ([]*ActPage, error) {

	const sqlstr = `SELECT ` +
		`id,state,stime,etime,ctime,mtime,name,author,pc_url,rank,h5_url,pc_cover,h5_cover,page_name,plat,desc,click,type,mold,series,dept,reply_id,tp_id,ptime,catalog,creator,spm_id,related_uid ` +
		`FROM  act_page ` +
		`WHERE stime = ? `
	q, err := db.Query(ctx, sqlstr, stime)
	if err != nil {
		log.Error("ActPageByStime Query err: %v", err)
		return nil, err
	}
	defer q.Close()

	res := []*ActPage{}
	for q.Next() {
		a := ActPage{}

		err = q.Scan(&a.ID, &a.State, &a.Stime, &a.Etime, &a.Ctime, &a.Mtime, &a.Name, &a.Author, &a.PcURL, &a.Rank, &a.H5URL, &a.PcCover, &a.H5Cover, &a.PageName, &a.Plat, &a.Desc, &a.Click, &a.Type, &a.Mold, &a.Series, &a.Dept, &a.ReplyID, &a.TpID, &a.Ptime, &a.Catalog, &a.Creator, &a.SpmID, &a.RelatedUID)
		if err != nil {
			log.Error("ActPageByStime Scan err: %v", err)
			return nil, err
		}
		res = append(res, &a)
	}
	if q.Err() != nil {
		log.Error("ActPageByStime Err() err: %v ", err)
		return nil, err
	}

	return res, nil
}
