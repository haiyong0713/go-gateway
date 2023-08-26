package thirdValidate

import (
	//nolint:gofmt
	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
)

type Validator interface {
	CompareLength(compared []int64) bool
	SpyOnDetail() error
}

type AccountCardValidator struct {
	AccountCards *accountgrpc.CardsReply
}

func (acv *AccountCardValidator) CompareLength(rawMids []int64) {
	//返回值为空
	if acv == nil || acv.AccountCards == nil || len(acv.AccountCards.Cards) == 0 {
		log.Error("CompareLength account is empty mids(%v)", rawMids)
		return
	}
	for _, v := range rawMids {
		av, ok := acv.AccountCards.Cards[v]
		if !ok {
			log.Error("CompareLength account not exist mid(%d)", v)
			continue
		}
		if av == nil {
			log.Error("CompareLength account is nil mid(%d)", v)
			continue
		}
		if av.Name == "" {
			log.Error("CompareLength account name is empty mid(%d) result(%v)", v, av)
			continue
		}
	}
}
