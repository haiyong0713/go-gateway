package player

import (
	"context"
	"go-common/library/database/sql"
	"go-gateway/app/app-svr/resource/service/conf"
	"go-gateway/app/app-svr/resource/service/model"
	"time"
)

// Dao struct user of color entry Dao.
type Dao struct {
	db *sql.DB
	c  *conf.Config
}

// New create a instance of color entry Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:  c,
		db: sql.NewMySQL(c.DB.Player),
	}
	return
}

// Close close db resource.
func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
}

// -------------------------- 业务dao --------------------------
const (
	dateFormat = "2006-01-02 15:04:05"
	//notPushedTime    = "2009-12-31 23:59:59"
	getEffectivePanelSQL = `SELECT 
		id, tids, btn_img, btn_text, text_color, link, label, display_stage, operator, priority 
	FROM 
		customized_panels 
	WHERE 
		online_status = 1 AND is_deprecated = 0 AND stime <= ? AND etime >= ?
	ORDER BY 
		display_stage desc, priority asc`
)

// APP_ENTRY
// 获取所有在线Entry
func (d *Dao) GetEffectivePanels(ctx context.Context) (result []*model.CustomizedPanel, err error) {
	now := time.Now().Format(dateFormat)
	rows, err := d.db.Query(ctx, getEffectivePanelSQL, now, now)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		panel := &model.CustomizedPanel{}
		if err = rows.Scan(
			&panel.ID, &panel.Tids, &panel.BtnImg, &panel.BtnText,
			&panel.TextColor, &panel.Link, &panel.Label, &panel.DisplayStage,
			&panel.Operator, &panel.Priority); err != nil {
			return
		}
		result = append(result, panel)
	}

	err = rows.Err()
	return result, err
}
