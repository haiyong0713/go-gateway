package dao

import (
	"context"

	"go-gateway/app/web-svr/space/interface/model"

	"github.com/pkg/errors"
)

const _photoMallListSQL = "SELECT id,product_name,price,ios_price,coin_type,vip_free,s_img,l_img,android_img,iphone_img,ipad_img,thumbnail_img,sort_num,type,is_disable FROM photo_mall WHERE is_disable = 0"

func (d *Dao) PhotoMallList(c context.Context) ([]*model.PhotoMall, error) {
	rows, err := d.db.Query(c, _photoMallListSQL)
	if err != nil {
		err = errors.Wrap(err, "PhotoMallList d.db.Query")
		return nil, err
	}
	defer rows.Close()
	var list []*model.PhotoMall
	for rows.Next() {
		r := new(model.PhotoMall)
		if err = rows.Scan(&r.Id, &r.ProductName, &r.Price, &r.IosPrice, &r.CoinType, &r.VipFree, &r.SImg, &r.LImg, &r.AndroidImg, &r.IphoneImg, &r.IpadImg, &r.ThumbnailImg, &r.SortNum, &r.Type, &r.IsDisable); err != nil {
			err = errors.Wrap(err, "PhotoMallList row.Scan")
			return nil, err
		}
		list = append(list, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "PhotoMallList rows.Err")
		return nil, err
	}
	return list, nil
}
