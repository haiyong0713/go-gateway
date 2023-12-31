package telecom

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
)

const (
	_phoneKey = "phone_%d"
	_payKey   = "pay_%d"
)

// AddPhoneCode
func (d *Dao) AddPhoneCode(c context.Context, phone int, captcha string) (err error) {
	conn := d.phoneRds.Conn(c)
	defer conn.Close()
	key := fmt.Sprintf(_phoneKey, phone)
	if _, err = conn.Do("SETEX", key, d.phoneKeyExpired, captcha); err != nil {
		log.Error("telecom_AddPhoneCode add conn.Do SETEX error(%v)", err)
		return
	}
	return
}

// PhoneCode
func (d *Dao) PhoneCode(c context.Context, phone int) (captcha string, err error) {
	conn := d.phoneRds.Conn(c)
	defer conn.Close()
	key := fmt.Sprintf(_phoneKey, phone)
	if captcha, err = redis.String(conn.Do("GET", key)); err != nil {
		log.Error("telecom_get conn.Do GET error(%v)", err)
		err = ecode.NotModified
		return
	}
	return
}

// AddPayPhone add phone and requestNo
func (d *Dao) AddPayPhone(c context.Context, requestNo int64, phone string) (err error) {
	conn := d.phoneRds.Conn(c)
	defer conn.Close()
	key := fmt.Sprintf(_payKey, requestNo)
	if _, err = conn.Do("SETEX", key, d.payKeyExpired, phone); err != nil {
		log.Error("telecom_AddPhoneCode add conn.Do SETEX error(%v)", err)
		return
	}
	return
}

// PayPhone requestNo by phone
func (d *Dao) PayPhone(c context.Context, requestNo int64) (phone string, err error) {
	conn := d.phoneRds.Conn(c)
	defer conn.Close()
	key := fmt.Sprintf(_payKey, requestNo)
	if phone, err = redis.String(conn.Do("GET", key)); err != nil {
		log.Error("telecom_get conn.Do GET requestNo(%v) error(%v)", requestNo, err)
		err = ecode.NothingFound
		return
	}
	return
}
