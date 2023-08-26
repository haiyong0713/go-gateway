package show

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/log"
	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
)

const (
	_getMaterial = "SELECT id,title,`desc`,cover,gifcover,corner,power_pic_sun,power_pic_night,width,height,reason,reason_content" +
		" FROM pos_material WHERE `state` = 1 "
)

func (d *Dao) GetMaterial(c context.Context) (rcs []*pb2.Material, err error) {
	rows, err := d.dbMgr.Query(c, _getMaterial)
	if err != nil {
		log.Error("dao.GetMaterial query error (%+v) sql(%+v)", err, _getMaterial)
		return
	}
	defer rows.Close()
	for rows.Next() {
		rc := &pb2.Material{}
		if err = rows.Scan(&rc.Id, &rc.Title, &rc.Desc, &rc.Cover, &rc.Gifcover, &rc.Corner, &rc.PowerPicSun,
			&rc.PowerPicNight, &rc.Width, &rc.Height, &rc.Reason, &rc.ReasonContent); err != nil {
			log.Error("dao.GetMaterial Scan err(%+v) sql(%+v)", err, _getMaterial)
			return
		}
		rcs = append(rcs, rc)
	}
	err = rows.Err()
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rcs, nil
}

func (d *Dao) GetMaterialMap(c context.Context) (materialMap map[int64]*pb2.Material, err error) {
	materialMap = make(map[int64]*pb2.Material)
	materialList, err := d.GetMaterial(c)
	if err != nil {
		log.Error("dao.GetMaterialMap GetMaterial err(%+v)", err)
		return
	}

	for _, v := range materialList {
		materialMap[v.Id] = v
	}
	return
}

func (d *Dao) SetMaterial2Cache(ctx context.Context, key string, expire int64, materialMap map[int64]*pb2.Material) (err error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	materialJsonStr, err := json.Marshal(materialMap)
	if err != nil {
		log.Error("dao:SetMaterial2Cache json.Marshal err(%+v)", err)
		return err
	}

	_, err = redis.String(conn.Do("SETEX", key, expire, materialJsonStr))
	if err != nil {
		log.Error("dao:SetMaterial2Cache SET cache err(%+v) KEY(%+v)", err, key)
		return err
	}
	return err
}

func (d *Dao) GetMaterialFromCache(ctx context.Context, key string) (materialMap map[int64]*pb2.Material, err error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	materialJsonStr, err := redis.String(conn.Do("GET", key))
	if err != nil {
		if err != redis.ErrNil {
			log.Error("dao:GetTabExtFromCache GET cache err(%+v) KEY(%+v)", err, key)
		}
		return
	}

	if materialJsonStr == "" {
		return
	}
	if err = json.Unmarshal([]byte(materialJsonStr), &materialMap); err != nil {
		log.Error("dao:GetTabExtFromCache json.Unmarshal err(%+v)", err)
	}
	return
}
