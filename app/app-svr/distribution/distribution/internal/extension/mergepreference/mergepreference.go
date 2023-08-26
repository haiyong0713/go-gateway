package mergepreference

import (
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/util"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"

	"go-common/library/log"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/dynamic"
)

func Init() {}

func MergePreference(dst *dynamic.Message, origin *dynamic.Message) {
	if dst.GetMessageDescriptor().GetFullyQualifiedName() != origin.GetMessageDescriptor().GetFullyQualifiedName() {
		log.Error("Mismatched message type: %q and %q", dst.GetMessageDescriptor().GetFullyQualifiedName(), origin.GetMessageDescriptor().GetFullyQualifiedName())
		return
	}

	for _, field := range dst.GetMessageDescriptor().GetFields() {
		if field.GetType() != dpb.FieldDescriptorProto_TYPE_MESSAGE {
			continue
		}
		if field.IsRepeated() {
			// skip merge in repeated field
			continue
		}
		fieldType := field.GetMessageType().GetFullyQualifiedName()
		if _, ok := preferenceproto.DistributionPrimitiveType.GetTypeDescriptor(fieldType); !ok {
			dmFieldV, err := util.GetFieldAsDynamicMessage(dst, field)
			if err != nil {
				log.Error("Failed to get dst field: %q: %+v", field.GetFullyQualifiedName(), err)
				continue
			}
			originDmFieldV, err := util.GetFieldAsDynamicMessage(origin, field)
			if err != nil {
				log.Error("Failed to get origin field: %q: %+v", field.GetFullyQualifiedName(), err)
				continue
			}

			if dmFieldV != nil {
				if originDmFieldV == nil {
					continue
				}
				MergePreference(dmFieldV, originDmFieldV)
				continue
			}

			if originDmFieldV == nil {
				continue
			}
			if err := dst.TrySetField(field, originDmFieldV); err != nil {
				log.Error("Failed to set field: %q as %+v: %+v", field.GetFullyQualifiedName(), originDmFieldV, err)
				continue
			}
			continue
		}

		dmFieldV, err := util.GetFieldAsDynamicMessage(dst, field)
		if err != nil {
			log.Error("Failed to get dst field: %q: %+v", field.GetFullyQualifiedName(), err)
			continue
		}
		if dmFieldV != nil {
			continue
		}

		originDmFieldV, err := util.GetFieldAsDynamicMessage(origin, field)
		if err != nil {
			log.Error("Failed to get origin field: %q: %+v", field.GetFullyQualifiedName(), err)
			continue
		}
		if originDmFieldV == nil {
			continue
		}
		if err := dst.TrySetField(field, originDmFieldV); err != nil {
			log.Error("Failed to set field: %q as %+v: %+v", field.GetFullyQualifiedName(), originDmFieldV, err)
			continue
		}
	}
}
