package model

import "go-common/library/log"

const (
	RedisWhiteListKey        = "ARCHIVE_WHITELIST"
	RedisAuthorWhiteListKey  = "USER_WHITELIST_%s"
	RedisBatchToPushKey      = "BATCH_TO_PUSH"
	RedisBatchToPushBVIDsKey = "BATCH_TO_PUSH_BVIDS_%d"
	RedisBatchToPushTimeKey  = "BATCH_TO_PUSH_TIME_%d"
	RedisBatchToPushLockKey  = "BATCH_TO_PUSH_LOCK"
	Deprecated               = 1
	NotDeprecated            = 0
	DefaultTimeLayout        = "2006-01-02 15:04:05"
)

type Page struct {
	Num   int   `json:"num"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

type Hosts struct {
	Manager string `json:"manager"`
	API     string `json:"api"`
}

type ApplicationConfig struct {
	Debug  bool
	Log    *log.Config
	Export *ExportConfig
	Push   *PushConfig
}

type ExportConfig struct {
	ArchivePushBatch *ArchivePushBatchExportConfig
}

type ArchivePushBatchExportConfig struct {
	FilenameFormat string
	Columns        []string
	Titles         []string
}

type PushConfig struct {
	MaxRetryCount int
}

type BaseHTTPResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
