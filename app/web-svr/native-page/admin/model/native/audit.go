package native

const (
	_auditBitModule      = 0
	_auditBitCollectTemp = 1
)

type AuditContent int64

func (ac AuditContent) IsModule() bool {
	return (ac>>_auditBitModule)&1 == 1
}

func (ac AuditContent) IsCollectTemp() bool {
	return (ac>>_auditBitCollectTemp)&1 == 1
}
