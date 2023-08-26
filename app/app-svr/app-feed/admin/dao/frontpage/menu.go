package frontpage

import (
	"go-gateway/app/app-svr/app-feed/admin/conf"
	model "go-gateway/app/app-svr/app-feed/admin/model/frontpage"
)

func (d *Dao) GetMenus() (res []*model.Menu, err error) {
	allMarks := d.GetCategoriesKeys()

	res = make([]*model.Menu, 0)
	if err = d.ORMResource.Model(model.Menu{}).Where("mark IN (?)", allMarks).Find(&res).Error; err != nil {
		return
	}
	for i := range res {
		res[i].Name = model.CategoriesMap[res[i].Mark]
	}
	res = append([]*model.Menu{conf.Conf.Frontpage.GlobalMenu}, res...)
	return
}

// GetCategoriesKeys
func (d *Dao) GetCategoriesKeys() (res []string) {
	res = make([]string, 0, len(model.CategoriesMap))
	for key := range model.CategoriesMap {
		res = append(res, key)
	}
	return
}
