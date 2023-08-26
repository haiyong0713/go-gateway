package fm

import (
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/job/model/fm"

	"github.com/pkg/errors"
)

// lock
func SetLock(conn redis.Conn, key string) (bool, error) {
	reply, err := redis.String(conn.Do("SET", key, time.Now().Unix(), "EX", 3, "NX"))
	if err != nil {
		if err == redis.ErrNil { // 锁已存在，未拿到
			return false, nil
		}
		return false, errors.Wrapf(err, "getLock key(%s)", key)
	}
	if reply != "OK" {
		return false, nil
	}
	return true, nil
}

// unlock
func DelLock(conn redis.Conn, key string) error {
	if _, err := conn.Do("DEL", key); err != nil {
		log.Error("DelLock error, key(%s)", key)
		return err
	}
	return nil
}

func SeasonInfoLock(req fm.SeasonInfoReq) string {
	if req.Scene == fm.SceneFm {
		if req.FmType == fm.AudioSeason {
			return fmt.Sprintf("lk_fmsi_%d", req.SeasonId)
		} else if req.FmType == fm.AudioSeasonUp {
			return fmt.Sprintf("lk_fmsiup_%d", req.SeasonId)
		}
	} else if req.Scene == fm.SceneVideo {
		return fmt.Sprintf("lk_vdsi_%d", req.SeasonId)
	}

	log.Error("lockSeasonInfo unknown req:%+v", req)
	return ""
}

func SeasonOidLock(scene fm.Scene, fmType fm.FmType, seasonId int64) string {
	if scene == fm.SceneFm {
		if fmType == fm.AudioSeason {
			return fmt.Sprintf("lk_fmso_%d", seasonId)
		} else if fmType == fm.AudioSeasonUp {
			return fmt.Sprintf("lk_fmsoup_%d", seasonId)
		}
	} else if scene == fm.SceneVideo {
		return fmt.Sprintf("lk_vdso_%d", seasonId)
	}
	log.Error("lockSeasonOid unknown scene:%s, fmType:%s", scene, fmType)
	return ""
}

func SeasonInfoKey(req fm.SeasonInfoReq) string {
	if req.Scene == fm.SceneFm {
		if req.FmType == fm.AudioSeason {
			return fmt.Sprintf("fmsi_%d", req.SeasonId)
		} else if req.FmType == fm.AudioSeasonUp {
			return fmt.Sprintf("fmsiup_%d", req.SeasonId)
		}
	} else if req.Scene == fm.SceneVideo {
		return fmt.Sprintf("vdsi_%d", req.SeasonId)
	}
	log.Error("seasonInfoKey unknown req:%+v", req)
	return ""
}

func SeasonOidKey(scene fm.Scene, fmType fm.FmType, seasonId int64) string {
	if scene == fm.SceneFm {
		if fmType == fm.AudioSeason {
			return fmt.Sprintf("fmso_%d", seasonId)
		} else if fmType == fm.AudioSeasonUp {
			return fmt.Sprintf("fmsoup_%d", seasonId)
		}
	} else if scene == fm.SceneVideo {
		return fmt.Sprintf("vdso_%d", seasonId)
	}
	log.Error("seasonOidKey unknown scene:%s, fmType:%s", scene, fmType)
	return ""
}
