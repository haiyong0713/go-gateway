package manager

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"

	pb "go-gateway/app/app-svr/resource/service/api/v1"
	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
)

const (
	_specialSQL        = "SELECT `id`,`title`,`desc`,`cover`,`scover`,`re_type`,`re_value`,`corner`,`size`,`card` FROM special_card WHERE `id` > ? ORDER BY `id` LIMIT ?"
	_getSpecialCardSQl = "SELECT id,title,`desc`,cover,scover,gifcover,bgcover,reason,tab_uri,re_type,re_value,corner,size,power_pic_sun,power_pic_night,width,height,url FROM special_card WHERE mtime >= ? AND id > ? ORDER BY id ASC LIMIT ? "
	_getSpecialByIdSQL = "SELECT id,title,`desc`,cover,scover,gifcover,bgcover,reason,tab_uri,re_type,re_value,corner,size,power_pic_sun,power_pic_night,width,height,url FROM special_card WHERE id = ? "
)

// Specials get specials cards from DB
func (d *Dao) Specials(c context.Context, offset int) (sps map[int64]*pb.SpecialReply, nextId int, err error) {
	rows, err := d.db.Query(c, _specialSQL, offset, 1000)
	if err != nil {
		return
	}
	defer rows.Close()
	sps = make(map[int64]*pb.SpecialReply, 1000)
	for rows.Next() {
		sc := &pb.SpecialReply{}
		if err = rows.Scan(&sc.Id, &sc.Title, &sc.Desc, &sc.Cover, &sc.Scover, &sc.ReType, &sc.ReValue, &sc.Corner, &sc.Siz, &sc.Card); err != nil {
			return
		}
		sps[sc.Id] = sc
		nextId = int(sc.Id)
	}
	err = rows.Err()
	return
}

// 分页获取特殊卡数据
func (d *Dao) GetSpecialCard(c context.Context, now xtime.Time, offset int64, pageSize int) (res map[int64]*pb2.AppSpecialCard, nextId int64, err error) {
	res = make(map[int64]*pb2.AppSpecialCard)
	rows, err := d.db.Query(c, _getSpecialCardSQl, now, offset, pageSize)
	if err != nil {
		log.Error("dao.GetSpecialCard Query now(%+v) offset(%d) pageSize(%d) err(%+v)", now, offset, pageSize, err)
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		special := &pb2.AppSpecialCard{}
		if err = rows.Scan(&special.Id, &special.Title, &special.Desc, &special.Cover, &special.Scover, &special.Gifcover,
			&special.Bgcover, &special.Reason, &special.TabUri, &special.ReType, &special.ReValue, &special.Corner,
			&special.Size_, &special.PowerPicSun, &special.PowerPicNight, &special.Width, &special.Height, &special.Url); err != nil {
			log.Error("dao.GetSpecialCard Scan err(%+v)", err)
			return nil, 0, err
		}
		res[special.Id] = special
		nextId = special.Id
	}
	if err = rows.Err(); err != nil {
		log.Error("dao.GetSpecialCard rows.err err(+%v)", err)
	}
	return
}

// 根据ID查询特殊卡信息
func (d *Dao) GetSpecialCardById(c context.Context, id int64) (res *pb2.AppSpecialCard, err error) {
	res = &pb2.AppSpecialCard{}
	row := d.db.QueryRow(c, _getSpecialByIdSQL, id)
	if err = row.Scan(&res.Id, &res.Title, &res.Desc, &res.Cover, &res.Scover, &res.Gifcover, &res.Bgcover, &res.Reason,
		&res.TabUri, &res.ReType, &res.ReValue, &res.Corner, &res.Size_, &res.PowerPicSun, &res.PowerPicNight, &res.Width,
		&res.Height, &res.Url); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("dao.GetSpecialCardById Scan id(%d) err(%+v)", id, err)
		return nil, err
	}
	return
}
