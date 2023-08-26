package service

import (
	"context"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	"github.com/pkg/errors"
)

var (
	LegalItemTypes = map[int32]struct{}{
		model.PlayItemUGC:   {},
		model.PlayItemOGV:   {},
		model.PlayItemAudio: {},
	}
)

func validatePlayItem(_ context.Context, p *v1.PlayItem, subIdLenLimit int) error {
	if p == nil {
		return errors.WithMessage(ecode.RequestErr, "unexpected nil PlayItem")
	}
	if _, ok := LegalItemTypes[p.ItemType]; !ok {
		return errors.WithMessagef(ecode.RequestErr, "unknown item type %d", p.ItemType)
	}
	if p.Oid <= 0 || len(p.SubId) < subIdLenLimit {
		return errors.WithMessagef(ecode.RequestErr, "malformed PlayItem(%+v)", p)
	}
	if subIdLenLimit > 0 {
		for i := range p.SubId {
			if p.SubId[i] <= 0 {
				return errors.WithMessagef(ecode.RequestErr, "malformed PlayItem SubId(%+v)", p)
			}
		}
	}
	return nil
}

func DevNetAuthFromCtx(ctx context.Context) (*device.Device, *network.Network, *auth.Auth) {
	n, ok := network.FromContext(ctx)
	if !ok {
		n = network.Network{}
	}
	d, ok := device.FromContext(ctx)
	if !ok {
		d = device.Device{}
	}
	a, ok := auth.FromContext(ctx)
	if !ok {
		a = auth.Auth{}
	}
	return &d, &n, &a
}

//nolint:deadcode,unused
func min(n1, n2 int) int {
	if n1 < n2 {
		return n1
	}
	return n2
}

//nolint:deadcode,unused
func max(n1, n2 int) int {
	if n1 > n2 {
		return n1
	}
	return n2
}
