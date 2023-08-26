package http

import (
	"strconv"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

func managerInfo(c *bm.Context) (uid int64, username string) {
	if nameInter, ok := c.Get("username"); ok {
		username = nameInter.(string)
	}
	if uidInter, ok := c.Get("uid"); ok {
		uid = uidInter.(int64)
	}
	if username == "" {
		cookie, err := c.Request.Cookie("username")
		if err != nil {
			log.Error("managerInfo get cookie error (%v)", err)
			return
		}
		username = cookie.Value
		c, err := c.Request.Cookie("uid")
		if err != nil {
			log.Error("managerInfo get cookie error (%v)", err)
			return
		}
		uidInt, _ := strconv.Atoi(c.Value)
		uid = int64(uidInt)
	}
	return
}
