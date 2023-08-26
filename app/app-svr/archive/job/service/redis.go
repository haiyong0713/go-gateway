package service

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	_maxMSET = 100
)

func (s *Service) redisMSetWithExp(c context.Context, kvMap map[string][]byte, exp int64) {
	if len(kvMap) == 0 {
		return
	}

	keys := make([]string, 0, len(kvMap))
	for key := range kvMap {
		keys = append(keys, key)
	}

	for k, pool := range s.arcRedises {
		k := k
		pool := pool
		if err := func() error {
			conn := pool.Get(c)
			defer conn.Close()

			// 使用redis pipeline批量提交
			for i := 0; i < len(keys); i += _maxMSET {
				var partKeys []string
				if i+_maxMSET > len(keys) {
					partKeys = keys[i:]
				} else {
					partKeys = keys[i : i+_maxMSET]
				}

				args := redis.Args{}
				for _, key := range partKeys {
					args = args.Add(key).Add(kvMap[key])
				}

				if err := conn.Send("MSET", args...); err != nil {
					log.Error("redisMSetWithExp conn.Send() MSET k(%+v) partKeys(%+v) err(%+v)", k, partKeys, err)
					return err
				}

				for _, key := range partKeys {
					if err := conn.Send("EXPIRE", key, exp); err != nil {
						log.Error("redisMSetWithExp conn.Send() EXPIRE k(%+v) key(%+v) err(%+v)", k, key, err)
						return err
					}
				}

				if err := conn.Flush(); err != nil {
					log.Error("redisMSetWithExp conn.Flush() k(%+v) partKeys(%+v) err(%+v)", k, partKeys, err)
					return err
				}

				for i := 0; i < len(partKeys)+1; i++ {
					if _, err := conn.Receive(); err != nil {
						log.Error("redisMSetWithExp conn.Receive() k(%+v) partKeys(%+v) err(%+v)", k, partKeys, err)
						return err
					}
				}
			}
			return nil
		}(); err != nil {
			log.Error("redisMSetWithExp fail keys(%+v) k(%+v) err(%+v)", keys, k, err)
			return
		}
	}
}
