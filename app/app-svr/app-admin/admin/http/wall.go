package http

import (
	"time"

	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-admin/admin/model/wall"
)

// walls select walls all
func walls(c *bm.Context) {
	c.JSON(wallSvc.Walls(c))
}

// wallByID select wall by id
func wallByID(c *bm.Context) {
	v := &wall.Param{}
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(wallSvc.WallByID(c, v.ID))
}

// saveWall insert or update wall
func saveWall(c *bm.Context) {
	var (
		err error
		v   = &wall.Param{}
	)
	if err = c.Bind(v); err != nil {
		return
	}
	if v.ID > 0 {
		err = wallSvc.UpdateWall(c, v, time.Now())
	} else {
		err = wallSvc.Insert(c, v, time.Now())
	}
	c.JSON(nil, err)
}

// publish update state
func publish(c *bm.Context) {
	v := &wall.Param{}
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, wallSvc.Publish(c, v.IDs, time.Now()))
}
