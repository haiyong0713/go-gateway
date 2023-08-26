package dynamic

import (
	"math"
)

const (
	_auditBitModule      = 0
	_auditBitCollectTemp = 1
)

type AuditContent int64

func (ac *AuditContent) IsModule() bool {
	return (*ac>>_auditBitModule)&1 == 1
}

func (ac *AuditContent) IsCollectTemp() bool {
	return (*ac>>_auditBitCollectTemp)&1 == 1
}

func (ac *AuditContent) SetModule() {
	*ac = *ac | (1 << _auditBitModule)
}

func (ac *AuditContent) SetCollectTemp() {
	*ac = *ac | (1 << _auditBitCollectTemp)
}

func (ac *AuditContent) UnsetCollectTemp() {
	*ac = *ac & (math.MaxInt64 - 1<<_auditBitCollectTemp)
}
