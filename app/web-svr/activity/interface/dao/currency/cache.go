package currency

import (
	"context"
	"fmt"

	"go-gateway/app/web-svr/activity/interface/model/currency"
)

func currencyKey(id int64) string {
	return fmt.Sprintf("curr_%d", id)
}

func relationKey(businessID, foreignID int64) string {
	return fmt.Sprintf("curr_rela_%d_%d", businessID, foreignID)
}

func userCurrKey(mid, id int64) string {
	return fmt.Sprintf("curr_user_%d_%d", mid, id)
}

//go:generate kratos tool btsgen
type _bts interface {
	// get currency data by id.
	// cache
	Currency(c context.Context, id int64) (*currency.Currency, error)
	// get a foreign business currency id data.
	// cache
	Relation(c context.Context, businessID int64, foreignID int64) (*currency.CurrencyRelation, error)
	// get user currency.
	// cache
	CurrencyUser(c context.Context, mid int64, id int64) (*currency.CurrencyUser, error)
}

//go:generate kratos tool mcgen
type _mc interface {
	// mc: -key=currencyKey
	CacheCurrency(c context.Context, id int64) (*currency.Currency, error)
	// mc: -key=currencyKey -expire=d.currencyExpire -encode=pb
	AddCacheCurrency(c context.Context, id int64, data *currency.Currency) error
	// mc: -key=relationKey
	CacheRelation(c context.Context, businessID int64, foreignID int64) (*currency.CurrencyRelation, error)
	// mc: -key=relationKey -expire=d.currencyExpire -encode=pb
	AddCacheRelation(c context.Context, businessID int64, data *currency.CurrencyRelation, foreignID int64) error
	// mc: -key=userCurrKey
	CacheCurrencyUser(c context.Context, mid int64, currID int64) (*currency.CurrencyUser, error)
	// mc: -key=userCurrKey -expire=d.currencyExpire -encode=pb
	AddCacheCurrencyUser(c context.Context, mid int64, data *currency.CurrencyUser, currID int64) error
	// mc: -key=userCurrKey
	DelCacheCurrencyUser(c context.Context, mid int64, currID int64) error
}
