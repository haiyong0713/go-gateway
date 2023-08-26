package utils

import (
	ipdb "github.com/ipipdotnet/ipdb-go"

	toolmdl "go-gateway/app/app-svr/fawkes/service/model/tool"
)

func ParseIP(addr, name string) (*toolmdl.IPInfo, error) {
	db, err := ipdb.NewCity(name)
	if err != nil {
		return nil, err
	}
	if err = db.Reload(name); err != nil {
		return nil, err
	}
	var cityInfo map[string]string
	if cityInfo, err = db.FindMap(addr, "CN"); err != nil {
		return nil, err
	}
	return &toolmdl.IPInfo{
		Country:   cityInfo["country_name"],
		Province:  cityInfo["region_name"],
		City:      cityInfo["city_name"],
		Latitude:  cityInfo["latitude"],
		Longitude: cityInfo["longitude"],
		ISP:       cityInfo["isp_domain"],
	}, nil
}
