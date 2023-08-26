package preferenceproto

import (
	"fmt"
	"strings"

	"go-gateway/app/app-svr/distribution/distribution/internal/distributionconst"
	"go-gateway/app/app-svr/distribution/distribution/internal/sessioncontext"

	"go-common/library/log"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

const (
	DoubleValue = "DoubleValue"
	FloatValue  = "FloatValue"
	Int64Value  = "Int64Value"
	UInt64Value = "UInt64Value"
	Int32Value  = "Int32Value"
	UInt32Value = "UInt32Value"
	BoolValue   = "BoolValue"
	StringValue = "StringValue"
	BytesValue  = "BytesValue"
)

var (
	DoubleValueFQN = distributionconst.MakeFullyQualifiedName(DoubleValue)
	FloatValueFQN  = distributionconst.MakeFullyQualifiedName(FloatValue)
	Int64ValueFQN  = distributionconst.MakeFullyQualifiedName(Int64Value)
	UInt64ValueFQN = distributionconst.MakeFullyQualifiedName(UInt64Value)
	Int32ValueFQN  = distributionconst.MakeFullyQualifiedName(Int32Value)
	UInt32ValueFQN = distributionconst.MakeFullyQualifiedName(UInt32Value)
	BoolValueFQN   = distributionconst.MakeFullyQualifiedName(BoolValue)
	StringValueFQN = distributionconst.MakeFullyQualifiedName(StringValue)
	BytesValueFQN  = distributionconst.MakeFullyQualifiedName(BytesValue)
)

type PreferenceMeta struct {
	ProtoDesc     *desc.MessageDescriptor
	storageDriver string
	disabled      bool
	feature       []string
	preference    string
}

type Preference struct {
	Meta    PreferenceMeta
	Message *dynamic.Message
}

func (p *PreferenceMeta) KeyBuilder() func(sessioncontext.SessionContext) string {
	return func(ctx sessioncontext.SessionContext) string {
		featureLabel := []string{
			fmt.Sprintf("{buvid:%s}", ctx.Device().Buvid),
			fmt.Sprintf("fp_local:%s", ctx.Device().FpLocal),
		}
		for _, f := range p.feature {
			switch f {
			case "mid":
				featureLabel = append(featureLabel, fmt.Sprintf("mid:%d", ctx.Mid()))
			default:
				value, ok := ctx.ExtraContextValue(f)
				if !ok {
					log.Warn("Unrecognized preference feature: %q in preference: %q", f, p.ProtoDesc.GetFullyQualifiedName())
					continue
				}
				featureLabel = append(featureLabel, fmt.Sprintf("%s:%s", f, value))
			}
		}
		featureLabel = append(featureLabel, p.ProtoDesc.GetFullyQualifiedName())
		return strings.Join(featureLabel, "/")
	}
}

func (p *PreferenceMeta) StorageDriver() string {
	return p.storageDriver
}

func (p *PreferenceMeta) Disabled() bool {
	return p.disabled
}

func (p *PreferenceMeta) Feature() []string {
	return p.feature
}

func (p *PreferenceMeta) Preference() string {
	return p.preference
}
