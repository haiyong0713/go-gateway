package feature

import (
	"go-gateway/app/app-svr/app-feed/admin/model/feature"
)

func (d *Dao) GroupAppPlats() map[string][]*feature.AppPlatItem {
	appPlats := make(map[string][]*feature.AppPlatItem, len(d.plats))
	for _, plat := range d.plats {
		if _, ok := appPlats[plat.Type]; !ok {
			appPlats[plat.Type] = make([]*feature.AppPlatItem, 0, len(d.plats))
		}
		item := &feature.AppPlatItem{}
		item.FormPlat(plat)
		appPlats[plat.Type] = append(appPlats[plat.Type], item)
	}
	return appPlats
}

func (d *Dao) Plats() map[string]*feature.Plat {
	plats := make(map[string]*feature.Plat, len(d.plats))
	for _, plat := range d.plats {
		plats[plat.MobiApp] = plat
	}
	return plats
}
