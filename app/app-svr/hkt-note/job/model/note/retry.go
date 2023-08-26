package note

import (
	"go-common/library/log"
	"strconv"
	"strings"
)

const (
	KeyRetryDetail       = "detail_cache_retry"
	KeyRetryDBDetail     = "detail_db_retry"
	KeyRetryContent      = "content_cache_retry"
	KeyRetryDBDelCont    = "content_db_del"
	KeyRetryDBDelDetail  = "detail_db_del"
	KeyRetryList         = "list_cache_retry"
	KeyRetryListRem      = "list_rem_cache_retry"
	KeyRetryUser         = "user_retry"
	KeyRetryAid          = "aid_cache_retry"
	KeyRetryDel          = "del_cache_retry"
	KeyRetryAudit        = "audit_retry"
	KeyRetryArtDetailDB  = "art_detail_db_retry"
	KeyRetryArtContDB    = "art_content_db_retry"
	KeyRetryArtDtlBinlog = "art_dtl_binlog_retry"
)

func ToIds(val string) []int64 {
	arr := strings.Split(val, "-")
	idsArr := make([]int64, 0)
	for _, a := range arr {
		id, err := strconv.ParseInt(a, 10, 64)
		if err != nil || id == 0 {
			log.Warn("noteWarn ToIds val(%s) format incorrect", val)
			return []int64{}
		}
		idsArr = append(idsArr, id)
	}
	return idsArr
}

// for retryDetailDBDel,
func ToStrIds(val string) (noteIdsStr string, mid int64) {
	arr := strings.Split(val, "-")
	if len(arr) != 2 { // nolint:gomnd
		return "", 0
	}
	noteIdsStr = arr[0]
	var err error
	if mid, err = strconv.ParseInt(arr[1], 10, 64); err != nil || mid == 0 {
		log.Warn("noteWarn ToStrIds val(%s) format incorrect", val)
		return "", 0
	}
	return noteIdsStr, mid
}
