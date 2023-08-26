package experimentalflag

import (
	"context"
	"fmt"
	"go-common/library/log"
	"strings"
	"time"

	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	"go-gateway/app/app-svr/distribution/distribution/internal/sessioncontext"

	parabox "git.bilibili.co/bapis/bapis-go/community/interface/parabox"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

const (
	expMessageFQN = "bilibili.app.distribution.experimental.v1.Exp"
)

// ExperimentalFlag is used to demostrace how external storage driver worked.
type ExperimentalFlag struct {
	parabox parabox.ParaboxClient
}

func New(parabox parabox.ParaboxClient) *ExperimentalFlag {
	return &ExperimentalFlag{
		parabox: parabox,
	}
}

func (ExperimentalFlag) Name() string {
	return "experimental-flag"
}

func castAsPreferenceExp(in []*parabox.Exp) ([]*dynamic.Message, string, error) {
	expDesc, ok := preferenceproto.TryGetMessage(expMessageFQN)
	if !ok {
		return nil, "", errors.Errorf("Failed to get %q descriptor", expMessageFQN)
	}
	now := time.Now()
	out := make([]*dynamic.Message, 0, len(in))
	for _, i := range in {
		dm := dynamic.NewMessage(expDesc)
		idValue := preferenceproto.DistributionPrimitiveType.NewInt64Value()
		if err := idValue.TrySetFieldByName("value", i.Id); err != nil {
			log.Error("Failed to set value id: %+v", err)
			continue
		}
		if err := idValue.TrySetFieldByName("last_modified", now.Unix()); err != nil {
			log.Error("Failed to set value id last_modified: %+v", err)
			continue
		}
		bucketValue := preferenceproto.DistributionPrimitiveType.NewInt32Value()
		if err := bucketValue.TrySetFieldByName("value", i.Bucket); err != nil {
			log.Error("Failed to set value bucket: %+v", err)
			continue
		}
		if err := bucketValue.TrySetFieldByName("last_modified", now.Unix()); err != nil {
			log.Error("Failed to set value bucket last_modified: %+v", err)
			continue
		}
		if err := dm.TrySetFieldByName("id", idValue); err != nil {
			log.Error("Failed to set exp.id: %+v", err)
			continue
		}
		if err := dm.TrySetFieldByName("bucket", bucketValue); err != nil {
			log.Error("Failed to set exp.bucket: %+v", err)
			continue
		}
		out = append(out, dm)
	}

	bucketFlag := []string{}
	for _, i := range in {
		bucketFlag = append(bucketFlag, fmt.Sprintf("%d=%d", i.Id, i.Bucket))
	}
	flag := strings.Join(bucketFlag, ",")
	return out, flag, nil
}

func (e ExperimentalFlag) GetUserPreference(ctx context.Context, metas []*preferenceproto.PreferenceMeta) ([]*preferenceproto.Preference, error) {
	_, ok := sessioncontext.FromContext(ctx)
	if !ok {
		return nil, errors.Errorf("Session context is required")
	}

	if len(metas) > 1 {
		return nil, errors.Errorf("Invalid experimental meta: %+v", metas)
	}

	reply, err := e.parabox.GetExpsAppColdStart(ctx, &parabox.GetExpsAppColdStartReq{})
	if err != nil {
		return nil, err
	}
	exps, flag, err := castAsPreferenceExp(reply.Exps)
	if err != nil {
		return nil, err
	}
	flagValue := preferenceproto.DistributionPrimitiveType.NewStringValue()
	if err := flagValue.TrySetFieldByName("value", flag); err != nil {
		return nil, err
	}

	out := make([]*preferenceproto.Preference, 0, len(metas))
	for _, m := range metas {
		dm := dynamic.NewMessage(m.ProtoDesc)
		if err := dm.TrySetFieldByName("exps", exps); err != nil {
			log.Error("Failed to set exps field: %+v: %+v", exps, err)
			continue
		}
		if err := dm.TrySetFieldByName("flag", flagValue); err != nil {
			log.Error("Failed to set flag field: %+v: %+v", exps, err)
			continue
		}
		out = append(out, &preferenceproto.Preference{
			Meta:    *m,
			Message: dm,
		})
	}
	return out, nil
}

func (ExperimentalFlag) SetUserPreference(ctx context.Context, preferences []*preferenceproto.Preference) error {
	return nil
}
