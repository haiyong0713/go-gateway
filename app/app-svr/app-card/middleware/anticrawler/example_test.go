package anticrawler

import (
	"go-common/library/ecode"
	"go-common/library/net/http/blademaster"

	"github.com/pkg/errors"
)

func Example() {
	c := &Config{
		LogID:  "009236",
		Worker: 10,
		Buffer: 10240,
		Infoc:  nil,
	}
	Init(c)
	engine := blademaster.Default()
	engine.GET("/users/profile", Report(), func(c *blademaster.Context) {
		values := c.Request.URL.Query()
		name := values.Get("name")
		age := values.Get("age")

		err := errors.New("error from others") // error from other call
		if err != nil {
			// mark this response should be degraded
			c.JSON(nil, ecode.Degrade)
			return
		}
		c.JSON(map[string]string{"name": name, "age": age}, nil)
	})
	engine.GET("/users/index", Report(), func(c *blademaster.Context) {
		c.String(200, "%s", "Title: User")
	})
	engine.GET("/users/list", Report(), func(c *blademaster.Context) {
		c.JSON([]string{"user1", "user2", "user3"}, nil)
	})
	engine.Run(":18080")
}
