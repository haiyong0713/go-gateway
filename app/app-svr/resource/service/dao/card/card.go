package card

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"

	api "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/conf"
)

const (
	// db show
	_followSQL = "SELECT `id`,`type`,`long_title`,`content` FROM `card_follow` WHERE `deleted`=0"
	// db manager
	_getSpecialSQL  = "SELECT s.id,s.title,s.desc,s.cover,s.scover,s.gifcover,s.bgcover,s.reason,s.tab_uri,s.re_type,s.re_value,s.corner,s.size,s.power_pic_sun,s.power_pic_night,s.width,s.height FROM special_card AS s,pos_rec AS p WHERE s.id=p.avid AND p.state=1"
	_getConvergeSQL = "SELECT id,re_type,re_value,title,cover,content FROM content_card"
	_getDownloadSQL = "SELECT `id`,`title`,`desc`,`icon`,`cover`,`url_type`,`url_value`,`btn_txt`,`re_type`,`re_value`,`number`,`double_cover` FROM download_card"
	// pos_rec card
	_getPosRecSQL = "SELECT id,title,card_desc,cover,gifcover,power_pic_sun,power_pic_night,width,height,bgcover,re_type,re_value FROM pos_rec WHERE state=1"
)

type Dao struct {
	db  *sql.DB
	dbm *sql.DB
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db:  sql.NewMySQL(c.DB.Show),
		dbm: sql.NewMySQL(c.DB.Manager),
	}
	return
}

func (d *Dao) Follow(c context.Context) (res []*api.CardFollow, err error) {
	rows, err := d.db.Query(c, _followSQL)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		c := &api.CardFollow{}
		if err = rows.Scan(&c.Id, &c.Type, &c.Title, &c.Content); err != nil {
			log.Error("%+v", err)
			return
		}
		res = append(res, c)
	}
	err = rows.Err()
	return
}

func (d *Dao) SpecialCard(c context.Context) (res []*api.SpecialCard, err error) {
	rows, err := d.dbm.Query(c, _getSpecialSQL)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		sc := &api.SpecialCard{}
		if err = rows.Scan(&sc.Id, &sc.Title, &sc.Desc, &sc.Cover, &sc.SingleCover, &sc.GifCover, &sc.BgCover, &sc.Reason, &sc.TabUri, &sc.ReType, &sc.ReValue, &sc.Badge, &sc.Size_, &sc.PowerPicSun, &sc.PowerPicNight, &sc.PowerPicWidth, &sc.PowerPicHeight); err != nil {
			return
		}
		res = append(res, sc)
	}
	err = rows.Err()
	return
}

func (d *Dao) ConvergeCards(c context.Context) (res []*api.ConvergeCard, err error) {
	rows, err := d.dbm.Query(c, _getConvergeSQL)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		c := &api.ConvergeCard{}
		if err = rows.Scan(&c.Id, &c.ReType, &c.ReValue, &c.Title, &c.Cover, &c.Content); err != nil {
			return
		}
		res = append(res, c)
	}
	err = rows.Err()
	return
}

func (d *Dao) DownLoad(c context.Context) (res []*api.DownLoadCard, err error) {
	rows, err := d.dbm.Query(c, _getDownloadSQL)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		d := &api.DownLoadCard{}
		if err = rows.Scan(&d.Id, &d.Title, &d.Desc, &d.Icon, &d.Cover, &d.UrlType, &d.UrlValue, &d.BtnTxt, &d.ReType, &d.ReValue, &d.Number, &d.DoubleCover); err != nil {
			return
		}
		res = append(res, d)
	}
	err = rows.Err()
	return
}

func (d *Dao) PosRec(c context.Context) (map[int64]*api.CardPosRec, error) {
	res := map[int64]*api.CardPosRec{}
	rows, err := d.dbm.Query(c, _getPosRecSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		p := &api.CardPosRec{}
		if err := rows.Scan(&p.Id, &p.Title, &p.CardDesc, &p.Cover, &p.Gifcover, &p.PowerPicSun, &p.PowerPicNight, &p.Width, &p.Height, &p.Bgcover, &p.ReType, &p.ReValue); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
		res[p.Id] = p
	}
	return res, rows.Err()
}
