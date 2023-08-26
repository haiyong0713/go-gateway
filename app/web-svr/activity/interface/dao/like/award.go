package like

import (
	"context"
	"database/sql"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-common/library/cache/redis"
	"go-common/library/log"
	lmdl "go-gateway/app/web-svr/activity/interface/model/like"
)

const (
	_awardSubjectSQL     = "SELECT id,name,etime,sid,type,source_id,source_expire,`state`,ctime,mtime,sid_type,other_sids,task_id FROM act_award_subject WHERE sid = ?"
	_awardSubjectByIDSQL = "SELECT id,name,etime,sid,type,source_id,source_expire,`state`,ctime,mtime,sid_type,other_sids,task_id FROM act_award_subject WHERE id = ?"
)

func keyAwardSubject(sid int64) string {
	return fmt.Sprintf("awa_sub_%d", sid)
}

func keyAwardSubjectByID(id int64) string {
	return fmt.Sprintf("awa_sub_id_%d", id)
}

func (d *Dao) RawAwardSubject(c context.Context, sid int64) (res *lmdl.AwardSubject, err error) {
	row := d.db.QueryRow(c, _awardSubjectSQL, sid)
	res = &lmdl.AwardSubject{}
	if err = row.Scan(&res.ID, &res.Name, &res.Etime, &res.Sid, &res.Type, &res.SourceId, &res.SourceExpire, &res.State, &res.Ctime, &res.Mtime, &res.SidType, &res.OtherSids, &res.TaskID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("RawAwardSubject sid(%d) error(%v)", sid, err)
		}
	}
	return
}

func (d *Dao) RawAwardSubjectByID(c context.Context, id int64) (res *lmdl.AwardSubject, err error) {
	row := d.db.QueryRow(c, _awardSubjectByIDSQL, id)
	res = &lmdl.AwardSubject{}
	if err = row.Scan(&res.ID, &res.Name, &res.Etime, &res.Sid, &res.Type, &res.SourceId, &res.SourceExpire, &res.State, &res.Ctime, &res.Mtime, &res.SidType, &res.OtherSids, &res.TaskID); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			log.Error("RawAwardSubjectByID id(%d) error(%v)", id, err)
		}
	}
	return
}

func (d *Dao) CacheAwardSubject(c context.Context, sid int64) (res *lmdl.AwardSubject, err error) {
	var (
		key = keyAwardSubject(sid)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheAwardSubject conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	res = new(lmdl.AwardSubject)
	if err = res.Unmarshal(bs); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (d *Dao) AddCacheAwardSubject(c context.Context, sid int64, data *lmdl.AwardSubject) (err error) {
	var (
		key = keyAwardSubject(sid)
		bs  []byte
	)
	if bs, err = data.Marshal(); err != nil {
		log.Error("json.Marshal(%+v) error (%v)", data, err)
		return
	}
	if _, err = component.GlobalRedis.Do(c, "SETEX", key, d.awardSubjectExpire, bs); err != nil {
		log.Error("conn.Do(SET, %s, %d) error(%v)", key, d.awardSubjectExpire, err)
	}
	return
}

func (d *Dao) DelCacheAwardSubject(c context.Context, sid, id int64) (err error) {
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	key := keyAwardSubject(sid)
	idKey := keyAwardSubjectByID(id)
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
	}
	if err = conn.Send("DEL", idKey); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", idKey, err)
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) CacheAwardSubjectByID(c context.Context, id int64) (res *lmdl.AwardSubject, err error) {
	var (
		key = keyAwardSubjectByID(id)
		bs  []byte
	)
	if bs, err = redis.Bytes(component.GlobalRedis.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheAwardSubject conn.Do(GET key(%v)) error(%v)", key, err)
		}
		return
	}
	res = new(lmdl.AwardSubject)
	if err = res.Unmarshal(bs); err != nil {
		log.Error("json.Unmarshal(%v) error(%v)", bs, err)
	}
	return
}

func (d *Dao) AddCacheAwardSubjectByID(c context.Context, id int64, data *lmdl.AwardSubject) (err error) {
	var (
		key = keyAwardSubjectByID(id)
		bs  []byte
	)
	if bs, err = data.Marshal(); err != nil {
		log.Error("json.Marshal(%+v) error (%v)", data, err)
		return
	}
	if _, err = component.GlobalRedis.Do(c, "SETEX", key, d.awardSubjectExpire, bs); err != nil {
		log.Error("conn.Do(SET, %s, %d) error(%v)", key, d.awardSubjectExpire, err)
	}
	return
}

func (d *Dao) DelCacheAwardSubjectByID(c context.Context, id int64) (err error) {
	key := keyAwardSubjectByID(id)
	_, err = component.GlobalRedis.Do(c, "DEL", key)
	return
}
