package fm

import (
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"

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

func SeasonInfoLock(req fm_v2.SeasonInfoReq) string {
	if req.Scene == fm_v2.SceneFm {
		if req.FmType == fm_v2.AudioSeason {
			return fmt.Sprintf("lk_fmsi_%d", req.SeasonId)
		} else if req.FmType == fm_v2.AudioSeasonUp {
			return fmt.Sprintf("lk_fmsiup_%d", req.SeasonId)
		}
	} else if req.Scene == fm_v2.SceneVideo {
		return fmt.Sprintf("lk_vdsi_%d", req.SeasonId)
	}

	log.Error("lockSeasonInfo unknown req:%+v", req)
	return ""
}

func SeasonOidLock(scene fm_v2.Scene, fmType fm_v2.FmType, seasonId int64) string {
	if scene == fm_v2.SceneFm {
		if fmType == fm_v2.AudioSeason {
			return fmt.Sprintf("lk_fmso_%d", seasonId)
		} else if fmType == fm_v2.AudioSeasonUp {
			return fmt.Sprintf("lk_fmsoup_%d", seasonId)
		}
	} else if scene == fm_v2.SceneVideo {
		return fmt.Sprintf("lk_vdso_%d", seasonId)
	}
	log.Error("lockSeasonOid unknown scene:%s, fmType:%s", scene, fmType)
	return ""
}

func SeasonInfoKey(req fm_v2.SeasonInfoReq) string {
	if req.Scene == fm_v2.SceneFm {
		if req.FmType == fm_v2.AudioSeason {
			return fmt.Sprintf("fmsi_%d", req.SeasonId)
		} else if req.FmType == fm_v2.AudioSeasonUp {
			return fmt.Sprintf("fmsiup_%d", req.SeasonId)
		}
	} else if req.Scene == fm_v2.SceneVideo {
		return fmt.Sprintf("vdsi_%d", req.SeasonId)
	}
	log.Error("seasonInfoKey unknown req:%+v", req)
	return ""
}

func SeasonOidKey(scene fm_v2.Scene, fmType fm_v2.FmType, seasonId int64) string {
	if scene == fm_v2.SceneFm {
		if fmType == fm_v2.AudioSeason {
			return fmt.Sprintf("fmso_%d", seasonId)
		} else if fmType == fm_v2.AudioSeasonUp {
			return fmt.Sprintf("fmsoup_%d", seasonId)
		}
	} else if scene == fm_v2.SceneVideo {
		return fmt.Sprintf("vdso_%d", seasonId)
	}
	log.Error("seasonOidKey unknown scene:%s, fmType:%s", scene, fmType)
	return ""
}
