package http

import (
	"go-common/library/ecode"

	bm "go-common/library/net/http/blademaster"

	dolbyMdl "go-gateway/app/app-svr/app-resource/interface/model/dolby"
)

func dolbyConfig(c *bm.Context) {
	var param = new(dolbyMdl.ConfigParam)
	if err := c.Bind(param); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(getDolbyCconfig(param), nil)
}

func getDolbyCconfig(req *dolbyMdl.ConfigParam) *dolbyMdl.ConfigReply {
	if config == nil || config.Dolby == nil {
		return nil
	}
	for _, v := range config.Dolby.DolbyConfig {
		if (v.Model == req.Model) && (v.Brand == req.Brand) {
			return &dolbyMdl.ConfigReply{
				File: v.File,
				Hash: v.Hash,
			}
		}
	}
	return nil
}
