package model

import (
	"encoding/json"
	"hash/crc64"

	v1 "go-gateway/app/app-svr/kvo/interface/api"
)

var crcTable = crc64.MakeTable(crc64.ECMA)

// Result get module message
func Result(player *v1.DanmuPlayerConfig) (rm json.RawMessage, checkSum int64, err error) {
	var (
		bs []byte
	)
	playerSha1 := player.ToPlayerSha1()
	bs, err = json.Marshal(player)
	if err != nil {
		return
	}
	rm = json.RawMessage(bs)
	// check_sum
	bs, err = json.Marshal(playerSha1)
	if err != nil {
		return
	}
	checkSum = int64(crc64.Checksum(bs, crcTable) >> 1)
	return
}
