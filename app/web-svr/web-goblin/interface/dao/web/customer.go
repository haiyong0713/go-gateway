package web

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-gateway/app/web-svr/web-goblin/interface/model/web"

	"github.com/pkg/errors"
)

const (
	_customerSQL = "SELECT c.id,c.customer_type,c.business_type,IFNULL(b.business , '') as business,IFNULL(b.logo,'') as logo,c.title,c.copywriting,c.highlight_title,c.image,c.web_url,c.h5_url,c.stime,c.etime,c.rank as c_rank,IFNULL(b.rank,0) as b_rank,IFNULL(b.customer_type,0) AS b_customer_type FROM gb_customer_centers AS c LEFT JOIN gb_customer_businesses AS b ON c.business_type = b.id  WHERE c.is_deleted = 0  ORDER BY customer_type ASC,b.rank DESC,c.rank DESC"
)

// CusCenter customer center.
func (d *Dao) CusCenter(ctx context.Context) (res map[int64][]*web.Customer, err error) {
	var rows *xsql.Rows
	res = map[int64][]*web.Customer{}
	if rows, err = d.db.Query(ctx, _customerSQL); err != nil {
		err = errors.Wrapf(err, "Customer d.db.Query error")
		return
	}
	defer rows.Close()
	for rows.Next() {
		c := &web.Customer{}
		if err = rows.Scan(&c.ID, &c.CustomerType, &c.BusinessType, &c.BusinessName, &c.Logo, &c.Title, &c.Copywriting, &c.HighlightTitle, &c.Image, &c.WebUrl, &c.H5Url, &c.Stime, &c.Etime, &c.CustomerRank, &c.BusinessRank, &c.BusinessCustomerType); err != nil {
			return
		}
		res[c.CustomerType] = append(res[c.CustomerType], c)
	}
	err = rows.Err()
	return
}
