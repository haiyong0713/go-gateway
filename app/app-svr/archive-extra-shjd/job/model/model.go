package model

const (
	BinlogUpdate = "update"
	BinlogInsert = "insert"

	TableArchiveExtraBiz = "archive_extra_biz"

	ArchiveExtraBinlog = "ArchiveExtra-Binlog"
)

type ArchiveExtraBizMsg struct {
	Action string           `json:"action"`
	Table  string           `json:"table"`
	New    *ArchiveExtraBiz `json:"new"`
	Old    *ArchiveExtraBiz `json:"old"`
}

type ArchiveExtraBiz struct {
	Aid       int64  `json:"aid"`
	BizType   string `json:"biz_type"`
	BizValue  string `json:"biz_value"`
	IsDeleted int    `json:"is_deleted"`
}
