package guess

import (
	"context"
	"fmt"

	"go-common/library/log"
)

func mdKey(mainID, business int64) string {
	return fmt.Sprintf("gmd_%d_%d", mainID, business)
}

func oKey(oid, business int64) string {
	return fmt.Sprintf("go_%d_%d", oid, business)
}

func muKey(mainID, mid int64) string {
	return fmt.Sprintf("mu_%d_%d", mainID, mid)
}

func mainKey(mainID int64) string {
	return fmt.Sprintf("gm_%d", mainID)
}

func statKey(mid, stakeType, business int64) string {
	return fmt.Sprintf("st_%d_%d_%d", mid, stakeType, business)
}

func userKey(mid, business int64) string {
	return fmt.Sprintf("gu_%d_%d", mid, business)
}

// DelGuessCache delete oid guess cache.
func (d *Dao) DelGuessCache(c context.Context, oid, business, mainID int64) (err error) {
	oidK := oKey(oid, business)
	mdK := mdKey(mainID, business)
	mK := mainKey(mainID)
	conn := d.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("DEL", oidK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", oidK, err)
		return
	}
	if err = conn.Send("DEL", mdK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", mdK, err)
		return
	}
	if err = conn.Send("DEL", mK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", mK, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) DeleteUserGuessLogCache(ctx context.Context, mid, business int64) (err error) {
	conn := d.redis.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Do("DEL", userKey(mid, business))

	return
}

// DelUserCache delete user cache.
func (d *Dao) DelUserCache(c context.Context, mid, stakeType, business, mainID int64) (err error) {
	statK := statKey(mid, stakeType, business)
	muK := muKey(mainID, mid)
	userK := userKey(mid, business)
	conn := d.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("DEL", statK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", statK, err)
		return
	}
	if err = conn.Send("DEL", muK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", muK, err)
		return
	}
	if err = conn.Send("DEL", userK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", userK, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 3; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// DelStatCache delete user stat cache.
func (d *Dao) DelStatCache(c context.Context, mid, stakeType, business int64) (err error) {
	statK := statKey(mid, stakeType, business)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", statK); err != nil {
		log.Error("d.DelStatCache,error(%v)", err)
	}
	return
}
