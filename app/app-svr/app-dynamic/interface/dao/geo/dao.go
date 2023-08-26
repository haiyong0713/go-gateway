package geo

import (
	"context"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	geogrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/lbs-geo"
)

type Dao struct {
	c         *conf.Config
	geoClient geogrpc.LbsGeoClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.geoClient, err = geogrpc.NewClient(c.GeoGRPC); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) GeoCoder(c context.Context, lat, lng float64, from string) (*geogrpc.GeoCoderReply, error) {
	res, err := d.geoClient.GeoCoder(c, &geogrpc.GeoCoderReq{
		Lat:  lat,
		Lng:  lng,
		From: from,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
